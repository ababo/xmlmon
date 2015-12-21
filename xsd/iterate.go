package xsd

import (
	"fmt"
)

type IterateFunc func(path string, element *Element) error

func (schema *Schema) Iterate(iterateFunc IterateFunc) error {
	for i := range schema.Elements {
		err := schema.Elements[i].iterate("", iterateFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (element *Element) iterate(
	path string, iterateFunc IterateFunc) error {
	if len(element.Ref) != 0 {
		if element.RefElement == nil {
			return fmt.Errorf("xsd: unresolved reference (%s) "+
				"in `element` (%s)", element.Ref, element.Name)
		}
		return element.RefElement.iterate(path, iterateFunc)
	}

	path += "/" + element.Name
	if err := iterateFunc(path, element); err != nil {
		return err
	}

	if element.ComplexType != nil {
		return element.ComplexType.iterate(path, iterateFunc)
	} else if element.SimpleType != nil {
	} else if _, err := xsdToValueType(element.Type); err == nil {
	} else {
		return fmt.Errorf("xsd: no `simpleType` or `complexType` "+
			"found for `element` (%s)", element.Name)
	}

	return nil
}

func (complexType *ComplexType) iterate(
	path string, iterateFunc IterateFunc) error {
	if complexType.Sequence != nil {
		return complexType.Sequence.iterate(path, iterateFunc)
	} else if complexType.SimpleContent != nil {
	} else {
		return fmt.Errorf("xsd: no `sequence` or `simpleContent` "+
			"found for `complexType` (%s)", complexType.Name)
	}
	return nil
}

func (sequence *Sequence) iterate(
	path string, iterateFunc IterateFunc) error {
	for i := range sequence.Elements {
		sequence.Elements[i].iterate(path, iterateFunc)
	}
	return nil
}

func (element *Element) Attributes() []Attribute {
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

func (simpleContent *SimpleContent) attributes() []Attribute {
	if simpleContent.Extension != nil {
		return simpleContent.Extension.Attributes
	}
	return nil
}
