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
	name string, xmlText io.Reader, snapshot bool) error {
	doc, err := FindDoc(handle, name)
	if err != nil {
		return err
	}

	schema, err := FindSchema(handle, doc.Schema)
	if err != nil {
		return err
	}

	var paths []path
	paths, err = findSchemaPaths(handle, schema.id)
	if err != nil {
		return err
	}

	var token interface{}
	var elt xml.StartElement
	decoder := xml.NewDecoder(xmlText)
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
	context := commitContext{handle, decoder, schema.id,
		doc.id, snapshot, time.Now(), make(docState)}
	err = commitPathTree(&context, "", paths, elt.Attr)
	if err != nil {
		return err
	}

	if err = commitRemovals(&context); err != nil {
		return err
	}

	return doc.Update(handle, context.now)
}

type commitContext struct {
	handle   data.Handle
	decoder  *xml.Decoder
	schema   int
	doc      int
	snapshot bool
	now      time.Time
	state    docState
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
	parent string, paths []path, attrs []xml.Attr) (string, error) {
	if len(paths) == 0 {
		return "", fmt.Errorf(
			"mon: element path (`%s`) not found", paths[0].path)
	}

	var monIdValue string
	if paths[0].monId.Valid {
		attr := findAttr(attrs, paths[0].monId.String)
		if attr == nil || len(attr.Value) == 0 {
			return "", fmt.Errorf("mon: `monId` attribute "+
				"(`%s`) not found for element path (`%s`)",
				paths[0].monId, paths[0].path)
		}
		_, ok := context.state[paths[0].path][parent][attr.Value]
		if ok {
			return "", fmt.Errorf("mon: non-unique `monId` "+
				"value (`%s`) for path (`%s`) and "+
				"parent (`%s`)", attr.Value, paths[0].path,
				parent)
		}
		monIdValue = attr.Value
	} else if _, ok := context.state[paths[0].path]; ok {
		return "", fmt.Errorf("mon: multiple elements for "+
			"path (`%s`) without `monId`", paths[0].path)
	}

	if _, ok := context.state[paths[0].path]; !ok {
		var err error
		context.state[paths[0].path], err = computePathState(
			context.handle, &paths[0], context.doc, time.Now())
		if err != nil {
			return "", err
		}
	}

	return monIdValue, nil
}

func commitPathTree(context *commitContext,
	parent string, paths []path, attrs []xml.Attr) error {
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
				value += " " + trimmed
			}
		case xml.EndElement:
			return commitPath(context,
				parent, monIdValue, &paths[0], attrs, value)
		}
	}

	return nil
}

func commitPath(context *commitContext,
	parent, monIdValue string, path *path,
	attrs []xml.Attr, value string) error {
	var err error
	if _, ok := context.state[path.path]; !ok {
		if !context.snapshot {
			context.state[path.path], err =
				computePathState(context.handle,
					path, context.doc, context.now)
			if err != nil {
				return err
			}
		}
		context.state[path.path] = make(pathState)
	}

	pathState := context.state[path.path]
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
	element.preserve = true

	return nil
}

const ( // events
	snapshot = iota
	addition = iota
	removal  = iota
	change   = iota
)

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

	for _, a := range attrs {
		columns["attr_"+a.Name.Local] = a.Value
	}

	_, err := data.InsertRow(context.handle,
		"mon_path_"+fmt.Sprint(path.id), columns, "")

	attrs2 := map[string]string{}
	for _, a := range attrs {
		attrs2[a.Name.Local] = a.Value
	}

	context.state[path.path][parent][monIdValue] =
		&element{attrs2, value, true}

	return err
}

func commitRemovals(context *commitContext) error {
	return nil
}
