package mon

import (
	"btc/data"
)

func Install(handle data.Handle) error {
	columns := []data.Column{
		{"id", data.Integer, data.PrimaryKey, "", ""},
		{"name", data.String, data.NotNull | data.Unique, "", ""},
		{"desc", data.String, 0, "", ""},
	}
	indexes := []data.Index{{[]string{"name"}}}
	if err := data.CreateTable(handle,
		"mon_schema", columns, indexes); err != nil {
		return err
	}

	columns = []data.Column{
		{"id", data.Integer, data.PrimaryKey, "", ""},
		{"name", data.String, data.NotNull | data.Unique, "", ""},
		{"schema", data.Integer, data.NotNull, "mon_schema", "id"},
		{"url", data.String, data.NotNull, "", ""},
		{"uperiod", data.Integer, data.NotNull, "", ""},
		{"speriod", data.Integer, data.NotNull, "", ""},
	}
	if err := data.CreateTable(
		handle, "mon_doc", columns, indexes); err != nil {
		return err
	}

	columns = []data.Column{
		{"id", data.Integer, data.PrimaryKey, "", ""},
		{"schema", data.Integer, data.NotNull, "mon_schema", "id"},
		{"path", data.String, data.NotNull, "", ""},
		{"mon_id", data.String, 0, "", ""},
	}
	indexes = []data.Index{{[]string{"schema"}}}
	if err := data.CreateTable(handle,
		"mon_path", columns, indexes); err != nil {
		return err
	}

	return nil
}
