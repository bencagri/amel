// Package lexer provides tokenization for the AMEL DSL.
package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_SingleCharacterTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{"+", []TokenType{TOKEN_PLUS, TOKEN_EOF}},
		{"-", []TokenType{TOKEN_MINUS, TOKEN_EOF}},
		{"*", []TokenType{TOKEN_STAR, TOKEN_EOF}},
		{"/", []TokenType{TOKEN_SLASH, TOKEN_EOF}},
		{"%", []TokenType{TOKEN_PERCENT, TOKEN_EOF}},
		{"(", []TokenType{TOKEN_LPAREN, TOKEN_EOF}},
		{")", []TokenType{TOKEN_RPAREN, TOKEN_EOF}},
		{"[", []TokenType{TOKEN_LBRACKET, TOKEN_EOF}},
		{"]", []TokenType{TOKEN_RBRACKET, TOKEN_EOF}},
		{",", []TokenType{TOKEN_COMMA, TOKEN_EOF}},
		{".", []TokenType{TOKEN_DOT, TOKEN_EOF}},
		{":", []TokenType{TOKEN_COLON, TOKEN_EOF}},
		{"$", []TokenType{TOKEN_DOLLAR, TOKEN_EOF}},
		{"!", []TokenType{TOKEN_BANG, TOKEN_EOF}},
		{"<", []TokenType{TOKEN_LT, TOKEN_EOF}},
		{">", []TokenType{TOKEN_GT, TOKEN_EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			for _, expected := range tt.expected {
				tok := l.NextToken()
				assert.Equal(t, expected, tok.Type, "input: %s", tt.input)
			}
		})
	}
}

func TestLexer_TwoCharacterTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
		literal  string
	}{
		{"==", TOKEN_EQ, "=="},
		{"!=", TOKEN_NEQ, "!="},
		{"<=", TOKEN_LTE, "<="},
		{">=", TOKEN_GTE, ">="},
		{"&&", TOKEN_LAND, "&&"},
		{"||", TOKEN_LOR, "||"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, tt.expected, tok.Type)
			assert.Equal(t, tt.literal, tok.Literal)
		})
	}
}

func TestLexer_Keywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"true", TOKEN_TRUE},
		{"false", TOKEN_FALSE},
		{"null", TOKEN_NULL},
		{"nil", TOKEN_NULL},
		{"IN", TOKEN_IN},
		{"in", TOKEN_IN},
		{"NOT", TOKEN_NOT},
		{"not", TOKEN_NOT},
		{"AND", TOKEN_AND},
		{"and", TOKEN_AND},
		{"OR", TOKEN_OR},
		{"or", TOKEN_OR},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, tt.expected, tok.Type)
		})
	}
}

func TestLexer_NotIn(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"NOT IN", TOKEN_NOT_IN},
		{"not in", TOKEN_NOT_IN},
		{"NOT  IN", TOKEN_NOT_IN},
		{"NOT\tIN", TOKEN_NOT_IN},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, tt.expected, tok.Type)
		})
	}
}

func TestLexer_Identifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"foo", "foo"},
		{"bar123", "bar123"},
		{"_private", "_private"},
		{"camelCase", "camelCase"},
		{"snake_case", "snake_case"},
		{"UPPERCASE", "UPPERCASE"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, TOKEN_IDENT, tok.Type)
			assert.Equal(t, tt.expected, tok.Literal)
		})
	}
}

func TestLexer_IntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", "0"},
		{"42", "42"},
		{"123456", "123456"},
		{"999999999", "999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, TOKEN_INT, tok.Type)
			assert.Equal(t, tt.expected, tok.Literal)
		})
	}
}

func TestLexer_FloatLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"3.14", "3.14"},
		{"0.5", "0.5"},
		{"123.456", "123.456"},
		{"1e10", "1e10"},
		{"1E10", "1E10"},
		{"1.5e10", "1.5e10"},
		{"1.5e+10", "1.5e+10"},
		{"1.5e-10", "1.5e-10"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, TOKEN_FLOAT, tok.Type)
			assert.Equal(t, tt.expected, tok.Literal)
		})
	}
}

