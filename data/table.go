package data

import (
	"fmt"
	"strings"
)

const ( // column types
	Int  = iota
	Str  = iota
	Time = iota
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

func (column *Column) sqlDesc() string {
	var desc = encodeName(column.Name)

	switch column.Type {
	case Int:
		if column.Flags&PrimaryKey != 0 {
			desc += " serial"
		} else {
			desc += " int"
		}
	case Str:
		desc += " varchar"
	case Time:
		desc += " timestamp with time zone"
	default:
		desc += " ?"
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

	return desc
}

type Index struct {
	Columns []string
}

func (index *Index) sqlDesc(table string) string {
	var cols []string
	for i := range index.Columns {
		cols = append(cols, encodeName(index.Columns[i]))
	}
	return fmt.Sprintf("INDEX ON %s(%s)",
		encodeName(table), strings.Join(cols, ", "))
}

func CreateTable(handle Handle, name string,
	columns []Column, indexes []Index) error {
	var names []string
	for i := range columns {
		names = append(names, columns[i].sqlDesc())
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

func DropTable(handle Handle, name string) error {
	_, err := handle.Query(fmt.Sprintf("DROP TABLE %s", encodeName(name)))
	return err
}
