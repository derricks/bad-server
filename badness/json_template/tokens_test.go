package json_template

import (
	"testing"
)

func TestStringToToken(test *testing.T) {
	tests := map[string]TokenType{
		"string":    "STRING",
		"int":       "INT",
		"bool":      "BOOL",
		"increment": "INCREMENT",
		"sniffle":   "KEY_NAME",
	}

	for input, expected := range tests {
		token := stringToToken(input)
		if token != expected {
			test.Errorf("Test for %s: expected %s, was %s", input, expected, token)
		}
	}
}

func TestIsDataType(test *testing.T) {
	tests := map[string]bool{
		"string":    true,
		"int":       true,
		"bool":      true,
		"increment": true,
		"sniffle":   false,
	}

	for input, expected := range tests {
		actual := isDataType(input)
		if expected != actual {
			test.Errorf("Test for %s: expected %v, got %v", input, expected, actual)
		}
	}
}
