package json_template

import (
	"errors"
	"fmt"
	"strconv"
)

// parse statements and construct a Template AST

type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	debug     bool
}

// NewParserWithString simplifies some of the work of creating a new parser
func NewParserWithString(input string) *Parser {
	lexer := newLexer(input)
	return NewParser(lexer)
}

func NewParser(lexer *Lexer) *Parser {
	parser := &Parser{lexer: lexer}

	parser.nextToken()
	parser.nextToken()

	return parser
}

func (parser *Parser) nextToken() {
	parser.curToken = parser.peekToken
	parser.peekToken = parser.lexer.NextToken()
	if parser.debug {
		fmt.Printf("curToken: %v, peekToken: %v\n", parser.curToken.Literal, parser.peekToken.Literal)
	}
}

func (parser *Parser) ParseTemplate() (*Template, error) {

	template := Template{make([]DataDeclaration, 0), make(map[string]DataDeclaration)}

	for parser.curToken.Type != EOF {
		if isDataType(parser.curToken.Literal) || parser.curToken.Type == KEY_NAME {

			if isDataType(parser.curToken.Literal) && parser.peekToken.Type == PIPE {
				// parseRawString will figure out if the string under the parser is an enum
				enum, err := parser.parseRawString()
				if err != nil {
					return nil, err
				}
				template.addDataDeclaration(enum)

			} else if parser.curToken.Type == KEY_NAME && parser.peekToken.Type == EQUAL {
				// this defines a custom object
				objectName := parser.curToken.Literal
				parser.nextToken()
				objectDeclaration, err := parser.parseObject()
				if err != nil {
					return nil, err
				}
				template.addDataDeclaration(objectDeclaration)
				template.addCustomType(objectName, objectDeclaration)

			} else {
				// either a data type as is or an object name
				dataDeclaration, err := parser.parseRawString()
				if err != nil {
					return nil, err
				}
				template.addDataDeclaration(dataDeclaration)
			}

		} else if parser.curToken.Type == LEFT_BRACKET {
			// an array
			dataDeclaration, err := parser.parseArray()
			if err != nil {
				return nil, err
			}
			template.addDataDeclaration(dataDeclaration)
		} else if parser.curToken.Type == SEMICOLON {
			// proceed
			parser.nextToken()
			continue
		} else {
			return nil, fmt.Errorf("Unexpected character at position %d: %v", parser.lexer.position, parser.curToken)
		}

		if err := parser.assertPeekTypeOneOf([]TokenType{SEMICOLON, EOF}); err != nil {
			return nil, err
		}

		parser.nextToken()
	}
	return &template, nil
}

// parses enum values into the appropriate Enum*DataType
func (parser *Parser) parseEnum(dataType string, stringValues []string) (DataDeclaration, error) {
	// depending on the data type, create the appropriate Enum*DataType struct
	switch dataType {
	case "string":
		enum := EnumStringDataType{}
		enum.Values = stringValues
		return enum, nil
	case "int":
		enum := EnumIntDataType{}
		enumValueInts := make([]int, 0, len(stringValues))
		for _, curString := range stringValues {
			intValue, err := strconv.Atoi(curString)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Invalid number in int enum: %v", curString))
			} else {
				enumValueInts = append(enumValueInts, intValue)
			}
		}
		enum.Values = enumValueInts
		return enum, nil
	case "float":
		enum := EnumFloatDataType{}
		enumValueFloats := make([]float64, 0, len(stringValues))
		for _, curString := range stringValues {
			floatValue, err := strconv.ParseFloat(curString, 64)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Invalid number in float enum: %v", curString))
			} else {
				enumValueFloats = append(enumValueFloats, floatValue)
			}
		}
		enum.Values = enumValueFloats
		return enum, nil
	default:
		// unlikely, given we have logic above ensuring it's a data type
		return nil, errors.New(fmt.Sprintf("Unrecognized data type: %v", dataType))
	}
}

// pulls enum string values out of a string by iterating on tokens and commas. It assumes
// the parser is sitting on the |
func (parser *Parser) extractEnumData() []string {
	enumValues := make([]string, 0)

	// we need to use KEY_NAME because we don't have another term for an arbitrary string
	for parser.peekToken.Type == COMMA || parser.peekToken.Type == KEY_NAME || parser.peekToken.Type == NUMBER {
		parser.nextToken() // advances
		if parser.curToken.Type != COMMA {
			enumValues = append(enumValues, parser.curToken.Literal)
		}
	}
	return enumValues

}

