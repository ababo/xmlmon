package xsd

import (
	"encoding/xml"
	"io"
	"os"
)

func (schema *Schema) setupTypes() {
	simpleTypes := map[string]*SimpleType{}
	for i := range schema.SimpleTypes {
		t := &schema.SimpleTypes[i]
		simpleTypes[t.Name] = t
	}

	complexTypes := map[string]*ComplexType{}
	for i := range schema.ComplexTypes {
		t := &schema.ComplexTypes[i]
		complexTypes[t.Name] = t
	}

	elements := map[string]*Element{}
	for i := range schema.Elements {
		e := &schema.Elements[i]
		elements[e.Name] = e
	}

	for i := range schema.Elements {
		schema.Elements[i].setup(
			simpleTypes, complexTypes, elements)
	}

	for i := range schema.ComplexTypes {
		schema.ComplexTypes[i].setup(
			simpleTypes, complexTypes, elements)
	}
}

func (element *Element) setup(
	simpleTypes map[string]*SimpleType,
	complexTypes map[string]*ComplexType,
	elements map[string]*Element) {

	if len(element.Ref) != 0 {
		element.RefElement = elements[element.Ref]
	}

	if element.SimpleType != nil {
		return
	} else if element.ComplexType != nil {
		element.ComplexType.setup(
			simpleTypes, complexTypes, elements)
		return
	}

	if t, ok := simpleTypes[element.Type]; ok {
		element.SimpleType = t
	} else if t, ok := complexTypes[element.Type]; ok {
		element.ComplexType = t
	}
}

func (complexType *ComplexType) setup(
	simpleTypes map[string]*SimpleType,
	complexTypes map[string]*ComplexType,
	elements map[string]*Element) {
	if complexType.Sequence != nil {
		complexType.Sequence.setup(
			simpleTypes, complexTypes, elements)
	}

	if complexType.SimpleContent != nil {
		complexType.SimpleContent.setup(simpleTypes)
	}

	for i := range complexType.Attributes {
		complexType.Attributes[i].setup(simpleTypes)
	}
}

func (sequence *Sequence) setup(
	simpleTypes map[string]*SimpleType,
	complexTypes map[string]*ComplexType,
	elements map[string]*Element) {
	for i := range sequence.Elements {
		sequence.Elements[i].setup(
			simpleTypes, complexTypes, elements)
	}
}

func (simpleContent *SimpleContent) setup(
	simpleTypes map[string]*SimpleType) {
	if simpleContent.Extension != nil {
		simpleContent.Extension.setup(simpleTypes)
	}
}

func (extension *Extension) setup(simpleTypes map[string]*SimpleType) {
	for i := range extension.Attributes {
		extension.Attributes[i].setup(simpleTypes)
	}
}

func (attribute *Attribute) setup(simpleTypes map[string]*SimpleType) {
	if t, ok := simpleTypes[attribute.Type]; ok {
		attribute.SimpleType = t
	}
}

func New(reader io.Reader) (*Schema, error) {
	var schema *Schema
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&schema); err != nil {
		return nil, err
	}

	schema.setupTypes()

	return schema, nil
}

func FromFile(filename string) (*Schema, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return New(file)
}
