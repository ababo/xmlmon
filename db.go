package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

type DBConnection struct {
	db *sql.DB
	tx *sql.Tx
}

func NewDBConnection(connectionString string) (*DBConnection, error) {
	db, err := sql.Open("postgres", connectionString)

	if err == nil {
		_, err = db.Exec("SELECT")
	}

	if err != nil {
		return nil, err
	}

	return &DBConnection{db, nil}, nil
}

func (connection *DBConnection) Close() {
	connection.RollbackTransaction()
	connection.db.Close()
}

func (connection *DBConnection) BeginTransaction() error {
	if connection.tx != nil {
		return fmt.Errorf("already in transaction")
	}

	tx, err := connection.db.Begin()
	if err == nil {
		connection.tx = tx
	}

	return err
}

func (connection *DBConnection) CommitTransaction() error {
	if connection.tx == nil {
		return fmt.Errorf("not in transaction")
	}

	err := connection.tx.Commit()
	if err == nil {
		connection.tx = nil
	}

	return err
}

func (connection *DBConnection) RollbackTransaction() error {
	if connection.tx == nil {
		return fmt.Errorf("not in transaction")
	}

	err := connection.tx.Rollback()
	if err == nil {
		connection.tx = nil
	}

	return err
}

func (connection *DBConnection) query(sql string) (*sql.Rows, error) {
	if connection.tx != nil {
		return connection.tx.Query(sql)
	} else {
		return connection.db.Query(sql)
	}
}

const (
	DBInteger = iota
	DBString  = iota
)

const (
	DBPrimaryKey = 1 << iota
	DBNotNull    = 1 << iota
)

type DBColumn struct {
	Name         string
	Type         int
	Flags        int
	ForeignTable string
	ForeignKey   string
}

func (column *DBColumn) sql() string {
	var sql = "\"" + column.Name + "\""

	switch column.Type {
	case DBInteger:
		if column.Flags&DBPrimaryKey != 0 {
			sql += " serial"
		} else {
			sql += " int"
		}
	case DBString:
		sql += " varchar"
	default:
		sql += " ?"
	}

	if column.Flags&DBPrimaryKey != 0 {
		sql += " PRIMARY KEY"
	}

	if column.Flags&DBNotNull != 0 {
		sql += " NOT NULL"
	}

	if len(column.ForeignKey) != 0 {
		sql += fmt.Sprintf(" REFERENCES \"%s\"(\"%s\")",
			column.ForeignTable, column.ForeignKey)
	}

	return sql
}

type DBIndex struct {
	Columns []string
}

func (index *DBIndex) sql(table string) string {
	var cols []string
	for i := range index.Columns {
		cols = append(cols, "\""+index.Columns[i]+"\"")
	}
	return fmt.Sprintf("INDEX ON \"%s\"(%s)",
		table, strings.Join(cols, ", "))
}

func (connection *DBConnection) CreateTable(
	name string, columns []DBColumn, indexes []DBIndex) error {
	var names []string
	for i := range columns {
		names = append(names, columns[i].sql())
	}

	sql := fmt.Sprintf("CREATE TABLE \"%s\"(%s)",
		name, strings.Join(names, ", "))
	if _, err := connection.query(sql); err != nil {
		return err
	}

	for i := range indexes {
		sql = "CREATE " + indexes[i].sql(name)
		if _, err := connection.query(sql); err != nil {
			return err
		}
	}

	return nil
}

func (connection *DBConnection) DropTable(name string) error {
	_, err := connection.query(fmt.Sprintf("DROP TABLE \"%s\"", name))
	return err
}

func (connection *DBConnection) InsertRow(table string,
	columns map[string]string, id_column string) (int, error) {
	var keys, values []string
	for k, v := range columns {
		keys = append(keys, "\""+k+"\"")
		values = append(values, "'"+v+"'")
	}

	ret := ""
	if len(id_column) != 0 {
		ret = " RETURNING \"" + id_column + "\""
	}

	sql := fmt.Sprintf("INSERT INTO \"%s\"(%s) VALUES(%s)%s",
		table, strings.Join(keys, ", "), strings.Join(values, ", "), ret)
	rows, err := connection.query(sql)
	if err != nil {
		return 0, err
	}

	if len(id_column) == 0 {
		return 0, nil
	}

	var id int
	rows.Next()
	defer rows.Next()
	if err = rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}