func TestLexer_StringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quotes", `"hello"`, "hello"},
		{"single quotes", `'hello'`, "hello"},
		{"empty string", `""`, ""},
		{"with spaces", `"hello world"`, "hello world"},
		{"escape newline", `"hello\nworld"`, "hello\nworld"},
		{"escape tab", `"hello\tworld"`, "hello\tworld"},
		{"escape carriage return", `"hello\rworld"`, "hello\rworld"},
		{"escape backslash", `"hello\\world"`, "hello\\world"},
		{"escape double quote", `"hello\"world"`, "hello\"world"},
		{"escape single quote", `'hello\'world'`, "hello'world"},
		{"unicode", `"你好世界"`, "你好世界"},
		{"emoji", `"👋🌍"`, "👋🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, TOKEN_STRING, tok.Type)
			assert.Equal(t, tt.expected, tok.Literal)
		})
	}
}

func TestLexer_JSONPath(t *testing.T) {
	input := "$.user.name"
	l := New(input)

	tok := l.NextToken()
	assert.Equal(t, TOKEN_DOLLAR, tok.Type)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_DOT, tok.Type)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_IDENT, tok.Type)
	assert.Equal(t, "user", tok.Literal)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_DOT, tok.Type)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_IDENT, tok.Type)
	assert.Equal(t, "name", tok.Literal)
}

func TestLexer_ComplexExpression(t *testing.T) {
	input := `$.user.age >= 18 && $.user.verified == true`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_DOLLAR, "$"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "user"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "age"},
		{TOKEN_GTE, ">="},
		{TOKEN_INT, "18"},
		{TOKEN_LAND, "&&"},
		{TOKEN_DOLLAR, "$"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "user"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "verified"},
		{TOKEN_EQ, "=="},
		{TOKEN_TRUE, "true"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_FunctionCall(t *testing.T) {
	input := `calculateDiscount($.order.total, 0.15)`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "calculateDiscount"},
		{TOKEN_LPAREN, "("},
		{TOKEN_DOLLAR, "$"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "order"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "total"},
		{TOKEN_COMMA, ","},
		{TOKEN_FLOAT, "0.15"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_ListLiteral(t *testing.T) {
	input := `[1, 2, 3, "admin", "user"]`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_LBRACKET, "["},
		{TOKEN_INT, "1"},
		{TOKEN_COMMA, ","},
		{TOKEN_INT, "2"},
		{TOKEN_COMMA, ","},
		{TOKEN_INT, "3"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, "admin"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, "user"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_InOperator(t *testing.T) {
	input := `$.user.role IN ["admin", "moderator"]`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_DOLLAR, "$"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "user"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "role"},
		{TOKEN_IN, "IN"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_STRING, "admin"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, "moderator"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_NotInOperator(t *testing.T) {
	input := `$.status NOT IN ["deleted", "archived"]`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_DOLLAR, "$"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "status"},
		{TOKEN_NOT_IN, "NOT IN"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_STRING, "deleted"},
		{TOKEN_COMMA, ","},
		{TOKEN_STRING, "archived"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_ArithmeticExpression(t *testing.T) {
	input := `($.price * 1.1) + 5 - $.discount / 2 % 10`
	l := New(input)

	expected := []TokenType{
		TOKEN_LPAREN,
		TOKEN_DOLLAR,
		TOKEN_DOT,
		TOKEN_IDENT, // price
		TOKEN_STAR,
		TOKEN_FLOAT, // 1.1
		TOKEN_RPAREN,
		TOKEN_PLUS,
		TOKEN_INT, // 5
		TOKEN_MINUS,
		TOKEN_DOLLAR,
		TOKEN_DOT,
		TOKEN_IDENT, // discount
		TOKEN_SLASH,
		TOKEN_INT, // 2
		TOKEN_PERCENT,
		TOKEN_INT, // 10
		TOKEN_EOF,
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp, tok.Type, "token %d", i)
	}
}

func TestLexer_Whitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"spaces", "a   +   b"},
		{"tabs", "a\t+\tb"},
		{"newlines", "a\n+\nb"},
		{"mixed", "a \t\n + \t\n b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)

			tok := l.NextToken()
			assert.Equal(t, TOKEN_IDENT, tok.Type)
			assert.Equal(t, "a", tok.Literal)

			tok = l.NextToken()
			assert.Equal(t, TOKEN_PLUS, tok.Type)

			tok = l.NextToken()
			assert.Equal(t, TOKEN_IDENT, tok.Type)
			assert.Equal(t, "b", tok.Literal)
		})
	}
}

