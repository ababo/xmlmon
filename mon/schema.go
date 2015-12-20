package mon

import (
	"btc/data"
	"btc/xsd"
	"fmt"
	"io"
)

type Schema struct {
	id   int
	Name string
	Desc string
}

func NewSchema(name, desc string) *Schema {
	return &Schema{0, name, desc}
}

func AddSchema(handle data.Handle, schema *Schema, xsdText io.Reader) error {
	columns := map[string]string{
		"name":        schema.Name,
		"description": schema.Desc,
	}
	var err error
	if schema.id, err = data.InsertRow(
		handle, "mon_schema", columns, "id"); err != nil {
		return err
	}

	var schema2 *xsd.Schema
	schema2, err = xsd.New(xsdText)
	if err != nil {
		return err
	}

	err = schema2.Iterate(func(path string, element *xsd.Element) error {
		columns := map[string]string{
			"schema":       fmt.Sprint(schema.id),
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

func FindSchema(handle data.Handle, name string) (*Schema, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{{"", "id"}, {"", "description"}},
		[]data.Join{{"", "mon_schema", ""}},
		data.Eq{data.ColName{"", "name"}, name}, nil, -1)
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("schema (%s) is not found", name)
	}

	var schema Schema
	schema.Name = name
	if err = rows.Scan(&schema.id, &schema.Desc); err != nil {
		return nil, err
	}

	return &schema, nil
}
