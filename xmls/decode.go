package xmls

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

func FromFile(xsdFilename string) (*Element, error) {
	file, err := os.Open(xsdFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return New(file)
}

func New(xsdText io.Reader) (*Element, error) {
	decoder := xml.NewDecoder(xsdText)

	var err error
	var root *Element
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		if elt.Name.Local != "schema" {
			msg := "xmls: expected `xs:schema` but found `%s`"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		root, err = decodeSchema(decoder, elt.Attr)
		return err
	})
	if err != io.EOF {
		return nil, err
	}

	return root, nil
}

func handleTokens(decoder *xml.Decoder,
	eltFunc func(elt *xml.StartElement) error) error {
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}

		switch token.(type) {
		case xml.StartElement:
			elt := token.(xml.StartElement)
			err = eltFunc(&elt)
		case xml.EndElement:
			return nil
		}

		if err != nil {
			return err
		}
	}
}

func newType(valueType int) *type_ {
	return &type_{nil, nil, nil, valueType}
}

func findType(types map[string]*type_, name string) *type_ {
	if type_, ok := types[name]; ok {
		return type_
	}

	type_ := newType(String)
	types[name] = type_

	return type_
}

func decodeSchema(decoder *xml.Decoder, attrs []xml.Attr) (*Element, error) {
	for _, a := range attrs {
		switch a.Name.Local {
		case "xs", "attributeFormDefault", "elementFormDefault":
		default:
			msg := "xmls: unsupported `schema` attribute (`%s`)"
			return nil, fmt.Errorf(msg, a.Name.Local)
		}
	}

	intType := newType(Integer)
	types := map[string]*type_{
		"string":        newType(String),
		"byte":          intType,
		"unsignedByte":  intType,
		"short":         intType,
		"unsignedShort": intType,
		"int":           intType,
		"float":         newType(Float),
		"time":          newType(Time),
	}

	var err error
	var root *Element
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "element":
			if root != nil {
				msg := "xmls: multiple root " +
					"elements not supported"
				return fmt.Errorf(msg)
			}
			root, err = decodeElement(decoder, elt.Attr, types)
		case "simpleType":
			_, err = decodeSimpleType(decoder, elt.Attr, types)
		case "complexType":
			_, err = decodeComplexType(decoder, elt.Attr, types)
		default:
			msg := "xmls: unsupported `schema` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return root, err
}

func decodeElement(decoder *xml.Decoder, attrs []xml.Attr,
	types map[string]*type_) (*Element, error) {
	var element Element
	for _, a := range attrs {
		switch a.Name.Local {
		case "name":
			element.Name = a.Value
		case "type":
			element.type_ = findType(types, a.Value)
		case "minOccurs", "maxOccurs":
		case "monId": // custom attribute (used in `btc/mon`)
			element.MonId = a.Value
		default:
			msg := "xmls: unsupported `element` attribute (`%s`)"
			return nil, fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	var type_ *type_
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "simpleType":
			type_, err = decodeSimpleType(
				decoder, elt.Attr, types)
			element.type_ = type_
		case "complexType":
			type_, err = decodeComplexType(
				decoder, elt.Attr, types)
			element.type_ = type_
		default:
			msg := "xmls: unsupported `element` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return &element, err
}

func decodeSimpleType(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_) (*type_, error) {
	var type_ *type_ = newType(String)
	for _, a := range attrs {
		switch a.Name.Local {
		case "name":
			type_ = findType(types, a.Value)
		default:
			msg := "xmls: unsupported " +
				"`simpleType` attribute (`%s`)"
			return nil, fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "restriction":
			err = decodeRestriction(
				decoder, elt.Attr, types, type_)
		default:
			msg := "xmls: unsupported `simpleType` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return type_, err
}

func decodeRestriction(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_, type_ *type_) error {
	for _, a := range attrs {
		switch a.Name.Local {
		case "base":
			type_.sourceType = findType(types, a.Value)
		default:
			msg := "xmls: unsupported " +
				"`restriction` attribute (`%s`)"
			return fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		default:
			msg := "xmls: unsupported " +
				"`restriction` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return err
}

func decodeComplexType(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_) (*type_, error) {
	var type_ *type_ = newType(String)
	for _, a := range attrs {
		switch a.Name.Local {
		case "name":
			type_ = findType(types, a.Value)
		default:
			msg := "xmls: unsupported " +
				"`complexType` attribute (`%s`)"
			return nil, fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	var attr *Attribute
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "simpleContent":
			err = decodeSimpleContent(
				decoder, elt.Attr, types, type_)
		case "sequence":
			err = decodeSequence(
				decoder, elt.Attr, types, type_)
		case "attribute":
			attr, err = decodeAttribute(decoder, elt.Attr, types)
			if err != nil {
				type_.attributes = append(
					type_.attributes, *attr)
			}
		default:
			msg := "xmls: unsupported " +
				"`complexType` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return type_, err
}

func decodeSimpleContent(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_, type_ *type_) error {
	for _, a := range attrs {
		switch a.Name.Local {
		default:
			msg := "xmls: unsupported " +
				"`simpleContent` attribute (`%s`)"
			return fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "extension":
			err = decodeExtension(
				decoder, elt.Attr, types, type_)
		case "restriction":
			err = decodeRestriction(
				decoder, elt.Attr, types, type_)
		default:
			msg := "xmls: unsupported " +
				"`simpleContent` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return err
}

func decodeExtension(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_, type_ *type_) error {
	for _, a := range attrs {
		switch a.Name.Local {
		case "base":
			type_.sourceType = findType(types, a.Value)
		default:
			msg := "xmls: unsupported " +
				"`extension` attribute (`%s`)"
			return fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	var attr *Attribute
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "attribute":
			attr, err = decodeAttribute(decoder, elt.Attr, types)
			if err != nil {
				type_.attributes = append(
					type_.attributes, *attr)
			}
		default:
			msg := "xmls: unsupported " +
				"`extension` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return err
}

func decodeAttribute(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_) (*Attribute, error) {
	attr := Attribute{"", String}
	type_ := newType(String)
	for _, a := range attrs {
		switch a.Name.Local {
		case "name":
			attr.Name = a.Value
		case "type":
			type_ = findType(types, a.Value)
		case "use":
		default:
			msg := "xmls: unsupported " +
				"`attribute` attribute (`%s`)"
			return nil, fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "simpleType":
			type_, err = decodeSimpleType(
				decoder, elt.Attr, types)
		default:
			msg := "xmls: unsupported " +
				"`attribute` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	attr.ValueType = type_.valueType
	return &attr, err
}

func decodeSequence(decoder *xml.Decoder,
	attrs []xml.Attr, types map[string]*type_, type_ *type_) error {
	for _, a := range attrs {
		switch a.Name.Local {
		default:
			msg := "xmls: unsupported " +
				"`sequence` attribute (`%s`)"
			return fmt.Errorf(msg, a.Name.Local)
		}
	}

	var err error
	var element *Element
	err = handleTokens(decoder, func(elt *xml.StartElement) error {
		switch elt.Name.Local {
		case "element":
			element, err = decodeElement(decoder, elt.Attr, types)
			if err == nil {
				type_.children = append(
					type_.children, *element)
			}
		default:
			msg := "xmls: unsupported " +
				"`sequence` element (`%s`)"
			return fmt.Errorf(msg, elt.Name.Local)
		}
		return err
	})

	return err
}
