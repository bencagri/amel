// Package functions provides function management for the AMEL DSL engine.
package functions

import (
	"math"
	"testing"

	"github.com/bencagri/amel/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Tests for Additional List Functions
// ============================================================================

func TestBuiltinIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		list     []types.Value
		value    types.Value
		expected int64
	}{
		{"found at start", []types.Value{types.Int(1), types.Int(2), types.Int(3)}, types.Int(1), 0},
		{"found in middle", []types.Value{types.Int(1), types.Int(2), types.Int(3)}, types.Int(2), 1},
		{"found at end", []types.Value{types.Int(1), types.Int(2), types.Int(3)}, types.Int(3), 2},
		{"not found", []types.Value{types.Int(1), types.Int(2), types.Int(3)}, types.Int(5), -1},
		{"string found", []types.Value{types.String("a"), types.String("b"), types.String("c")}, types.String("b"), 1},
		{"empty list", []types.Value{}, types.Int(1), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinIndexOf(types.List(tt.list...), tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinSortAsc(t *testing.T) {
	t.Run("sort integers ascending", func(t *testing.T) {
		list := types.List(types.Int(3), types.Int(1), types.Int(4), types.Int(1), types.Int(5))
		result, err := builtinSortAsc(list)
		require.NoError(t, err)

		sorted, ok := result.AsList()
		require.True(t, ok)
		require.Len(t, sorted, 5)
		assert.Equal(t, int64(1), sorted[0].Raw)
		assert.Equal(t, int64(1), sorted[1].Raw)
		assert.Equal(t, int64(3), sorted[2].Raw)
		assert.Equal(t, int64(4), sorted[3].Raw)
		assert.Equal(t, int64(5), sorted[4].Raw)
	})

	t.Run("sort strings ascending", func(t *testing.T) {
		list := types.List(types.String("banana"), types.String("apple"), types.String("cherry"))
		result, err := builtinSortAsc(list)
		require.NoError(t, err)

		sorted, ok := result.AsList()
		require.True(t, ok)
		assert.Equal(t, "apple", sorted[0].Raw)
		assert.Equal(t, "banana", sorted[1].Raw)
		assert.Equal(t, "cherry", sorted[2].Raw)
	})

	t.Run("empty list", func(t *testing.T) {
		result, err := builtinSortAsc(types.List())
		require.NoError(t, err)
		sorted, ok := result.AsList()
		require.True(t, ok)
		assert.Len(t, sorted, 0)
	})
}

func TestBuiltinSortDesc(t *testing.T) {
	t.Run("sort integers descending", func(t *testing.T) {
		list := types.List(types.Int(3), types.Int(1), types.Int(4), types.Int(1), types.Int(5))
		result, err := builtinSortDesc(list)
		require.NoError(t, err)

		sorted, ok := result.AsList()
		require.True(t, ok)
		require.Len(t, sorted, 5)
		assert.Equal(t, int64(5), sorted[0].Raw)
		assert.Equal(t, int64(4), sorted[1].Raw)
		assert.Equal(t, int64(3), sorted[2].Raw)
		assert.Equal(t, int64(1), sorted[3].Raw)
		assert.Equal(t, int64(1), sorted[4].Raw)
	})
}

func TestBuiltinAll(t *testing.T) {
	tests := []struct {
		name     string
		list     types.Value
		expected bool
	}{
		{"all true", types.List(types.Bool(true), types.Bool(true), types.Bool(true)), true},
		{"one false", types.List(types.Bool(true), types.Bool(false), types.Bool(true)), false},
		{"all false", types.List(types.Bool(false), types.Bool(false)), false},
		{"all truthy numbers", types.List(types.Int(1), types.Int(2), types.Int(3)), true},
		{"contains zero", types.List(types.Int(1), types.Int(0), types.Int(3)), false},
		{"empty list", types.List(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinAll(tt.list)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinAny(t *testing.T) {
	tests := []struct {
		name     string
		list     types.Value
		expected bool
	}{
		{"all true", types.List(types.Bool(true), types.Bool(true)), true},
		{"one true", types.List(types.Bool(false), types.Bool(true), types.Bool(false)), true},
		{"all false", types.List(types.Bool(false), types.Bool(false)), false},
		{"one truthy number", types.List(types.Int(0), types.Int(0), types.Int(1)), true},
		{"all zero", types.List(types.Int(0), types.Int(0)), false},
		{"empty list", types.List(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinAny(tt.list)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

// ============================================================================
// Tests for Additional Numeric Functions
// ============================================================================

func TestBuiltinClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    types.Value
		min      types.Value
		max      types.Value
		expected interface{}
	}{
		{"value in range", types.Int(5), types.Int(0), types.Int(10), int64(5)},
		{"value below min", types.Int(-5), types.Int(0), types.Int(10), int64(0)},
		{"value above max", types.Int(15), types.Int(0), types.Int(10), int64(10)},
		{"float in range", types.Float(5.5), types.Float(0.0), types.Float(10.0), float64(5.5)},
		{"float below min", types.Float(-5.5), types.Float(0.0), types.Float(10.0), float64(0.0)},
		{"float above max", types.Float(15.5), types.Float(0.0), types.Float(10.0), float64(10.0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinClamp(tt.value, tt.min, tt.max)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinBetween(t *testing.T) {
	tests := []struct {
		name     string
		value    types.Value
		min      types.Value
		max      types.Value
		expected bool
	}{
		{"value in range", types.Int(5), types.Int(0), types.Int(10), true},
		{"value at min", types.Int(0), types.Int(0), types.Int(10), true},
		{"value at max", types.Int(10), types.Int(0), types.Int(10), true},
		{"value below min", types.Int(-5), types.Int(0), types.Int(10), false},
		{"value above max", types.Int(15), types.Int(0), types.Int(10), false},
		{"float in range", types.Float(5.5), types.Float(0.0), types.Float(10.0), true},
		{"string in range", types.String("b"), types.String("a"), types.String("c"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinBetween(tt.value, tt.min, tt.max)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

// ============================================================================
// Tests for Additional Utility Functions
// ============================================================================

func TestBuiltinDefaultVal(t *testing.T) {
	t.Run("non-null value", func(t *testing.T) {
		result, err := builtinDefaultVal(types.Int(42), types.Int(0))
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Raw)
	})

	t.Run("null value returns default", func(t *testing.T) {
		result, err := builtinDefaultVal(types.Null(), types.Int(99))
		require.NoError(t, err)
		assert.Equal(t, int64(99), result.Raw)
	})

	t.Run("string default", func(t *testing.T) {
		result, err := builtinDefaultVal(types.Null(), types.String("default"))
		require.NoError(t, err)
		assert.Equal(t, "default", result.Raw)
	})
}

func TestBuiltinFormat(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected string
	}{
		{"simple string", []types.Value{types.String("Hello, {0}!"), types.String("World")}, "Hello, World!"},
		{"multiple args", []types.Value{types.String("{0} + {1} = {2}"), types.Int(2), types.Int(3), types.Int(5)}, "2 + 3 = 5"},
		{"float arg", []types.Value{types.String("Pi is approximately {0}"), types.Float(3.14)}, "Pi is approximately 3.14"},
		{"bool arg", []types.Value{types.String("The answer is {0}"), types.Bool(true)}, "The answer is true"},
		{"null arg", []types.Value{types.String("Value: {0}"), types.Null()}, "Value: null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinFormat(tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

// ============================================================================
// Tests for Additional String Functions
// ============================================================================

func TestBuiltinTrimLeft(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello", "hello"},
		{"\t\nhello", "hello"},
		{"hello  ", "hello  "},
		{"  hello  ", "hello  "},
		{"hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := builtinTrimLeft(types.String(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinTrimRight(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello  ", "hello"},
		{"hello\t\n", "hello"},
		{"  hello", "  hello"},
		{"  hello  ", "  hello"},
		{"hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := builtinTrimRight(types.String(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinPadLeft(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		length   int64
		pad      string
		expected string
	}{
		{"pad with zeros", "42", 5, "0", "00042"},
		{"pad with spaces", "hi", 5, " ", "   hi"},
		{"no padding needed", "hello", 3, "x", "llo"},
		{"exact length", "abc", 3, "x", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinPadLeft(types.String(tt.str), types.Int(tt.length), types.String(tt.pad))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinPadRight(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		length   int64
		pad      string
		expected string
	}{
		{"pad with dots", "hi", 5, ".", "hi..."},
		{"pad with spaces", "hi", 5, " ", "hi   "},
		{"no padding needed", "hello", 3, "x", "hel"},
		{"exact length", "abc", 3, "x", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinPadRight(types.String(tt.str), types.Int(tt.length), types.String(tt.pad))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinRepeat(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		count    int64
		expected string
	}{
		{"repeat 3 times", "ab", 3, "ababab"},
		{"repeat 1 time", "hello", 1, "hello"},
		{"repeat 0 times", "hello", 0, ""},
		{"empty string", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinRepeat(types.String(tt.str), types.Int(tt.count))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestNewDefaultRegistry(t *testing.T) {
	r, err := NewDefaultRegistry()
	require.NoError(t, err)
	require.NotNil(t, r)

	// Check that all expected functions are registered
	expectedFunctions := []string{
		// Aggregate
		"count", "sum", "avg", "min", "max",
		// Math
		"abs", "ceil", "floor", "round", "pow", "sqrt", "mod",
		// String
		"len", "lower", "upper", "trim", "contains", "startsWith", "endsWith",
		"substr", "replace", "split", "join", "concat", "match",
		// Type conversion
		"int", "float", "string", "bool",
		// List
		"first", "last", "at", "reverse", "unique", "flatten", "slice",
		// Utility
		"coalesce", "ifThenElse", "isNull", "isNotNull", "isEmpty", "typeOf",
	}

	for _, name := range expectedFunctions {
		assert.True(t, r.Has(name), "expected function %s to be registered", name)
	}
}

// ============================================================================
// Aggregate Function Tests
// ============================================================================

func TestBuiltinCount(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected int64
	}{
		{
			name:     "empty args",
			args:     []types.Value{},
			expected: 0,
		},
		{
			name:     "count list elements",
			args:     []types.Value{types.List(types.Int(1), types.Int(2), types.Int(3))},
			expected: 3,
		},
		{
			name:     "count variadic args",
			args:     []types.Value{types.Int(1), types.Int(2), types.Int(3), types.Int(4)},
			expected: 4,
		},
		{
			name:     "empty list",
			args:     []types.Value{types.List()},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinCount(tt.args...)
			require.NoError(t, err)
			assert.Equal(t, types.TypeInt, result.Type)
			got, _ := result.AsInt()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinSum(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected float64
	}{
		{
			name:     "sum integers",
			args:     []types.Value{types.Int(1), types.Int(2), types.Int(3)},
			expected: 6,
		},
		{
			name:     "sum floats",
			args:     []types.Value{types.Float(1.5), types.Float(2.5), types.Float(3.0)},
			expected: 7.0,
		},
		{
			name:     "sum list",
			args:     []types.Value{types.List(types.Int(10), types.Int(20), types.Int(30))},
			expected: 60,
		},
		{
			name:     "sum mixed types",
			args:     []types.Value{types.Int(1), types.Float(2.5), types.Int(3)},
			expected: 6.5,
		},
		{
			name:     "empty",
			args:     []types.Value{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinSum(tt.args...)
			require.NoError(t, err)
			got, _ := result.AsFloat()
			assert.InDelta(t, tt.expected, got, 0.001)
		})
	}
}

func TestBuiltinAvg(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected float64
		isNull   bool
	}{
		{
			name:     "avg integers",
			args:     []types.Value{types.Int(2), types.Int(4), types.Int(6)},
			expected: 4.0,
		},
		{
			name:     "avg floats",
			args:     []types.Value{types.Float(1.0), types.Float(2.0), types.Float(3.0)},
			expected: 2.0,
		},
		{
			name:     "avg list",
			args:     []types.Value{types.List(types.Int(10), types.Int(20), types.Int(30))},
			expected: 20.0,
		},
		{
			name:   "empty returns null",
			args:   []types.Value{},
			isNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinAvg(tt.args...)
			require.NoError(t, err)
			if tt.isNull {
				assert.True(t, result.IsNull())
			} else {
				got, _ := result.AsFloat()
				assert.InDelta(t, tt.expected, got, 0.001)
			}
		})
	}
}

func TestBuiltinMin(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected interface{}
		isNull   bool
	}{
		{
			name:     "min integers",
			args:     []types.Value{types.Int(5), types.Int(2), types.Int(8)},
			expected: int64(2),
		},
		{
			name:     "min floats",
			args:     []types.Value{types.Float(3.5), types.Float(1.5), types.Float(2.5)},
			expected: 1.5,
		},
		{
			name:     "min strings",
			args:     []types.Value{types.String("banana"), types.String("apple"), types.String("cherry")},
			expected: "apple",
		},
		{
			name:     "min list",
			args:     []types.Value{types.List(types.Int(100), types.Int(50), types.Int(75))},
			expected: int64(50),
		},
		{
			name:   "empty returns null",
			args:   []types.Value{},
			isNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinMin(tt.args...)
			require.NoError(t, err)
			if tt.isNull {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

func TestBuiltinMax(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected interface{}
		isNull   bool
	}{
		{
			name:     "max integers",
			args:     []types.Value{types.Int(5), types.Int(2), types.Int(8)},
			expected: int64(8),
		},
		{
			name:     "max floats",
			args:     []types.Value{types.Float(3.5), types.Float(1.5), types.Float(2.5)},
			expected: 3.5,
		},
		{
			name:     "max strings",
			args:     []types.Value{types.String("banana"), types.String("apple"), types.String("cherry")},
			expected: "cherry",
		},
		{
			name:     "max list",
			args:     []types.Value{types.List(types.Int(100), types.Int(50), types.Int(75))},
			expected: int64(100),
		},
		{
			name:   "empty returns null",
			args:   []types.Value{},
			isNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinMax(tt.args...)
			require.NoError(t, err)
			if tt.isNull {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

// ============================================================================
// Math Function Tests
// ============================================================================

func TestBuiltinAbs(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected float64
	}{
		{"positive int", types.Int(5), 5},
		{"negative int", types.Int(-5), 5},
		{"positive float", types.Float(3.14), 3.14},
		{"negative float", types.Float(-3.14), 3.14},
		{"zero", types.Int(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinAbs(tt.input)
			require.NoError(t, err)
			got, _ := result.AsFloat()
			assert.InDelta(t, tt.expected, got, 0.001)
		})
	}
}

func TestBuiltinCeil(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected int64
	}{
		{"positive float", types.Float(3.2), 4},
		{"negative float", types.Float(-3.2), -3},
		{"whole number", types.Float(5.0), 5},
		{"integer", types.Int(7), 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinCeil(tt.input)
			require.NoError(t, err)
			got, _ := result.AsInt()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinFloor(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected int64
	}{
		{"positive float", types.Float(3.8), 3},
		{"negative float", types.Float(-3.2), -4},
		{"whole number", types.Float(5.0), 5},
		{"integer", types.Int(7), 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinFloor(tt.input)
			require.NoError(t, err)
			got, _ := result.AsInt()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinRound(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected int64
	}{
		{"round down", types.Float(3.4), 3},
		{"round up", types.Float(3.6), 4},
		{"round half", types.Float(3.5), 4},
		{"negative round", types.Float(-3.5), -4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinRound(tt.input)
			require.NoError(t, err)
			got, _ := result.AsInt()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinPow(t *testing.T) {
	tests := []struct {
		name     string
		base     types.Value
		exp      types.Value
		expected float64
	}{
		{"2^3", types.Int(2), types.Int(3), 8},
		{"3^2", types.Int(3), types.Int(2), 9},
		{"2.5^2", types.Float(2.5), types.Int(2), 6.25},
		{"4^0.5", types.Int(4), types.Float(0.5), 2},
		{"any^0", types.Int(100), types.Int(0), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinPow(tt.base, tt.exp)
			require.NoError(t, err)
			got, _ := result.AsFloat()
			assert.InDelta(t, tt.expected, got, 0.001)
		})
	}
}

func TestBuiltinSqrt(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected float64
		hasError bool
	}{
		{"sqrt 4", types.Int(4), 2, false},
		{"sqrt 9", types.Int(9), 3, false},
		{"sqrt 2", types.Float(2), math.Sqrt(2), false},
		{"sqrt negative", types.Int(-1), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinSqrt(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				got, _ := result.AsFloat()
				assert.InDelta(t, tt.expected, got, 0.001)
			}
		})
	}
}

func TestBuiltinMod(t *testing.T) {
	tests := []struct {
		name     string
		a        types.Value
		b        types.Value
		expected int64
		hasError bool
	}{
		{"10 % 3", types.Int(10), types.Int(3), 1, false},
		{"15 % 4", types.Int(15), types.Int(4), 3, false},
		{"negative mod", types.Int(-7), types.Int(3), -1, false},
		{"div by zero", types.Int(5), types.Int(0), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinMod(tt.a, tt.b)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				got, _ := result.AsInt()
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

// ============================================================================
// String Function Tests
// ============================================================================

func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected int64
	}{
		{"string length", types.String("hello"), 5},
		{"empty string", types.String(""), 0},
		{"unicode string", types.String("héllo"), 5}, // len counts runes, not bytes
		{"list length", types.List(types.Int(1), types.Int(2), types.Int(3)), 3},
		{"empty list", types.List(), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinLen(tt.input)
			require.NoError(t, err)
			got, _ := result.AsInt()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinLower(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected string
	}{
		{"uppercase", types.String("HELLO"), "hello"},
		{"mixed case", types.String("HeLLo WoRLD"), "hello world"},
		{"already lower", types.String("hello"), "hello"},
		{"with numbers", types.String("Hello123"), "hello123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinLower(tt.input)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected string
	}{
		{"lowercase", types.String("hello"), "HELLO"},
		{"mixed case", types.String("HeLLo WoRLD"), "HELLO WORLD"},
		{"already upper", types.String("HELLO"), "HELLO"},
		{"with numbers", types.String("Hello123"), "HELLO123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinUpper(tt.input)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected string
	}{
		{"leading spaces", types.String("   hello"), "hello"},
		{"trailing spaces", types.String("hello   "), "hello"},
		{"both sides", types.String("   hello   "), "hello"},
		{"tabs and newlines", types.String("\t\nhello\n\t"), "hello"},
		{"no whitespace", types.String("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinTrim(tt.input)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinContains(t *testing.T) {
	tests := []struct {
		name     string
		str      types.Value
		substr   types.Value
		expected bool
	}{
		{"contains", types.String("hello world"), types.String("world"), true},
		{"not contains", types.String("hello world"), types.String("xyz"), false},
		{"contains at start", types.String("hello world"), types.String("hello"), true},
		{"case sensitive", types.String("hello world"), types.String("WORLD"), false},
		{"empty substr", types.String("hello"), types.String(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinContains(tt.str, tt.substr)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinStartsWith(t *testing.T) {
	tests := []struct {
		name     string
		str      types.Value
		prefix   types.Value
		expected bool
	}{
		{"starts with", types.String("hello world"), types.String("hello"), true},
		{"not starts with", types.String("hello world"), types.String("world"), false},
		{"full match", types.String("hello"), types.String("hello"), true},
		{"case sensitive", types.String("Hello"), types.String("hello"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinStartsWith(tt.str, tt.prefix)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinEndsWith(t *testing.T) {
	tests := []struct {
		name     string
		str      types.Value
		suffix   types.Value
		expected bool
	}{
		{"ends with", types.String("hello world"), types.String("world"), true},
		{"not ends with", types.String("hello world"), types.String("hello"), false},
		{"full match", types.String("hello"), types.String("hello"), true},
		{"case sensitive", types.String("Hello"), types.String("ELLO"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinEndsWith(tt.str, tt.suffix)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinSubstr(t *testing.T) {
	tests := []struct {
		name     string
		str      types.Value
		start    types.Value
		length   types.Value
		expected string
	}{
		{"basic", types.String("hello world"), types.Int(0), types.Int(5), "hello"},
		{"middle", types.String("hello world"), types.Int(6), types.Int(5), "world"},
		{"negative start", types.String("hello world"), types.Int(-5), types.Int(5), "world"},
		{"beyond length", types.String("hello"), types.Int(0), types.Int(100), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinSubstr(tt.str, tt.start, tt.length)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinReplace(t *testing.T) {
	tests := []struct {
		name     string
		str      types.Value
		old      types.Value
		new      types.Value
		expected string
	}{
		{"basic", types.String("hello world"), types.String("world"), types.String("there"), "hello there"},
		{"multiple", types.String("aaa"), types.String("a"), types.String("b"), "bbb"},
		{"no match", types.String("hello"), types.String("x"), types.String("y"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinReplace(tt.str, tt.old, tt.new)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinSplit(t *testing.T) {
	t.Run("basic split", func(t *testing.T) {
		result, err := builtinSplit(types.String("a,b,c"), types.String(","))
		require.NoError(t, err)
		list, ok := result.AsList()
		require.True(t, ok)
		require.Len(t, list, 3)
		s0, _ := list[0].AsString()
		s1, _ := list[1].AsString()
		s2, _ := list[2].AsString()
		assert.Equal(t, "a", s0)
		assert.Equal(t, "b", s1)
		assert.Equal(t, "c", s2)
	})

	t.Run("split with space", func(t *testing.T) {
		result, err := builtinSplit(types.String("hello world test"), types.String(" "))
		require.NoError(t, err)
		list, ok := result.AsList()
		require.True(t, ok)
		assert.Len(t, list, 3)
	})
}

func TestBuiltinJoin(t *testing.T) {
	t.Run("basic join", func(t *testing.T) {
		list := types.List(types.String("a"), types.String("b"), types.String("c"))
		result, err := builtinJoin(list, types.String(","))
		require.NoError(t, err)
		got, _ := result.AsString()
		assert.Equal(t, "a,b,c", got)
	})

	t.Run("join with mixed types", func(t *testing.T) {
		list := types.List(types.String("a"), types.Int(1), types.String("b"))
		result, err := builtinJoin(list, types.String("-"))
		require.NoError(t, err)
		got, _ := result.AsString()
		assert.Equal(t, "a-1-b", got)
	})
}

func TestBuiltinConcat(t *testing.T) {
	result, err := builtinConcat(types.String("hello"), types.String(" "), types.String("world"))
	require.NoError(t, err)
	got, _ := result.AsString()
	assert.Equal(t, "hello world", got)
}

func TestBuiltinMatch(t *testing.T) {
	tests := []struct {
		name     string
		str      types.Value
		pattern  types.Value
		expected bool
		hasError bool
	}{
		{"simple match", types.String("hello123"), types.String(`\d+`), true, false},
		{"no match", types.String("hello"), types.String(`\d+`), false, false},
		{"full pattern", types.String("test@email.com"), types.String(`^[\w]+@[\w]+\.[\w]+$`), true, false},
		{"invalid regex", types.String("hello"), types.String(`[`), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinMatch(tt.str, tt.pattern)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				got, _ := result.AsBool()
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

// ============================================================================
// Type Conversion Tests
// ============================================================================

func TestBuiltinInt(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected int64
		hasError bool
	}{
		{"from int", types.Int(42), 42, false},
		{"from float", types.Float(3.7), 3, false},
		{"from string", types.String("123"), 123, false},
		{"from bool true", types.Bool(true), 1, false},
		{"from bool false", types.Bool(false), 0, false},
		{"from invalid string", types.String("abc"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinInt(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				got, _ := result.AsInt()
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestBuiltinFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected float64
		hasError bool
	}{
		{"from float", types.Float(3.14), 3.14, false},
		{"from int", types.Int(42), 42.0, false},
		{"from string", types.String("3.14"), 3.14, false},
		{"from bool true", types.Bool(true), 1.0, false},
		{"from bool false", types.Bool(false), 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinFloat(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				got, _ := result.AsFloat()
				assert.InDelta(t, tt.expected, got, 0.001)
			}
		})
	}
}

func TestBuiltinString(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected string
	}{
		{"from string", types.String("hello"), "hello"},
		{"from int", types.Int(42), "42"},
		{"from float", types.Float(3.14), "3.14"},
		{"from bool", types.Bool(true), "true"},
		{"from null", types.Null(), "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinString(tt.input)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinBool(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected bool
	}{
		{"from true", types.Bool(true), true},
		{"from false", types.Bool(false), false},
		{"from non-zero int", types.Int(42), true},
		{"from zero int", types.Int(0), false},
		{"from non-empty string", types.String("hello"), true},
		{"from empty string", types.String(""), false},
		{"from null", types.Null(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinBool(tt.input)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ============================================================================
// List Function Tests
// ============================================================================

func TestBuiltinFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected interface{}
		isNull   bool
	}{
		{
			name:     "first of list",
			input:    types.List(types.Int(1), types.Int(2), types.Int(3)),
			expected: int64(1),
		},
		{
			name:   "empty list",
			input:  types.List(),
			isNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinFirst(tt.input)
			require.NoError(t, err)
			if tt.isNull {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

func TestBuiltinLast(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected interface{}
		isNull   bool
	}{
		{
			name:     "last of list",
			input:    types.List(types.Int(1), types.Int(2), types.Int(3)),
			expected: int64(3),
		},
		{
			name:   "empty list",
			input:  types.List(),
			isNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinLast(tt.input)
			require.NoError(t, err)
			if tt.isNull {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

func TestBuiltinAt(t *testing.T) {
	list := types.List(types.String("a"), types.String("b"), types.String("c"))

	tests := []struct {
		name     string
		index    types.Value
		expected string
		hasError bool
	}{
		{"index 0", types.Int(0), "a", false},
		{"index 1", types.Int(1), "b", false},
		{"negative index", types.Int(-1), "c", false},
		{"out of bounds", types.Int(10), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinAt(list, tt.index)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				got, _ := result.AsString()
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestBuiltinReverse(t *testing.T) {
	list := types.List(types.Int(1), types.Int(2), types.Int(3))
	result, err := builtinReverse(list)
	require.NoError(t, err)

	reversed, ok := result.AsList()
	require.True(t, ok)
	require.Len(t, reversed, 3)

	v0, _ := reversed[0].AsInt()
	v1, _ := reversed[1].AsInt()
	v2, _ := reversed[2].AsInt()
	assert.Equal(t, int64(3), v0)
	assert.Equal(t, int64(2), v1)
	assert.Equal(t, int64(1), v2)
}

func TestBuiltinUnique(t *testing.T) {
	list := types.List(types.Int(1), types.Int(2), types.Int(1), types.Int(3), types.Int(2))
	result, err := builtinUnique(list)
	require.NoError(t, err)

	unique, ok := result.AsList()
	require.True(t, ok)
	assert.Len(t, unique, 3) // Should have 1, 2, 3
}

func TestBuiltinFlatten(t *testing.T) {
	nested := types.List(
		types.Int(1),
		types.List(types.Int(2), types.Int(3)),
		types.List(types.Int(4), types.List(types.Int(5))),
	)
	result, err := builtinFlatten(nested)
	require.NoError(t, err)

	flat, ok := result.AsList()
	require.True(t, ok)
	assert.Len(t, flat, 5)
}

func TestBuiltinSlice(t *testing.T) {
	list := types.List(types.Int(1), types.Int(2), types.Int(3), types.Int(4), types.Int(5))

	tests := []struct {
		name     string
		start    types.Value
		end      types.Value
		expected int
	}{
		{"basic slice", types.Int(1), types.Int(4), 3},     // [2,3,4]
		{"negative end", types.Int(0), types.Int(-1), 4},   // [1,2,3,4]
		{"full slice", types.Int(0), types.Int(5), 5},      // [1,2,3,4,5]
		{"empty slice", types.Int(2), types.Int(2), 0},     // []
		{"negative start", types.Int(-3), types.Int(5), 3}, // [3,4,5]
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinSlice(list, tt.start, tt.end)
			require.NoError(t, err)
			sliced, ok := result.AsList()
			require.True(t, ok)
			assert.Len(t, sliced, tt.expected)
		})
	}
}

// ============================================================================
// Utility Function Tests
// ============================================================================

func TestBuiltinCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		args     []types.Value
		expected interface{}
		isNull   bool
	}{
		{
			name:     "first non-null",
			args:     []types.Value{types.Null(), types.Int(1), types.Int(2)},
			expected: int64(1),
		},
		{
			name:   "all null",
			args:   []types.Value{types.Null(), types.Null()},
			isNull: true,
		},
		{
			name:     "first is not null",
			args:     []types.Value{types.String("hello"), types.Null()},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinCoalesce(tt.args...)
			require.NoError(t, err)
			if tt.isNull {
				assert.True(t, result.IsNull())
			} else {
				assert.Equal(t, tt.expected, result.Raw)
			}
		})
	}
}

func TestBuiltinIfThenElse(t *testing.T) {
	tests := []struct {
		name      string
		condition types.Value
		thenVal   types.Value
		elseVal   types.Value
		expected  interface{}
	}{
		{"true condition", types.Bool(true), types.String("yes"), types.String("no"), "yes"},
		{"false condition", types.Bool(false), types.String("yes"), types.String("no"), "no"},
		{"truthy int", types.Int(1), types.Int(100), types.Int(200), int64(100)},
		{"falsy int", types.Int(0), types.Int(100), types.Int(200), int64(200)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinIfThenElse(tt.condition, tt.thenVal, tt.elseVal)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestBuiltinIsNull(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected bool
	}{
		{"null value", types.Null(), true},
		{"non-null int", types.Int(0), false},
		{"non-null string", types.String(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinIsNull(tt.input)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinIsNotNull(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected bool
	}{
		{"null value", types.Null(), false},
		{"non-null int", types.Int(0), true},
		{"non-null string", types.String(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinIsNotNull(tt.input)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected bool
	}{
		{"null", types.Null(), true},
		{"empty string", types.String(""), true},
		{"non-empty string", types.String("hello"), false},
		{"empty list", types.List(), true},
		{"non-empty list", types.List(types.Int(1)), false},
		{"zero int (not empty)", types.Int(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinIsEmpty(tt.input)
			require.NoError(t, err)
			got, _ := result.AsBool()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBuiltinTypeOf(t *testing.T) {
	tests := []struct {
		name     string
		input    types.Value
		expected string
	}{
		{"int", types.Int(42), "int"},
		{"float", types.Float(3.14), "float"},
		{"string", types.String("hello"), "string"},
		{"bool", types.Bool(true), "bool"},
		{"null", types.Null(), "null"},
		{"list", types.List(types.Int(1)), "list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builtinTypeOf(tt.input)
			require.NoError(t, err)
			got, _ := result.AsString()
			assert.Equal(t, tt.expected, got)
		})
	}
}

// ============================================================================
// Registry Integration Tests
// ============================================================================

func TestRegistryCall(t *testing.T) {
	r, err := NewDefaultRegistry()
	require.NoError(t, err)

	// Test calling through registry
	result, err := r.Call("sum", types.Int(1), types.Int(2), types.Int(3))
	require.NoError(t, err)
	got, _ := result.AsFloat()
	assert.InDelta(t, 6.0, got, 0.001)

	// Test undefined function
	_, err = r.Call("undefined_function")
	assert.Error(t, err)
}

func TestComplexExpressions(t *testing.T) {
	r, err := NewDefaultRegistry()
	require.NoError(t, err)

	// Test: avg of max and min
	list := types.List(types.Int(1), types.Int(5), types.Int(10))
	minResult, _ := r.Call("min", list)
	maxResult, _ := r.Call("max", list)
	avgResult, err := r.Call("avg", minResult, maxResult)
	require.NoError(t, err)
	got, _ := avgResult.AsFloat()
	assert.InDelta(t, 5.5, got, 0.001) // (1 + 10) / 2 = 5.5

	// Test: nested function calls - len(split("a,b,c", ","))
	splitResult, _ := r.Call("split", types.String("a,b,c"), types.String(","))
	lenResult, err := r.Call("len", splitResult)
	require.NoError(t, err)
	count, _ := lenResult.AsInt()
	assert.Equal(t, int64(3), count)
}
