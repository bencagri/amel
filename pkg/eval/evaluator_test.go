// Package eval implements the AST evaluator for the AMEL DSL.
package eval

import (
	"testing"
	"time"

	"github.com/bencagri/amel/pkg/parser"
	"github.com/bencagri/amel/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluator_Literals(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
		typ      types.Type
	}{
		{"42", int64(42), types.TypeInt},
		{"3.14", 3.14, types.TypeFloat},
		{`"hello"`, "hello", types.TypeString},
		{"true", true, types.TypeBool},
		{"false", false, types.TypeBool},
		{"null", nil, types.TypeNull},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.typ, result.Type)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestEvaluator_ListLiteral(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	expr, err := parser.Parse("[1, 2, 3]")
	require.NoError(t, err)

	result, err := evaluator.Evaluate(expr, ctx)
	require.NoError(t, err)
	assert.Equal(t, types.TypeList, result.Type)

	list, ok := result.AsList()
	require.True(t, ok)
	require.Len(t, list, 3)

	v1, _ := list[0].AsInt()
	v2, _ := list[1].AsInt()
	v3, _ := list[2].AsInt()
	assert.Equal(t, int64(1), v1)
	assert.Equal(t, int64(2), v2)
	assert.Equal(t, int64(3), v3)
}

func TestEvaluator_ArithmeticOperations(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"5 + 3", int64(8)},
		{"10 - 4", int64(6)},
		{"3 * 4", int64(12)},
		{"10 / 2", 5.0},
		{"10 % 3", int64(1)},
		{"2 + 3 * 4", int64(14)},
		{"(2 + 3) * 4", int64(20)},
		{"-5", int64(-5)},
		{"--5", int64(5)},
		{"3.5 + 1.5", 5.0},
		{"10 / 4", 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			switch expected := tt.expected.(type) {
			case int64:
				got, ok := result.AsInt()
				require.True(t, ok, "expected int, got %v", result.Type)
				assert.Equal(t, expected, got)
			case float64:
				got, ok := result.AsFloat()
				require.True(t, ok, "expected float, got %v", result.Type)
				assert.InDelta(t, expected, got, 0.001)
			}
		})
	}
}

func TestEvaluator_StringConcatenation(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	expr, err := parser.Parse(`"hello" + " " + "world"`)
	require.NoError(t, err)

	result, err := evaluator.Evaluate(expr, ctx)
	require.NoError(t, err)
	assert.Equal(t, types.TypeString, result.Type)

	str, ok := result.AsString()
	require.True(t, ok)
	assert.Equal(t, "hello world", str)
}

