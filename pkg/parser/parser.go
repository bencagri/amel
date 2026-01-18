// Package parser implements a recursive descent parser for the AMEL DSL.
package parser

import (
	"fmt"
	"strconv"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/ast"
	"github.com/bencagri/amel/pkg/lexer"
)

// Precedence levels for operators (lower number = lower precedence)
const (
	_ int = iota
	LOWEST
	LAMBDA      // =>
	OR          // ||, OR
	AND         // &&, AND
	NOT         // ! (unary)
	EQUALS      // ==, !=
	LESSGREATER // <, >, <=, >=
	REGEX       // =~, !~
	IN          // IN, NOT IN
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // -X, !X
	CALL        // function(X)
	INDEX       // array[index], obj.property
)

// Operator precedence mapping
var precedences = map[lexer.TokenType]int{
	lexer.TOKEN_ARROW:     LAMBDA,
	lexer.TOKEN_LOR:       OR,
	lexer.TOKEN_OR:        OR,
	lexer.TOKEN_LAND:      AND,
	lexer.TOKEN_AND:       AND,
	lexer.TOKEN_EQ:        EQUALS,
	lexer.TOKEN_NEQ:       EQUALS,
	lexer.TOKEN_LT:        LESSGREATER,
	lexer.TOKEN_GT:        LESSGREATER,
	lexer.TOKEN_LTE:       LESSGREATER,
	lexer.TOKEN_GTE:       LESSGREATER,
	lexer.TOKEN_MATCH:     REGEX,
	lexer.TOKEN_NOT_MATCH: REGEX,
	lexer.TOKEN_IN:        IN,
	lexer.TOKEN_NOT_IN:    IN,
	lexer.TOKEN_PLUS:      SUM,
	lexer.TOKEN_MINUS:     SUM,
	lexer.TOKEN_STAR:      PRODUCT,
	lexer.TOKEN_SLASH:     PRODUCT,
	lexer.TOKEN_PERCENT:   PRODUCT,
	lexer.TOKEN_LPAREN:    CALL,
	lexer.TOKEN_LBRACKET:  INDEX,
	lexer.TOKEN_DOT:       INDEX,
}

