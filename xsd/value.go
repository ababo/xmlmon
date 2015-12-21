package xsd

import (
	"fmt"
)

const ( // value types
	String  = iota
	Integer = iota
	Float   = iota
	Time    = iota
)

func (element *Element) ValueType() (int, error) {
	if element.SimpleType != nil {
		return element.SimpleType.valueType()
	} else if element.ComplexType != nil {
		return element.ComplexType.valueType()
	}
	return xsdToValueType(element.Type)
}

func (attribute *Attribute) ValueType() (int, error) {
	if attribute.SimpleType != nil {
		return attribute.SimpleType.valueType()
	}
	return xsdToValueType(attribute.Type)
}

func (simpleType *SimpleType) valueType() (int, error) {
	if simpleType.Restriction == nil {
		return 0, fmt.Errorf("xsd: no `restriction` found "+
			"for `simpleType` (%s)", simpleType.Name)
	}
	return simpleType.Restriction.valueType()
}

func (restriction *Restriction) valueType() (int, error) {
	if restriction.Choice != nil {
		return 0, choiceNotSupported
	}
	return xsdToValueType(restriction.Base)
}

func (complexType *ComplexType) valueType() (int, error) {
	if complexType.SimpleContent != nil {
		return complexType.SimpleContent.valueType()
	}
	return String, nil
}

func (simpleContent *SimpleContent) valueType() (int, error) {
	if simpleContent.Restriction != nil {
		return simpleContent.Restriction.valueType()
	}
	return String, nil
}

func xsdToValueType(xsdType string) (int, error) {
	switch xsdType {
	case "xs:string", "":
		return String, nil
	case "xs:byte", "unsignedByte", "xs:short", "unsignedShort", "xs:int":
		return Integer, nil
	case "xs:float":
		return Float, nil
	case "xs:time":
		return Time, nil
	default:
		return 0, fmt.Errorf("xsd: unknown value type (%s)", xsdType)
	}
}
