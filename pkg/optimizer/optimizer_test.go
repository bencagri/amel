// Package optimizer provides AST optimization for the AMEL DSL.
package optimizer

import (
	"testing"

	"github.com/bencagri/amel/pkg/ast"
	"github.com/bencagri/amel/pkg/lexer"
	"github.com/bencagri/amel/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstantFolding(t *testing.T) {
	opt := New()

	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		// Integer arithmetic
		{"add integers", "2 + 3", int64(5)},
		{"subtract integers", "10 - 4", int64(6)},
		{"multiply integers", "3 * 4", int64(12)},
		{"divide integers", "20 / 4", int64(5)},
		{"modulo integers", "17 % 5", int64(2)},

		// Float arithmetic
		{"add floats", "2.5 + 3.5", float64(6.0)},
		{"subtract floats", "10.5 - 4.5", float64(6.0)},
		{"multiply floats", "2.5 * 4.0", float64(10.0)},
		{"divide floats", "10.0 / 4.0", float64(2.5)},

		// Mixed numeric
		{"int + float", "2 + 3.5", float64(5.5)},
		{"float - int", "10.5 - 4", float64(6.5)},

		// String concatenation
		{"concat strings", `"hello" + " " + "world"`, "hello world"},

		// Comparison operators
		{"less than true", "3 < 5", true},
		{"less than false", "5 < 3", false},
		{"greater than true", "5 > 3", true},
		{"greater than false", "3 > 5", false},
		{"less than or equal", "3 <= 3", true},
		{"greater than or equal", "5 >= 5", true},
		{"equal integers", "5 == 5", true},
		{"not equal integers", "5 != 3", true},
		{"equal strings", `"foo" == "foo"`, true},

		// Logical operators
		{"and true", "true && true", true},
		{"and false", "true && false", false},
		{"or true", "false || true", true},
		{"or false", "false || false", false},

		// Unary operators
		{"negate integer", "-5", int64(-5)},
		{"negate float", "-3.14", float64(-3.14)},
		{"not true", "!true", false},
		{"not false", "!false", true},

		// Complex expressions
		{"nested arithmetic", "(2 + 3) * 4", int64(20)},
		{"nested comparison", "(5 > 3) && (2 < 4)", true},
		{"triple add", "1 + 2 + 3", int64(6)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			optimized := opt.Optimize(expr)
			val := getLiteralValue(optimized)

			assert.Equal(t, tt.expected, val, "input: %s", tt.input)
		})
	}
}

func TestConstantFoldingDivisionByZero(t *testing.T) {
	opt := New()

	// Division by zero should not be folded
	expr, err := parser.Parse("10 / 0")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	// Should still be a binary expression (not folded)
	_, isBinary := optimized.(*ast.BinaryExpression)
	assert.True(t, isBinary, "division by zero should not be folded")
}

func TestConstantFoldingPreservesNonConstants(t *testing.T) {
	opt := New()

	tests := []struct {
		name  string
		input string
	}{
		{"identifier", "x + 2"},
		{"jsonpath", "$.value + 5"},
		{"function call", "max(1, 2) + 3"},
		{"mixed", "1 + x + 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			optimized := opt.Optimize(expr)

			// Result should not be a simple literal
			assert.False(t, isLiteral(optimized), "non-constant expression should not be fully folded")
		})
	}
}

func TestConstantFoldingPartialOptimization(t *testing.T) {
	opt := New()

	// "x + (2 + 3)" should become "x + 5"
	expr, err := parser.Parse("x + (2 + 3)")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	binary, ok := optimized.(*ast.BinaryExpression)
	require.True(t, ok)

	// Left should still be identifier
	_, isIdent := binary.Left.(*ast.Identifier)
	assert.True(t, isIdent)

	// Right should be folded to 5
	intLit, isInt := binary.Right.(*ast.IntegerLiteral)
	require.True(t, isInt)
	assert.Equal(t, int64(5), intLit.Value)
}

func TestConstantFoldingList(t *testing.T) {
	opt := New()

	// List elements should be folded
	expr, err := parser.Parse("[1 + 1, 2 + 2, 3 + 3]")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	list, ok := optimized.(*ast.ListLiteral)
	require.True(t, ok)
	require.Len(t, list.Elements, 3)

	// Each element should be folded
	for i, elem := range list.Elements {
		intLit, ok := elem.(*ast.IntegerLiteral)
		require.True(t, ok, "element %d should be integer literal", i)
		assert.Equal(t, int64((i+1)*2), intLit.Value)
	}
}

func TestConstantFoldingIndexExpression(t *testing.T) {
	opt := New()

	// Constant list indexing should be folded
	expr, err := parser.Parse("[10, 20, 30][1]")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	intLit, ok := optimized.(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(20), intLit.Value)
}

func TestConstantFoldingIndexExpressionOutOfBounds(t *testing.T) {
	opt := New()

	// Out of bounds index should not be folded
	expr, err := parser.Parse("[1, 2, 3][10]")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	// Should still be an index expression
	_, isIndex := optimized.(*ast.IndexExpression)
	assert.True(t, isIndex, "out of bounds index should not be folded")
}

func TestConstantFoldingInExpression(t *testing.T) {
	opt := New()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"value in list - found", "2 IN [1, 2, 3]", true},
		{"value in list - not found", "5 IN [1, 2, 3]", false},
		{"value not in list - found", "2 NOT IN [1, 2, 3]", false},
		{"value not in list - not found", "5 NOT IN [1, 2, 3]", true},
		{"string in list", `"b" IN ["a", "b", "c"]`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			optimized := opt.Optimize(expr)

			boolLit, ok := optimized.(*ast.BooleanLiteral)
			require.True(t, ok)
			assert.Equal(t, tt.expected, boolLit.Value)
		})
	}
}

