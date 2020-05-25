// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refactor

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"gitee.com/azhai/xorm-refactor/config"
	"gitee.com/azhai/xorm-refactor/language"
	"gitee.com/azhai/xorm-refactor/rewrite"
	"github.com/azhai/gozzo-utils/filesystem"
	"github.com/gobwas/glob"
	"github.com/grsmv/inflect"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

var (
	formatters   = map[string]language.Formatter{}
	importters   = map[string]language.Importter{}
	defaultFuncs = template.FuncMap{
		"Lower":            strings.ToLower,
		"Upper":            strings.ToUpper,
		"Title":            strings.Title,
		"Camelize":         inflect.Camelize,
		"Underscore":       inflect.Underscore,
		"Singularize":      inflect.Singularize,
		"Pluralize":        inflect.Pluralize,
		"DiffPluralize":    DiffPluralize,
		"GetSinglePKey":    GetSinglePKey,
		"GetCreatedColumn": GetCreatedColumn,
	}
)

// 如果复数形式和单数相同，人为增加后缀
func DiffPluralize(word, suffix string) string {
	words := inflect.Pluralize(word)
	if words == word {
		words += suffix
	}
	return words
}

func GetSinglePKey(table *schemas.Table) string {
	if cols := table.PKColumns(); len(cols) == 1 {
		return cols[0].FieldName
	}
	return ""
}

func GetCreatedColumn(table *schemas.Table) string {
	for name, ok := range table.Created {
		if ok {
			return table.GetColumn(name).Name
		}
	}
	if col := table.GetColumn("created_at"); col != nil {
		if col.SQLType.IsTime() {
			return "created_at"
		}
	}
	return ""
}

func GetTableSchemas(source *config.ReverseSource, target *config.ReverseTarget, verbose bool) []*schemas.Table {
	var tableSchemas []*schemas.Table
	engine, err := source.Connect(verbose)
	if err != nil {
		panic(err)
	}
	tableSchemas, _ = engine.DBMetas()
	return filterTables(tableSchemas, target.IncludeTables, target.ExcludeTables)
}

func filterTables(tables []*schemas.Table, includes, excludes []string) []*schemas.Table {
	res := make([]*schemas.Table, 0, len(tables))
	for _, tb := range tables {
		var remove bool
		for _, exclude := range excludes {
			s, _ := glob.Compile(exclude)
			remove = s.Match(tb.Name)
			if remove {
				break
			}
		}
		if remove {
			continue
		}
		if len(includes) == 0 {
			res = append(res, tb)
			continue
		}

		var keep bool
		for _, include := range includes {
			s, _ := glob.Compile(include)
			keep = s.Match(tb.Name)
			if keep {
				break
			}
		}
		if keep {
			res = append(res, tb)
		}
	}
	return res
}

func newFuncs() template.FuncMap {
	m := make(template.FuncMap)
	for k, v := range defaultFuncs {
		m[k] = v
	}
	return m
}

func convertMapper(mapname string) names.Mapper {
	switch mapname {
	case "gonic":
		return names.LintGonicMapper
	case "same":
		return names.SameMapper{}
	default:
		return names.SnakeMapper{}
	}
}

func Reverse(target *config.ReverseTarget, source *config.DataSource, verbose bool) error {
	formatter := formatters[target.Formatter]
	lang := language.GetLanguage(target.Language)
	if lang != nil {
		lang.FixTarget(target)
		formatter = lang.Formatter
	}
	if formatter == nil {
		formatter = rewrite.WriteCodeFile
	}

	isRedis := true
	if source.ReverseSource.Database != "redis" {
		isRedis = false
		tableSchemas := GetTableSchemas(source.ReverseSource, target, verbose)
		err := RunReverse(target, tableSchemas)
		if err != nil {
			return err
		}
	}
	if target.Language != "golang" {
		return nil
	}

	var tmpl *template.Template
	if isRedis {
		tmpl = language.GetGolangTemplate("cache", nil)
	} else {
		tmpl = language.GetGolangTemplate("conn", nil)
	}
	buf := new(bytes.Buffer)
	data := map[string]interface{}{
		"Target":       target,
		"NameSpace":    target.NameSpace,
		"ConnKey":      source.ConnKey,
		"ImporterPath": source.ImporterPath,
	}
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	fileName := target.GetOutFileName(config.CONN_FILE_NAME)
	_, err := formatter(fileName, buf.Bytes())

	if target.ApplyMixins {
		_err := ExecApplyMixins(target)
		if _err != nil {
			err = _err
		}
	}
	return err
}

