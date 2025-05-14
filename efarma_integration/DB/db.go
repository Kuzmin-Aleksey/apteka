package DB

import (
	"apteka_client/config"
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
)

func Connect(cnf *config.DbConfig) (*sql.DB, error) {
	db, err := sql.Open("mssql", fmt.Sprintf("sqlserver://%s:%s@%s/SQLExpress?database=%s&encrypt=%s&connection+timeout=%d",
		cnf.Username, cnf.Password, cnf.Host, cnf.DBName, cnf.Encrypt, cnf.ConnectTimeout))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
