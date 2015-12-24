package xmls

const ( // value types
	String  = iota
	Integer = iota
	Float   = iota
	Time    = iota
)

type Attribute struct {
	Name      string
	ValueType int
}

type Element struct {
	Name  string
	type_ *type_
	MonId string
}

type type_ struct {
	sourceType *type_
	attributes []Attribute
	children   []Element
	valueType  int
}

func (element *Element) Attributes() []Attribute {
	var attrs []Attribute
	for type_ := element.type_; type_ != nil; type_ = type_.sourceType {
		attrs = append(attrs, type_.attributes...)
	}
	return attrs
}

func (element *Element) Children() []Element {
	var children []Element
	for type_ := element.type_; type_ != nil; type_ = type_.sourceType {
		children = append(children, type_.children...)
	}
	return children
}

func (element *Element) ValueType() int {
	vtype := String
	for type_ := element.type_; type_ != nil; type_ = type_.sourceType {
		vtype = type_.valueType
	}
	return vtype
}

func (element *Element) MonIdAttr() *Attribute {
	attrs := element.Attributes()
	for i := range attrs {
		if attrs[i].Name == element.MonId {
			return &attrs[i]
		}
	}
	return nil
}

type TraverseFunc func(element, parent *Element, path string) error

func (element *Element) Traverse(traverseFunc TraverseFunc) error {
	var traverse func(element, parent *Element,
		path string, traverseFunc TraverseFunc) error
	traverse = func(element, parent *Element,
		path string, traverseFunc TraverseFunc) error {
		path += "/" + element.Name
		if err := traverseFunc(element, parent, path); err != nil {
			return err
		}

		children := element.Children()
		for i := range children {
			if err := traverse(&children[i], element,
				path, traverseFunc); err != nil {
				return err
			}
		}

		return nil
	}

	return traverse(element, nil, "", traverseFunc)
}
