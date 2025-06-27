package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
	"server/internal/config"
	"time"
)

func Connect(cnf *config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cnf.User, cnf.Password, cnf.Addr, cnf.Schema))
	if err != nil {
		return nil, errors.New("open connection failed: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, errors.New("ping connection failed: " + err.Error())
	}
	return db, nil
}
