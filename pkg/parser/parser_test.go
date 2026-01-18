// Package parser implements a recursive descent parser for the AMEL DSL.
package parser

import (
	"testing"

	"github.com/bencagri/amel/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"42", 42},
		{"0", 0},
		{"999999", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			lit, ok := expr.(*ast.IntegerLiteral)
			require.True(t, ok, "expected IntegerLiteral, got %T", expr)
			assert.Equal(t, tt.expected, lit.Value)
		})
	}
}

func TestParseFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"0.5", 0.5},
		{"123.456", 123.456},
		{"1e10", 1e10},
		{"1.5e-3", 1.5e-3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			lit, ok := expr.(*ast.FloatLiteral)
			require.True(t, ok, "expected FloatLiteral, got %T", expr)
			assert.InDelta(t, tt.expected, lit.Value, 0.0001)
		})
	}
}

func TestParseStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`'single quotes'`, "single quotes"},
		{`""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			lit, ok := expr.(*ast.StringLiteral)
			require.True(t, ok, "expected StringLiteral, got %T", expr)
			assert.Equal(t, tt.expected, lit.Value)
		})
	}
}

func TestParseBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			lit, ok := expr.(*ast.BooleanLiteral)
			require.True(t, ok, "expected BooleanLiteral, got %T", expr)
			assert.Equal(t, tt.expected, lit.Value)
		})
	}
}

func TestParseNullLiteral(t *testing.T) {
	tests := []string{"null", "nil"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			expr, err := Parse(input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			_, ok := expr.(*ast.NullLiteral)
			require.True(t, ok, "expected NullLiteral, got %T", expr)
		})
	}
}

func TestParseListLiteral(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int
	}{
		{"[]", 0},
		{"[1]", 1},
		{"[1, 2, 3]", 3},
		{`["a", "b"]`, 2},
		{"[1, 2.5, true, null]", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			list, ok := expr.(*ast.ListLiteral)
			require.True(t, ok, "expected ListLiteral, got %T", expr)
			assert.Len(t, list.Elements, tt.expectedCount)
		})
	}
}

func TestParsePrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"-5", "-"},
		{"!true", "!"},
		{"not false", "not"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			unary, ok := expr.(*ast.UnaryExpression)
			require.True(t, ok, "expected UnaryExpression, got %T", expr)
			assert.Equal(t, tt.operator, unary.Operator)
		})
	}
}

func TestParseInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5", int64(5), "+", int64(5)},
		{"5 - 5", int64(5), "-", int64(5)},
		{"5 * 5", int64(5), "*", int64(5)},
		{"5 / 5", int64(5), "/", int64(5)},
		{"5 % 3", int64(5), "%", int64(3)},
		{"5 > 5", int64(5), ">", int64(5)},
		{"5 < 5", int64(5), "<", int64(5)},
		{"5 == 5", int64(5), "==", int64(5)},
		{"5 != 5", int64(5), "!=", int64(5)},
		{"5 >= 5", int64(5), ">=", int64(5)},
		{"5 <= 5", int64(5), "<=", int64(5)},
		{"true && false", true, "&&", false},
		{"true || false", true, "||", false},
		{"true and false", true, "and", false},
		{"true or false", true, "or", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			binary, ok := expr.(*ast.BinaryExpression)
			require.True(t, ok, "expected BinaryExpression, got %T", expr)
			assert.Equal(t, tt.operator, binary.Operator)

			testLiteralExpression(t, binary.Left, tt.left)
			testLiteralExpression(t, binary.Right, tt.right)
		})
	}
}

func TestParseInExpression(t *testing.T) {
	tests := []struct {
		input   string
		negated bool
	}{
		{`"admin" IN ["admin", "user"]`, false},
		{`5 IN [1, 2, 3, 4, 5]`, false},
		{`"guest" NOT IN ["admin", "user"]`, true},
		{`x NOT IN [1, 2, 3]`, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			inExpr, ok := expr.(*ast.InExpression)
			require.True(t, ok, "expected InExpression, got %T", expr)
			assert.Equal(t, tt.negated, inExpr.Negated)
			require.NotNil(t, inExpr.Left)
			require.NotNil(t, inExpr.Right)
		})
	}
}

func TestParseGroupedExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(5)", "5"},
		{"(5 + 5)", "(5 + 5)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 * (5 + 5)", "(2 * (5 + 5))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func TestParseOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2 + 3", "((1 + 2) + 3)"},
		{"1 + 2 * 3", "(1 + (2 * 3))"},
		{"1 * 2 + 3", "((1 * 2) + 3)"},
		{"1 * 2 * 3", "((1 * 2) * 3)"},
		{"-1 * 2", "((-1) * 2)"},
		{"!true == false", "((!true) == false)"},
		{"1 + 2 == 3", "((1 + 2) == 3)"},
		{"a && b || c", "((a && b) || c)"},
		{"a || b && c", "(a || (b && c))"},
		{"1 > 2 == false", "((1 > 2) == false)"},
		{"1 < 2 && 3 > 4", "((1 < 2) && (3 > 4))"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		name     string
		argCount int
	}{
		{"max(1, 2)", "max", 2},
		{"min(1, 2, 3)", "min", 3},
		{"count()", "count", 0},
		{"sum(a, b)", "sum", 2},
		{"len(list)", "len", 1},
		{`contains("hello", "ell")`, "contains", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			call, ok := expr.(*ast.FunctionCall)
			require.True(t, ok, "expected FunctionCall, got %T", expr)
			assert.Equal(t, tt.name, call.Name)
			assert.Len(t, call.Arguments, tt.argCount)
		})
	}
}

func TestParseNestedFunctionCalls(t *testing.T) {
	input := "max(min(1, 2), 3)"
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	outer, ok := expr.(*ast.FunctionCall)
	require.True(t, ok, "expected FunctionCall, got %T", expr)
	assert.Equal(t, "max", outer.Name)
	require.Len(t, outer.Arguments, 2)

	inner, ok := outer.Arguments[0].(*ast.FunctionCall)
	require.True(t, ok, "expected nested FunctionCall, got %T", outer.Arguments[0])
	assert.Equal(t, "min", inner.Name)
}

func TestParseJSONPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$", "$"},
		{"$.name", "$.name"},
		{"$.user.name", "$.user.name"},
		{"$.users[0]", "$.users[0]"},
		{"$.users[0].name", "$.users[0].name"},
		{`$.data["key"]`, `$.data["key"]`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)

			jp, ok := expr.(*ast.JSONPathExpression)
			require.True(t, ok, "expected JSONPathExpression, got %T", expr)
			assert.Equal(t, tt.expected, jp.Path)
		})
	}
}

func TestParseJSONPathInExpression(t *testing.T) {
	input := "$.user.age >= 18"
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	binary, ok := expr.(*ast.BinaryExpression)
	require.True(t, ok, "expected BinaryExpression, got %T", expr)
	assert.Equal(t, ">=", binary.Operator)

	jp, ok := binary.Left.(*ast.JSONPathExpression)
	require.True(t, ok, "expected JSONPathExpression, got %T", binary.Left)
	assert.Equal(t, "$.user.age", jp.Path)

	lit, ok := binary.Right.(*ast.IntegerLiteral)
	require.True(t, ok, "expected IntegerLiteral, got %T", binary.Right)
	assert.Equal(t, int64(18), lit.Value)
}

func TestParseComplexExpression(t *testing.T) {
	input := `$.user.age >= 18 && $.user.verified == true`
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	and, ok := expr.(*ast.BinaryExpression)
	require.True(t, ok, "expected BinaryExpression, got %T", expr)
	assert.Equal(t, "&&", and.Operator)

	// Left: $.user.age >= 18
	left, ok := and.Left.(*ast.BinaryExpression)
	require.True(t, ok)
	assert.Equal(t, ">=", left.Operator)

	// Right: $.user.verified == true
	right, ok := and.Right.(*ast.BinaryExpression)
	require.True(t, ok)
	assert.Equal(t, "==", right.Operator)
}

func TestParseRoleCheckExpression(t *testing.T) {
	input := `$.user.role IN ["admin", "moderator"] || $.user.reputation >= 1000`
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	or, ok := expr.(*ast.BinaryExpression)
	require.True(t, ok, "expected BinaryExpression, got %T", expr)
	assert.Equal(t, "||", or.Operator)

	// Left should be IN expression
	inExpr, ok := or.Left.(*ast.InExpression)
	require.True(t, ok, "expected InExpression, got %T", or.Left)
	assert.False(t, inExpr.Negated)

	// Right should be comparison
	cmp, ok := or.Right.(*ast.BinaryExpression)
	require.True(t, ok, "expected BinaryExpression, got %T", or.Right)
	assert.Equal(t, ">=", cmp.Operator)
}

func TestParseFunctionCallWithJSONPath(t *testing.T) {
	input := `len($.items) > 0`
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	cmp, ok := expr.(*ast.BinaryExpression)
	require.True(t, ok, "expected BinaryExpression, got %T", expr)
	assert.Equal(t, ">", cmp.Operator)

	call, ok := cmp.Left.(*ast.FunctionCall)
	require.True(t, ok, "expected FunctionCall, got %T", cmp.Left)
	assert.Equal(t, "len", call.Name)
	require.Len(t, call.Arguments, 1)

	jp, ok := call.Arguments[0].(*ast.JSONPathExpression)
	require.True(t, ok, "expected JSONPathExpression, got %T", call.Arguments[0])
	assert.Equal(t, "$.items", jp.Path)
}

func TestParseIndexExpression(t *testing.T) {
	input := "list[0]"
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	idx, ok := expr.(*ast.IndexExpression)
	require.True(t, ok, "expected IndexExpression, got %T", expr)

	ident, ok := idx.Left.(*ast.Identifier)
	require.True(t, ok)
	assert.Equal(t, "list", ident.Value)

	index, ok := idx.Index.(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(0), index.Value)
}

func TestParseMemberExpression(t *testing.T) {
	input := "obj.property"
	expr, err := Parse(input)
	require.NoError(t, err)
	require.NotNil(t, expr)

	member, ok := expr.(*ast.MemberExpression)
	require.True(t, ok, "expected MemberExpression, got %T", expr)

	obj, ok := member.Object.(*ast.Identifier)
	require.True(t, ok)
	assert.Equal(t, "obj", obj.Value)
	assert.Equal(t, "property", member.Property.Value)
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
	}{
		{"(5", true},        // Missing closing paren
		{"5 +", true},       // Missing operand
		{"[1, 2,", true},    // Incomplete list
		{"func(", true},     // Incomplete function call
		{"", true},          // Empty input
		{"@ invalid", true}, // Invalid character
		{"5 5", true},       // Two expressions without operator
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(tt.input)
			_, err := p.Parse()
			if tt.expectError {
				assert.Error(t, err, "expected error for input: %s", tt.input)
			}
		})
	}
}

func TestParseASTString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "5"},
		{"-5", "(-5)"},
		{"!true", "(!true)"},
		{"5 + 5", "(5 + 5)"},
		{"max(1, 2)", "max(1, 2)"},
		{"[1, 2, 3]", "[1, 2, 3]"},
		{`"hello"`, `"hello"`},
		{"$.user.name", "$.user.name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := Parse(tt.input)
			require.NoError(t, err)
			require.NotNil(t, expr)
			assert.Equal(t, tt.expected, expr.String())
		})
	}
}

// Helper function to test literal expressions
func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) {
	t.Helper()
	switch v := expected.(type) {
	case int:
		testIntegerLiteral(t, exp, int64(v))
	case int64:
		testIntegerLiteral(t, exp, v)
	case float64:
		testFloatLiteral(t, exp, v)
	case string:
		testStringLiteral(t, exp, v)
	case bool:
		testBooleanLiteral(t, exp, v)
	}
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) {
	t.Helper()
	lit, ok := exp.(*ast.IntegerLiteral)
	require.True(t, ok, "expected IntegerLiteral, got %T", exp)
	assert.Equal(t, value, lit.Value)
}

func testFloatLiteral(t *testing.T, exp ast.Expression, value float64) {
	t.Helper()
	lit, ok := exp.(*ast.FloatLiteral)
	require.True(t, ok, "expected FloatLiteral, got %T", exp)
	assert.InDelta(t, value, lit.Value, 0.0001)
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) {
	t.Helper()
	lit, ok := exp.(*ast.StringLiteral)
	require.True(t, ok, "expected StringLiteral, got %T", exp)
	assert.Equal(t, value, lit.Value)
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) {
	t.Helper()
	lit, ok := exp.(*ast.BooleanLiteral)
	require.True(t, ok, "expected BooleanLiteral, got %T", exp)
	assert.Equal(t, value, lit.Value)
}
