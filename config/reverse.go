// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	//_ "github.com/mattn/go-oci8"
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

// ReverseConfig represents a reverse configuration
type ReverseConfig struct {
	Kind    string          `yaml:"kind"`
	Name    string          `yaml:"name"`
	Source  ReverseSource   `yaml:"source"`
	Targets []ReverseTarget `yaml:"targets"`
}

// ReverseSource represents a reverse source which should be a database connection
type ReverseSource struct {
	Database string `yaml:"database"`
	ConnStr  string `yaml:"conn_str"`
	OptStr   string `yaml:"opt_str"`
}

func (r ReverseSource) Connect(verbose bool) (*xorm.Engine, error) {
	if r.Database == "" || r.ConnStr == "" {
		return nil, fmt.Errorf("the config of connection is empty")
	} else if verbose {
		fmt.Println("Connect:", r.Database, r.ConnStr)
	}
	engine, err := xorm.NewEngine(r.Database, r.ConnStr)
	if err == nil {
		engine.ShowSQL(verbose)
	}
	return engine, err
}

// ReverseTarget represents a reverse target
type ReverseTarget struct {
	Type              string   `yaml:"type"`
	IncludeTables     []string `yaml:"include_tables"`
	ExcludeTables     []string `yaml:"exclude_tables"`
	TableMapper       string   `yaml:"table_mapper"`
	ColumnMapper      string   `yaml:"column_mapper"`
	TemplatePath      string   `yaml:"template_path"`
	QueryTemplatePath string   `yaml:"query_template_path"`
	InitTemplatePath  string   `yaml:"init_template_path"`
	InitNameSpace     string   `yaml:"init_name_space"`
	MultipleFiles     bool     `yaml:"multiple_files"`
	OutputDir         string   `yaml:"output_dir"`
	TablePrefix       string   `yaml:"table_prefix"`
	Language          string   `yaml:"language"`

	Funcs     map[string]string `yaml:"funcs"`
	Formatter string            `yaml:"formatter"`
	Importter string            `yaml:"importter"`
	ExtName   string            `yaml:"ext_name"`

	NameSpace       string `yaml:"name_space"`
	GenJsonTag      bool   `yaml:"gen_json_tag"`
	GenTableName    bool   `yaml:"gen_table_name"`
	GenQueryMethods bool   `yaml:"gen_query_methods"`
	ApplyMixins     bool   `yaml:"apply_mixins"`
	MixinDirPath    string `yaml:"mixin_dir_path"`
	MixinNameSpace  string `yaml:"mixin_name_space"`
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

func (t ReverseTarget) MergeOptions(name string, part PartConfig) ReverseTarget {
	if t.Type == "codes" && t.Language == "" {
		t.Language = "golang"
	}
	if name != "" {
		t.OutputDir = filepath.Join(t.OutputDir, name)
	}
	if t.TablePrefix == "" {
		t.TablePrefix = part.TablePrefix
	}
	if t.IncludeTables == nil {
		t.IncludeTables = part.IncludeTables
	}
	if t.ExcludeTables == nil {
		t.ExcludeTables = part.ExcludeTables
	}
	return t
}
