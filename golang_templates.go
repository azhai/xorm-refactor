package refactor

import (
	"fmt"
	"strings"
	"text/template"
)

var (
	golangModelTemplate = fmt.Sprintf(`package {{.Target.NameSpace}}

{{$ilen := len .Imports}}{{if gt $ilen 0 -}}
import (
	{{range $imp, $al := .Imports}}{{$al}} "{{$imp}}"{{end}}
)
{{end -}}

{{range $table_name, $table := .Tables}}
{{$class := TableMapper $table.Name}}
// {{$class}} {{$table.Comment}}
type {{$class}} struct { {{- range $table.ColumnsSeq}}{{$col := $table.GetColumn .}}
	{{ColumnMapper $col.Name}} {{Type $col}} %s{{Tag $table $col true}}%s{{end}}
}

func ({{$class}}) TableName() string {
	return "{{$table_name}}"
}
{{end}}
`, "`", "`")

	golangCacheTemplate = `package {{.Target.NameSpace}}

import (
	"gitee.com/azhai/xorm-refactor/base"
	"gitee.com/azhai/xorm-refactor/setting"
	"github.com/azhai/gozzo-utils/redisw"
	"github.com/gomodule/redigo/redis"
	"xorm.io/xorm/log"
)

const (
	SESS_RESCUE_TIMEOUT = 3600                    // 过期前1个小时，重设会话生命期为5个小时
	SESS_CREATE_TIMEOUT = SESS_RESCUE_TIMEOUT * 5 // 最后一次请求后4到5小时会话过期
)

var (
	sessreg *base.SessionRegistry
)

// Initialize 初始化、连接缓存
func Initialize(c setting.ConnConfig, verbose bool) {
	var wrapper *redisw.RedisWrapper
	if conn, err := c.ConnectRedis(verbose); err == nil {
		wrapper = redisw.NewRedisConnMux(conn)
		wrapper.MaxReadTime = 0 // 不支持 ConnWithTimeout 和 DoWithTimeout
	} else {
		dial := func() (redis.Conn, error) {
			return c.ConnectRedis(verbose)
		}
		wrapper = redisw.NewRedisPool(dial, -1)
	}
	sessreg = base.NewRegistry(wrapper)
}

// Registry 获得当前会话管理器
func Registry() *base.SessionRegistry {
	return sessreg
}

// Session 获得用户会话
func Session(token string) *base.Session {
	if sessreg == nil {
		return nil
	}
	sess := sessreg.GetSession(token, SESS_CREATE_TIMEOUT)
	timeout := sess.GetTimeout(false)
	if timeout >= 0 && timeout < SESS_RESCUE_TIMEOUT {
		sess.Expire(SESS_CREATE_TIMEOUT) // 重设会话生命期
	}
	return sess
}

// DelSession 删除会话
func DelSession(token string) bool {
	if sessreg == nil {
		return false
	}
	return sessreg.DelSession(token)
}
`

	golangConnTemplate = `package {{.Target.NameSpace}}

import (
	"gitee.com/azhai/xorm-refactor/base"
	"gitee.com/azhai/xorm-refactor/setting"
	_ "{{.ImporterPath}}"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var (
	engine  *xorm.Engine
)

// Initialize 初始化、连接数据库
func Initialize(c setting.ConnConfig, verbose bool) {
	var err error
	if engine, err = c.ConnectXorm(verbose); err != nil {
		panic(err)
	}
	if c.LogFile != "" {
		logger := setting.NewSqlLogger(c.LogFile)
		engine.SetLogger(logger)
	}
}

// Engine 获取当前数据库连接
func Engine() *xorm.Engine {
	return engine
}

// Quote 转义表名或字段名
func Quote(value string) string {
	if engine == nil {
		return value
	}
	return engine.Quote(value)
}

// Table 查询某张数据表
func Table(args ...interface{}) *xorm.Session {
	if engine == nil {
		return nil
	}
	if args == nil {
		return engine.NewSession()
	}
	return engine.Table(args[0])
}

// QueryAll 查询多行数据
func QueryAll(filter base.FilterFunc, pages ...int) *xorm.Session {
	query := engine.NewSession()
	if filter != nil {
		query = filter(query)
	}
	pageno, pagesize := 0, -1
	if len(pages) >= 1 {
		pageno = pages[0]
		if len(pages) >= 2 {
			pagesize = pages[1]
		}
	}
	return base.Paginate(query, pageno, pagesize)
}

// ExecTx 执行事务
func ExecTx(modify base.ModifyFunc) error {
	tx := engine.NewSession() // 必须是新的session
	defer tx.Close()
	_ = tx.Begin()
	if _, err := modify(tx); err != nil {
		_ = tx.Rollback() // 失败回滚
		return err
	}
	return tx.Commit()
}

// InsertBatch 写入多行数据
func InsertBatch(tableName string, rows []map[string]interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	return ExecTx(func(tx *xorm.Session) (int64, error) {
		return tx.Table(tableName).Insert(rows)
	})
}
`

	golangInitTemplate = `package models

{{$initns := .Target.InitNameSpace -}}
import (
	"strings"

	"gitee.com/azhai/xorm-refactor/cmd"
	"gitee.com/azhai/xorm-refactor/setting"

	{{range $dir, $al := .Imports}}
	{{if ne $al $dir}}{{$al}} {{end -}}
	"{{$initns}}/{{$dir}}"{{end}}
)

var (
	configFiles = []string{ // 设置多个路径，方便从子目录下运行
		"./databases.json", "../databases.json", "../../databases.json",
		"./databases.yml", "../databases.yml", "../../databases.yml",
		"./settings.yml", "../settings.yml", "../../settings.yml",
	}
)

func init() {
	confs := make(map[string]setting.ConnConfig)
	fileName := cmd.FindFirstFile(configFiles, 0)
	if strings.Contains(fileName, "database") {
		_, err := setting.ReadSettingsExt(fileName, &confs)
		if err != nil {
			panic(err)
		}
	} else {
		settings, err := cmd.Prepare(fileName, "{{.ProjNameSpace}}")
		if err != nil {
			panic(err)
		}
		confs = settings.GetConnConfigMap()
	}
	ConnectDatabases(confs)
}

func ConnectDatabases(confs map[string]setting.ConnConfig) {
	verbose := cmd.Verbose()
	for key, c := range confs {
		switch key {
		{{- range $dir, $al := .Imports}}
			case "{{$dir}}":
			{{$al}}.Initialize(c, verbose){{end}}
		}
	}
}
`

	golangQueryTemplate = `{{if not .Target.MultipleFiles}}package {{.Target.NameSpace}}

import (
	"time"

	{{range $imp, $al := .Imports}}{{$al}} "{{$imp}}"{{end}}
)
{{end -}}

{{range .Tables}}
{{$class := TableMapper .Name -}}
{{$pkey := GetSinglePKey . -}}
{{$created := GetCreatedColumn . -}}
// the queries of {{$class}}

func (m *{{$class}}) Load(where interface{}, args ...interface{}) (bool, error) {
	return Table().Where(where, args...).Get(m)
}

{{if ne $pkey "" -}}
func (m *{{$class}}) Save(changes map[string]interface{}) error {
	return ExecTx(func(tx *xorm.Session) (int64, error) {
		if changes == nil || m.{{$pkey}} == 0 {
			{{if ne $created "" -}}changes["{{$created}}"] = time.Now()
			{{else}}{{end -}}
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.{{$pkey}}).Update(changes)
		}
	})
}
{{end}}
{{end -}}
`
)

func GetGolangTemplate(name string, funcs template.FuncMap) *template.Template {
	var content string
	switch strings.ToLower(name) {
	case "cache":
		name, content = "cache", golangCacheTemplate
	case "conn":
		name, content = "conn", golangConnTemplate
	case "init":
		name, content = "init", golangInitTemplate
	case "query":
		name, content = "query", golangQueryTemplate
	default:
		name, content = "model", golangModelTemplate
	}
	if tmpl := GetPresetTemplate(name); tmpl != nil {
		return tmpl
	}
	return NewTemplate(name, content, funcs)
}
