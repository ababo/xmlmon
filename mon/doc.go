package mon

import (
	"btc/data"
	"fmt"
)

type Doc struct {
	id             int
	Name           string
	Schema         string
	Url            string
	UpdatePeriod   int
	SnapshotPeriod int
}

func NewDoc(name, schema, url string,
	updatePeriod, snapshotPeriod int) *Doc {
	return &Doc{0, name, schema, url, updatePeriod, snapshotPeriod}
}

func AddDoc(handle data.Handle, doc *Doc) error {
	schema, err := FindSchema(handle, doc.Schema)
	if err != nil {
		return err
	}

	columns := map[string]string{
		"name":            doc.Name,
		"schema":          fmt.Sprint(schema.id),
		"url":             doc.Url,
		"update_period":   fmt.Sprint(doc.UpdatePeriod),
		"snapshot_period": fmt.Sprint(doc.SnapshotPeriod),
	}
	doc.id, err = data.InsertRow(handle, "mon_document", columns, "id")

	return err
}

func FindDoc(handle data.Handle, name string) (*Doc, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{
			{"mon_document", "id"},
			{"mon_schema", "name"},
			{"", "url"},
			{"", "update_period"},
			{"", "snapshot_period"}},
		[]data.Join{
			{"", "mon_document", "schema"},
			{"id", "mon_schema", ""}},
		data.Eq{data.ColName{"mon_document", "name"}, name}, nil, -1)
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("document `%s` is not found", name)
	}

	var doc Doc
	doc.Name = name
	if err = rows.Scan(&doc.id, &doc.Schema, &doc.Url,
		&doc.UpdatePeriod, &doc.SnapshotPeriod); err != nil {
		return nil, err
	}

	return &doc, nil
}
