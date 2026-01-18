// Package lexer provides tokenization for the AMEL DSL.
package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bencagri/amel/internal/errors"
)

// Lexer tokenizes AMEL DSL input strings.
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           rune // current character under examination
	line         int  // current line number (1-based)
	column       int  // current column number (1-based)
	startColumn  int  // column at the start of the current token
	startLine    int  // line at the start of the current token
	errors       []error
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// Errors returns any errors encountered during lexing.
func (l *Lexer) Errors() []error {
	return l.errors
}

// Position returns the current line and column.
func (l *Lexer) Position() (line, column int) {
	return l.line, l.column
}

// readChar advances to the next character in the input.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
		l.position = l.readPosition
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
	}

	// Update position tracking
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar looks at the next character without advancing.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

// skipWhitespace skips whitespace characters (space, tab, newline, carriage return).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	l.startColumn = l.column
	l.startLine = l.line

	var tok Token

	switch l.ch {
	case '+':
		tok = l.newToken(TOKEN_PLUS, string(l.ch))
		l.readChar()
	case '-':
		tok = l.newToken(TOKEN_MINUS, string(l.ch))
		l.readChar()
	case '*':
		tok = l.newToken(TOKEN_STAR, string(l.ch))
		l.readChar()
	case '/':
		// Check for comment
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		}
		tok = l.newToken(TOKEN_SLASH, string(l.ch))
		l.readChar()
	case '%':
		tok = l.newToken(TOKEN_PERCENT, string(l.ch))
		l.readChar()
	case '(':
		tok = l.newToken(TOKEN_LPAREN, string(l.ch))
		l.readChar()
	case ')':
		tok = l.newToken(TOKEN_RPAREN, string(l.ch))
		l.readChar()
	case '[':
		tok = l.newToken(TOKEN_LBRACKET, string(l.ch))
		l.readChar()
	case ']':
		tok = l.newToken(TOKEN_RBRACKET, string(l.ch))
		l.readChar()
	case ',':
		tok = l.newToken(TOKEN_COMMA, string(l.ch))
		l.readChar()
	case '.':
		tok = l.newToken(TOKEN_DOT, string(l.ch))
		l.readChar()
	case ':':
		tok = l.newToken(TOKEN_COLON, string(l.ch))
		l.readChar()
	case '$':
		tok = l.newToken(TOKEN_DOLLAR, string(l.ch))
		l.readChar()
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_EQ, string(ch)+string(l.ch))
			l.readChar()
		} else if l.peekChar() == '~' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_MATCH, string(ch)+string(l.ch))
			l.readChar()
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_ARROW, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
			l.addError(errors.NewAtf(errors.ErrUnexpectedCharacter, l.line, l.startColumn,
				"unexpected character '=', did you mean '==', '=~', or '=>'?"))
			l.readChar()
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_NEQ, string(ch)+string(l.ch))
			l.readChar()
		} else if l.peekChar() == '~' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_NOT_MATCH, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.newToken(TOKEN_BANG, string(l.ch))
			l.readChar()
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_LTE, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.newToken(TOKEN_LT, string(l.ch))
			l.readChar()
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_GTE, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.newToken(TOKEN_GT, string(l.ch))
			l.readChar()
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_LAND, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
			l.addError(errors.NewAtf(errors.ErrUnexpectedCharacter, l.line, l.startColumn,
				"unexpected character '&', did you mean '&&'?"))
			l.readChar()
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_LOR, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
			l.addError(errors.NewAtf(errors.ErrUnexpectedCharacter, l.line, l.startColumn,
				"unexpected character '|', did you mean '||'?"))
			l.readChar()
		}
	case '"':
		tok = l.readString('"')
	case '\'':
		tok = l.readString('\'')
	case 0:
		tok = Token{Type: TOKEN_EOF, Literal: "", Line: l.line, Column: l.column}
	default:
		if isDigit(l.ch) {
			tok = l.readNumber()
		} else if isLetter(l.ch) {
			tok = l.readIdentifier()
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
			l.addError(errors.NewAtf(errors.ErrUnexpectedCharacter, l.line, l.startColumn,
				"unexpected character '%c'", l.ch))
			l.readChar()
		}
	}

	return tok
}

