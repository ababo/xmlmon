package mon

import (
	"btc/data"
	"strings"
	"time"
)

type path struct {
	id     int
	schema int
	path   string
	monId  string
}

func findSchemaPaths(handle data.Handle, schemaId int) ([]path, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{{"", "id"}, {"", "path"}, {"", "mon_id"}},
		[]data.Join{{"", "mon_path", ""}},
		data.Eq{data.ColName{"", "schema"}, schemaId},
		[]data.Order{{"", "path", false}}, -1)
	if err != nil {
		return nil, err
	}

	var paths []path
	for rows.Next() {
		var path path
		if err = rows.Scan(
			&path.id, &path.path, &path.monId); err != nil {
			return nil, err
		}
		path.schema = schemaId
		paths = append(paths, path)
	}

	return paths, nil
}

func filterPaths(paths []path, prefix string) []path {
	var filtered []path
	for _, p := range paths {
		if p.path == prefix || strings.HasPrefix(p.path, prefix+"/") {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

type element struct {
	attrs map[string]string
	value string
}

// element's monId value => element
type parentState map[string]element

// parent's monId value => parentState
type pathState map[string]parentState

// doc path => pathState
type docState map[string]pathState

func computePathState(handle data.Handle,
	path *path, doc int, timestamp time.Time) (pathState, error) {
	return nil, nil
}
