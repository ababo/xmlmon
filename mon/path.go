package mon

import (
	"btc/data"
	"database/sql"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type path struct {
	id     int
	schema int
	path   string
	monId  sql.NullString
}

func findSchemaPaths(handle data.Handle, schemaId int) ([]path, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{{"", "id"}, {"", "path"}, {"", "mon_id"}},
		[]data.Join{{"", "mon_path", ""}},
		data.Eq{data.ColName{"", "schema"}, schemaId},
		nil, []data.Order{{"", "path", false}}, -1)
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
	attrs    map[string]string
	value    string
	preserve bool
}

func (element *element) isChanged(attrs []xml.Attr, value string) bool {
	return false
}

// element's monId value => element
type parentState map[string]*element

// parent's monId value => parentState
type pathState map[string]parentState

// doc path => pathState
type docState map[string]pathState

func findSnapshot(handle data.Handle,
	path *path, doc *Doc, from time.Time) (time.Time, error) {
	docWhere := data.Eq{doc.id, data.ColName{"", "doc"}}
	timeWhere := data.Gr{from, data.ColName{"", "time"}}
	eventWhere := data.Eq{snapshot, data.ColName{"", "event"}}
	snapshotWhere := data.And{docWhere, data.And{timeWhere, eventWhere}}

	rows, err := data.SelectRows(handle,
		[]data.Aggr{{"", "time", data.Max}},
		[]data.Join{{"", "mon_path_" + fmt.Sprint(path.id), ""}},
		snapshotWhere, nil, nil, -1)
	if err != nil {
		return from, err
	}
	defer rows.Close()

	rows.Next()
	var stime data.NullTime
	if err = rows.Scan(&stime); err != nil {
		return from, err
	}
	if !stime.Valid {
		msg := "data: no snapshot found for document (`%s`)"
		return from, fmt.Errorf(msg, doc.Name)
	}

	return stime.Time, nil
}

func computePathState(handle data.Handle,
	path *path, doc int, from, to time.Time) (pathState, error) {
	return make(pathState), nil
}
