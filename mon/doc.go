package mon

import (
	"btc/data"
	"fmt"
	"time"
)

type Doc struct {
	id             int
	Name           string
	Schema         string
	Url            string
	UpdatePeriod   int
	SnapshotPeriod int
	UpdateTime     data.NullTime
}

func NewDoc(name, schema, url string,
	updatePeriod, snapshotPeriod int) *Doc {
	return &Doc{0, name, schema, url,
		updatePeriod, snapshotPeriod, data.TimeAsNull()}
}

func AddDoc(handle data.Handle, doc *Doc) error {
	schema, err := FindSchema(handle, doc.Schema)
	if err != nil {
		return err
	}

	columns := map[string]interface{}{
		"name":    doc.Name,
		"schema":  schema.id,
		"url":     doc.Url,
		"uperiod": doc.UpdatePeriod,
		"speriod": doc.SnapshotPeriod,
		"utime":   doc.UpdateTime,
	}
	doc.id, err = data.InsertRow(handle, "mon_doc", columns, "id")

	return err
}

func FindDoc(handle data.Handle, name string) (*Doc, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{
			{"mon_doc", "id"},
			{"mon_schema", "name"},
			{"", "url"},
			{"", "uperiod"},
			{"", "speriod"},
			{"", "utime"}},
		[]data.Join{
			{"", "mon_doc", "schema"},
			{"id", "mon_schema", ""}},
		data.Eq{data.ColName{"mon_doc", "name"}, name}, nil, -1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("mon: document (`%s`) not found", name)
	}

	var doc Doc
	doc.Name = name
	if err = rows.Scan(&doc.id, &doc.Schema, &doc.Url, &doc.UpdatePeriod,
		&doc.SnapshotPeriod, &doc.UpdateTime); err != nil {
		return nil, err
	}

	return &doc, nil
}

func (doc *Doc) Update(handle data.Handle) error {
	now := time.Now()
	if doc.UpdateTime.Valid && doc.UpdateTime.Time.After(now) {
		return fmt.Errorf("mon: document (`%s`) "+
			"last update time (`%s`) in future",
			doc.Name, doc.UpdateTime.Time.String())
	}

	return data.UpdateRows(handle, "mon_doc",
		map[string]interface{}{"utime": now},
		data.Eq{data.ColName{"", "id"}, doc.id})
}
