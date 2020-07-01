package dialect

import (
	"fmt"
)

const MYSQL_DEFAULT_PORT uint16 = 3306

type Mysql struct {
}

func (Mysql) Name() string {
	return "mysql"
}

func (Mysql) ImporterPath() string {
	return "github.com/go-sql-driver/mysql"
}

func (Mysql) QuoteIdent(ident string) string {
	return WrapWith(ident, "`", "`")
}

func (Mysql) ParseDSN(params ConnParams) string {
	user := ConcatWith(params.Username, params.Password)
	addr := params.GetAddr("127.0.0.1", MYSQL_DEFAULT_PORT)
	dsn := user + "@"
	if addr != "" {
		dsn += "tcp(" + addr + ")"
	}
	if params.Database != "" {
		dsn += "/" + params.Database
	}
	dsn += "?parseTime=true&loc=Local"
	if charset, ok := params.Options["charset"]; ok {
		dsn += fmt.Sprintf("&charset=%s", charset)
	}
	return dsn
}
