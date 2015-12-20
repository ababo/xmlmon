package mon

import (
	"btc/data"
)

func Install(handle data.Handle) error {
	columns := []data.Column{
		{"id", data.Int, data.PrimaryKey, "", ""},
		{"name", data.Str, data.NotNull | data.Unique, "", ""},
		{"description", data.Str, 0, "", ""},
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