// parses a string such as either "string" or "other key name"
func (parser *Parser) parseRawString() (DataDeclaration, error) {
	if isDataType(parser.curToken.Literal) {

		if parser.peekToken.Type == PIPE {
			dataType := parser.curToken.Literal
			parser.nextToken()
			if err := parser.assertPeekTypeOneOf([]TokenType{KEY_NAME, NUMBER}); err != nil {
				return nil, errors.New(fmt.Sprintf("No values found for enum"))
			}
			enumValues := parser.extractEnumData()
			return parser.parseEnum(dataType, enumValues)
		}

		return PrimitiveDataType{Literal: parser.curToken.Literal}, nil
	} else {
		return KeyNameDataType{Literal: parser.curToken.Literal}, nil
	}
}

// parses an array. parser is currently at [
func (parser *Parser) parseArray() (DataDeclaration, error) {
	array := ArrayDataType{Length: 10000}
	if err := parser.assertPeekTypeOneOf([]TokenType{KEY_NAME, STRING_DATA_TYPE, BOOL_DATA_TYPE, INT_DATA_TYPE, INCREMENT_DATA_TYPE, FLOAT_DATA_TYPE}); err != nil {
		return nil, err
	}
	parser.nextToken()

	// figure out the nested data declaration
	// note that parseRawString will return an enum type if that's what it is
	dataDeclaration, err := parser.parseRawString()
	if err != nil {
		return nil, err
	}
	array.NestedType = dataDeclaration

	// make sure there's an ] coming up
	if err = parser.assertPeekTypeOneOf([]TokenType{RIGHT_BRACKET}); err != nil {
		return nil, err
	}

	parser.nextToken()

	// if there's a colon, we need to parse the length. Otherwise we can return
	if parser.peekToken.Type == COLON {
		parser.nextToken()

		// check there's a number there
		if err = parser.assertPeekType(NUMBER); err != nil {
			return nil, err
		}

		parser.nextToken()
		length, err := strconv.Atoi(parser.curToken.Literal)
		if err != nil {
			return nil, err
		} else {
			array.Length = length
		}
	}

	return array, nil
}

// the parser is at an = when this is called
func (parser *Parser) parseObject() (DataDeclaration, error) {
	// format is key/value,key/value
	if err := parser.assertPeekType(KEY_NAME); err != nil {
		return nil, err
	}

	keyValues := make([]KeyValueDataType, 0)

	parser.nextToken()

	for parser.curToken.Type != EOF && parser.curToken.Type != SEMICOLON {
		// each start of the loop should place the parser at a key token
		key := parser.curToken.Literal

		if err := parser.assertPeekType(SLASH); err != nil {
			return nil, err
		}

		parser.nextToken()

		if err := parser.assertPeekTypeOneOf([]TokenType{LEFT_BRACKET, KEY_NAME, STRING_DATA_TYPE, INT_DATA_TYPE, BOOL_DATA_TYPE, INCREMENT_DATA_TYPE}); err != nil {
			return nil, err
		}
		parser.nextToken()

		var valueData DataDeclaration
		var parseErr error

		switch parser.curToken.Type {
		case LEFT_BRACKET:
			valueData, parseErr = parser.parseArray()
		case KEY_NAME, STRING_DATA_TYPE, INT_DATA_TYPE, BOOL_DATA_TYPE, INCREMENT_DATA_TYPE:
			valueData, parseErr = parser.parseRawString()
		default:
			valueData = nil
			parseErr = fmt.Errorf("Position %d: Unexpected token %s", parser.lexer.position, parser.curToken.Type)
		}

		if parseErr != nil {
			return nil, parseErr
		}

		if err := parser.assertPeekTypeOneOf([]TokenType{COMMA, SEMICOLON, EOF}); err != nil {
			return nil, err
		}

		keyValues = append(keyValues, KeyValueDataType{Key: key, Value: valueData})

		// exit early if a semicolon/eof is upcoming
		if parser.peekToken.Type == SEMICOLON || parser.peekToken.Type == EOF {
			return ObjectDataType{Members: keyValues}, nil
		}

		if parser.peekToken.Type == COMMA {
			// advance again to get past the comma
			parser.nextToken()
			if err := parser.assertPeekType(KEY_NAME); err != nil {
				return nil, err
			}
			parser.nextToken()
		}

	}

	return ObjectDataType{Members: keyValues}, nil
}

func (parser *Parser) assertPeekTypeOneOf(tokenTypes []TokenType) error {
	for _, tokenType := range tokenTypes {
		err := parser.assertPeekType(tokenType)
		if err == nil {
			// found a match
			return nil
		}
	}
	return fmt.Errorf("Expected one of %v at position %d, but was %s", tokenTypes, parser.lexer.position, parser.peekToken.Type)
}

func (parser *Parser) assertPeekType(tokenType TokenType) error {
	if parser.peekToken.Type != tokenType {
		return fmt.Errorf("Expected %s at %d, got %s", tokenType, parser.lexer.position, parser.peekToken.Type)
	}
	return nil
}