// Parser parses AMEL DSL expressions into an AST.
type Parser struct {
	lexer  *lexer.Lexer
	errors []error

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new Parser for the given input string.
func New(input string) *Parser {
	l := lexer.New(input)
	p := &Parser{
		lexer:  l,
		errors: []error{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.TOKEN_IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.TOKEN_INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TOKEN_FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.TOKEN_STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TOKEN_TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TOKEN_FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TOKEN_NULL, p.parseNullLiteral)
	p.registerPrefix(lexer.TOKEN_BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.TOKEN_LBRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.TOKEN_DOLLAR, p.parseJSONPath)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_STAR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_NEQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LTE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GTE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LAND, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LOR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_AND, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_OR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_IN, p.parseInExpression)
	p.registerInfix(lexer.TOKEN_NOT_IN, p.parseInExpression)
	p.registerInfix(lexer.TOKEN_MATCH, p.parseRegexExpression)
	p.registerInfix(lexer.TOKEN_NOT_MATCH, p.parseRegexExpression)
	p.registerInfix(lexer.TOKEN_ARROW, p.parseLambdaExpression)
	p.registerInfix(lexer.TOKEN_LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.TOKEN_LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.TOKEN_DOT, p.parseMemberExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// NewFromLexer creates a new Parser using an existing lexer.
func NewFromLexer(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []error{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.TOKEN_IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.TOKEN_INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TOKEN_FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.TOKEN_STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TOKEN_TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TOKEN_FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.TOKEN_NULL, p.parseNullLiteral)
	p.registerPrefix(lexer.TOKEN_BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.TOKEN_LBRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.TOKEN_DOLLAR, p.parseJSONPath)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_STAR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_NEQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LTE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GTE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LAND, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LOR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_AND, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_OR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_IN, p.parseInExpression)
	p.registerInfix(lexer.TOKEN_NOT_IN, p.parseInExpression)
	p.registerInfix(lexer.TOKEN_MATCH, p.parseRegexExpression)
	p.registerInfix(lexer.TOKEN_NOT_MATCH, p.parseRegexExpression)
	p.registerInfix(lexer.TOKEN_ARROW, p.parseLambdaExpression)
	p.registerInfix(lexer.TOKEN_LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.TOKEN_LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.TOKEN_DOT, p.parseMemberExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Parse parses the input and returns the AST root expression.
func (p *Parser) Parse() (ast.Expression, error) {
	expr := p.parseExpression(LOWEST)

	// After parsing, we should be at EOF or have consumed everything
	// Check peekToken since the last parsed token's next should be EOF
	if !p.peekTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_EOF) {
		p.addError(errors.NewAtf(errors.ErrUnexpectedToken, p.peekToken.Line, p.peekToken.Column,
			"unexpected token %s after expression", p.peekToken.Type))
	}

	// Collect lexer errors
	for _, err := range p.lexer.Errors() {
		p.errors = append(p.errors, err)
	}

	if len(p.errors) > 0 {
		return expr, p.errors[0]
	}

	return expr, nil
}

// Errors returns all parsing errors encountered.
func (p *Parser) Errors() []error {
	return p.errors
}

// ============================================================================
// Internal helpers
// ============================================================================

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t lexer.TokenType) {
	p.addError(errors.NewAtf(errors.ErrUnexpectedToken, p.peekToken.Line, p.peekToken.Column,
		"expected %s, got %s", t, p.peekToken.Type))
}

func (p *Parser) addError(err error) {
	p.errors = append(p.errors, err)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	p.addError(errors.NewAtf(errors.ErrUnexpectedToken, p.curToken.Line, p.curToken.Column,
		"unexpected token %s", t))
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// ============================================================================
// Expression parsing
// ============================================================================

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TOKEN_EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// ============================================================================
// Prefix parsers
// ============================================================================

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.addError(errors.NewAtf(errors.ErrInvalidNumber, p.curToken.Line, p.curToken.Column,
			"could not parse %q as integer", p.curToken.Literal))
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.addError(errors.NewAtf(errors.ErrInvalidNumber, p.curToken.Line, p.curToken.Column,
			"could not parse %q as float", p.curToken.Literal))
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curTokenIs(lexer.TOKEN_TRUE),
	}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Operand = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	openParen := p.curToken
	p.nextToken()

	// Check if this could be a multi-parameter lambda: (a, b) => expr
	// First, try to parse as identifier list
	if p.curTokenIs(lexer.TOKEN_IDENT) {
		firstIdent := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// Check if next is comma (multi-param) or rparen (single param that might be lambda)
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			// Multi-parameter case: (a, b, ...) => expr
			params := []*ast.Identifier{firstIdent}

			for p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken() // consume comma
				if !p.expectPeek(lexer.TOKEN_IDENT) {
					return nil
				}
				params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
			}

			if !p.expectPeek(lexer.TOKEN_RPAREN) {
				return nil
			}

			// Check if this is followed by '=>'
			if p.peekTokenIs(lexer.TOKEN_ARROW) {
				p.nextToken() // consume '=>'
				return p.parseMultiParamLambda(params, p.curToken)
			}

			// Not a lambda, but we parsed identifiers with commas - this is invalid
			p.addError(errors.NewAtf(errors.ErrInvalidSyntax, openParen.Line, openParen.Column,
				"unexpected comma in grouped expression"))
			return nil
		} else if p.peekTokenIs(lexer.TOKEN_RPAREN) {
			// Single identifier in parens - could be (x) => expr or just (x)
			p.nextToken() // consume ')'

			if p.peekTokenIs(lexer.TOKEN_ARROW) {
				p.nextToken() // consume '=>'
				return p.parseMultiParamLambda([]*ast.Identifier{firstIdent}, p.curToken)
			}

			// Just a grouped identifier
			return firstIdent
		}
	}

	// Regular grouped expression
	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseListLiteral() ast.Expression {
	list := &ast.ListLiteral{Token: p.curToken}
	list.Elements = p.parseExpressionList(lexer.TOKEN_RBRACKET)
	return list
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseJSONPath() ast.Expression {
	jp := &ast.JSONPathExpression{
		Token: p.curToken,
		Path:  "$",
	}

	// Parse the path segments
	for p.peekTokenIs(lexer.TOKEN_DOT) || p.peekTokenIs(lexer.TOKEN_LBRACKET) {
		if p.peekTokenIs(lexer.TOKEN_DOT) {
			p.nextToken() // consume '.'
			jp.Path += "."

			if !p.peekTokenIs(lexer.TOKEN_IDENT) {
				p.addError(errors.NewAtf(errors.ErrInvalidJSONPath, p.curToken.Line, p.curToken.Column,
					"expected identifier after '.' in JSON path"))
				return jp
			}
			p.nextToken() // consume identifier
			jp.Path += p.curToken.Literal
		} else if p.peekTokenIs(lexer.TOKEN_LBRACKET) {
			p.nextToken() // consume '['
			jp.Path += "["

			if p.peekTokenIs(lexer.TOKEN_INT) {
				p.nextToken()
				jp.Path += p.curToken.Literal
			} else if p.peekTokenIs(lexer.TOKEN_STRING) {
				p.nextToken()
				jp.Path += fmt.Sprintf("%q", p.curToken.Literal)
			} else {
				p.addError(errors.NewAtf(errors.ErrInvalidJSONPath, p.curToken.Line, p.curToken.Column,
					"expected integer or string in JSON path bracket"))
				return jp
			}

			if !p.expectPeek(lexer.TOKEN_RBRACKET) {
				return jp
			}
			jp.Path += "]"
		}
	}

	return jp
}

// ============================================================================
// Infix parsers
// ============================================================================

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	expression := &ast.InExpression{
		Token:   p.curToken,
		Left:    left,
		Negated: p.curTokenIs(lexer.TOKEN_NOT_IN),
	}

	p.nextToken()
	expression.Right = p.parseExpression(IN)

	return expression
}

