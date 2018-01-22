package json_template

import (
	"fmt"
)

// ast for json_template code. A template is defined by a series of
// data declarations.
type DataDeclaration interface {
	TokenLiteral() string
}

type Template struct {
	Declarations []DataDeclaration
	CustomTypes  map[string]DataDeclaration
}

func (template *Template) TokenLiteral() string {
	if len(template.Declarations) > 0 {
		return template.Declarations[0].TokenLiteral()
	} else {
		return ""
	}
}

func (template *Template) addDataDeclaration(def DataDeclaration) {
	template.Declarations = append(template.Declarations, def)
}

func (template *Template) addCustomType(name string, definition DataDeclaration) {
	template.CustomTypes[name] = definition
}

type PrimitiveDataType struct {
	DataDeclaration
	Literal string
}

func (primitive PrimitiveDataType) TokenLiteral() string {
	return primitive.Literal
}

type KeyNameDataType struct {
	DataDeclaration
	Literal string
}

func (keyName KeyNameDataType) TokenLiteral() string {
	return keyName.Literal
}

type ArrayDataType struct {
	DataDeclaration
	NestedType DataDeclaration
	Length     int
}

func (array ArrayDataType) TokenLiteral() string {
	return fmt.Sprintf("[%s]:%d", array.NestedType.TokenLiteral(), array.Length)
}

type KeyValueDataType struct {
	DataDeclaration
	Key   string
	Value DataDeclaration
}

func (keyValue KeyValueDataType) TokenLiteral() string {
	return fmt.Sprintf("%s: %s", keyValue.Key, keyValue.Value.TokenLiteral())
}

type ObjectDataType struct {
	DataDeclaration
	Members []KeyValueDataType
}

func (object ObjectDataType) TokenLiteral() string {
	returnString := "{"

	for index, member := range object.Members {
		returnString += member.TokenLiteral()
		if index != len(object.Members)-1 {
			returnString += ", "
		}
	}

	returnString += "}"
	return returnString
}
