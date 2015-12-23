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

type TraverseFunc func(path string, element *Element) error

func (element *Element) Traverse(
	rootPath string, traverseFunc TraverseFunc) error {
	path := rootPath + "/" + element.Name
	if err := traverseFunc(path, element); err != nil {
		return err
	}

	for _, e := range element.Children() {
		if err := e.Traverse(path, traverseFunc); err != nil {
			return err
		}
	}

	return nil
}
