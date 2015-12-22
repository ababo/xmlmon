package mon

import (
	"btc/data"
	"encoding/xml"
	"io"
)

func CommitDoc(handle data.Handle,
	name string, xmlText io.Reader, snapshot bool) error {
	return nil
}

func commitPath(handle data.Handle, decoder *xml.Decoder,
	schemaId, docId int, path string, snapshot bool) error {
	return nil
}
