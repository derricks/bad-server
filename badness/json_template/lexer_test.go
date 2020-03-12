package json_template

import (
	"testing"
)

type tokenExpect struct {
	expectedType    TokenType
	expectedLiteral string
}

func TestNextToken(test *testing.T) {
	input := "=:,/[];string;bookcase;increment;int;bool;1234|1.234"

	expects := []tokenExpect{
		{EQUAL, "="},
		{COLON, ":"},
		{COMMA, ","},
		{SLASH, "/"},
		{LEFT_BRACKET, "["},
		{RIGHT_BRACKET, "]"},
		{SEMICOLON, ";"},
		{STRING_DATA_TYPE, "string"},
		{SEMICOLON, ";"},
		{KEY_NAME, "bookcase"},
		{SEMICOLON, ";"},
		{INCREMENT_DATA_TYPE, "increment"},
		{SEMICOLON, ";"},
		{INT_DATA_TYPE, "int"},
		{SEMICOLON, ";"},
		{BOOL_DATA_TYPE, "bool"},
		{SEMICOLON, ";"},
		{NUMBER, "1234"},
		{PIPE, "|"},
		{NUMBER, "1.234"},
	}
	lexer := newLexer(input)

	for index, expect := range expects {
		token := lexer.NextToken()

		if token.Type != expect.expectedType {
			test.Errorf("Test case %d: Expected %s but got %s", index, expect.expectedType, token.Type)
		}

		if token.Literal != expect.expectedLiteral {
			test.Errorf("Test case %d: Expected %s but got %s", index, expect.expectedLiteral, token.Literal)
		}
	}
}

type isCharExpect struct {
	charTest byte
	lower    byte
	higher   byte
	result   bool
}

func TestIsCharBetween(test *testing.T) {
	expects := []isCharExpect{
		isCharExpect{'q', 'a', 'z', true},
		isCharExpect{'a', 'a', 'z', true},
		isCharExpect{'z', 'a', 'z', true},
		isCharExpect{'A', 'a', 'z', false},
	}

	for index, expect := range expects {
		metExpectation := isCharBetween(expect.charTest, expect.lower, expect.higher)
		if metExpectation != expect.result {
			test.Errorf("Test %d: Expected %v but got %v", index, expect.result, metExpectation)
		}
	}
}

func TestIsLetter(test *testing.T) {
	tests := map[byte]bool{
		'a': true,
		'm': true,
		'Q': true,
		'-': true,
		'_': true,
		'%': false,
	}

	for testChar, expected := range tests {
		isLetter := isLetter(testChar)
		if isLetter != expected {
			test.Errorf("%v should have been %v, was %v", string(testChar), expected, isLetter)
		}
	}
}

func TestIsDigit(test *testing.T) {
	tests := map[byte]bool{
		'0': true,
		'9': true,
		'5': true,
		'a': false,
	}
	for input, expectation := range tests {
		actual := isDigit(input)
		if actual != expectation {
			test.Errorf("For input %v expected %v but got %v", input, expectation, actual)
		}
	}
}

func TestReadString(test *testing.T) {
	// this is testing a method in a lexer, so it's assumed that the lexer
	// will be at the beginning of a string when read (covered in other tests)
	tests := map[string]string{
		"air-brushed":       "air-brushed",
		"snake_case":        "snake_case",
		"camelCase":         "camelCase",
		"string]":           "string",
		"number123sandwich": "number",
	}

	for input, expected := range tests {
		lexer := newLexer(input)
		result := lexer.readString()
		if result != expected {
			test.Errorf("Parsing %s, expected %s got %s", input, expected, result)
		}
	}
}

func TestReadNumber(test *testing.T) {
	tests := map[string]string{
		"01234":      "01234",
		"123456789]": "123456789",
		"5678[901]":  "5678",
	}
	for input, expected := range tests {
		lexer := newLexer(input)
		actual := lexer.readNumber()
		if actual != expected {
			test.Errorf("Parsing %s, expected %s, got %s", input, expected, actual)
		}
	}
}
