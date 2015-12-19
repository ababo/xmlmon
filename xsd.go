package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type XSDSchema struct {
	Elements     []XSDElement     `xml:"element"`
	SimpleTypes  []XSDSimpleType  `xml:"simpleType"`
	ComplexTypes []XSDComplexType `xml:"complexType"`
}

func (schema *XSDSchema) setupTypes() {
	simpleTypes := map[string]*XSDSimpleType{}
	for i := range schema.SimpleTypes {
		t := &schema.SimpleTypes[i]
		simpleTypes[t.Name] = t
	}

	complexTypes := map[string]*XSDComplexType{}
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

type NextXSDElementFunc func(path string, element *XSDElement) error

func (schema *XSDSchema) IterateElements(nextFunc NextXSDElementFunc) error {
	for i := range schema.Elements {
		err := schema.Elements[i].iterateElements("", nextFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

type XSDElement struct {
	Name        string          `xml:"name,attr"`
	Type        string          `xml:"type,attr"`
	MaxOccurs   string          `xml:"maxOccurs,attr"`
	MinOccurs   string          `xml:"minOccurs,attr"`
	IdAttribute string          `xml:"idAttribute,attr"` // extension
	SimpleType  *XSDSimpleType  `xml:"simpleType"`
	ComplexType *XSDComplexType `xml:"complexType"`
}

func (element *XSDElement) setupTypes(
	simpleTypes map[string]*XSDSimpleType,
	complexTypes map[string]*XSDComplexType) {

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

func (element *XSDElement) iterateElements(
	path string, nextFunc NextXSDElementFunc) error {
	path += "/" + element.Name
	if err := nextFunc(path, element); err != nil {
		return err
	}

	if element.ComplexType != nil {
		return element.ComplexType.iterateElements(path, nextFunc)
	} else if element.SimpleType != nil {
	} else {
		return fmt.Errorf("type of `element` (%s) neither corresponds "+
			"to `simpleType` nor to `complexType` (not supported)",
			element.Name)
	}

	return nil
}

func (element *XSDElement) Attributes() []XSDAttribute {
	if element.ComplexType == nil {
		return nil
	}
	if element.ComplexType.Attributes != nil {
		return element.ComplexType.Attributes
	}
	if element.ComplexType.SimpleContent != nil {
		return element.ComplexType.SimpleContent.attributes()
	}
	return nil
}

type XSDSimpleType struct {
	Name string `xml:"name,attr"`
}

type XSDComplexType struct {
	Name          string            `xml:"name,attr"`
	Sequence      *XSDSequence      `xml:"sequence"`
	SimpleContent *XSDSimpleContent `xml:"simpleContent"`
	Attributes    []XSDAttribute    `xml:"attribute"`
}

func (complexType *XSDComplexType) setupTypes(
	simpleTypes map[string]*XSDSimpleType,
	complexTypes map[string]*XSDComplexType) {
	if complexType.Sequence != nil {
		complexType.Sequence.setupTypes(simpleTypes, complexTypes)
	}
}

func (complexType *XSDComplexType) iterateElements(
	path string, nextFunc NextXSDElementFunc) error {
	if complexType.Sequence != nil {
		return complexType.Sequence.iterateElements(path, nextFunc)
	} else {
		return fmt.Errorf("`complexType` (%s) doesn't "+
			"contain `sequence` (not supported)", complexType.Name)
	}
}

type XSDSequence struct {
	Elements []XSDElement `xml:"element"`
}

func (sequence *XSDSequence) setupTypes(
	simpleTypes map[string]*XSDSimpleType,
	complexTypes map[string]*XSDComplexType) {
	for i := range sequence.Elements {
		sequence.Elements[i].setupTypes(simpleTypes, complexTypes)
	}
}

func (sequence *XSDSequence) iterateElements(
	path string, nextFunc NextXSDElementFunc) error {
	for i := range sequence.Elements {
		sequence.Elements[i].iterateElements(path, nextFunc)
	}
	return nil
}

type XSDSimpleContent struct {
	Extension *XSDExtension `xml:"extension"`
}

func (simpleContent *XSDSimpleContent) attributes() []XSDAttribute {
	if simpleContent.Extension != nil {
		return simpleContent.Extension.Attributes
	}
	return nil
}

type XSDExtension struct {
	Attributes []XSDAttribute `xml:"attribute"`
}

type XSDAttribute struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Use  string `xml:"type,use"`
}

func NewXSDSchema(reader io.Reader) (*XSDSchema, error) {
	var schema *XSDSchema
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&schema); err != nil {
		return nil, err
	}

	schema.setupTypes()

	return schema, nil
}

func NewXSDSchemaFromFile(filename string) (*XSDSchema, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return NewXSDSchema(file)
}
