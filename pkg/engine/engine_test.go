// Package engine provides the main AMEL engine facade.
package engine

import (
	"testing"
	"time"

	"github.com/bencagri/amel/pkg/functions"
	"github.com/bencagri/amel/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_New(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestEngine_NewWithOptions(t *testing.T) {
	engine, err := New(
		WithTimeout(200*time.Millisecond),
		WithExplainMode(true),
		WithCaching(true),
		WithStrictTypes(true),
	)
	require.NoError(t, err)
	require.NotNil(t, engine)
	assert.Equal(t, 200*time.Millisecond, engine.timeout)
	assert.True(t, engine.explainMode)
	assert.True(t, engine.caching)
	assert.True(t, engine.strictTypes)
}

func TestEngine_Compile(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	tests := []struct {
		name    string
		dsl     string
		wantErr bool
	}{
		{"simple literal", "42", false},
		{"comparison", "$.age >= 18", false},
		{"complex expression", `$.role IN ["admin", "user"] && $.verified == true`, false},
		{"function call", "max(1, 2, 3)", false},
		{"invalid syntax", "(5 +", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := engine.Compile(tt.dsl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, compiled)
				assert.Equal(t, tt.dsl, compiled.Source)
			}
		})
	}
}

func TestEngineJSFunctions(t *testing.T) {
	t.Run("register and call JS function", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		// Register a simple JS function
		err = engine.RegisterFunction(`function double(x) { return x * 2; }`)
		require.NoError(t, err)

		// Compile and evaluate expression using the JS function
		compiled, err := engine.Compile("double(5)")
		require.NoError(t, err)

		result, err := engine.Evaluate(compiled, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(10), result.Raw)
	})

	t.Run("JS function with string manipulation", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		err = engine.RegisterFunction(`function greet(name) { return "Hello, " + name + "!"; }`)
		require.NoError(t, err)

		result, err := engine.EvaluateDirect(`greet("World")`, nil)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", result.Raw)
	})

	t.Run("JS function with array operations", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		err = engine.RegisterFunction(`function sumArray(arr) { return arr.reduce((a, b) => a + b, 0); }`)
		require.NoError(t, err)

		result, err := engine.EvaluateDirect("sumArray([1, 2, 3, 4, 5])", nil)
		require.NoError(t, err)
		assert.Equal(t, int64(15), result.Raw)
	})

	t.Run("JS function with payload access", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		err = engine.RegisterFunction(`function checkAge(age) { return age >= 18; }`)
		require.NoError(t, err)

		payload := map[string]interface{}{
			"user": map[string]interface{}{
				"age": 25,
			},
		}

		result, err := engine.EvaluateDirect("checkAge($.user.age)", payload)
		require.NoError(t, err)
		assert.Equal(t, true, result.Raw)
	})

	t.Run("JS function registration error - invalid syntax", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		err = engine.RegisterFunction(`not a valid function`)
		require.Error(t, err)
	})

	t.Run("combined built-in and JS functions", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		err = engine.RegisterFunction(`function square(x) { return x * x; }`)
		require.NoError(t, err)

		// Use both built-in max() and custom square()
		result, err := engine.EvaluateDirect("max(square(3), square(2))", nil)
		require.NoError(t, err)
		assert.Equal(t, int64(9), result.Raw)
	})
}

func TestEngineSandboxOptions(t *testing.T) {
	t.Run("custom sandbox config", func(t *testing.T) {
		config := &functions.SandboxConfig{
			Timeout:       200 * time.Millisecond,
			MemoryLimit:   5 * 1024 * 1024,
			MaxStackDepth: 50,
		}

		engine, err := New(WithSandboxConfig(config))
		require.NoError(t, err)

		sandbox := engine.GetSandbox()
		require.NotNil(t, sandbox)
		assert.Equal(t, 200*time.Millisecond, sandbox.Config().Timeout)
	})

	t.Run("custom sandbox instance", func(t *testing.T) {
		sandbox := functions.NewSandbox(&functions.SandboxConfig{
			Timeout:       300 * time.Millisecond,
			MaxStackDepth: 75,
		})

		engine, err := New(WithSandbox(sandbox))
		require.NoError(t, err)

		assert.Equal(t, sandbox, engine.GetSandbox())
	})
}

