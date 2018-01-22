package json_template

import (
	"fmt"
	"strconv"
)

// parse statements and construct a Template AST

type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
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
}

func (parser *Parser) ParseTemplate() (*Template, error) {

	template := Template{make([]DataDeclaration, 0), make(map[string]DataDeclaration)}

	for parser.curToken.Type != EOF {
		if isDataType(parser.curToken.Literal) || parser.curToken.Type == KEY_NAME {

			if parser.curToken.Type == KEY_NAME && parser.peekToken.Type == EQUAL {
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
				dataDeclaration, err := parser.parseRawString()
				if err != nil {
					return nil, err
				}
				template.addDataDeclaration(dataDeclaration)
			}

		} else if parser.curToken.Type == LEFT_BRACKET {
			dataDeclaration, err := parser.parseArray()
			if err != nil {
				return nil, err
			}
			template.addDataDeclaration(dataDeclaration)
		} else if parser.curToken.Type == SEMICOLON {
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

// parses a string such as either "string" or "other key name"
func (parser *Parser) parseRawString() (DataDeclaration, error) {
	if isDataType(parser.curToken.Literal) {
		return PrimitiveDataType{Literal: parser.curToken.Literal}, nil
	} else {
		return KeyNameDataType{Literal: parser.curToken.Literal}, nil
	}
}

// parses an array. parser is currently at [
func (parser *Parser) parseArray() (DataDeclaration, error) {
	array := ArrayDataType{Length: 10000}
	if err := parser.assertPeekTypeOneOf([]TokenType{KEY_NAME, STRING_DATA_TYPE, BOOL_DATA_TYPE, INT_DATA_TYPE, INCREMENT_DATA_TYPE}); err != nil {
		return nil, err
	}
	parser.nextToken()

	// figure out the nested data declaration
	dataDeclaration, err := parser.parseRawString()
	if err != nil {
		return nil, err
	}
	array.NestedType = dataDeclaration

	// make sure there's an ] coming up
	if err = parser.assertPeekType(RIGHT_BRACKET); err != nil {
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
