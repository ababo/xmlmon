package mon

import (
	"btc/data"
	"btc/xmls"
	"fmt"
)

type Schema struct {
	id   int
	Name string
	Desc string
}

func NewSchema(name, desc string) *Schema {
	return &Schema{0, name, desc}
}

func AddSchema(handle data.Handle,
	schema *Schema, root *xmls.Element) error {
	columns := map[string]string{
		"name": schema.Name,
		"desc": schema.Desc,
	}
	var err error
	if schema.id, err = data.InsertRow(
		handle, "mon_schema", columns, "id"); err != nil {
		return err
	}

	traverseFunc := func(path string, element *xmls.Element) error {
		columns := map[string]string{
			"schema": fmt.Sprint(schema.id),
			"path":   path,
			"mon_id": element.MonId,
		}
		id, err := data.InsertRow(handle, "mon_path", columns, "id")
		if err != nil {
			return err
		}

		vtype := element.ValueType()
		columns2 := []data.Column{
			{"doc", data.Integer,
				data.NotNull, "mon_doc", "id"},
			{"time", data.Time, data.NotNull, "", ""},
			{"event", data.Integer, data.NotNull, "", ""},
			{"value", valueToDataType(vtype), 0, "", ""},
		}
		for _, a := range element.Attributes() {
			vtype := a.ValueType
			columns2 = append(columns2,
				data.Column{"attr_" + a.Name,
					valueToDataType(vtype), 0, "", ""})
		}
		indexes := []data.Index{
			{[]string{"doc", "time"}},
		}
		if err := data.CreateTable(handle, "mon_path_"+fmt.Sprint(id),
			columns2, indexes); err != nil {
			return err
		}

		return nil
	}

	return root.Traverse("", traverseFunc)
}

func valueToDataType(xsdType int) int {
	switch xsdType {
	case xmls.String:
		return data.String
	case xmls.Integer:
		return data.Integer
	case xmls.Float:
		return data.Float
	case xmls.Time:
		return data.Time
	default:
		return data.String
	}
}

func FindSchema(handle data.Handle, name string) (*Schema, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{{"", "id"}, {"", "desc"}},
		[]data.Join{{"", "mon_schema", ""}},
		data.Eq{data.ColName{"", "name"}, name}, nil, -1)
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("mon: schema (`%s`) not found", name)
	}

	var schema Schema
	schema.Name = name
	if err = rows.Scan(&schema.id, &schema.Desc); err != nil {
		return nil, err
	}

	return &schema, nil
}