func TestEngineOptimization(t *testing.T) {
	t.Run("optimization enabled by default", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		optimizer := engine.GetOptimizer()
		require.NotNil(t, optimizer)
	})

	t.Run("optimization disabled", func(t *testing.T) {
		engine, err := New(WithOptimization(false))
		require.NoError(t, err)

		optimizer := engine.GetOptimizer()
		assert.Nil(t, optimizer)
	})

	t.Run("constant folding optimization", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		// Expression with constants that can be folded
		compiled, err := engine.Compile("(2 + 3) * 4 + $.value")
		require.NoError(t, err)

		// The optimized AST should have folded (2 + 3) * 4 = 20
		payload := map[string]interface{}{"value": 5}
		result, err := engine.Evaluate(compiled, payload)
		require.NoError(t, err)
		assert.Equal(t, int64(25), result.Raw)
	})

	t.Run("fully constant expression", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)

		compiled, err := engine.Compile("(10 + 5) * 2 - 10")
		require.NoError(t, err)

		result, err := engine.Evaluate(compiled, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(20), result.Raw)
	})
}

func TestEngineExplanationMode(t *testing.T) {
	t.Run("explanation with binary expression", func(t *testing.T) {
		engine, err := New(WithExplainMode(true))
		require.NoError(t, err)

		compiled, err := engine.Compile("$.a + $.b")
		require.NoError(t, err)

		payload := map[string]interface{}{"a": 10, "b": 20}
		result, explanation, err := engine.EvaluateWithExplanation(compiled, payload)
		require.NoError(t, err)

		assert.Equal(t, int64(30), result.Raw)
		assert.NotNil(t, explanation)
		assert.NotEmpty(t, explanation.Expression)
		assert.NotEmpty(t, explanation.Reason)
	})

	t.Run("explanation with function call", func(t *testing.T) {
		engine, err := New(WithExplainMode(true))
		require.NoError(t, err)

		compiled, err := engine.Compile("max($.x, $.y)")
		require.NoError(t, err)

		payload := map[string]interface{}{"x": 5, "y": 10}
		result, explanation, err := engine.EvaluateWithExplanation(compiled, payload)
		require.NoError(t, err)

		assert.Equal(t, int64(10), result.Raw)
		assert.NotNil(t, explanation)
		assert.Contains(t, explanation.Reason, "max")
	})

	t.Run("explanation with complex expression", func(t *testing.T) {
		engine, err := New(WithExplainMode(true))
		require.NoError(t, err)

		compiled, err := engine.Compile("$.age >= 18 && $.verified == true")
		require.NoError(t, err)

		payload := map[string]interface{}{"age": 25, "verified": true}
		result, explanation, err := engine.EvaluateWithExplanation(compiled, payload)
		require.NoError(t, err)

		assert.Equal(t, true, result.Raw)
		assert.NotNil(t, explanation)
		assert.NotNil(t, explanation.Children)
	})
}

