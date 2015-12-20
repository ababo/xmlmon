package mon

import (
	"btc/data"
	"fmt"
)

func AddDoc(handle data.Handle, name string, schema string,
	url string, updatePeriod int, snapshotPeriod int) error {
	id, err := findSchemeId(handle, schema)
	if err != nil {
		return err
	}

	columns := map[string]string{
		"name":            name,
		"schema":          fmt.Sprint(id),
		"url":             url,
		"update_period":   fmt.Sprint(updatePeriod),
		"snapshot_period": fmt.Sprint(snapshotPeriod),
	}
	_, err = data.InsertRow(handle, "mon_document", columns, "")

	return err
}

func findSchemeId(handle data.Handle, name string) (int, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{{"", "id"}},
		[]data.Join{{"", "mon_schema", ""}},
		data.Eq{data.ColName{"", "name"}, name}, nil, -1)
	defer rows.Close()

	if !rows.Next() {
		return 0, fmt.Errorf("document scheme `%s` not found", name)
	}

	var id int
	if err = rows.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}
