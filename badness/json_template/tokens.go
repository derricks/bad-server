package json_template

// json_template provides a small parser for converting a program into a JSON template
// that can in turn be used to create a string of json generators for writing back
// a response that meets the template
//
// Examples:
// [string]:1000 will create an array of 1000 strings
// [book]:1000;book=title/string,author/string will create an array of book objects
// title/string,author/string,chapters/[string]:6 will return an object with a title and an author and a 6-item array of strings

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL       = "ILLEGAL"
	EOF           = "EOF"
	SLASH         = "/"
	LEFT_BRACKET  = "["
	RIGHT_BRACKET = "]"

	// data types
	STRING_DATA_TYPE    = "STRING"
	INT_DATA_TYPE       = "INT"
	INCREMENT_DATA_TYPE = "INCREMENT"
	BOOL_DATA_TYPE      = "BOOL"

	// keys or sizes
	KEY_NAME = "KEY_NAME"
	NUMBER   = "NUMBER"

	EQUAL = "="
	COLON = ":"

	COMMA     = ","
	SEMICOLON = ";"

	PIPE = "|"
)

var dataTypes = map[string]TokenType{
	"string":    STRING_DATA_TYPE,
	"int":       INT_DATA_TYPE,
	"increment": INCREMENT_DATA_TYPE,
	"bool":      BOOL_DATA_TYPE,
}

// stringToToken decides if the passed-string is a known datatype or not
// and returns the appropriate token
func stringToToken(input string) TokenType {
	if isDataType(input) {
		return dataTypes[input]
	}
	return KEY_NAME
}

// determines if the given TokenType is a data type
func isDataType(input string) bool {
	_, found := dataTypes[input]
	return found
}