func TestEngineRequest(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"name":     "John",
			"age":      25,
			"verified": true,
			"role":     "admin",
		},
	}

	tests := []struct {
		name     string
		dsl      string
		expected interface{}
	}{
		{"literal int", "42", int64(42)},
		{"literal string", `"hello"`, "hello"},
		{"literal bool", "true", true},
		{"json path", "$.user.name", "John"},
		{"json path int", "$.user.age", int64(25)},
		{"comparison true", "$.user.age >= 18", true},
		{"comparison false", "$.user.age < 18", false},
		{"logical and", "$.user.age >= 18 && $.user.verified == true", true},
		{"in expression", `$.user.role IN ["admin", "moderator"]`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := engine.Compile(tt.dsl)
			require.NoError(t, err)

			result, err := engine.Evaluate(compiled, payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestEngine_EvaluateBool(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"age":      25,
		"verified": true,
	}

	tests := []struct {
		name     string
		dsl      string
		expected bool
	}{
		{"true condition", "$.age >= 18", true},
		{"false condition", "$.age < 18", false},
		{"and true", "$.age >= 18 && $.verified == true", true},
		{"and false", "$.age >= 30 && $.verified == true", false},
		{"or true", "$.age >= 30 || $.verified == true", true},
		{"not true", "!false", true},
		{"literal true", "true", true},
		{"literal false", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := engine.Compile(tt.dsl)
			require.NoError(t, err)

			result, err := engine.EvaluateBool(compiled, payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEngine_EvaluateDirect(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"x": 10,
		"y": 20,
	}

	result, err := engine.EvaluateDirect("$.x + $.y", payload)
	require.NoError(t, err)

	got, ok := result.AsInt()
	require.True(t, ok)
	assert.Equal(t, int64(30), got)
}

func TestEngine_EvaluateDirectBool(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"count": 5,
	}

	result, err := engine.EvaluateDirectBool("$.count > 0", payload)
	require.NoError(t, err)
	assert.True(t, result)
}

func TestEngine_EvaluateWithExplanation(t *testing.T) {
	engine, err := New(WithExplainMode(true))
	require.NoError(t, err)

	payload := map[string]interface{}{
		"age": 25,
	}

	compiled, err := engine.Compile("$.age >= 18")
	require.NoError(t, err)

	result, explanation, err := engine.EvaluateWithExplanation(compiled, payload)
	require.NoError(t, err)

	assert.True(t, result.IsTruthy())
	assert.NotNil(t, explanation)
	assert.Equal(t, "($.age >= 18)", explanation.Expression)
}

func TestEngine_Caching(t *testing.T) {
	engine, err := New(WithCaching(true))
	require.NoError(t, err)

	dsl := "$.value > 10"

	// First compile
	compiled1, err := engine.Compile(dsl)
	require.NoError(t, err)

	// Second compile should return cached
	compiled2, err := engine.Compile(dsl)
	require.NoError(t, err)

	// Should be the same instance
	assert.Same(t, compiled1, compiled2)

	// Clear cache
	engine.ClearCache()

	// Third compile should create new
	compiled3, err := engine.Compile(dsl)
	require.NoError(t, err)

	// Should be different instance after clear
	assert.NotSame(t, compiled1, compiled3)
}

func TestEngine_FunctionCalls(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	tests := []struct {
		name     string
		dsl      string
		expected interface{}
	}{
		{"max", "max(1, 5, 3)", int64(5)},
		{"min", "min(1, 5, 3)", int64(1)},
		{"sum", "sum(1, 2, 3)", 6.0},
		{"avg", "avg(2, 4, 6)", 4.0},
		{"abs", "abs(-5)", 5.0},
		{"ceil", "ceil(3.2)", int64(4)},
		{"floor", "floor(3.8)", int64(3)},
		{"len", `len("hello")`, int64(5)},
		{"upper", `upper("hello")`, "HELLO"},
		{"lower", `lower("WORLD")`, "world"},
		{"contains", `contains("hello world", "world")`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.EvaluateDirect(tt.dsl, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Raw)
		})
	}
}

func TestEngine_ListFunctions(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	funcs := engine.ListFunctions()
	assert.NotEmpty(t, funcs)

	// Check for some expected functions
	funcSet := make(map[string]bool)
	for _, f := range funcs {
		funcSet[f] = true
	}

	expectedFuncs := []string{"max", "min", "sum", "avg", "len", "upper", "lower", "contains"}
	for _, expected := range expectedFuncs {
		assert.True(t, funcSet[expected], "expected function %s to be registered", expected)
	}
}

func TestEngine_RegisterBuiltIn(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	// Register a custom function
	err = engine.RegisterBuiltIn("double", func(args ...types.Value) (types.Value, error) {
		if len(args) == 0 {
			return types.Null(), nil
		}
		if n, ok := args[0].AsInt(); ok {
			return types.Int(n * 2), nil
		}
		if f, ok := args[0].AsFloat(); ok {
			return types.Float(f * 2), nil
		}
		return types.Null(), nil
	}, types.NewFunctionSignature("double", types.TypeAny, types.Param("value", types.TypeAny)))
	require.NoError(t, err)

	// Use the custom function
	result, err := engine.EvaluateDirect("double(21)", nil)
	require.NoError(t, err)

	got, ok := result.AsInt()
	require.True(t, ok)
	assert.Equal(t, int64(42), got)
}

func TestEngine_ComplexExpressions(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"name":       "Alice",
			"age":        30,
			"role":       "admin",
			"verified":   true,
			"reputation": 2500,
		},
		"order": map[string]interface{}{
			"total": 150.50,
			"items": []interface{}{
				map[string]interface{}{"name": "Item 1", "price": 50.00},
				map[string]interface{}{"name": "Item 2", "price": 100.50},
			},
		},
	}

	tests := []struct {
		name     string
		dsl      string
		expected bool
	}{
		{
			name:     "admin access check",
			dsl:      `$.user.role == "admin" && $.user.verified == true`,
			expected: true,
		},
		{
			name:     "role or reputation",
			dsl:      `$.user.role IN ["admin", "moderator"] || $.user.reputation >= 1000`,
			expected: true,
		},
		{
			name:     "complex access control",
			dsl:      `($.user.age >= 18 && $.user.verified == true) && ($.user.role == "admin" || $.user.reputation >= 2000)`,
			expected: true,
		},
		{
			name:     "order threshold",
			dsl:      `$.order.total > 100`,
			expected: true,
		},
		{
			name:     "name length check",
			dsl:      `len($.user.name) >= 3`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.EvaluateDirectBool(tt.dsl, payload)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEngine_EvaluateRequest(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	req := &EvalRequest{
		Payload: map[string]interface{}{
			"value": 42,
		},
		DSL: "$.value * 2",
	}

	resp := engine.EvaluateRequest(req)
	assert.Empty(t, resp.Error)
	assert.Equal(t, int64(84), resp.Result)
	assert.Equal(t, "int", resp.Type)
}

func TestEngine_EvaluateRequestWithError(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	req := &EvalRequest{
		Payload: map[string]interface{}{},
		DSL:     "(invalid syntax",
	}

	resp := engine.EvaluateRequest(req)
	assert.NotEmpty(t, resp.Error)
}

func TestEngine_EvaluateRequestWithExplanation(t *testing.T) {
	engine, err := New(WithExplainMode(true))
	require.NoError(t, err)

	req := &EvalRequest{
		Payload: map[string]interface{}{
			"x": 5,
		},
		DSL: "$.x > 3",
	}

	resp := engine.EvaluateRequest(req)
	assert.Empty(t, resp.Error)
	assert.True(t, resp.Result.(bool))
	assert.NotNil(t, resp.Explanation)
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("Eval", func(t *testing.T) {
		result, err := Eval("5 + 3", nil)
		require.NoError(t, err)
		got, ok := result.AsInt()
		require.True(t, ok)
		assert.Equal(t, int64(8), got)
	})

	t.Run("EvalBool", func(t *testing.T) {
		result, err := EvalBool("5 > 3", nil)
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("MustEval", func(t *testing.T) {
		result := MustEval("10 * 2", nil)
		got, ok := result.AsInt()
		require.True(t, ok)
		assert.Equal(t, int64(20), got)
	})

	t.Run("MustEvalBool", func(t *testing.T) {
		result := MustEvalBool("true && true", nil)
		assert.True(t, result)
	})
}

func TestMustEval_Panic(t *testing.T) {
	assert.Panics(t, func() {
		MustEval("(invalid", nil)
	})
}

func TestMustEvalBool_Panic(t *testing.T) {
	assert.Panics(t, func() {
		MustEvalBool("(invalid", nil)
	})
}

// Benchmark tests
func BenchmarkEngine_Compile(b *testing.B) {
	engine, _ := New()
	dsl := `$.user.age >= 18 && $.user.role IN ["admin", "user"]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Compile(dsl)
	}
}

func BenchmarkEngine_CompileWithCaching(b *testing.B) {
	engine, _ := New(WithCaching(true))
	dsl := `$.user.age >= 18 && $.user.role IN ["admin", "user"]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Compile(dsl)
	}
}

func BenchmarkEngine_Evaluate(b *testing.B) {
	engine, _ := New()
	compiled, _ := engine.Compile(`$.user.age >= 18`)
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"age": 25,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.EvaluateBool(compiled, payload)
	}
}

func BenchmarkEngine_EvaluateComplex(b *testing.B) {
	engine, _ := New()
	compiled, _ := engine.Compile(`($.user.role IN ["admin", "moderator"] || $.user.reputation >= 1000) && $.user.verified == true && $.user.age >= 18`)
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"role":       "admin",
			"reputation": 1500,
			"verified":   true,
			"age":        25,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.EvaluateBool(compiled, payload)
	}
}
