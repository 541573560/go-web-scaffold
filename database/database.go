package database

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type DatabaseConfig struct {
	Host         string
	User         string
	Password     string
	DatabaseName string
}

type DBC struct {
	sqldb *bun.DB
}

// NewDB creates a new database client
func NewDB(dbc DatabaseConfig) (db *DBC, err error) {
	sqldb, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbc.User, dbc.Password, dbc.Host, dbc.DatabaseName))
	if err != nil {
		err = fmt.Errorf("Open: %v", err)
		return
	}

	dbcHandler := bun.NewDB(sqldb, mysqldialect.New())

	// To see executed queries in stdout
	dbcHandler.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))

	db = &DBC{sqldb: dbcHandler}
	return
}

// Close closes the database client, releasing any open resources.
func (db *DBC) Close() (errs []error) {
	err := db.sqldb.Close()
	if err != nil {
		err = fmt.Errorf("db.sqldb.Close: %w", err)
		errs = append(errs, err)
	}

	return
}
