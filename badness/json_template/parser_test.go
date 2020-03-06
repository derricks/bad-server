package json_template

import "testing"

func TestPlainDataTypes(test *testing.T) {
	input := "string;int;bool;increment"
	lexer := newLexer(input)
	parser := NewParser(lexer)

	template, err := parser.ParseTemplate()
	if err != nil {
		test.Fatalf("Unexpected error, %v", err)
	}

	if len(template.Declarations) != 4 {
		test.Errorf("Expected program of 4 but got %d", len(template.Declarations))
	}
}

func TestAssertPeekToken(test *testing.T) {
	input := "string;int"
	lexer := newLexer(input)
	parser := NewParser(lexer)

	err := parser.assertPeekType(INT_DATA_TYPE)
	if err == nil {
		test.Errorf("Expected error but didn't get one")
	}

	err = parser.assertPeekType(SEMICOLON)
	if err != nil {
		test.Errorf("Unexpected error %v", err)
	}
}

func TestAssertPeekOneOf(test *testing.T) {
	input := "string;"
	lexer := newLexer(input)
	parser := NewParser(lexer)

	err := parser.assertPeekTypeOneOf([]TokenType{INT_DATA_TYPE, EQUAL})
	if err == nil {
		test.Errorf("Expected error but didn't get one")
	}

	err = parser.assertPeekTypeOneOf([]TokenType{SEMICOLON, EOF})
	if err != nil {
		test.Errorf("Unexpected error %v", err)
	}
}

// verify the TokenLiteral output for different simple parse strings
type tokenLiteralTest struct {
	input    string
	expected string
}

func TestTokenLiterals(test *testing.T) {
	tests := []tokenLiteralTest{
		{"string", "string"},
		{"int", "int"},
		{"bool", "bool"},
		{"increment", "increment"},
		{"[string]", "[string]:10000"},
		{"[string]:100", "[string]:100"},
		{"book=title/string", "{title: string}"},
		{"book=title/string,pages/[string]", "{title: string, pages: [string]:10000}"},
		{"[string|a,b,c]", "[(a|b|c)]:10000"},
		{"int|1,2,3", "(1|2|3)"},
	}

	for testNumber, testCase := range tests {
		lexer := newLexer(testCase.input)
		parser := NewParser(lexer)
		template, err := parser.ParseTemplate()
		if err != nil {
			test.Fatalf("Test case %d: Unexpected error in parsing: %v", testNumber, err)
		}

		if len(template.Declarations) == 0 {
			test.Errorf("Test case %d: Template is empty", testNumber)
		} else {
			output := template.Declarations[0].TokenLiteral()
			if testCase.expected != output {
				test.Errorf("Test case %d: Expected %s but got %s", testNumber, testCase.expected, output)
			}
		}
	}
}

func TestParseMulti(test *testing.T) {
	tests := map[string][]string{
		"book;book=title/string":                     []string{"book", "{title: string}"},
		"book;book=pages/[page]:1;page=text/string":  []string{"book", "{pages: [page]:1}", "{text: string}"},
		"book;book=pages/[page]:1;;page=text/string": []string{"book", "{pages: [page]:1}", "{text: string}"},
	}

	for input, expected := range tests {
		parser := NewParserWithString(input)
		template, err := parser.ParseTemplate()
		if err != nil {
			test.Fatalf("Unexpected error in parsing %s: %v", input, err)
		}

		for index, expectedToken := range expected {
			actual := template.Declarations[index].TokenLiteral()
			if actual != expectedToken {
				test.Errorf("For input %s, expected %s in position %d but got %s", input, expectedToken, index, actual)
			}
		}

	}
}

func TestParseErrors(test *testing.T) {
	// all of these should generate errors
	tests := []string{
		"string&",
		"[123]",
		"[string]:gh",
		"[123",
		"book=string",
		"book=title",
		"book=title/string/isbn/string",
		"book=pages/[page]:100;%;page=text/string",
		"int|1,a,3",
		"int|",
	}

	for testNumber, testCase := range tests {
		lexer := newLexer(testCase)
		parser := NewParser(lexer)
		_, err := parser.ParseTemplate()
		if err == nil {
			test.Errorf("Test %d (%s) should have produced an error but did not", testNumber, testCase)
		}
	}
}

func TestExtractEnumValues(test *testing.T) {
	testString := "|1,a,c"
	parser := NewParserWithString(testString)
	values := parser.extractEnumData()
	if len(values) != 3 {
		test.Errorf("Expected 3 strings but got %v", len(values))
	}

	expectedValues := []string{"1", "a", "c"}
	for index, extractedString := range values {
		if expectedValues[index] != extractedString {
			test.Errorf("Expected string %v but got %v", expectedValues[index], extractedString)
		}
	}
}
