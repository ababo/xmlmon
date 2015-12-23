package mon

import (
	"btc/data"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
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
			err = commitPathTree(handle, decoder, doc.id,
				paths, pathStr, elt.Attr, snapshot)
		}
		break
	}

	return err
}

type path struct {
	id    int
	path  string
	monId string
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

func commitPathTree(handle data.Handle, decoder *xml.Decoder, docId int,
	paths []path, pathStr string, attrs []xml.Attr, snapshot bool) error {
	if len(paths) == 0 {
		return fmt.Errorf(
			"mon: element path (`%s`) not found", pathStr)
	}

	var value string
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch token.(type) {
		case xml.StartElement:
			elt := token.(xml.StartElement)
			pathStr2 := pathStr + "/" + elt.Name.Local
			paths2 := filterPaths(paths, pathStr2)

			err = commitPathTree(handle, decoder, docId,
				paths2, pathStr2, elt.Attr, snapshot)
			if err != nil {
				return err
			}
		case xml.CharData:
			data := string(token.(xml.CharData))
			trimmed := strings.Trim(data, " \t\r\n")
			if len(paths) > 1 && len(trimmed) != 0 {
				return fmt.Errorf("mon: no value "+
					"expected for element path (`%s`)",
					pathStr)
			} else {
				value += " " + trimmed
			}
		case xml.EndElement:
			return commitPath(handle, docId,
				&paths[0], attrs, value, snapshot)
		}
	}

	return nil
}

func commitPath(handle data.Handle, docId int, path *path,
	attrs []xml.Attr, value string, snapshot bool) error {
	fmt.Printf("commitPath path: %s\n", path.path)
	return nil
}