func RunReverse(target *config.ReverseTarget, tableSchemas []*schemas.Table) error {
	// load configuration from language
	lang := language.GetLanguage(target.Language)
	funcs := newFuncs()
	formatter := formatters[target.Formatter]
	importter := importters[target.Importter]

	// load template
	var bs []byte
	if lang != nil {
		bs = []byte(lang.Template)
		for k, v := range lang.Funcs {
			funcs[k] = v
		}
		if formatter == nil {
			formatter = lang.Formatter
		}
		if importter == nil {
			importter = lang.Importter
		}
	}

	tableMapper := convertMapper(target.TableMapper)
	colMapper := convertMapper(target.ColumnMapper)
	funcs["TableMapper"] = tableMapper.Table2Obj
	funcs["ColumnMapper"] = colMapper.Table2Obj

	// 配置模板优先于语言模板
	var tmplQuery *template.Template
	if target.QueryTemplatePath != "" {
		qt, err := ioutil.ReadFile(target.QueryTemplatePath)
		if err == nil && len(qt) > 0 {
			tmplQuery = language.NewTemplate("custom-query", string(qt), funcs)
		} else {
			target.GenQueryMethods = false
		}
	} else {
		tmplQuery = language.GetGolangTemplate("query", funcs)
	}
	var err error
	if target.TemplatePath != "" {
		bs, err = ioutil.ReadFile(target.TemplatePath)
		if err != nil {
			return err
		}
	}

	if bs == nil {
		return errors.New("you have to indicate template / template path or a language")
	}
	tmpl := language.NewTemplate("custom-model", string(bs), funcs)
	queryImports := map[string]string{"xorm.io/xorm": ""}

	tables := make(map[string]*schemas.Table)
	for _, table := range tableSchemas {
		tableName := table.Name
		if target.TablePrefix != "" {
			table.Name = strings.TrimPrefix(table.Name, target.TablePrefix)
		}
		for _, col := range table.Columns() {
			col.FieldName = colMapper.Table2Obj(col.Name)
		}
		tables[tableName] = table
	}

	err = os.MkdirAll(target.OutputDir, os.ModePerm)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if !target.MultipleFiles {
		packages := importter(tables)
		data := map[string]interface{}{
			"Target":  target,
			"Tables":  tables,
			"Imports": packages,
		}
		if err = tmpl.Execute(buf, data); err != nil {
			return err
		}
		fileName := target.GetOutFileName(config.SINGLE_FILE_NAME)
		if _, err = formatter(fileName, buf.Bytes()); err != nil {
			return err
		}
		if target.GenQueryMethods && tmplQuery != nil {
			buf.Reset()
			data["Imports"] = queryImports
			if err = tmplQuery.Execute(buf, data); err != nil {
				return err
			}
			fileName := target.GetOutFileName(config.QUERY_FILE_NAME)
			if _, err = formatter(fileName, buf.Bytes()); err != nil {
				return err
			}
		}
	} else {
		for tableName, table := range tables {
			tbs := map[string]*schemas.Table{tableName: table}
			packages := importter(tbs)
			data := map[string]interface{}{
				"Target":  target,
				"Tables":  tbs,
				"Imports": packages,
			}
			buf.Reset()
			if err = tmpl.Execute(buf, data); err != nil {
				return err
			}
			if target.GenQueryMethods && tmplQuery != nil {
				data["Imports"] = queryImports
				if err = tmplQuery.Execute(buf, data); err != nil {
					return err
				}
			}
			fileName := target.GetOutFileName(table.Name)
			if _, err = formatter(fileName, buf.Bytes()); err != nil {
				return err
			}
		}
	}
	return nil
}

func ExecReverseSettings(cfg config.IReverseSettings, verbose bool, names ...string) error {
	conns := cfg.GetConnConfigMap(names...)
	targets := cfg.GetReverseTargets()
	if len(targets) == 0 {
		return nil
	}
	var target config.ReverseTarget
	imports := make(map[string]string)
	for key, conf := range conns {
		d := config.NewDataSource(conf, key)
		if d.ReverseSource == nil {
			continue
		}
		for _, target = range targets {
			target = target.MergeOptions(d.ConnKey, d.PartConfig)
			if err := Reverse(&target, d, verbose); err != nil {
				return err
			}
			imports[d.ConnKey] = target.NameSpace
		}
	}
	return GenModelInitFile(target, imports)
}

func GenModelInitFile(target config.ReverseTarget, imports map[string]string) error {
	var tmpl *template.Template
	if target.InitTemplatePath != "" {
		it, err := ioutil.ReadFile(target.InitTemplatePath)
		if err != nil || len(it) == 0 {
			return err
		}
		tmpl = language.NewTemplate("custom-init", string(it), nil)
	} else {
		tmpl = language.GetGolangTemplate("init", nil)
	}
	buf := new(bytes.Buffer)
	data := map[string]interface{}{
		"Target":  target,
		"Imports": imports,
	}
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	fileName := target.GetParentOutFileName(config.INIT_FILE_NAME, 1)
	_, err := rewrite.CleanImportsWriteGolangFile(fileName, buf.Bytes())
	return err
}

func ExecApplyMixins(target *config.ReverseTarget) error {
	if target.MixinDirPath != "" {
		files, _ := filesystem.FindFiles(target.MixinDirPath, ".go")
		for fileName := range files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			_ = rewrite.AddFormerMixins(fileName, target.MixinNameSpace, "")
		}
	}
	files, _ := filesystem.FindFiles(target.OutputDir, ".go")
	var err error
	for fileName := range files {
		_err := rewrite.ParseAndMixinFile(fileName, true)
		if _err != nil {
			err = _err
		}
	}
	return err
}