func TestLexer_LineComments(t *testing.T) {
	input := `a + b // this is a comment
c + d`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "a"},
		{TOKEN_PLUS, "+"},
		{TOKEN_IDENT, "b"},
		{TOKEN_IDENT, "c"},
		{TOKEN_PLUS, "+"},
		{TOKEN_IDENT, "d"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_BlockComments(t *testing.T) {
	input := `a + /* inline comment */ b`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "a"},
		{TOKEN_PLUS, "+"},
		{TOKEN_IDENT, "b"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_MultiLineBlockComments(t *testing.T) {
	input := `a + /* this is a
multi-line
comment */ b`
	l := New(input)

	tok := l.NextToken()
	assert.Equal(t, TOKEN_IDENT, tok.Type)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_PLUS, tok.Type)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_IDENT, tok.Type)
	assert.Equal(t, "b", tok.Literal)
}

func TestLexer_Position(t *testing.T) {
	input := `a + b
c + d`
	l := New(input)

	tok := l.NextToken() // a
	assert.Equal(t, 1, tok.Line)
	assert.Equal(t, 1, tok.Column)

	tok = l.NextToken() // +
	assert.Equal(t, 1, tok.Line)
	assert.Equal(t, 3, tok.Column)

	tok = l.NextToken() // b
	assert.Equal(t, 1, tok.Line)
	assert.Equal(t, 5, tok.Column)

	tok = l.NextToken() // c (on line 2)
	assert.Equal(t, 2, tok.Line)
	assert.Equal(t, 1, tok.Column)

	tok = l.NextToken() // + (on line 2)
	assert.Equal(t, 2, tok.Line)
	assert.Equal(t, 3, tok.Column)

	tok = l.NextToken() // d (on line 2)
	assert.Equal(t, 2, tok.Line)
	assert.Equal(t, 5, tok.Column)
}

func TestLexer_UnterminatedString(t *testing.T) {
	input := `"hello`
	l := New(input)

	tok := l.NextToken()
	assert.Equal(t, TOKEN_ILLEGAL, tok.Type)

	errors := l.Errors()
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "unterminated string")
}

func TestLexer_InvalidEscapeSequence(t *testing.T) {
	input := `"hello\x"`
	l := New(input)

	l.NextToken()

	errors := l.Errors()
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "invalid escape sequence")
}

func TestLexer_IllegalCharacters(t *testing.T) {
	tests := []struct {
		input string
		char  string
	}{
		{"@", "@"},
		{"#", "#"},
		{"^", "^"},
		{"`", "`"},
		{"~", "~"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()
			assert.Equal(t, TOKEN_ILLEGAL, tok.Type)

			errors := l.Errors()
			require.Len(t, errors, 1)
			assert.Contains(t, errors[0].Error(), "unexpected character")
		})
	}
}

func TestLexer_SingleAmpersand(t *testing.T) {
	input := `a & b`
	l := New(input)

	l.NextToken() // a
	tok := l.NextToken()
	assert.Equal(t, TOKEN_ILLEGAL, tok.Type)

	errors := l.Errors()
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "did you mean '&&'")
}

func TestLexer_SinglePipe(t *testing.T) {
	input := `a | b`
	l := New(input)

	l.NextToken() // a
	tok := l.NextToken()
	assert.Equal(t, TOKEN_ILLEGAL, tok.Type)

	errors := l.Errors()
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "did you mean '||'")
}

func TestLexer_SingleEquals(t *testing.T) {
	input := `a = b`
	l := New(input)

	l.NextToken() // a
	tok := l.NextToken()
	assert.Equal(t, TOKEN_ILLEGAL, tok.Type)

	errors := l.Errors()
	require.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "did you mean '=='")
}

func TestLexer_Peek(t *testing.T) {
	input := `a + b`
	l := New(input)

	// Peek should not advance
	peeked := l.Peek()
	assert.Equal(t, TOKEN_IDENT, peeked.Type)
	assert.Equal(t, "a", peeked.Literal)

	// Peek again should return same token
	peeked2 := l.Peek()
	assert.Equal(t, peeked.Type, peeked2.Type)
	assert.Equal(t, peeked.Literal, peeked2.Literal)

	// Now advance
	tok := l.NextToken()
	assert.Equal(t, TOKEN_IDENT, tok.Type)
	assert.Equal(t, "a", tok.Literal)

	// Next peek should be +
	peeked = l.Peek()
	assert.Equal(t, TOKEN_PLUS, peeked.Type)
}

