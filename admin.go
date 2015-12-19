package main

import (
	"bytes"
	"fmt"
	"io"
)

func InstallDatamon(connection *DBConnection) error {
	if err := connection.BeginTransaction(); err != nil {
		return err
	}
	defer connection.RollbackTransaction()

	columns := []DBColumn{
		{"id", DBInteger, DBPrimaryKey, "", ""},
		{"name", DBString, DBNotNull, "", ""},
		{"description", DBString, 0, "", ""},
		{"xsd", DBString, DBNotNull, "", ""},
	}
	indexes := []DBIndex{
		{[]string{"name"}},
	}
	if err := connection.CreateTable(
		"dm_schema", columns, indexes); err != nil {
		return err
	}

	columns = []DBColumn{
		{"id", DBInteger, DBPrimaryKey, "", ""},
		{"name", DBString, DBNotNull, "", ""},
		{"schema", DBInteger, DBNotNull, "dm_schema", "id"},
		{"url", DBString, DBNotNull, "", ""},
		{"update_period", DBInteger, DBNotNull, "", ""},
		{"snapshot_period", DBInteger, DBNotNull, "", ""},
	}
	if err := connection.CreateTable(
		"dm_document", columns, indexes); err != nil {
		return err
	}

	columns = []DBColumn{
		{"id", DBInteger, DBPrimaryKey, "", ""},
		{"schema", DBInteger, DBNotNull, "dm_schema", "id"},
		{"path", DBString, DBNotNull, "", ""},
		{"id_attribute", DBString, 0, "", ""},
	}
	indexes = []DBIndex{
		{[]string{"schema", "path"}},
	}
	if err := connection.CreateTable("dm_path", columns, indexes); err != nil {
		return err
	}

	return connection.CommitTransaction()
}

func UninstallDatamon(connection *DBConnection) error {
	if err := connection.BeginTransaction(); err != nil {
		return err
	}
	defer connection.RollbackTransaction()

	tables := []string{"dm_path", "dm_document", "dm_schema"}
	for _, t := range tables {
		if err := connection.DropTable(t); err != nil {
			return err
		}
	}

	return connection.CommitTransaction()
}

func InstallSchema(connection *DBConnection,
	name string, description string, xsd io.Reader) error {
	if err := connection.BeginTransaction(); err != nil {
		return err
	}
	defer connection.RollbackTransaction()

	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(xsd); err != nil {
		return err
	}

	columns := map[string]string{
		"name":        name,
		"description": description,
		"xsd":         buf.String(),
	}
	id, err := connection.InsertRow("dm_schema", columns, "id")
	if err != nil {
		return err
	}

	var schema *XSDSchema
	schema, err = NewXSDSchema(buf)
	if err != nil {
		return err
	}

	err = schema.IterateElements(func(path string, element *XSDElement) error {
		columns := map[string]string{
			"schema":       fmt.Sprint(id),
			"path":         path,
			"id_attribute": element.IdAttribute,
		}
		id, err := connection.InsertRow("dm_path", columns, "id")
		if err != nil {
			return err
		}

		columns2 := []DBColumn{
			{"document", DBInteger, DBNotNull, "dm_document", "id"},
			{"time", DBTime, DBNotNull, "", ""},
			{"event", DBInteger, DBNotNull, "", ""},
			{"text", DBString, 0, "", ""},
		}
		for _, a := range element.Attributes() {
			columns2 = append(columns2, DBColumn{a.Name, DBString, 0, "", ""})
		}
		indexes := []DBIndex{
			{[]string{"document", "time"}},
		}
		if err := connection.CreateTable(
			"dm_path_"+fmt.Sprint(id), columns2, indexes); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return connection.CommitTransaction()
}

func UninstallSchema(connection *DBConnection, name string) error {

	return nil
}
