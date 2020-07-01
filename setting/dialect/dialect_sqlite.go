package dialect

type Sqlite struct {
}

func (Sqlite) Name() string {
	return "sqlite"
}

func (Sqlite) ImporterPath() string {
	return "github.com/mattn/go-sqlite3"
}

func (Sqlite) QuoteIdent(ident string) string {
	return WrapWith(ident, "`", "`")
}

func (Sqlite) ParseDSN(params ConnParams) string {
	user := ConcatWith(params.Username, params.Password)
	var dsn string
	if user != "" {
		dsn = user + "@"
	}
	dsn += params.Database + "?cache=shared&mode=rwc"
	return dsn
}

type Sqlite3 struct {
	Sqlite
}

func (Sqlite3) Name() string {
	return "sqlite3"
}
