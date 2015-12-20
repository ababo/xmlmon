package mon

import (
	"btc/data"
	"encoding/xml"
	"io"
)

func CommitDoc(handle data.Handle,
	name string, xmlText io.Reader, snapshot bool) error {
	/*
		findSchemeId(db, name)

		decoder := xml.NewDecoder(xml)

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		//

		return tx.Commit()*/

	return nil
}

func commitPath(handle data.Handle, decoder *xml.Decoder,
	schemaId, docId int, path string, snapshot bool) error {
	/*
		for {
			token, err := decoder.Token()
			if err != nil {
				if err != io.EOF {
					return err
				}
				break
			}

			switch t := token.(type) {
			case xml.StartElement:
				elmt := xml.StartElement(t)
				name := elmt.Name.Local

			case xml.EndElement:
				depth--
				elmt := xml.EndElement(t)
				name := elmt.Name.Local
			case xml.CharData:
				bytes := xml.CharData(t)
			default:
			}
		}
	*/
	return nil
}
