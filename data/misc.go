package data

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type ColName struct {
	Table  string
	Column string
}

type Eq struct {
	Left  interface{}
	Right interface{}
}

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

func sqlWhere(where interface{}) (string, error) {
	var left, right string
	var err error

	switch where.(type) {
	case int, string:
		return encodeValue(where), nil
	case ColName:
		colName := where.(ColName)
		return colName.sqlDesc(), nil
	case Eq:
		var eq Eq = where.(Eq)
		if left, err = sqlWhere(eq.Left); err != nil {
			return "", err
		}
		if right, err = sqlWhere(eq.Right); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s = %s", left, right), nil
	default:
		return "", fmt.Errorf("data: unknown type (`%T`) "+
			"in `where` clause of SelectRows", where)
	}
}
