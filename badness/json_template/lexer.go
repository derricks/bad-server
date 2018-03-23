package json_template

type Lexer struct {
	input        string
	position     int  // current position in input
	readPosition int  // next read position
	ch           byte // current char
}

func newLexer(input string) *Lexer {
	lexer := &Lexer{input: input}
	lexer.readChar()
	return lexer
}

func (lexer *Lexer) readChar() {
	if lexer.readPosition >= len(lexer.input) {
		lexer.ch = 0
	} else {
		lexer.ch = lexer.input[lexer.readPosition]
	}
	lexer.position = lexer.readPosition
	lexer.readPosition += 1
}

func (lexer *Lexer) NextToken() Token {
	var token Token

	switch lexer.ch {
	case 0:
		token = newToken(EOF, 0)
	case '=':
		token = newToken(EQUAL, lexer.ch)
	case ';':
		token = newToken(SEMICOLON, lexer.ch)
	case '/':
		token = newToken(SLASH, lexer.ch)
	case '[':
		token = newToken(LEFT_BRACKET, lexer.ch)
	case ']':
		token = newToken(RIGHT_BRACKET, lexer.ch)
	case ':':
		token = newToken(COLON, lexer.ch)
	case ',':
		token = newToken(COMMA, lexer.ch)
	default:
		if isLetter(lexer.ch) {
			fullString := lexer.readString()
			token.Type = stringToToken(fullString)
			token.Literal = fullString
			return token
		} else if isDigit(lexer.ch) {
			number := lexer.readNumber()
			token.Type = NUMBER
			token.Literal = string(number)
			return token
		} else {
			token = newToken(ILLEGAL, lexer.ch)
		}
	}

	lexer.readChar()
	return token
}

func (lexer *Lexer) readString() string {
	position := lexer.position
	for isLetter(lexer.ch) {
		lexer.readChar()
	}
	return lexer.input[position:lexer.position]
}

func (lexer *Lexer) readNumber() string {
	position := lexer.position
	for isDigit(lexer.ch) {
		lexer.readChar()
	}
	return lexer.input[position:lexer.position]
}

func isLetter(char byte) bool {
	return isCharBetween(char, 'a', 'z') || isCharBetween(char, 'A', 'Z') || char == '-' || char == '_'
}

func isDigit(char byte) bool {
	return isCharBetween(char, '0', '9')
}

// isCharBetween determines if char is between lower and higher inclusively
func isCharBetween(char byte, lower byte, higher byte) bool {
	return lower <= char && char <= higher
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{tokenType, string(ch)}
}
