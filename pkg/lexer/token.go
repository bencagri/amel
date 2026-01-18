// Package lexer provides tokenization for the AMEL DSL.
package lexer

import "fmt"

// TokenType represents the type of a token.
type TokenType int

const (
	// Special tokens
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_COMMENT

	// Literals
	TOKEN_IDENT  // identifier
	TOKEN_INT    // integer literal
	TOKEN_FLOAT  // float literal
	TOKEN_STRING // string literal

	// Keywords
	TOKEN_TRUE  // true
	TOKEN_FALSE // false
	TOKEN_NULL  // null
	TOKEN_IN    // IN
	TOKEN_NOT   // NOT
	TOKEN_AND   // AND (alternative to &&)
	TOKEN_OR    // OR (alternative to ||)

	// Operators
	TOKEN_PLUS    // +
	TOKEN_MINUS   // -
	TOKEN_STAR    // *
	TOKEN_SLASH   // /
	TOKEN_PERCENT // %

	// Comparison operators
	TOKEN_EQ        // ==
	TOKEN_NEQ       // !=
	TOKEN_LT        // <
	TOKEN_GT        // >
	TOKEN_LTE       // <=
	TOKEN_GTE       // >=
	TOKEN_NOT_IN    // NOT IN
	TOKEN_MATCH     // =~ (regex match)
	TOKEN_NOT_MATCH // !~ (regex not match)

	// Logical operators
	TOKEN_LAND // &&
	TOKEN_LOR  // ||
	TOKEN_BANG // !

	// Delimiters
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_COMMA    // ,
	TOKEN_DOT      // .
	TOKEN_COLON    // :
	TOKEN_ARROW    // =>

	// JSONPath
	TOKEN_DOLLAR // $
)

var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL: "ILLEGAL",
	TOKEN_EOF:     "EOF",
	TOKEN_COMMENT: "COMMENT",

	TOKEN_IDENT:  "IDENT",
	TOKEN_INT:    "INT",
	TOKEN_FLOAT:  "FLOAT",
	TOKEN_STRING: "STRING",

	TOKEN_TRUE:  "TRUE",
	TOKEN_FALSE: "FALSE",
	TOKEN_NULL:  "NULL",
	TOKEN_IN:    "IN",
	TOKEN_NOT:   "NOT",
	TOKEN_AND:   "AND",
	TOKEN_OR:    "OR",

	TOKEN_PLUS:    "+",
	TOKEN_MINUS:   "-",
	TOKEN_STAR:    "*",
	TOKEN_SLASH:   "/",
	TOKEN_PERCENT: "%",

	TOKEN_EQ:        "==",
	TOKEN_NEQ:       "!=",
	TOKEN_LT:        "<",
	TOKEN_GT:        ">",
	TOKEN_LTE:       "<=",
	TOKEN_GTE:       ">=",
	TOKEN_NOT_IN:    "NOT IN",
	TOKEN_MATCH:     "=~",
	TOKEN_NOT_MATCH: "!~",

	TOKEN_LAND: "&&",
	TOKEN_LOR:  "||",
	TOKEN_BANG: "!",

	TOKEN_LPAREN:   "(",
	TOKEN_RPAREN:   ")",
	TOKEN_LBRACKET: "[",
	TOKEN_RBRACKET: "]",
	TOKEN_COMMA:    ",",
	TOKEN_DOT:      ".",
	TOKEN_COLON:    ":",
	TOKEN_ARROW:    "=>",

	TOKEN_DOLLAR: "$",
}

// String returns the string representation of a token type.
func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("TOKEN(%d)", t)
}

// keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
	"null":  TOKEN_NULL,
	"nil":   TOKEN_NULL, // alias for null
	"IN":    TOKEN_IN,
	"in":    TOKEN_IN, // case insensitive
	"NOT":   TOKEN_NOT,
	"not":   TOKEN_NOT, // case insensitive
	"AND":   TOKEN_AND,
	"and":   TOKEN_AND, // case insensitive
	"OR":    TOKEN_OR,
	"or":    TOKEN_OR, // case insensitive
}

// LookupIdent checks if an identifier is a keyword.
// If it is, it returns the keyword token type.
// Otherwise, it returns TOKEN_IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

// Token represents a lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// String returns a string representation of the token for debugging.
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Line: %d, Column: %d}",
		t.Type, t.Literal, t.Line, t.Column)
}

// Is checks if the token is of the given type.
func (t Token) Is(tt TokenType) bool {
	return t.Type == tt
}

// IsOneOf checks if the token is one of the given types.
func (t Token) IsOneOf(types ...TokenType) bool {
	for _, tt := range types {
		if t.Type == tt {
			return true
		}
	}
	return false
}

// IsLiteral checks if the token is a literal (int, float, string, bool, null).
func (t Token) IsLiteral() bool {
	return t.IsOneOf(TOKEN_INT, TOKEN_FLOAT, TOKEN_STRING, TOKEN_TRUE, TOKEN_FALSE, TOKEN_NULL)
}

// IsComparisonOperator checks if the token is a comparison operator.
func (t Token) IsComparisonOperator() bool {
	return t.IsOneOf(TOKEN_EQ, TOKEN_NEQ, TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE, TOKEN_IN, TOKEN_NOT_IN, TOKEN_MATCH, TOKEN_NOT_MATCH)
}

// IsArithmeticOperator checks if the token is an arithmetic operator.
func (t Token) IsArithmeticOperator() bool {
	return t.IsOneOf(TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT)
}

// IsLogicalOperator checks if the token is a logical operator.
func (t Token) IsLogicalOperator() bool {
	return t.IsOneOf(TOKEN_LAND, TOKEN_LOR, TOKEN_BANG, TOKEN_AND, TOKEN_OR, TOKEN_NOT)
}

// NewToken creates a new token with the given type, literal, and position.
func NewToken(tokenType TokenType, literal string, line, column int) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    line,
		Column:  column,
	}
}

// EOF creates an EOF token at the given position.
func EOF(line, column int) Token {
	return Token{
		Type:   TOKEN_EOF,
		Line:   line,
		Column: column,
	}
}

// Illegal creates an illegal token at the given position.
func Illegal(literal string, line, column int) Token {
	return Token{
		Type:    TOKEN_ILLEGAL,
		Literal: literal,
		Line:    line,
		Column:  column,
	}
}
