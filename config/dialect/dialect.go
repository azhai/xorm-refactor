package dialect

import (
	"strconv"
	"strings"

	utils "github.com/azhai/gozzo-utils/common"
)

var (
	ConcatWith = utils.ConcatWith
	WrapWith   = utils.WrapWith
)

var dialects = map[string]Dialect{
	"mssql":    &Mssql{},
	"mysql":    &Mysql{},
	"oracle":   &Oracle{},
	"postgres": &Postgres{},
	"redis":    &Redis{},
	"sqlite":   &Sqlite{},
	"sqlite3":  &Sqlite3{},
}

type Dialect interface {
	Name() string
	ImporterPath() string
	QuoteIdent(ident string) string
	ParseDSN(params ConnParams) string
}

func GetDialectByName(name string) Dialect {
	name = strings.ToLower(name)
	if d, ok := dialects[name]; ok {
		return d
	}
	return nil
}

// 连接配置
type ConnParams struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Options  map[string]interface{}
}

func (p ConnParams) GetAddr(defaultHost string, defaultPort uint16) string {
	if p.Host != "" {
		defaultHost = p.Host
	}
	return ConcatWith(defaultHost, p.StrPort(defaultPort))
}

func (p ConnParams) StrPort(defaultPort uint16) string {
	if p.Port > 0 {
		return strconv.Itoa(p.Port)
	}
	return strconv.Itoa(int(defaultPort))
}
