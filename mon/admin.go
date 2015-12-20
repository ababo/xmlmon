package mon

import (
	"btc/data"
	"btc/xsd"
	"bytes"
	"fmt"
	"io"
)

func Install(handle data.Handle) error {
	columns := []data.Column{
		{"id", data.Int, data.PrimaryKey, "", ""},
		{"name", data.Str, data.NotNull | data.Unique, "", ""},
		{"description", data.Str, 0, "", ""},
		{"xsd", data.Str, data.NotNull, "", ""},
	}
	indexes := []data.Index{
		{[]string{"name"}},
	}
	if err := data.CreateTable(handle,
		"mon_schema", columns, indexes); err != nil {
		return err
	}

	columns = []data.Column{
		{"id", data.Int, data.PrimaryKey, "", ""},
		{"name", data.Str, data.NotNull | data.Unique, "", ""},
		{"schema", data.Int, data.NotNull, "mon_schema", "id"},
		{"url", data.Str, data.NotNull, "", ""},
		{"update_period", data.Int, data.NotNull, "", ""},
		{"snapshot_period", data.Int, data.NotNull, "", ""},
	}
	if err := data.CreateTable(
		handle, "mon_document", columns, indexes); err != nil {
		return err
	}

	columns = []data.Column{
		{"id", data.Int, data.PrimaryKey, "", ""},
		{"schema", data.Int, data.NotNull, "mon_schema", "id"},
		{"path", data.Str, data.NotNull, "", ""},
		{"id_attribute", data.Str, 0, "", ""},
	}
	indexes = []data.Index{
		{[]string{"schema", "path"}},
	}
	if err := data.CreateTable(handle,
		"mon_path", columns, indexes); err != nil {
		return err
	}

	return nil
}

func InstallSchema(handle data.Handle,
	name string, desc string, xsdText io.Reader) error {
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(xsdText); err != nil {
		return err
	}

	columns := map[string]string{
		"name":        name,
		"description": desc,
		"xsd":         buf.String(),
	}
	id, err := data.InsertRow(handle, "mon_schema", columns, "id")
	if err != nil {
		return err
	}

	var schema *xsd.Schema
	schema, err = xsd.New(buf)
	if err != nil {
		return err
	}

	err = schema.Iterate(func(path string, element *xsd.Element) error {
		columns := map[string]string{
			"schema":       fmt.Sprint(id),
			"path":         path,
			"id_attribute": element.IdAttribute,
		}
		id, err := data.InsertRow(handle, "mon_path", columns, "id")
		if err != nil {
			return err
		}

		columns2 := []data.Column{
			{"document", data.Int,
				data.NotNull, "mon_document", "id"},
			{"time", data.Time, data.NotNull, "", ""},
			{"event", data.Int, data.NotNull, "", ""},
			{"text", data.Str, 0, "", ""},
		}
		for _, a := range element.Attributes() {
			columns2 = append(columns2,
				data.Column{a.Name, data.Str, 0, "", ""})
		}
		indexes := []data.Index{
			{[]string{"document", "time"}},
		}
		if err := data.CreateTable(handle, "mon_path_"+fmt.Sprint(id),
			columns2, indexes); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
