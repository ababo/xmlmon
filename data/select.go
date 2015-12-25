package data

import (
	"database/sql"
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

type Gr struct {
	Left  interface{}
	Right interface{}
}

type And struct {
	Left  interface{}
	Right interface{}
}

const ( // aggregate types
	Max = iota
	Min = iota
)

type Aggr struct {
	Table  string
	Column string
	Type   int
}

type Join struct {
	LeftColumn  string
	Table       string
	RightColumn string
}

type Order struct {
	Table      string
	Column     string
	Descending bool
}

func SelectRows(handle Handle,
	columns interface{}, from []Join, where interface{},
	groupBy []ColName, orderBy []Order, limit int) (*sql.Rows, error) {
	cols, err := sqlExprList(columns)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", cols, sqlFrom(from))

	if where != nil {
		expr, err := sqlExpr(where)
		if err != nil {
			return nil, err
		}
		sql += " WHERE " + expr
	}

	if groupBy != nil {
		cols, _ := sqlExprList(groupBy)
		sql += " GROUP BY " + cols
	}

	if orderBy != nil {
		sql += " " + sqlOrderBy(orderBy)
	}

	if limit >= 0 {
		sql += fmt.Sprintf(" LIMIT %d", limit)
	}

	return handle.Query(sql)
}

func sqlExprList(list interface{}) (string, error) {
	var strs []string
	switch list.(type) {
	case []interface{}:
		lst := list.([]interface{})
		for _, e := range lst {
			str, err := sqlExpr(e)
			if err != nil {
				return "", err
			}
			strs = append(strs, str)
		}
	case []ColName:
		lst := list.([]ColName)
		for _, e := range lst {
			str, _ := sqlExpr(e)
			strs = append(strs, str)
		}
	case []Aggr:
		lst := list.([]Aggr)
		for _, e := range lst {
			str, err := sqlExpr(e)
			if err != nil {
				return "", err
			}
			strs = append(strs, str)
		}
	default:
		return "", fmt.Errorf("data: unsupported "+
			"type (`%T`) of expression list", list)
	}
	return strings.Join(strs, ", "), nil
}

func sqlExpr(expr interface{}) (string, error) {
	binaryOp := func(format string,
		left, right interface{}) (string, error) {
		var err error
		var lstr, rstr string
		if lstr, err = sqlExpr(left); err != nil {
			return "", err
		}
		if rstr, err = sqlExpr(right); err != nil {
			return "", err
		}
		return fmt.Sprintf(format, lstr, rstr), nil
	}

	switch expr.(type) {
	case int, string, time.Time:
		return encodeValue(expr), nil
	case ColName:
		colName := expr.(ColName)
		return colName.sqlDesc(), nil
	case Aggr:
		aggr := expr.(Aggr)
		return aggr.sqlDesc()
	case Eq:
		eq := expr.(Eq)
		return binaryOp("(%s = %s)", eq.Left, eq.Right)
	case Gr:
		gr := expr.(Gr)
		return binaryOp("(%s > %s)", gr.Left, gr.Right)
	case And:
		and := expr.(And)
		return binaryOp("(%s AND %s)", and.Left, and.Right)
	default:
		return "", fmt.Errorf(
			"data: unknown type (`%T`) in expression", expr)
	}
}

func sqlFrom(joins []Join) string {
	var sql string
	for i, j := range joins {
		if i == 0 {
			sql = encodeName(j.Table)
		} else {
			p := joins[i-1]
			jt := encodeName(j.Table)
			sql += fmt.Sprintf(" JOIN %s ON %s.%s = %s.%s",
				jt, encodeName(p.Table),
				encodeName(p.RightColumn),
				jt, encodeName(j.LeftColumn))
		}
	}
	return sql
}

func sqlOrderBy(orders []Order) string {
	var cols []string
	for _, o := range orders {
		cols = append(cols, o.sqlDesc())
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(cols, ", "))
}

func (colName *ColName) sqlDesc() string {
	if len(colName.Table) == 0 {
		if len(colName.Column) == 0 {
			return "*"
		}
		return encodeName(colName.Column)
	}
	return encodeName(colName.Table) + "." + encodeName(colName.Column)
}

func (aggr *Aggr) sqlDesc() (string, error) {
	col := (&ColName{aggr.Table, aggr.Column}).sqlDesc()
	var format string
	switch aggr.Type {
	case Max:
		format = "MAX(%s)"
	case Min:
		format = "MIN(%s)"
	default:
		return "", fmt.Errorf("data: unknown aggregate type "+
			"(%d) for column (`%s`)", aggr.Type, aggr.Column)
	}
	return fmt.Sprintf(format, col), nil
}

func (order *Order) sqlDesc() string {
	desc := (&ColName{order.Table, order.Column}).sqlDesc()
	if order.Descending {
		desc += " DESC"
	}
	return desc
}
