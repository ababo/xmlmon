package mon

import (
	"btc/data"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"
)

func CommitDoc(handle data.Handle,
	name string, reader io.Reader, snapshot bool) error {
	doc, err := FindDoc(handle, name)
	if err != nil {
		return err
	}

	schema, err := FindSchema(handle, doc.Schema)
	if err != nil {
		return err
	}

	var paths []*path
	paths, err = findSchemaPaths(handle, schema.id)
	if err != nil {
		return err
	}

	now := time.Now()
	var lastSnapshot time.Time
	if !snapshot {
		lastSnapshot, err = findSnapshot(
			handle, paths[0], doc, now)
		if err != nil {
			return err
		}
	}

	var token interface{}
	var elt xml.StartElement
	decoder := xml.NewDecoder(reader)
	token, err = decoder.Token()
L:
	for ; err == nil; token, err = decoder.Token() {
		switch token.(type) {
		case xml.StartElement:
			elt = token.(xml.StartElement)
			break L
		}
	}

	pathStr := "/" + elt.Name.Local
	paths = filterPaths(paths, pathStr)
	if len(paths) == 0 {
		msg := "mon: element path (`%s`) not found"
		return fmt.Errorf(msg, pathStr)
	}

	context := commitContext{handle, decoder, schema.id,
		doc.id, snapshot, lastSnapshot, now, make(docState)}
	err = commitPathTree(&context, "", paths, elt.Attr)
	if err != nil {
		return err
	}

	if !snapshot {
		if err = commitRemovals(&context); err != nil {
			return err
		}
	}

	return doc.Update(handle, context.now)
}

type commitContext struct {
	handle       data.Handle
	decoder      *xml.Decoder
	schema       int
	doc          int
	snapshot     bool
	lastSnapshot time.Time
	now          time.Time
	state        docState
}

func findAttr(attrs []xml.Attr, name string) *xml.Attr {
	for i := range attrs {
		if attrs[i].Name.Local == name {
			return &attrs[i]
		}
	}
	return nil
}

func getMonIdValue(context *commitContext,
	parent string, paths []*path, attrs []xml.Attr) (string, error) {
	var monIdValue string
	if paths[0].monId.Valid {
		attr := findAttr(attrs, paths[0].monId.String)
		if attr == nil || len(attr.Value) == 0 {
			return "", fmt.Errorf("mon: `monId` attribute "+
				"(`%s`) not found for element path (`%s`)",
				paths[0].monId.String, paths[0].path)
		}
		elt, ok := context.state[paths[0]][parent][attr.Value]
		if ok && elt.preserve {
			return "", fmt.Errorf("mon: non-unique `monId` "+
				"value (`%s`) for path (`%s`) and "+
				"parent (`%s`)", attr.Value, paths[0].path,
				parent)
		}
		monIdValue = attr.Value
	} else if _, ok := context.state[paths[0]]; ok {
		return "", fmt.Errorf("mon: multiple elements for "+
			"path (`%s`) without `monId`", paths[0].path)
	}

	return monIdValue, nil
}

func commitPathTree(context *commitContext,
	parent string, paths []*path, attrs []xml.Attr) error {
	monIdValue, err := getMonIdValue(context, parent, paths, attrs)
	if err != nil {
		return err
	}

	var value string
	for {
		token, err := context.decoder.Token()
		if err != nil {
			return err
		}

		switch token.(type) {
		case xml.StartElement:
			elt := token.(xml.StartElement)
			path := paths[0].path + "/" + elt.Name.Local
			paths2 := filterPaths(paths, path)
			if len(paths2) == 0 {
				msg := "mon: element path (`%s`) not found"
				return fmt.Errorf(msg, path)
			}

			err = commitPathTree(context,
				monIdValue, paths2, elt.Attr)
			if err != nil {
				return err
			}
		case xml.CharData:
			data := string(token.(xml.CharData))
			trimmed := strings.Trim(data, " \t\r\n")
			if len(paths) > 1 && len(trimmed) != 0 {
				return fmt.Errorf("mon: no value "+
					"expected for element path (`%s`)",
					paths[0].path)
			} else if len(trimmed) != 0 {
				if len(value) != 0 {
					value += " " + trimmed
				} else {
					value = trimmed
				}
			}
		case xml.EndElement:
			return commitPath(context,
				parent, monIdValue, paths[0], attrs, value)
		}
	}

	return nil
}

func commitPath(context *commitContext,
	parent, monIdValue string, path *path,
	attrs []xml.Attr, value string) error {
	var err error
	if _, ok := context.state[path]; !ok {
		if context.snapshot {
			context.state[path] = make(pathState)
		} else {
			context.state[path], err = computePathState(
				context.handle, path, context.doc,
				context.lastSnapshot, context.now)
			if err != nil {
				return err
			}
		}
	}

	pathState := context.state[path]
	if _, ok := pathState[parent]; !ok {
		pathState[parent] = make(parentState)
	}

	if context.snapshot {
		return addEvent(context, path, snapshot,
			parent, monIdValue, attrs, value)
	}

	parentState := pathState[parent]
	if _, ok := parentState[monIdValue]; !ok {
		return addEvent(context, path, addition,
			parent, monIdValue, attrs, value)
	}

	element := parentState[monIdValue]
	if element.isChanged(attrs, value) {
		return addEvent(context, path, change,
			parent, monIdValue, attrs, value)
	}
	context.state[path][parent][monIdValue].preserve = true

	return nil
}

func addEvent(context *commitContext, path *path, event int, parent,
	monIdValue string, attrs []xml.Attr, value string) error {
	columns := map[string]interface{}{
		"doc":   context.doc,
		"time":  context.now,
		"event": event,
	}

	if len(parent) != 0 {
		columns["parent"] = parent
	}

	if len(value) != 0 {
		columns["value"] = value
	}

	// handy for removal
	if len(monIdValue) != 0 {
		columns["attr_"+path.monId.String] = monIdValue
	}

	for _, a := range attrs {
		columns["attr_"+a.Name.Local] = a.Value
	}

	_, err := data.InsertRow(context.handle,
		"mon_path_"+fmt.Sprint(path.id), columns, "")
	if err != nil {
		return err
	}

	switch event {
	case snapshot:
		break
	case addition:
		attrs2 := map[string]string{}
		for _, a := range attrs {
			attrs2[a.Name.Local] = a.Value
		}
		context.state[path][parent][monIdValue] =
			&element{attrs2, value, true}
	case change, removal:
		context.state[path][parent][monIdValue].preserve = true
	}

	return nil
}

func commitRemovals(context *commitContext) error {
	remove := func(path *path, parent, monIdValue string) error {
		return addEvent(context, path,
			removal, parent, monIdValue, nil, "")
	}

	for path, pathState := range context.state {
		for parent, parentState := range pathState {
			for monIdValue, element := range parentState {
				if !element.preserve {
					err := remove(
						path, parent, monIdValue)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
