package mon

import (
	"btc/data"
	"encoding/xml"
	//"fmt"
	"io"
	"time"
)

func CheckoutDoc(handle data.Handle,
	name string, timestamp time.Time,
	writer io.Writer, prefix, indent string) error {
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

	var snapshot time.Time
	snapshot, err = findSnapshot(handle, paths[0], doc, timestamp)
	if err != nil {
		return err
	}

	encoder := xml.NewEncoder(writer)
	encoder.Indent(prefix, indent)

	docState := make(docState)
	for _, p := range paths {
		docState[p], err = computePathState(
			handle, p, doc.id, snapshot, timestamp)
		if err != nil {
			return err
		}
	}

	context := checkoutContext{handle, doc.id,
		writer, snapshot, timestamp, encoder, docState}
	err = checkoutPathTree(&context, paths, "")
	if err != nil {
		return err
	}

	return context.encoder.Flush()
}

type checkoutContext struct {
	handle       data.Handle
	doc          int
	writer       io.Writer
	lastSnapshot time.Time
	timestamp    time.Time
	encoder      *xml.Encoder
	state        docState
}

func checkoutPathTree(
	context *checkoutContext, paths []*path, parent string) error {
	base, pathGroups := groupPaths(paths)
	start := xml.StartElement{xml.Name{"", base}, nil}
	end := xml.EndElement{xml.Name{"", base}}

	for monIdVal, element := range context.state[paths[0]][parent] {
		start.Attr = nil
		for n, v := range element.attrs {
			attr := xml.Attr{xml.Name{"", n}, v}
			start.Attr = append(start.Attr, attr)
		}

		if err := context.encoder.EncodeToken(start); err != nil {
			return err
		}

		for _, g := range pathGroups {
			err := checkoutPathTree(context, g, monIdVal)
			if err != nil {
				return err
			}
		}

		if len(element.value) != 0 {
			data := xml.CharData(element.value)
			err := context.encoder.EncodeToken(data)
			if err != nil {
				return err
			}
		}

		if err := context.encoder.EncodeToken(end); err != nil {
			return err
		}
	}

	return nil
}