// Peek returns the next token without advancing the lexer.
func (l *Lexer) Peek() Token {
	// Save state
	position := l.position
	readPosition := l.readPosition
	ch := l.ch
	line := l.line
	column := l.column
	startColumn := l.startColumn
	errLen := len(l.errors)

	// Get next token
	tok := l.NextToken()

	// Restore state
	l.position = position
	l.readPosition = readPosition
	l.ch = ch
	l.line = line
	l.column = column
	l.startColumn = startColumn
	l.errors = l.errors[:errLen]

	return tok
}

// newToken creates a new token with the current position.
func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.startLine,
		Column:  l.startColumn,
	}
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() Token {
	startPos := l.position
	startCol := l.column

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	literal := l.input[startPos:l.position]
	tokenType := LookupIdent(literal)

	// Handle "NOT IN" as a compound token
	if tokenType == TOKEN_NOT {
		// Save state in case we need to backtrack
		savedPos := l.position
		savedReadPos := l.readPosition
		savedCh := l.ch
		savedLine := l.line
		savedCol := l.column

		l.skipWhitespace()
		if isLetter(l.ch) {
			nextStart := l.position
			for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			nextLiteral := l.input[nextStart:l.position]
			if strings.ToUpper(nextLiteral) == "IN" {
				return Token{
					Type:    TOKEN_NOT_IN,
					Literal: "NOT IN",
					Line:    l.startLine,
					Column:  startCol,
				}
			}
		}

		// Backtrack if it wasn't "NOT IN"
		l.position = savedPos
		l.readPosition = savedReadPos
		l.ch = savedCh
		l.line = savedLine
		l.column = savedCol
	}

	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.startLine,
		Column:  startCol,
	}
}

// readNumber reads a number (integer or float).
func (l *Lexer) readNumber() Token {
	startPos := l.position
	startCol := l.column
	isFloat := false

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'

		// Read fractional part
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Check for exponent (scientific notation)
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar() // consume 'e' or 'E'

		// Optional sign
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}

		// Exponent digits
		if !isDigit(l.ch) {
			l.addError(errors.NewAtf(errors.ErrInvalidNumber, l.line, l.column,
				"expected digits after exponent"))
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[startPos:l.position]
	tokenType := TOKEN_INT
	if isFloat {
		tokenType = TOKEN_FLOAT
	}

	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.startLine,
		Column:  startCol,
	}
}

// readString reads a string literal.
func (l *Lexer) readString(quote rune) Token {
	startCol := l.column
	var sb strings.Builder

	l.readChar() // consume opening quote

	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '\'':
				sb.WriteRune('\'')
			case '0':
				sb.WriteRune('\x00')
			default:
				l.addError(errors.NewAtf(errors.ErrInvalidEscape, l.line, l.column,
					"invalid escape sequence '\\%c'", l.ch))
				sb.WriteRune(l.ch)
			}
		} else if l.ch == '\n' {
			l.addError(errors.NewAtf(errors.ErrUnterminatedString, l.line, startCol,
				"newline in string literal"))
			break
		} else {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}

	if l.ch != quote {
		l.addError(errors.NewAtf(errors.ErrUnterminatedString, l.line, startCol,
			"unterminated string literal"))
		return Token{
			Type:    TOKEN_ILLEGAL,
			Literal: sb.String(),
			Line:    l.line,
			Column:  startCol,
		}
	}

	l.readChar() // consume closing quote

	return Token{
		Type:    TOKEN_STRING,
		Literal: sb.String(),
		Line:    l.line,
		Column:  startCol,
	}
}

// skipLineComment skips a line comment (// ...).
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips a block comment (/* ... */).
func (l *Lexer) skipBlockComment() {
	l.readChar() // consume '/'
	l.readChar() // consume '*'

	for {
		if l.ch == 0 {
			l.addError(errors.NewAtf(errors.ErrUnterminatedString, l.line, l.column,
				"unterminated block comment"))
			return
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume '*'
			l.readChar() // consume '/'
			return
		}
		l.readChar()
	}
}

// addError adds an error to the lexer's error list.
func (l *Lexer) addError(err error) {
	l.errors = append(l.errors, err)
}

// isLetter checks if a rune is a letter or underscore.
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isDigit checks if a rune is a digit.
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// Tokenize returns all tokens from the input.
func Tokenize(input string) ([]Token, []error) {
	l := New(input)
	var tokens []Token

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}

	return tokens, l.Errors()
}
