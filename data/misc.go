package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nullTime *NullTime) Scan(value interface{}) error {
	nullTime.Time, nullTime.Valid = value.(time.Time)
	return nil
}

func (nullTime NullTime) Value() (driver.Value, error) {
	if !nullTime.Valid {
		return nil, nil
	}
	return nullTime.Time, nil
}

func TimeAsNull() NullTime {
	return NullTime{time.Unix(0, 0), false}
}

func ToNullString(str string) sql.NullString {
	return sql.NullString{str, len(str) != 0}
}

func encodeName(name string) string {
	return "\"" + strings.Replace(name, "\"", "\"\"", -1) + "\""
}

func encodeValue(value interface{}) string {
	encodeStr := func(value string) string {
		return "'" + strings.Replace(value, "'", "''", -1) + "'"
	}

	// time.String() result is recognizable by PostgreSQL
	encodeTime := func(value time.Time) string {
		return encodeStr(value.Format(time.RFC3339))
	}

	if time_, ok := value.(time.Time); ok {
		return encodeTime(time_)
	}

	valuer, ok := value.(driver.Valuer)
	if !ok {
		return encodeStr(fmt.Sprint(value))
	}

	val, err := valuer.Value()
	if err != nil {
		return encodeStr(fmt.Sprint(value))
	}

	if time_, ok := val.(time.Time); ok {
		return encodeTime(time_)
	}

	if val != nil {
		return encodeStr(fmt.Sprint(val))
	}

	return "NULL"
}

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
