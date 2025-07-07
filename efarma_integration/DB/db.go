package DB

import (
	"database/sql"
	"efarma_integration/config"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"strings"
)

func Connect(cnf *config.DbConfig) (*sql.DB, error) {
	sqlExpressUrl := ""

	if cnf.SQLExpress {
		sqlExpressUrl = "/SQLExpress"
	}

	var args string

	if len(cnf.Args) > 0 {
		for k, v := range cnf.Args {
			args += fmt.Sprintf("%s=%v&", k, v)
		}
		args = strings.TrimRight(args, "&")
	}

	db, err := sql.Open("mssql", fmt.Sprintf("sqlserver://%s:%s@%s%s?database=%s&%s",
		cnf.Username, cnf.Password, cnf.Host, sqlExpressUrl, cnf.DBName, args))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
