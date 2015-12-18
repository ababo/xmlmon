package main

import (
	"fmt"
)

func main() {
	schema, err := NewXSDSchemaFromFile("data.xsd")
	if err != nil {
		fmt.Printf("failure: %s\n", err)
		return
	}

	fmt.Printf("%+v\n", *schema)
}
