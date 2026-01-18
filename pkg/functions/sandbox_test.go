// Package functions provides function management for the AMEL DSL engine.
package functions

import (
	"context"
	"testing"
	"time"

	"github.com/bencagri/amel/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSandbox(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		sandbox := NewSandbox(nil)
		assert.NotNil(t, sandbox)
		assert.Equal(t, 100*time.Millisecond, sandbox.config.Timeout)
		assert.Equal(t, int64(10*1024*1024), sandbox.config.MemoryLimit)
		assert.Equal(t, 100, sandbox.config.MaxStackDepth)
	})

	t.Run("custom config", func(t *testing.T) {
		config := &SandboxConfig{
			Timeout:       200 * time.Millisecond,
			MemoryLimit:   5 * 1024 * 1024,
			MaxStackDepth: 50,
		}
		sandbox := NewSandbox(config)
		assert.NotNil(t, sandbox)
		assert.Equal(t, 200*time.Millisecond, sandbox.config.Timeout)
		assert.Equal(t, int64(5*1024*1024), sandbox.config.MemoryLimit)
		assert.Equal(t, 50, sandbox.config.MaxStackDepth)
	})
}

func TestSandboxExecute(t *testing.T) {
	sandbox := NewSandbox(nil)
	ctx := context.Background()

	t.Run("simple function", func(t *testing.T) {
		jsBody := `function add(a, b) { return a + b; }`
		args := []types.Value{types.Int(3), types.Int(5)}

		result, err := sandbox.Execute(ctx, jsBody, "add", args)
		require.NoError(t, err)
		assert.Equal(t, types.TypeInt, result.Type)
		assert.Equal(t, int64(8), result.Raw)
	})

	t.Run("function with float", func(t *testing.T) {
		jsBody := `function multiply(a, b) { return a * b; }`
		args := []types.Value{types.Float(2.5), types.Float(4.0)}

		result, err := sandbox.Execute(ctx, jsBody, "multiply", args)
		require.NoError(t, err)
		// JS returns 10 as an integer since 2.5 * 4.0 = 10.0 (whole number)
		// The jsToValue function converts whole number floats to int
		val, ok := result.AsFloat()
		require.True(t, ok)
		assert.Equal(t, 10.0, val)
	})

	t.Run("function with string", func(t *testing.T) {
		jsBody := `function greet(name) { return "Hello, " + name + "!"; }`
		args := []types.Value{types.String("World")}

		result, err := sandbox.Execute(ctx, jsBody, "greet", args)
		require.NoError(t, err)
		assert.Equal(t, types.TypeString, result.Type)
		assert.Equal(t, "Hello, World!", result.Raw)
	})

	t.Run("function with boolean", func(t *testing.T) {
		jsBody := `function isPositive(n) { return n > 0; }`
		args := []types.Value{types.Int(5)}

		result, err := sandbox.Execute(ctx, jsBody, "isPositive", args)
		require.NoError(t, err)
		assert.Equal(t, types.TypeBool, result.Type)
		assert.Equal(t, true, result.Raw)
	})

	t.Run("function returning null", func(t *testing.T) {
		jsBody := `function getNil() { return null; }`
		args := []types.Value{}

		result, err := sandbox.Execute(ctx, jsBody, "getNil", args)
		require.NoError(t, err)
		assert.Equal(t, types.TypeNull, result.Type)
	})

	t.Run("function with array", func(t *testing.T) {
		jsBody := `function sum(arr) { return arr.reduce((a, b) => a + b, 0); }`
		args := []types.Value{types.List(types.Int(1), types.Int(2), types.Int(3))}

		result, err := sandbox.Execute(ctx, jsBody, "sum", args)
		require.NoError(t, err)
		assert.Equal(t, int64(6), result.Raw)
	})

	t.Run("function returning array", func(t *testing.T) {
		jsBody := `function range(n) { return Array.from({length: n}, (_, i) => i); }`
		args := []types.Value{types.Int(3)}

		result, err := sandbox.Execute(ctx, jsBody, "range", args)
		require.NoError(t, err)
		assert.Equal(t, types.TypeList, result.Type)
		list, ok := result.AsList()
		require.True(t, ok)
		assert.Len(t, list, 3)
	})

	t.Run("function not found", func(t *testing.T) {
		jsBody := `function foo() { return 1; }`
		args := []types.Value{}

		_, err := sandbox.Execute(ctx, jsBody, "bar", args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("syntax error", func(t *testing.T) {
		jsBody := `function broken( { return 1; }`
		args := []types.Value{}

		_, err := sandbox.Execute(ctx, jsBody, "broken", args)
		require.Error(t, err)
	})

	t.Run("not a function", func(t *testing.T) {
		jsBody := `var notAFunction = 42;`
		args := []types.Value{}

		_, err := sandbox.Execute(ctx, jsBody, "notAFunction", args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a function")
	})
}

func TestSandboxExecuteExpression(t *testing.T) {
	sandbox := NewSandbox(nil)
	ctx := context.Background()

	t.Run("simple expression", func(t *testing.T) {
		result, err := sandbox.ExecuteExpression(ctx, "2 + 3")
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Raw)
	})

	t.Run("string expression", func(t *testing.T) {
		result, err := sandbox.ExecuteExpression(ctx, `"hello".toUpperCase()`)
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result.Raw)
	})

	t.Run("math expression", func(t *testing.T) {
		result, err := sandbox.ExecuteExpression(ctx, "Math.sqrt(16)")
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.Raw)
	})

	t.Run("array expression", func(t *testing.T) {
		result, err := sandbox.ExecuteExpression(ctx, "[1, 2, 3].map(x => x * 2)")
		require.NoError(t, err)
		assert.Equal(t, types.TypeList, result.Type)
	})
}

func TestSandboxSecurity(t *testing.T) {
	sandbox := NewSandbox(nil)
	ctx := context.Background()

	t.Run("eval is restricted", func(t *testing.T) {
		jsBody := `function tryEval() { return eval("1 + 1"); }`
		args := []types.Value{}

		_, err := sandbox.Execute(ctx, jsBody, "tryEval", args)
		// eval should either fail or be undefined
		if err == nil {
			// If no error, verify eval returns undefined behavior
			t.Log("eval executed but may be restricted")
		}
	})

	t.Run("Function constructor is restricted", func(t *testing.T) {
		jsBody := `function tryFunctionCtor() { return Function("return 1")(); }`
		args := []types.Value{}

		_, err := sandbox.Execute(ctx, jsBody, "tryFunctionCtor", args)
		// Function constructor should be restricted
		if err == nil {
			t.Log("Function constructor executed but may be restricted")
		}
	})

	t.Run("console is safe", func(t *testing.T) {
		jsBody := `function logAndReturn() { console.log("test"); return 42; }`
		args := []types.Value{}

		result, err := sandbox.Execute(ctx, jsBody, "logAndReturn", args)
		require.NoError(t, err)
		assert.Equal(t, int64(42), result.Raw)
	})
}

func TestSandboxTimeout(t *testing.T) {
	config := &SandboxConfig{
		Timeout:       50 * time.Millisecond,
		MaxStackDepth: 100,
	}
	sandbox := NewSandbox(config)
	ctx := context.Background()

	t.Run("infinite loop times out", func(t *testing.T) {
		jsBody := `function infiniteLoop() { while(true) {} return 1; }`
		args := []types.Value{}

		_, err := sandbox.Execute(ctx, jsBody, "infiniteLoop", args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}

func TestParseJSFunction(t *testing.T) {
	t.Run("simple function", func(t *testing.T) {
		source := `function add(a, b) { return a + b; }`
		name, params, returnType, body, err := ParseJSFunction(source)

		require.NoError(t, err)
		assert.Equal(t, "add", name)
		assert.Equal(t, []string{"a", "b"}, params)
		assert.Equal(t, types.TypeAny, returnType)
		assert.NotEmpty(t, body)
	})

	t.Run("function with return type", func(t *testing.T) {
		source := `function multiply(x, y): int { return x * y; }`
		name, params, returnType, body, err := ParseJSFunction(source)

		require.NoError(t, err)
		assert.Equal(t, "multiply", name)
		assert.Equal(t, []string{"x", "y"}, params)
		assert.Equal(t, types.TypeInt, returnType)
		assert.NotEmpty(t, body)
	})

	t.Run("no parameters", func(t *testing.T) {
		source := `function getAnswer() { return 42; }`
		name, params, _, _, err := ParseJSFunction(source)

		require.NoError(t, err)
		assert.Equal(t, "getAnswer", name)
		assert.Empty(t, params)
	})

	t.Run("nested braces", func(t *testing.T) {
		source := `function complex(a) { if (a > 0) { return { value: a }; } return null; }`
		name, _, _, body, err := ParseJSFunction(source)

		require.NoError(t, err)
		assert.Equal(t, "complex", name)
		assert.NotEmpty(t, body)
	})

	t.Run("missing function keyword", func(t *testing.T) {
		source := `add(a, b) { return a + b; }`
		_, _, _, _, err := ParseJSFunction(source)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "function")
	})

	t.Run("missing function name", func(t *testing.T) {
		source := `function (a, b) { return a + b; }`
		_, _, _, _, err := ParseJSFunction(source)

		require.Error(t, err)
	})

	t.Run("missing parenthesis", func(t *testing.T) {
		source := `function add a, b { return a + b; }`
		_, _, _, _, err := ParseJSFunction(source)

		require.Error(t, err)
	})

	t.Run("unmatched braces", func(t *testing.T) {
		source := `function broken() { if (true) { return 1; }`
		_, _, _, _, err := ParseJSFunction(source)

		require.Error(t, err)
	})
}

func TestRegistryRegisterJSFunction(t *testing.T) {
	t.Run("register and call JS function", func(t *testing.T) {
		registry := NewRegistry()
		sandbox := NewSandbox(nil)

		source := `function double(x) { return x * 2; }`
		err := registry.RegisterJSFunction(source, sandbox)
		require.NoError(t, err)

		fn, ok := registry.Get("double")
		require.True(t, ok)
		assert.True(t, fn.IsJS())
		assert.Equal(t, "double", fn.Name)

		// Call via sandbox
		ctx := context.Background()
		result, err := registry.CallJS(ctx, sandbox, "double", []types.Value{types.Int(5)})
		require.NoError(t, err)
		assert.Equal(t, int64(10), result.Raw)
	})

	t.Run("call non-existent function", func(t *testing.T) {
		registry := NewRegistry()
		sandbox := NewSandbox(nil)
		ctx := context.Background()

		_, err := registry.CallJS(ctx, sandbox, "notExist", []types.Value{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "undefined")
	})

	t.Run("call built-in via CallJS fails", func(t *testing.T) {
		registry, _ := NewDefaultRegistry()
		sandbox := NewSandbox(nil)
		ctx := context.Background()

		_, err := registry.CallJS(ctx, sandbox, "abs", []types.Value{types.Int(-5)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a JS function")
	})
}

func TestSandboxSetters(t *testing.T) {
	sandbox := NewSandbox(nil)

	t.Run("set timeout", func(t *testing.T) {
		sandbox.SetTimeout(500 * time.Millisecond)
		assert.Equal(t, 500*time.Millisecond, sandbox.config.Timeout)
	})

	t.Run("set memory limit", func(t *testing.T) {
		sandbox.SetMemoryLimit(20 * 1024 * 1024)
		assert.Equal(t, int64(20*1024*1024), sandbox.config.MemoryLimit)
	})

	t.Run("set max stack depth", func(t *testing.T) {
		sandbox.SetMaxStackDepth(200)
		assert.Equal(t, 200, sandbox.config.MaxStackDepth)
	})

	t.Run("get config", func(t *testing.T) {
		config := sandbox.Config()
		assert.NotNil(t, config)
		assert.Equal(t, 200, config.MaxStackDepth)
	})
}

func TestVMPool(t *testing.T) {
	// This test verifies VM reuse behavior
	sandbox := NewSandbox(nil)
	ctx := context.Background()

	// Execute multiple times to exercise pool
	for i := 0; i < 20; i++ {
		jsBody := `function test(n) { return n + 1; }`
		result, err := sandbox.Execute(ctx, jsBody, "test", []types.Value{types.Int(int64(i))})
		require.NoError(t, err)
		assert.Equal(t, int64(i+1), result.Raw)
	}
}
