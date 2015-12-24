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
	decoder := xml.NewDecoder(xmlText)
	token, err = decoder.Token()
	for ; err == nil; token, err = decoder.Token() {
		switch token.(type) {
		case xml.StartElement:
			elt := token.(xml.StartElement)
			pathStr := "/" + elt.Name.Local
			paths = filterPaths(paths, pathStr)
			context := commitContext{handle, decoder,
				schema.id, doc.id, snapshot, make(docState)}
			return commitPathTree(&context, "", paths, elt.Attr)
		}
	}

	return err
}

type commitContext struct {
	handle   data.Handle
	decoder  *xml.Decoder
	schema   int
	doc      int
	snapshot bool
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
	if len(paths[0].monId) != 0 {
		attr := findAttr(attrs, paths[0].monId)
		if attr == nil || len(attr.Value) == 0 {
			return "", fmt.Errorf("mon: `monId` attribute "+
				"(`%s`) not found for element path (`%s`)",
				paths[0].monId, paths[0])
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
			"path (`%x`) without `monId`", paths[0].path)
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
			} else {
				value += " " + trimmed
			}
		case xml.EndElement:
			return commitPath(context,
				parent, &paths[0], attrs, value)
		}
	}

	return nil
}

func commitPath(context *commitContext, parent string,
	path *path, attrs []xml.Attr, value string) error {
	fmt.Printf("commitPath path: %s\n", path.path)
	return nil
}
