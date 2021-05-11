// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package setting

import (
	"fmt"
	"os"
	"path/filepath"

	"gitee.com/azhai/xorm-refactor/setting/dialect"
	"github.com/gomodule/redigo/redis"

	//_ "github.com/mattn/go-oci8"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

const ( // 约定大于配置
	INIT_FILE_NAME   = "init"
	CONN_FILE_NAME   = "conn"
	SINGLE_FILE_NAME = "models"
	QUERY_FILE_NAME  = "queries"

	XORM_TAG_NAME        = "xorm"
	XORM_TAG_NOT_NULL    = "notnull"
	XORM_TAG_AUTO_INCR   = "autoincr"
	XORM_TAG_PRIMARY_KEY = "pk"
	XORM_TAG_UNIQUE      = "unique"
	XORM_TAG_INDEX       = "index"
)

// ReverseSource represents a reverse source which should be a database connection
type ReverseSource struct {
	DriverName   string             `json:"driver_name" yaml:"driver_name"`
	TablePrefix  string             `json:"table_prefix" yaml:"table_prefix"`
	ImporterPath string             `json:"importer_path" yaml:"importer_path"`
	ConnStr      string             `json:"conn_str" yaml:"conn_str"`
	OptStr       string             `json:"opt_str" yaml:"opt_str"`
	options      []redis.DialOption `json:"-" yaml:"-"`
}

func NewReverseSource(c ConnConfig) (*ReverseSource, dialect.Dialect) {
	d := dialect.GetDialectByName(c.DriverName)
	r := &ReverseSource{
		DriverName:   c.DriverName,
		TablePrefix:  c.TablePrefix,
		ConnStr:      d.ParseDSN(c.Params),
		ImporterPath: d.ImporterPath(),
	}
	if dr, ok := d.(*dialect.Redis); ok {
		r.options = dr.GetOptions()
		r.OptStr = dr.Values.Encode()
	}
	return r, d
}

func (r ReverseSource) Connect(verbose bool) (*xorm.Engine, redis.Conn, error) {
	if r.DriverName == "" || r.ConnStr == "" {
		err := fmt.Errorf("the config of connection is empty")
		return nil, nil, err
	} else if verbose {
		fmt.Println("Connect:", r.DriverName, r.ConnStr)
	}
	if r.DriverName == "redis" {
		conn, err := redis.Dial("tcp", r.ConnStr, r.options...)
		return nil, conn, err
	}
	engine, err := xorm.NewEngine(r.DriverName, r.ConnStr)
	if err == nil {
		engine.ShowSQL(verbose)
	}
	return engine, nil, err
}

// ReverseTarget represents a reverse target
type ReverseTarget struct {
	Language          string   `json:"language" yaml:"language"`
	IncludeTables     []string `json:"include_tables" yaml:"include_tables"`
	ExcludeTables     []string `json:"exclude_tables" yaml:"exclude_tables"`
	InitNameSpace     string   `json:"init_name_space" yaml:"init_name_space"`
	OutputDir         string   `json:"output_dir" yaml:"output_dir"`
	TemplatePath      string   `json:"template_path" yaml:"template_path"`
	QueryTemplatePath string   `json:"query_template_path" yaml:"query_template_path"`
	InitTemplatePath  string   `json:"init_template_path" yaml:"init_template_path"`

	TableMapper  string            `json:"table_mapper" yaml:"table_mapper"`
	ColumnMapper string            `json:"column_mapper" yaml:"column_mapper"`
	Funcs        map[string]string `json:"funcs" yaml:"funcs"`
	Formatter    string            `json:"formatter" yaml:"formatter"`
	Importter    string            `json:"importter" yaml:"importter"`
	ExtName      string            `json:"-" yaml:"-"`
	NameSpace    string            `json:"-" yaml:"-"`

	MultipleFiles  bool   `json:"multiple_files" yaml:"multiple_files"`
	ApplyMixins    bool   `json:"apply_mixins" yaml:"apply_mixins"`
	MixinDirPath   string `json:"mixin_dir_path" yaml:"mixin_dir_path"`
	MixinNameSpace string `json:"mixin_name_space" yaml:"mixin_name_space"`
}

func DefaultReverseTarget(nameSpace string) ReverseTarget {
	return ReverseTarget{
		Language:      "golang",
		InitNameSpace: nameSpace + "/models",
		OutputDir:     "./models",
	}
}

func DefaultMixinReverseTarget(nameSpace string) ReverseTarget {
	rt := DefaultReverseTarget(nameSpace)
	rt.ApplyMixins = true
	rt.MixinDirPath = filepath.Join(rt.OutputDir, "mixins")
	rt.MixinNameSpace = rt.InitNameSpace + "/mixins"
	return rt
}

func (t ReverseTarget) GetFileName(dir, name string) string {
	_ = os.MkdirAll(dir, DEFAULT_DIR_MODE)
	return filepath.Join(dir, name+t.ExtName)
}

func (t ReverseTarget) GetOutFileName(name string) string {
	return t.GetFileName(t.OutputDir, name)
}

func (t ReverseTarget) GetParentOutFileName(name string, backward int) string {
	outDir := t.OutputDir
	for i := 0; i < backward; i++ {
		outDir = filepath.Dir(outDir)
	}
	return t.GetFileName(outDir, name)
}

func (t ReverseTarget) MergeOptions(key string, src *ReverseSource) ReverseTarget {
	if t.Language == "" {
		t.Language = "golang"
	}
	if key != "" {
		t.OutputDir = filepath.Join(t.OutputDir, key)
	}
	return t
}
