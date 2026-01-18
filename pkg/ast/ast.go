// Package ast defines the Abstract Syntax Tree nodes for the AMEL DSL.
package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bencagri/amel/pkg/lexer"
)

// Node represents a node in the AST.
type Node interface {
	// TokenLiteral returns the literal value of the token associated with this node.
	TokenLiteral() string
	// String returns a string representation of the node for debugging.
	String() string
}

// Expression represents an expression node that produces a value.
type Expression interface {
	Node
	expressionNode()
}

// ============================================================================
// Literal Expressions
// ============================================================================

// IntegerLiteral represents an integer literal (e.g., 42).
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a floating-point literal (e.g., 3.14).
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal (e.g., "hello").
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("%q", sl.Value) }

// BooleanLiteral represents a boolean literal (true or false).
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NullLiteral represents a null literal.
type NullLiteral struct {
	Token lexer.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

// ListLiteral represents a list literal (e.g., [1, 2, 3]).
type ListLiteral struct {
	Token    lexer.Token // The '[' token
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) String() string {
	var out bytes.Buffer
	elements := make([]string, len(ll.Elements))
	for i, el := range ll.Elements {
		elements[i] = el.String()
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// ============================================================================
// Identifier and Path Expressions
// ============================================================================

// Identifier represents an identifier (e.g., variable name, function name).
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// JSONPathExpression represents a JSONPath expression (e.g., $.user.name).
type JSONPathExpression struct {
	Token lexer.Token // The '$' token
	Path  string      // The full path including $
}

func (jp *JSONPathExpression) expressionNode()      {}
func (jp *JSONPathExpression) TokenLiteral() string { return jp.Token.Literal }
func (jp *JSONPathExpression) String() string       { return jp.Path }

// ============================================================================
// Operator Expressions
// ============================================================================

// BinaryExpression represents a binary operation (e.g., a + b, x && y).
type BinaryExpression struct {
	Token    lexer.Token // The operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }
func (be *BinaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" ")
	out.WriteString(be.Operator)
	out.WriteString(" ")
	out.WriteString(be.Right.String())
	out.WriteString(")")
	return out.String()
}

// UnaryExpression represents a unary operation (e.g., !x, -5).
type UnaryExpression struct {
	Token    lexer.Token // The operator token (!, -)
	Operator string
	Operand  Expression
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }
func (ue *UnaryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ue.Operator)
	out.WriteString(ue.Operand.String())
	out.WriteString(")")
	return out.String()
}

// ============================================================================
// Function Call Expression
// ============================================================================

// FunctionCall represents a function call (e.g., max(a, b)).
type FunctionCall struct {
	Token     lexer.Token // The function name token
	Name      string
	Arguments []Expression
}

func (fc *FunctionCall) expressionNode()      {}
func (fc *FunctionCall) TokenLiteral() string { return fc.Token.Literal }
func (fc *FunctionCall) String() string {
	var out bytes.Buffer
	args := make([]string, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		args[i] = arg.String()
	}
	out.WriteString(fc.Name)
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// ============================================================================
// Index Expression
// ============================================================================

// IndexExpression represents an index access (e.g., list[0]).
type IndexExpression struct {
	Token lexer.Token // The '[' token
	Left  Expression  // The expression being indexed
	Index Expression  // The index expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// ============================================================================
// Member Access Expression
// ============================================================================

// MemberExpression represents member access (e.g., obj.property).
type MemberExpression struct {
	Token    lexer.Token // The '.' token
	Object   Expression  // The object being accessed
	Property *Identifier // The property name
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(me.Object.String())
	out.WriteString(".")
	out.WriteString(me.Property.String())
	out.WriteString(")")
	return out.String()
}

// ============================================================================
// Conditional Expression (Ternary)
// ============================================================================

// ConditionalExpression represents a ternary conditional (condition ? then : else).
// Note: This is for future use if we add ternary operator support.
type ConditionalExpression struct {
	Token       lexer.Token // The '?' token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (ce *ConditionalExpression) expressionNode()      {}
func (ce *ConditionalExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *ConditionalExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ce.Condition.String())
	out.WriteString(" ? ")
	out.WriteString(ce.Consequence.String())
	out.WriteString(" : ")
	out.WriteString(ce.Alternative.String())
	out.WriteString(")")
	return out.String()
}

// ============================================================================
// Grouped Expression
// ============================================================================

// GroupedExpression represents a parenthesized expression (e.g., (a + b)).
type GroupedExpression struct {
	Token      lexer.Token // The '(' token
	Expression Expression
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupedExpression) String() string {
	return ge.Expression.String()
}

// ============================================================================
// IN Expression
// ============================================================================

// InExpression represents an IN membership test (e.g., x IN [1, 2, 3]).
type InExpression struct {
	Token   lexer.Token // The 'IN' or 'NOT IN' token
	Left    Expression  // The value to check
	Right   Expression  // The list/collection to check against
	Negated bool        // true for NOT IN
}

func (ie *InExpression) expressionNode()      {}
func (ie *InExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	if ie.Negated {
		out.WriteString(" NOT IN ")
	} else {
		out.WriteString(" IN ")
	}
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// ============================================================================
// Regex Expression
// ============================================================================

// RegexExpression represents a regex match operation (e.g., name =~ "^John").
type RegexExpression struct {
	Token   lexer.Token // The '=~' or '!~' token
	Left    Expression  // The string to match against
	Pattern Expression  // The regex pattern (usually a string literal)
	Negated bool        // true for !~ (not match)
}

func (re *RegexExpression) expressionNode()      {}
func (re *RegexExpression) TokenLiteral() string { return re.Token.Literal }
func (re *RegexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(re.Left.String())
	if re.Negated {
		out.WriteString(" !~ ")
	} else {
		out.WriteString(" =~ ")
	}
	out.WriteString(re.Pattern.String())
	out.WriteString(")")
	return out.String()
}

// ============================================================================
// Lambda Expression (for map, filter, reduce)
// ============================================================================

// LambdaExpression represents an inline function (e.g., x => x * 2).
type LambdaExpression struct {
	Token      lexer.Token   // The '=>' token
	Parameters []*Identifier // Parameter names
	Body       Expression    // The expression body
}

func (le *LambdaExpression) expressionNode()      {}
func (le *LambdaExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LambdaExpression) String() string {
	var out bytes.Buffer
	if len(le.Parameters) == 1 {
		out.WriteString(le.Parameters[0].String())
	} else {
		out.WriteString("(")
		for i, p := range le.Parameters {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(p.String())
		}
		out.WriteString(")")
	}
	out.WriteString(" => ")
	out.WriteString(le.Body.String())
	return out.String()
}
