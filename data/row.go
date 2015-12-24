package data

import (
	"fmt"
	"strings"
)

func InsertRow(handle Handle, table string,
	columns map[string]interface{}, id_column string) (int, error) {
	var keys, values []string
	for k, v := range columns {
		keys = append(keys, encodeName(k))
		values = append(values, encodeValue(v))
	}

	ret := ""
	if len(id_column) != 0 {
		ret = " RETURNING  " + encodeName(id_column)
	}

	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)%s",
		encodeName(table), strings.Join(keys, ", "),
		strings.Join(values, ", "), ret)
	rows, err := handle.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if len(id_column) == 0 {
		return 0, nil
	}

	var id int
	rows.Next()
	if err = rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func UpdateRows(handle Handle, table string,
	columns map[string]interface{}, where interface{}) error {
	var assts []string
	for k, v := range columns {
		assts = append(assts, encodeName(k)+" = "+encodeValue(v))
	}

	sql := fmt.Sprintf("UPDATE %s SET %s",
		encodeName(table), strings.Join(assts, ", "))

	if where != nil {
		expr, err := sqlExpr(where)
		if err != nil {
			return err
		}
		sql += " WHERE " + expr
	}

	_, err := handle.Query(sql)
	return err
}
