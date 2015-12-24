package data

import (
	"fmt"
	"strings"
)

const ( // column types
	String  = iota
	Integer = iota
	Time    = iota
)

const ( // column flags
	PrimaryKey = 1 << iota
	NotNull    = 1 << iota
	Unique     = 1 << iota
)

type Column struct {
	Name         string
	Type         int
	Flags        int
	ForeignTable string
	ForeignKey   string
}

type Index struct {
	Columns []string
}

func CreateTable(handle Handle, name string,
	columns []Column, indexes []Index) error {
	var names []string
	for i := range columns {
		desc, err := columns[i].sqlDesc()
		if err != nil {
			return err
		}
		names = append(names, desc)
	}

	sql := fmt.Sprintf("CREATE TABLE %s(%s)",
		encodeName(name), strings.Join(names, ", "))
	if _, err := handle.Query(sql); err != nil {
		return err
	}

	for i := range indexes {
		sql = "CREATE " + indexes[i].sqlDesc(name)
		if _, err := handle.Query(sql); err != nil {
			return err
		}
	}

	return nil
}

func (column *Column) sqlDesc() (string, error) {
	var desc = encodeName(column.Name)

	switch column.Type {
	case String:
		desc += " varchar"
	case Integer:
		if column.Flags&PrimaryKey != 0 {
			desc += " serial"
		} else {
			desc += " int"
		}
	case Time:
		desc += " timestamp with time zone"
	default:
		return "", fmt.Errorf("data: unknown type (%d) "+
			"for column (`%s`)", column.Type, column.Name)
	}

	if column.Flags&PrimaryKey != 0 {
		desc += " PRIMARY KEY"
	}

	if column.Flags&NotNull != 0 {
		desc += " NOT NULL"
	}

	if column.Flags&Unique != 0 {
		desc += " UNIQUE"
	}

	if len(column.ForeignTable) != 0 {
		desc += fmt.Sprintf(" REFERENCES %s(%s)",
			encodeName(column.ForeignTable),
			encodeName(column.ForeignKey))
	}

	return desc, nil
}

func (index *Index) sqlDesc(table string) string {
	var cols []string
	for i := range index.Columns {
		cols = append(cols, encodeName(index.Columns[i]))
	}
	return fmt.Sprintf("INDEX ON %s(%s)",
		encodeName(table), strings.Join(cols, ", "))
}

func DropTable(handle Handle, name string) error {
	sql := fmt.Sprintf("DROP TABLE %s", encodeName(name))
	_, err := handle.Query(sql)
	return err
}
