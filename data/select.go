package data

import (
	"database/sql"
	"fmt"
	"strings"
)

type ColName struct {
	Table  string
	Column string
}

type Join struct {
	LeftColumn  string
	Table       string
	RightColumn string
}

type Eq struct {
	Left  interface{}
	Right interface{}
}

type Order struct {
	ColName
	Descending bool
}

func SelectRows(handle Handle, columns []ColName, joins []Join,
	where interface{}, orders []Order, limit int) (*sql.Rows, error) {
	var cols []string
	for _, c := range columns {
		cols = append(cols, c.sqlDesc())
	}

	sql := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(cols, ", "), sqlJoins(joins))

	if where != nil {
		expr, err := sqlWhere(where)
		if err != nil {
			return nil, err
		}
		sql += " WHERE " + expr
	}

	if orders != nil {
		sql += " " + sqlOrders(orders)
	}

	if limit >= 0 {
		sql += fmt.Sprintf(" LIMIT %d", limit)
	}

	return handle.Query(sql)
}

func (colName *ColName) sqlDesc() string {
	if len(colName.Table) == 0 {
		return encodeName(colName.Column)
	}
	return encodeName(colName.Table) + "." + encodeName(colName.Column)
}

func sqlJoins(joins []Join) string {
	var sql string
	for i, j := range joins {
		if i == 0 {
			sql = encodeName(j.Table)
		} else {
			p := joins[i-1]
			jt := encodeName(j.Table)
			sql += fmt.Sprintf(" JOIN %s ON %s.%s = %s.%s", jt,
				encodeName(p.Table), encodeName(p.RightColumn),
				jt, encodeName(j.LeftColumn))
		}
	}
	return sql
}

func sqlWhere(where interface{}) (string, error) {
	var left, right string
	var err error

	switch where.(type) {
	case int:
		return encodeValue(fmt.Sprint(where)), nil
	case string:
		return encodeValue(where.(string)), nil
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
		return "", fmt.Errorf(
			"select unknown type in `where` of select: %T", where)
	}
}

func sqlOrders(orders []Order) string {
	var cols []string
	for _, o := range orders {
		col := o.sqlDesc()
		if o.Descending {
			col += " DESC"
		}
		cols = append(cols, col)
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(cols, ", "))
}