func TestConstantFoldingGroupedExpression(t *testing.T) {
	opt := New()

	// Grouped constant should be unwrapped
	expr, err := parser.Parse("(42)")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	intLit, ok := optimized.(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(42), intLit.Value)
}

func TestConstantFoldingFunctionArgs(t *testing.T) {
	opt := New()

	// Function arguments should be folded, but function call itself preserved
	expr, err := parser.Parse("max(1 + 1, 2 + 2)")
	require.NoError(t, err)

	optimized := opt.Optimize(expr)

	funcCall, ok := optimized.(*ast.FunctionCall)
	require.True(t, ok)
	require.Len(t, funcCall.Arguments, 2)

	// First arg should be 2
	arg1, ok := funcCall.Arguments[0].(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(2), arg1.Value)

	// Second arg should be 4
	arg2, ok := funcCall.Arguments[1].(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(4), arg2.Value)
}

func TestOptimizeWithStats(t *testing.T) {
	opt := New()

	// Expression: (2 + 3) * (4 - 1) should fold multiple constants
	expr, err := parser.Parse("(2 + 3) * (4 - 1)")
	require.NoError(t, err)

	optimized, stats := opt.OptimizeWithStats(expr)

	// Should be folded to a single integer
	intLit, ok := optimized.(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(15), intLit.Value)

	// Stats should reflect optimizations
	assert.Greater(t, stats.ConstantsFolded, 0)
	assert.Greater(t, stats.ExpressionsTotal, 0)
}

func TestIsConstant(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"integer literal", "42", true},
		{"float literal", "3.14", true},
		{"string literal", `"hello"`, true},
		{"boolean literal", "true", true},
		{"null literal", "null", true},
		{"constant list", "[1, 2, 3]", true},
		{"constant arithmetic", "2 + 3", true},
		{"constant comparison", "5 > 3", true},
		{"constant unary", "-5", true},
		{"constant grouped", "(2 + 3)", true},
		{"constant in", "1 IN [1, 2, 3]", true},
		{"identifier", "x", false},
		{"jsonpath", "$.foo", false},
		{"function call", "max(1, 2)", false},
		{"mixed", "x + 2", false},
		{"list with non-constant", "[1, x, 3]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result := IsConstant(expr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluateConstant(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer", "42", int64(42)},
		{"arithmetic", "2 + 3 * 4", int64(14)},
		{"string concat", `"a" + "b"`, "ab"},
		{"boolean", "true && false", false},
		{"null", "null", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result := EvaluateConstant(expr)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestEvaluateConstantNonConstant(t *testing.T) {
	expr, err := parser.Parse("x + 2")
	require.NoError(t, err)

	result := EvaluateConstant(expr)
	assert.Nil(t, result, "non-constant expression should return nil")
}

func TestOptimizerOptions(t *testing.T) {
	t.Run("constant folding disabled", func(t *testing.T) {
		opt := New(WithConstantFolding(false))

		expr, err := parser.Parse("2 + 3")
		require.NoError(t, err)

		optimized := opt.Optimize(expr)

		// Should still be binary expression (not folded)
		_, isBinary := optimized.(*ast.BinaryExpression)
		assert.True(t, isBinary)
	})

	t.Run("constant folding enabled", func(t *testing.T) {
		opt := New(WithConstantFolding(true))

		expr, err := parser.Parse("2 + 3")
		require.NoError(t, err)

		optimized := opt.Optimize(expr)

		// Should be folded
		intLit, ok := optimized.(*ast.IntegerLiteral)
		require.True(t, ok)
		assert.Equal(t, int64(5), intLit.Value)
	})
}

func TestOptimizeForCache(t *testing.T) {
	expr, err := parser.Parse("(1 + 2) * (3 + 4)")
	require.NoError(t, err)

	optimized := OptimizeForCache(expr)

	intLit, ok := optimized.(*ast.IntegerLiteral)
	require.True(t, ok)
	assert.Equal(t, int64(21), intLit.Value)
}

func TestValueToLiteral(t *testing.T) {
	token := lexer.Token{Type: lexer.TOKEN_INT, Literal: "0"}

	tests := []struct {
		name  string
		val   interface{}
		check func(t *testing.T, expr ast.Expression)
	}{
		{"int64", int64(42), func(t *testing.T, expr ast.Expression) {
			lit, ok := expr.(*ast.IntegerLiteral)
			require.True(t, ok)
			assert.Equal(t, int64(42), lit.Value)
		}},
		{"float64", float64(3.14), func(t *testing.T, expr ast.Expression) {
			lit, ok := expr.(*ast.FloatLiteral)
			require.True(t, ok)
			assert.Equal(t, 3.14, lit.Value)
		}},
		{"string", "hello", func(t *testing.T, expr ast.Expression) {
			lit, ok := expr.(*ast.StringLiteral)
			require.True(t, ok)
			assert.Equal(t, "hello", lit.Value)
		}},
		{"bool true", true, func(t *testing.T, expr ast.Expression) {
			lit, ok := expr.(*ast.BooleanLiteral)
			require.True(t, ok)
			assert.True(t, lit.Value)
		}},
		{"bool false", false, func(t *testing.T, expr ast.Expression) {
			lit, ok := expr.(*ast.BooleanLiteral)
			require.True(t, ok)
			assert.False(t, lit.Value)
		}},
		{"nil", nil, func(t *testing.T, expr ast.Expression) {
			_, ok := expr.(*ast.NullLiteral)
			require.True(t, ok)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueToLiteral(tt.val, token)
			require.NotNil(t, result)
			tt.check(t, result)
		})
	}
}
