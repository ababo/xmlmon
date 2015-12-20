package data

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Handle interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func Open(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	_, err = db.Query("SELECT")
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