func TestLexer_ArrayAccess(t *testing.T) {
	input := `$.users[0].name`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_DOLLAR, "$"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "users"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_INT, "0"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_DOT, "."},
		{TOKEN_IDENT, "name"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		assert.Equal(t, exp.tokenType, tok.Type, "token %d", i)
		assert.Equal(t, exp.literal, tok.Literal, "token %d", i)
	}
}

func TestLexer_NegativeNumbers(t *testing.T) {
	// Note: Negative numbers are handled as unary minus in the parser,
	// so -5 is tokenized as MINUS followed by INT
	input := `-5`
	l := New(input)

	tok := l.NextToken()
	assert.Equal(t, TOKEN_MINUS, tok.Type)

	tok = l.NextToken()
	assert.Equal(t, TOKEN_INT, tok.Type)
	assert.Equal(t, "5", tok.Literal)
}

func TestTokenize(t *testing.T) {
	input := `$.age >= 18`
	tokens, errs := Tokenize(input)

	require.Empty(t, errs)
	require.Len(t, tokens, 6) // $, ., age, >=, 18, EOF

	assert.Equal(t, TOKEN_DOLLAR, tokens[0].Type)
	assert.Equal(t, TOKEN_DOT, tokens[1].Type)
	assert.Equal(t, TOKEN_IDENT, tokens[2].Type)
	assert.Equal(t, TOKEN_GTE, tokens[3].Type)
	assert.Equal(t, TOKEN_INT, tokens[4].Type)
	assert.Equal(t, TOKEN_EOF, tokens[5].Type)
}

func TestToken_Is(t *testing.T) {
	tok := Token{Type: TOKEN_PLUS, Literal: "+"}
	assert.True(t, tok.Is(TOKEN_PLUS))
	assert.False(t, tok.Is(TOKEN_MINUS))
}

func TestToken_IsOneOf(t *testing.T) {
	tok := Token{Type: TOKEN_PLUS, Literal: "+"}
	assert.True(t, tok.IsOneOf(TOKEN_PLUS, TOKEN_MINUS))
	assert.False(t, tok.IsOneOf(TOKEN_STAR, TOKEN_SLASH))
}

func TestToken_IsLiteral(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{TOKEN_INT, true},
		{TOKEN_FLOAT, true},
		{TOKEN_STRING, true},
		{TOKEN_TRUE, true},
		{TOKEN_FALSE, true},
		{TOKEN_NULL, true},
		{TOKEN_IDENT, false},
		{TOKEN_PLUS, false},
	}

	for _, tt := range tests {
		tok := Token{Type: tt.tokenType}
		assert.Equal(t, tt.expected, tok.IsLiteral(), "token type: %v", tt.tokenType)
	}
}

func TestToken_IsComparisonOperator(t *testing.T) {
	compOps := []TokenType{TOKEN_EQ, TOKEN_NEQ, TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE, TOKEN_IN, TOKEN_NOT_IN}
	for _, op := range compOps {
		tok := Token{Type: op}
		assert.True(t, tok.IsComparisonOperator(), "token type: %v", op)
	}

	nonCompOps := []TokenType{TOKEN_PLUS, TOKEN_MINUS, TOKEN_LAND, TOKEN_LOR}
	for _, op := range nonCompOps {
		tok := Token{Type: op}
		assert.False(t, tok.IsComparisonOperator(), "token type: %v", op)
	}
}

func TestToken_IsArithmeticOperator(t *testing.T) {
	arithOps := []TokenType{TOKEN_PLUS, TOKEN_MINUS, TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT}
	for _, op := range arithOps {
		tok := Token{Type: op}
		assert.True(t, tok.IsArithmeticOperator(), "token type: %v", op)
	}
}

func TestToken_IsLogicalOperator(t *testing.T) {
	logicalOps := []TokenType{TOKEN_LAND, TOKEN_LOR, TOKEN_BANG, TOKEN_AND, TOKEN_OR, TOKEN_NOT}
	for _, op := range logicalOps {
		tok := Token{Type: op}
		assert.True(t, tok.IsLogicalOperator(), "token type: %v", op)
	}
}