func (p *Parser) parseRegexExpression(left ast.Expression) ast.Expression {
	expression := &ast.RegexExpression{
		Token:   p.curToken,
		Left:    left,
		Negated: p.curTokenIs(lexer.TOKEN_NOT_MATCH),
	}

	p.nextToken()
	expression.Pattern = p.parseExpression(REGEX)

	return expression
}

func (p *Parser) parseLambdaExpression(left ast.Expression) ast.Expression {
	// The left side should be an identifier (single parameter) or a grouped expression with identifiers
	token := p.curToken

	var params []*ast.Identifier

	switch l := left.(type) {
	case *ast.Identifier:
		params = []*ast.Identifier{l}
	default:
		p.addError(errors.NewAtf(errors.ErrInvalidSyntax, token.Line, token.Column,
			"expected identifier before '=>'"))
		return nil
	}

	p.nextToken() // move past '=>'
	body := p.parseExpression(LAMBDA)

	return &ast.LambdaExpression{
		Token:      token,
		Parameters: params,
		Body:       body,
	}
}

// parseMultiParamLambda parses a lambda with multiple parameters: (a, b) => expr
// This is called from parseGroupedExpression when we detect '=>' after the params
func (p *Parser) parseMultiParamLambda(params []*ast.Identifier, arrowToken lexer.Token) ast.Expression {
	p.nextToken() // move past '=>'
	body := p.parseExpression(LAMBDA)

	return &ast.LambdaExpression{
		Token:      arrowToken,
		Parameters: params,
		Body:       body,
	}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// The function should be an identifier
	ident, ok := function.(*ast.Identifier)
	if !ok {
		p.addError(errors.NewAtf(errors.ErrInvalidSyntax, p.curToken.Line, p.curToken.Column,
			"expected function name before '('"))
		return nil
	}

	exp := &ast.FunctionCall{
		Token: p.curToken,
		Name:  ident.Value,
	}
	exp.Arguments = p.parseExpressionList(lexer.TOKEN_RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{
		Token:  p.curToken,
		Object: left,
	}

	if !p.expectPeek(lexer.TOKEN_IDENT) {
		return nil
	}

	exp.Property = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	return exp
}

// ============================================================================
// Convenience functions
// ============================================================================

// Parse parses the input string and returns the AST.
func Parse(input string) (ast.Expression, error) {
	p := New(input)
	return p.Parse()
}