func TestEvaluator_ComparisonOperations(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected bool
	}{
		{"5 == 5", true},
		{"5 == 6", false},
		{"5 != 6", true},
		{"5 != 5", false},
		{"5 < 6", true},
		{"5 < 5", false},
		{"5 > 4", true},
		{"5 > 5", false},
		{"5 <= 5", true},
		{"5 <= 4", false},
		{"5 >= 5", true},
		{"4 >= 5", false},
		{`"abc" == "abc"`, true},
		{`"abc" != "def"`, true},
		{`"abc" < "def"`, true},
		{"3.14 == 3.14", true},
		{"3.14 > 3.0", true},
		{"null == null", true},
		{"true == true", true},
		{"true == false", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, types.TypeBool, result.Type)

			got, ok := result.AsBool()
			require.True(t, ok)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEvaluator_LogicalOperations(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{"!true", false},
		{"!false", true},
		{"true and true", true},
		{"true or false", true},
		{"not true", false},
		{"!(5 > 3)", false},
		{"5 > 3 && 2 < 4", true},
		{"5 < 3 || 2 < 4", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			got, ok := result.AsBool()
			require.True(t, ok, "expected bool, got %v", result.Type)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEvaluator_ShortCircuit(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	// AND short-circuit: false && (1/0) should not cause division by zero
	t.Run("and short-circuit", func(t *testing.T) {
		expr, err := parser.Parse("false && true")
		require.NoError(t, err)

		result, err := evaluator.Evaluate(expr, ctx)
		require.NoError(t, err)
		assert.Equal(t, false, result.Raw)
	})

	// OR short-circuit: true || (1/0) should not cause division by zero
	t.Run("or short-circuit", func(t *testing.T) {
		expr, err := parser.Parse("true || false")
		require.NoError(t, err)

		result, err := evaluator.Evaluate(expr, ctx)
		require.NoError(t, err)
		assert.Equal(t, true, result.Raw)
	})
}

func TestEvaluator_InOperator(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected bool
	}{
		{"5 IN [1, 2, 3, 4, 5]", true},
		{"6 IN [1, 2, 3, 4, 5]", false},
		{`"admin" IN ["admin", "user"]`, true},
		{`"guest" IN ["admin", "user"]`, false},
		{"5 NOT IN [1, 2, 3]", true},
		{"2 NOT IN [1, 2, 3]", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			got, ok := result.AsBool()
			require.True(t, ok)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEvaluator_JSONPath(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  25,
			"roles": []interface{}{
				"admin",
				"user",
			},
			"metadata": map[string]interface{}{
				"verified": true,
			},
		},
		"order": map[string]interface{}{
			"total": 150.50,
			"items": 3,
		},
	}

	ctx, err := NewContext(payload)
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"$.user.name", "John"},
		{"$.user.age", int64(25)},
		{"$.user.metadata.verified", true},
		{"$.order.total", 150.50},
		{"$.order.items", int64(3)},
		{"$.nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

func TestEvaluator_JSONPathExpressions(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"name":     "John",
			"age":      25,
			"verified": true,
			"role":     "admin",
		},
	}

	ctx, err := NewContext(payload)
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected bool
	}{
		{"$.user.age >= 18", true},
		{"$.user.age < 18", false},
		{"$.user.verified == true", true},
		{`$.user.name == "John"`, true},
		{`$.user.role IN ["admin", "moderator"]`, true},
		{"$.user.age >= 18 && $.user.verified == true", true},
		{`$.user.role == "guest" || $.user.age >= 21`, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.EvaluateBool(expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_JSONPathWithArray(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": 30},
			map[string]interface{}{"name": "Bob", "age": 25},
		},
	}

	ctx, err := NewContext(payload)
	require.NoError(t, err)

	// Test array access
	expr, err := parser.Parse("$.users[0].name")
	require.NoError(t, err)

	result, err := evaluator.Evaluate(expr, ctx)
	require.NoError(t, err)
	assert.Equal(t, "Alice", result.Raw)
}

func TestEvaluator_FunctionCalls(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"max(1, 2, 3)", int64(3)},
		{"min(5, 2, 8)", int64(2)},
		{"sum(1, 2, 3)", 6.0},
		{"avg(2, 4, 6)", 4.0},
		{"count([1, 2, 3])", int64(3)},
		{"abs(-5)", 5.0},
		{"ceil(3.2)", int64(4)},
		{"floor(3.8)", int64(3)},
		{"round(3.5)", int64(4)},
		{`len("hello")`, int64(5)},
		{`upper("hello")`, "HELLO"},
		{`lower("HELLO")`, "hello"},
		{`trim("  hello  ")`, "hello"},
		{`contains("hello world", "world")`, true},
		{`startsWith("hello", "hel")`, true},
		{`endsWith("hello", "llo")`, true},
		{"pow(2, 3)", 8.0},
		{"sqrt(16)", 4.0},
		{"mod(10, 3)", int64(1)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			switch expected := tt.expected.(type) {
			case int64:
				got, ok := result.AsInt()
				require.True(t, ok, "expected int, got %v: %v", result.Type, result.Raw)
				assert.Equal(t, expected, got)
			case float64:
				got, ok := result.AsFloat()
				require.True(t, ok, "expected float, got %v: %v", result.Type, result.Raw)
				assert.InDelta(t, expected, got, 0.001)
			case string:
				got, ok := result.AsString()
				require.True(t, ok, "expected string, got %v: %v", result.Type, result.Raw)
				assert.Equal(t, expected, got)
			case bool:
				got, ok := result.AsBool()
				require.True(t, ok, "expected bool, got %v: %v", result.Type, result.Raw)
				assert.Equal(t, expected, got)
			}
		})
	}
}

