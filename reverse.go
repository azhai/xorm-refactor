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

	"gitee.com/azhai/xorm-refactor/rewrite"
	"gitee.com/azhai/xorm-refactor/setting"
	"github.com/azhai/gozzo-utils/filesystem"
	"github.com/grsmv/inflect"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

var (
	formatters   = map[string]Formatter{}
	importters   = map[string]Importter{}
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

func filterTables(tables []*schemas.Table, includes, excludes []string) []*schemas.Table {
	res := make([]*schemas.Table, 0, len(tables))
	incl_matchers, excl_matchers := setting.NewGlobs(includes), setting.NewGlobs(excludes)
	for _, tb := range tables {
		if excl_matchers.MatchAny(tb.Name, false) {
			continue
		}
		if incl_matchers.MatchAny(tb.Name, true) {
			res = append(res, tb)
		}
	}
	return res
}

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

func GetTableSchemas(source *setting.ReverseSource, target *setting.ReverseTarget, verbose bool) []*schemas.Table {
	var tableSchemas []*schemas.Table
	engine, _, err := source.Connect(verbose)
	if err != nil {
		panic(err)
	}
	tableSchemas, _ = engine.DBMetas()
	return filterTables(tableSchemas, target.IncludeTables, target.ExcludeTables)
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

func Reverse(target *setting.ReverseTarget, source *setting.ReverseSource, verbose bool) error {
	formatter := formatters[target.Formatter]
	lang := GetLanguage(target.Language)
	if lang != nil {
		lang.FixTarget(target)
		formatter = lang.Formatter
	}
	if formatter == nil {
		formatter = rewrite.WriteCodeFile
	}

	isRedis := true
	if source.DriverName != "redis" {
		isRedis = false
		tableSchemas := GetTableSchemas(source, target, verbose)
		err := RunReverse(source.TablePrefix, target, tableSchemas)
		if err != nil {
			return err
		}
	}
	if target.Language != "golang" {
		return nil
	}

	var tmpl *template.Template
	if isRedis {
		tmpl = GetGolangTemplate("cache", nil)
	} else {
		tmpl = GetGolangTemplate("conn", nil)
	}
	buf := new(bytes.Buffer)
	data := map[string]interface{}{
		"Target":       target,
		"NameSpace":    target.NameSpace,
		"ImporterPath": source.ImporterPath,
	}
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	fileName := target.GetOutFileName(setting.CONN_FILE_NAME)
	_, err := formatter(fileName, buf.Bytes())

	if target.ApplyMixins {
		_err := ExecApplyMixins(target, verbose)
		if _err != nil {
			err = _err
		}
	}
	return err
}

func RunReverse(tablePrefix string, target *setting.ReverseTarget, tableSchemas []*schemas.Table) error {
	// load configuration from language
	lang := GetLanguage(target.Language)
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
			tmplQuery = NewTemplate("custom-query", string(qt), funcs)
		}
	} else {
		tmplQuery = GetGolangTemplate("query", funcs)
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
	tmpl := NewTemplate("custom-model", string(bs), funcs)
	queryImports := map[string]string{"xorm.io/xorm": ""}

	tables := make(map[string]*schemas.Table)
	for _, table := range tableSchemas {
		tableName := table.Name
		if tablePrefix != "" {
			table.Name = strings.TrimPrefix(table.Name, tablePrefix)
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
		fileName := target.GetOutFileName(setting.SINGLE_FILE_NAME)
		if _, err = formatter(fileName, buf.Bytes()); err != nil {
			return err
		}
		if tmplQuery != nil {
			buf.Reset()
			data["Imports"] = queryImports
			if err = tmplQuery.Execute(buf, data); err != nil {
				return err
			}
			fileName := target.GetOutFileName(setting.QUERY_FILE_NAME)
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
			if tmplQuery != nil {
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

func ExecReverseSettings(cfg setting.IReverseConfig, verbose bool, names ...string) error {
	target := cfg.GetReverseTarget("*")
	if target.OutputDir == "/dev/null" {
		return nil
	}
	conns := cfg.GetConnConfigMap(names...)
	imports := make(map[string]string)
	for key, conf := range conns {
		src, _ := setting.NewReverseSource(conf)
		target = target.MergeOptions(key, src)
		if err := Reverse(&target, src, verbose); err != nil {
			return err
		}
		imports[key] = target.NameSpace
	}
	if target.InitNameSpace != "" {
		return GenModelInitFile(target, imports)
	}
	return nil
}

func GenModelInitFile(target setting.ReverseTarget, imports map[string]string) error {
	var tmpl *template.Template
	if target.InitTemplatePath != "" {
		it, err := ioutil.ReadFile(target.InitTemplatePath)
		if err != nil || len(it) == 0 {
			return err
		}
		tmpl = NewTemplate("custom-init", string(it), nil)
	} else {
		tmpl = GetGolangTemplate("init", nil)
	}
	buf := new(bytes.Buffer)
	data := map[string]interface{}{
		"Target":  target,
		"Imports": imports,
	}
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	fileName := target.GetParentOutFileName(setting.INIT_FILE_NAME, 1)
	_, err := rewrite.CleanImportsWriteGolangFile(fileName, buf.Bytes())
	return err
}

func ExecApplyMixins(target *setting.ReverseTarget, verbose bool) error {
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
		_err := rewrite.ParseAndMixinFile(fileName, verbose)
		if _err != nil {
			err = _err
		}
	}
	return err
}
