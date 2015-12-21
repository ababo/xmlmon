package xsd

type Schema struct {
	Elements     []Element     `xml:"element"`
	SimpleTypes  []SimpleType  `xml:"simpleType"`
	ComplexTypes []ComplexType `xml:"complexType"`
}

type Element struct {
	Name        string       `xml:"name,attr"`
	Type        string       `xml:"type,attr"`
	MaxOccurs   string       `xml:"maxOccurs,attr"`
	MinOccurs   string       `xml:"minOccurs,attr"`
	Ref         string       `xml:"ref,attr"`
	IdAttribute string       `xml:"idAttribute,attr"` // extension
	SimpleType  *SimpleType  `xml:"simpleType"`
	ComplexType *ComplexType `xml:"complexType"`
	reference   *Element
}

type SimpleType struct {
	Name        string       `xml:"name,attr"`
	Restriction *Restriction `xml:"restriction"`
}

type Restriction struct {
	Base string `xml:"base,attr"`
}

type ComplexType struct {
	Name          string         `xml:"name,attr"`
	Sequence      *Sequence      `xml:"sequence"`
	SimpleContent *SimpleContent `xml:"simpleContent"`
	Attributes    []Attribute    `xml:"attribute"`
}

type Sequence struct {
	Elements []Element `xml:"element"`
}

type SimpleContent struct {
	Extension   *Extension   `xml:"extension"`
	Restriction *Restriction `xml:"restriction"`
}

type Extension struct {
	Attributes []Attribute `xml:"attribute"`
}

type Attribute struct {
	Name       string      `xml:"name,attr"`
	Type       string      `xml:"type,attr"`
	Use        string      `xml:"use,attr"`
	SimpleType *SimpleType `xml:"simpleType"`
}
