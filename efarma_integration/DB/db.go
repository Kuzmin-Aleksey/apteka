package DB

import (
	"database/sql"
	"efarma_integration/config"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
)

func Connect(cnf *config.DbConfig) (*sql.DB, error) {
	sqlExpressUrl := ""

	if cnf.SQLExpress {
		sqlExpressUrl = "/SQLExpress"
	}

	db, err := sql.Open("mssql", fmt.Sprintf("sqlserver://%s:%s@%s%s?database=%s&encrypt=%s&connection+timeout=%d",
		cnf.Username, cnf.Password, cnf.Host, sqlExpressUrl, cnf.DBName, cnf.Encrypt, cnf.ConnectTimeout))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
