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

	for i := range schema.Elements {
		schema.Elements[i].setupTypes(simpleTypes, complexTypes)
	}

	for i := range schema.ComplexTypes {
		schema.ComplexTypes[i].setupTypes(simpleTypes, complexTypes)
	}
}

func (element *Element) setupTypes(
	simpleTypes map[string]*SimpleType,
	complexTypes map[string]*ComplexType) {

	if element.SimpleType != nil {
		return
	} else if element.ComplexType != nil {
		element.ComplexType.setupTypes(simpleTypes, complexTypes)
		return
	}

	if t, ok := simpleTypes[element.Type]; ok {
		element.SimpleType = t
	} else if t, ok := complexTypes[element.Type]; ok {
		element.ComplexType = t
	}
}

func (complexType *ComplexType) setupTypes(
	simpleTypes map[string]*SimpleType,
	complexTypes map[string]*ComplexType) {
	if complexType.Sequence != nil {
		complexType.Sequence.setupTypes(simpleTypes, complexTypes)
	}
}

func (sequence *Sequence) setupTypes(
	simpleTypes map[string]*SimpleType,
	complexTypes map[string]*ComplexType) {
	for i := range sequence.Elements {
		sequence.Elements[i].setupTypes(simpleTypes, complexTypes)
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