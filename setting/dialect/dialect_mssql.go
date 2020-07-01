package dialect

const MSSQL_DEFAULT_PORT uint16 = 1433

type Mssql struct {
}

func (Mssql) Name() string {
	return "mssql"
}

func (Mssql) ImporterPath() string {
	return "github.com/denisenkom/go-mssqldb"
}

func (Mssql) QuoteIdent(ident string) string {
	return WrapWith(ident, "[", "]")
}

func (Mssql) ParseDSN(params ConnParams) string {
	dsn := "sqlserver://"
	user := ConcatWith(params.Username, params.Password)
	if user != "" {
		dsn += user + "@"
	}
	dsn += params.GetAddr("127.0.0.1", MSSQL_DEFAULT_PORT)
	if params.Database != "" {
		dsn += "?database" + params.Database
	}
	return dsn
}
