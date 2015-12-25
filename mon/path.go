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

func findSchemaPaths(handle data.Handle, schemaId int) ([]*path, error) {
	rows, err := data.SelectRows(handle,
		[]data.ColName{{"", "id"}, {"", "path"}, {"", "mon_id"}},
		[]data.Join{{"", "mon_path", ""}},
		data.Eq{data.ColName{"", "schema"}, schemaId},
		nil, []data.Order{{"", "path", false}}, -1)
	if err != nil {
		return nil, err
	}

	var paths []*path
	for rows.Next() {
		var path path
		if err = rows.Scan(
			&path.id, &path.path, &path.monId); err != nil {
			return nil, err
		}
		path.schema = schemaId
		paths = append(paths, &path)
	}

	return paths, nil
}

func filterPaths(paths []*path, prefix string) []*path {
	var filtered []*path
	for _, p := range paths {
		if p.path == prefix || strings.HasPrefix(p.path, prefix+"/") {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func groupPaths(paths []*path) (string, [][]*path) {
	first := func(path string) string {
		return strings.SplitN(path[1:], "/", 2)[0]
	}
	last := func(path string) string {
		return path[strings.LastIndex(path, "/")+1:]
	}

	var base string
	var groups [][]*path
	for i := 1; i < len(paths); i += 1 {
		comp := first(paths[i].path[len(paths[0].path):])
		if comp != base {
			base = comp
			groups = append(groups, []*path{paths[i]})
		} else {
			l := &groups[len(groups)-1]
			*l = append(*l, paths[i])
		}
	}

	return last(paths[0].path), groups
}

type element struct {
	attrs    map[string]string
	value    string
	preserve bool
}

func (element *element) isChanged(attrs []xml.Attr, value string) bool {
	if element.value != value || len(element.attrs) != len(attrs) {
		return true
	}

	for _, a := range attrs {
		if element.attrs[a.Name.Local] != a.Value {
			return true
		}
	}

	return false
}

// element's monId value => element
type parentState map[string]*element

// parent's monId value => parentState
type pathState map[string]parentState

// doc path => pathState
type docState map[*path]pathState

func findSnapshot(handle data.Handle,
	path *path, doc *Doc, from time.Time) (time.Time, error) {
	docWhere := data.Eq{doc.id, data.ColName{"", "doc"}}
	timeWhere := data.Ge{from, data.ColName{"", "time"}}
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
		return from, fmt.Errorf("mon: no snapshot found "+
			"for document (`%s`) before `%s`",
			doc.Name, from.String())
	}

	return stime.Time, nil
}

const ( // event types
	snapshot = iota
	addition = iota
	change   = iota
	removal  = iota
)

type event struct {
	doc    int
	time   time.Time
	event  int
	parent string
	value  string
	attrs  map[string]string
}

func findPathEvents(handle data.Handle,
	path, doc int, from, to time.Time) ([]event, error) {
	docWhere := data.Eq{doc, data.ColName{"", "doc"}}
	fromWhere := data.Ge{data.ColName{"", "time"}, from}
	toWhere := data.Ge{to, data.ColName{"", "time"}}
	eventsWhere := data.And{docWhere, data.And{fromWhere, toWhere}}

	rows, err := data.SelectRows(handle, []data.ColName{{"", ""}},
		[]data.Join{{"", "mon_path_" + fmt.Sprint(path), ""}},
		eventsWhere, nil, []data.Order{{"", "time", false}}, -1)
	if err != nil {
		return nil, err
	}

	var cols []string
	if cols, err = rows.Columns(); err != nil {
		return nil, err
	}
	for i := range cols {
		if strings.HasPrefix(cols[i], "attr_") {
			cols[i] = cols[i][5:]
		}
	}

	fixedCount := 3
	params := make([]interface{}, len(cols))
	values := make([]sql.NullString, len(cols)-fixedCount)
	for i := 0; i < len(cols)-fixedCount; i += 1 {
		params[fixedCount+i] = &values[i]
	}

	var events []event
	for rows.Next() {
		var event event
		event.attrs = make(map[string]string)

		params[0], params[1], params[2] =
			&event.doc, &event.time, &event.event
		if err = rows.Scan(params...); err != nil {
			return nil, err
		}

		i := 0
		if fixedCount+i < len(cols) &&
			cols[fixedCount+i] == "parent" {
			if values[i].Valid {
				event.parent = values[i].String
			}
			i += 1
		}
		if fixedCount+i < len(cols) &&
			cols[fixedCount+i] == "value" {
			if values[i].Valid {
				event.value = values[i].String
			}
			i += 1
		}

		for ; i < len(values); i += 1 {
			if values[i].Valid {
				event.attrs[cols[fixedCount+i]] =
					values[i].String
			}
		}

		events = append(events, event)
	}

	return events, nil
}

func computePathState(handle data.Handle,
	path *path, doc int, from, to time.Time) (pathState, error) {
	events, err := findPathEvents(handle, path.id, doc, from, to)
	if err != nil {
		return nil, err
	}

	state := make(pathState)
	for _, e := range events {
		monIdVal := ""
		if path.monId.Valid {
			monIdVal = e.attrs[path.monId.String]
		}

		switch e.event {
		case snapshot, addition, change:
			if _, ok := state[e.parent]; !ok {
				state[e.parent] = make(parentState)
			}
			state[e.parent][monIdVal] =
				&element{e.attrs, e.value, false}
		case removal:
			delete(state[e.parent], monIdVal)
		}
	}

	return state, nil
}