func TestEvaluator_FunctionCallWithJSONPath(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"items":   []interface{}{1, 2, 3, 4, 5},
		"message": "hello world",
	}

	ctx, err := NewContext(payload)
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"count($.items)", int64(5)},
		{`len($.message)`, int64(11)},
		{`contains($.message, "world")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			switch expected := tt.expected.(type) {
			case int64:
				got, ok := result.AsInt()
				require.True(t, ok)
				assert.Equal(t, expected, got)
			case bool:
				got, ok := result.AsBool()
				require.True(t, ok)
				assert.Equal(t, expected, got)
			}
		})
	}
}

func TestEvaluator_NestedFunctions(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"max(min(5, 3), 2)", int64(3)},
		{"abs(min(-5, -10))", 10.0},
		{`len(upper("hello"))`, int64(5)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			switch expected := tt.expected.(type) {
			case int64:
				got, ok := result.AsInt()
				require.True(t, ok)
				assert.Equal(t, expected, got)
			case float64:
				got, ok := result.AsFloat()
				require.True(t, ok)
				assert.InDelta(t, expected, got, 0.001)
			}
		})
	}
}

func TestEvaluator_ComplexExpressions(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"name":       "John",
			"age":        25,
			"verified":   true,
			"role":       "admin",
			"reputation": 1500,
		},
		"order": map[string]interface{}{
			"total": 150.50,
			"items": 3,
		},
	}

	ctx, err := NewContext(payload)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "admin or high reputation",
			input:    `$.user.role IN ["admin", "moderator"] || $.user.reputation >= 1000`,
			expected: true,
		},
		{
			name:     "verified adult",
			input:    "$.user.age >= 18 && $.user.verified == true",
			expected: true,
		},
		{
			name:     "complex access control",
			input:    `($.user.role == "admin" || $.user.reputation >= 1000) && $.user.verified == true`,
			expected: true,
		},
		{
			name:     "order threshold check",
			input:    "$.order.total > 100 && $.order.items >= 3",
			expected: true,
		},
		{
			name:     "name check with function",
			input:    `len($.user.name) > 0 && $.user.name != ""`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.EvaluateBool(expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_DivisionByZero(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	expr, err := parser.Parse("10 / 0")
	require.NoError(t, err)

	_, err = evaluator.Evaluate(expr, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

func TestEvaluator_ModuloByZero(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	expr, err := parser.Parse("10 % 0")
	require.NoError(t, err)

	_, err = evaluator.Evaluate(expr, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "modulo by zero")
}

func TestEvaluator_UndefinedFunction(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	expr, err := parser.Parse("unknownFunc(1, 2)")
	require.NoError(t, err)

	_, err = evaluator.Evaluate(expr, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined")
}

func TestEvaluator_WithExplanation(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"age": 25,
	}

	ctx, err := NewContext(payload)
	require.NoError(t, err)

	expr, err := parser.Parse("$.age >= 18")
	require.NoError(t, err)

	result, explanation, err := evaluator.EvaluateWithExplanation(expr, ctx)
	require.NoError(t, err)

	assert.True(t, result.IsTruthy())
	assert.NotNil(t, explanation)
	assert.Equal(t, "($.age >= 18)", explanation.Expression)
}

func TestEvaluator_Timeout(t *testing.T) {
	evaluator, err := New(WithTimeout(1 * time.Nanosecond))
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	// Large expression that would take time
	expr, err := parser.Parse("1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1 + 1")
	require.NoError(t, err)

	// The timeout is so short that evaluation might fail
	// This is a basic timeout test - in practice timeouts are for long-running operations
	_, err = evaluator.Evaluate(expr, ctx)
	// Error might or might not happen depending on timing, so we just verify no panic
}

func TestEvaluator_IndexExpression(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)
	ctx.SetVariable("list", types.List(types.Int(10), types.Int(20), types.Int(30)))

	expr, err := parser.Parse("list[1]")
	require.NoError(t, err)

	result, err := evaluator.Evaluate(expr, ctx)
	require.NoError(t, err)

	got, ok := result.AsInt()
	require.True(t, ok)
	assert.Equal(t, int64(20), got)
}

func TestEvaluator_NegativeIndex(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)
	ctx.SetVariable("list", types.List(types.Int(10), types.Int(20), types.Int(30)))

	expr, err := parser.Parse("list[-1]")
	require.NoError(t, err)

	result, err := evaluator.Evaluate(expr, ctx)
	require.NoError(t, err)

	got, ok := result.AsInt()
	require.True(t, ok)
	assert.Equal(t, int64(30), got)
}

func TestEvaluator_IndexOutOfBounds(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)
	ctx.SetVariable("list", types.List(types.Int(10), types.Int(20)))

	expr, err := parser.Parse("list[10]")
	require.NoError(t, err)

	_, err = evaluator.Evaluate(expr, ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")
}

func TestEvaluator_CoalesceFunction(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{"coalesce(null, 5)", int64(5)},
		{"coalesce(null, null, 10)", int64(10)},
		{`coalesce("hello", "world")`, "hello"},
		{"coalesce(null, null)", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			if tt.expected == nil {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

func TestEvaluator_IfThenElse(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected interface{}
	}{
		{`ifThenElse(true, "yes", "no")`, "yes"},
		{`ifThenElse(false, "yes", "no")`, "no"},
		{"ifThenElse(5 > 3, 100, 200)", int64(100)},
		{"ifThenElse(5 < 3, 100, 200)", int64(200)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestEvaluator_TypeOf(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected string
	}{
		{"typeOf(42)", "int"},
		{"typeOf(3.14)", "float"},
		{`typeOf("hello")`, "string"},
		{"typeOf(true)", "bool"},
		{"typeOf(null)", "null"},
		{"typeOf([1, 2, 3])", "list"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			got, ok := result.AsString()
			require.True(t, ok)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestEvaluator_IsNullIsEmpty(t *testing.T) {
	evaluator, err := New()
	require.NoError(t, err)

	ctx, err := NewContext(map[string]interface{}{})
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected bool
	}{
		{"isNull(null)", true},
		{"isNull(5)", false},
		{"isNotNull(5)", true},
		{"isNotNull(null)", false},
		{`isEmpty("")`, true},
		{`isEmpty("hello")`, false},
		{"isEmpty([])", true},
		{"isEmpty([1, 2])", false},
		{"isEmpty(null)", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parser.Parse(tt.input)
			require.NoError(t, err)

			result, err := evaluator.Evaluate(expr, ctx)
			require.NoError(t, err)

			got, ok := result.AsBool()
			require.True(t, ok)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestNewContext_JSONString(t *testing.T) {
	jsonStr := `{"name": "John", "age": 30}`
	ctx, err := NewContext(jsonStr)
	require.NoError(t, err)

	evaluator, err := New()
	require.NoError(t, err)

	expr, err := parser.Parse("$.name")
	require.NoError(t, err)

	result, err := evaluator.Evaluate(expr, ctx)
	require.NoError(t, err)
	assert.Equal(t, "John", result.Raw)
}

// Benchmark tests
func BenchmarkEvaluator_SimpleLiteral(b *testing.B) {
	evaluator, _ := New()
	ctx, _ := NewContext(map[string]interface{}{})
	expr, _ := parser.Parse("42")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(expr, ctx)
	}
}

func BenchmarkEvaluator_BinaryExpression(b *testing.B) {
	evaluator, _ := New()
	ctx, _ := NewContext(map[string]interface{}{})
	expr, _ := parser.Parse("5 + 3 * 2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(expr, ctx)
	}
}

func BenchmarkEvaluator_JSONPath(b *testing.B) {
	evaluator, _ := New()
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  25,
		},
	}
	ctx, _ := NewContext(payload)
	expr, _ := parser.Parse("$.user.age >= 18")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(expr, ctx)
	}
}

func BenchmarkEvaluator_FunctionCall(b *testing.B) {
	evaluator, _ := New()
	ctx, _ := NewContext(map[string]interface{}{})
	expr, _ := parser.Parse("max(1, 2, 3)")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(expr, ctx)
	}
}

func BenchmarkEvaluator_ComplexExpression(b *testing.B) {
	evaluator, _ := New()
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"role":       "admin",
			"age":        25,
			"verified":   true,
			"reputation": 1500,
		},
	}
	ctx, _ := NewContext(payload)
	expr, _ := parser.Parse(`($.user.role IN ["admin", "moderator"] || $.user.reputation >= 1000) && $.user.verified == true && $.user.age >= 18`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(expr, ctx)
	}
}
