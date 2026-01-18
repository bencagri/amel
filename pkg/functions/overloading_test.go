package functions

import (
	"testing"

	"github.com/bencagri/amel/pkg/types"
)

func TestRegisterOverload(t *testing.T) {
	r := NewRegistry()

	// Register first overload: add(int, int) -> int
	fn1 := &Function{
		Name: "add",
		Signature: types.NewFunctionSignature("add", types.TypeInt,
			types.Param("a", types.TypeInt),
			types.Param("b", types.TypeInt),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			a, _ := args[0].AsInt()
			b, _ := args[1].AsInt()
			return types.Int(a + b), nil
		},
	}

	err := r.RegisterOverload(fn1)
	if err != nil {
		t.Fatalf("failed to register first overload: %v", err)
	}

	// Register second overload: add(string, string) -> string
	fn2 := &Function{
		Name: "add",
		Signature: types.NewFunctionSignature("add", types.TypeString,
			types.Param("a", types.TypeString),
			types.Param("b", types.TypeString),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			a, _ := args[0].AsString()
			b, _ := args[1].AsString()
			return types.String(a + b), nil
		},
	}

	err = r.RegisterOverload(fn2)
	if err != nil {
		t.Fatalf("failed to register second overload: %v", err)
	}

	// Verify overloads are registered
	if !r.IsOverloaded("add") {
		t.Error("expected 'add' to be overloaded")
	}

	overloads := r.ListOverloads("add")
	if len(overloads) != 2 {
		t.Errorf("expected 2 overloads, got %d", len(overloads))
	}
}

func TestGetBestMatch(t *testing.T) {
	r := NewRegistry()

	// Register int version
	intFn := &Function{
		Name: "process",
		Signature: types.NewFunctionSignature("process", types.TypeInt,
			types.Param("val", types.TypeInt),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			v, _ := args[0].AsInt()
			return types.Int(v * 2), nil
		},
	}
	r.RegisterOverload(intFn)

	// Register string version
	strFn := &Function{
		Name: "process",
		Signature: types.NewFunctionSignature("process", types.TypeString,
			types.Param("val", types.TypeString),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			v, _ := args[0].AsString()
			return types.String(v + v), nil
		},
	}
	r.RegisterOverload(strFn)

	// Test with int argument
	intArgs := []types.Value{types.Int(5)}
	fn, ok := r.GetBestMatch("process", intArgs)
	if !ok {
		t.Fatal("expected to find match for int argument")
	}
	if fn.Signature.Parameters[0].Type != types.TypeInt {
		t.Error("expected int overload to be selected for int argument")
	}

	// Test with string argument
	strArgs := []types.Value{types.String("hello")}
	fn, ok = r.GetBestMatch("process", strArgs)
	if !ok {
		t.Fatal("expected to find match for string argument")
	}
	if fn.Signature.Parameters[0].Type != types.TypeString {
		t.Error("expected string overload to be selected for string argument")
	}
}

func TestCallOverloadedFunction(t *testing.T) {
	r := NewRegistry()

	// Register int version
	r.RegisterOverload(&Function{
		Name: "double",
		Signature: types.NewFunctionSignature("double", types.TypeInt,
			types.Param("val", types.TypeInt),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			v, _ := args[0].AsInt()
			return types.Int(v * 2), nil
		},
	})

	// Register float version
	r.RegisterOverload(&Function{
		Name: "double",
		Signature: types.NewFunctionSignature("double", types.TypeFloat,
			types.Param("val", types.TypeFloat),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			v, _ := args[0].AsFloat()
			return types.Float(v * 2), nil
		},
	})

	// Call with int
	result, err := r.Call("double", types.Int(5))
	if err != nil {
		t.Fatalf("failed to call double with int: %v", err)
	}
	if result.Raw != int64(10) {
		t.Errorf("expected 10, got %v", result.Raw)
	}

	// Call with float
	result, err = r.Call("double", types.Float(3.5))
	if err != nil {
		t.Fatalf("failed to call double with float: %v", err)
	}
	if result.Raw != 7.0 {
		t.Errorf("expected 7.0, got %v", result.Raw)
	}
}

func TestDuplicateSignatureError(t *testing.T) {
	r := NewRegistry()

	fn1 := &Function{
		Name: "test",
		Signature: types.NewFunctionSignature("test", types.TypeInt,
			types.Param("a", types.TypeInt),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			return types.Int(1), nil
		},
	}

	err := r.RegisterOverload(fn1)
	if err != nil {
		t.Fatalf("failed to register first function: %v", err)
	}

	// Try to register same signature again
	fn2 := &Function{
		Name: "test",
		Signature: types.NewFunctionSignature("test", types.TypeInt,
			types.Param("b", types.TypeInt), // Same type, different param name
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			return types.Int(2), nil
		},
	}

	err = r.RegisterOverload(fn2)
	if err == nil {
		t.Error("expected error when registering duplicate signature")
	}
}

func TestConvertRegularToOverloaded(t *testing.T) {
	r := NewRegistry()

	// Register a regular function first
	err := r.RegisterBuiltIn("myFunc", func(args ...types.Value) (types.Value, error) {
		return types.Int(1), nil
	}, types.NewFunctionSignature("myFunc", types.TypeInt, types.Param("a", types.TypeInt)))
	if err != nil {
		t.Fatalf("failed to register regular function: %v", err)
	}

	// Now add an overload
	err = r.RegisterOverload(&Function{
		Name: "myFunc",
		Signature: types.NewFunctionSignature("myFunc", types.TypeString,
			types.Param("a", types.TypeString),
		),
		BuiltIn: func(args ...types.Value) (types.Value, error) {
			return types.String("hello"), nil
		},
	})
	if err != nil {
		t.Fatalf("failed to register overload: %v", err)
	}

	// Should now be overloaded
	if !r.IsOverloaded("myFunc") {
		t.Error("expected function to be overloaded after adding overload")
	}

	overloads := r.ListOverloads("myFunc")
	if len(overloads) != 2 {
		t.Errorf("expected 2 overloads, got %d", len(overloads))
	}
}

func TestSignaturesMatch(t *testing.T) {
	tests := []struct {
		name     string
		a        *types.FunctionSignature
		b        *types.FunctionSignature
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        types.NewFunctionSignature("f", types.TypeInt),
			b:        nil,
			expected: false,
		},
		{
			name:     "same signature no params",
			a:        types.NewFunctionSignature("f", types.TypeInt),
			b:        types.NewFunctionSignature("f", types.TypeInt),
			expected: true,
		},
		{
			name:     "same signature with params",
			a:        types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			b:        types.NewFunctionSignature("f", types.TypeInt, types.Param("b", types.TypeInt)),
			expected: true,
		},
		{
			name:     "different param types",
			a:        types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			b:        types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeString)),
			expected: false,
		},
		{
			name:     "different param count",
			a:        types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			b:        types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt), types.Param("b", types.TypeInt)),
			expected: false,
		},
		{
			name:     "variadic vs non-variadic",
			a:        types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			b:        types.NewVariadicSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := signaturesMatch(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("signaturesMatch() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchScore(t *testing.T) {
	tests := []struct {
		name          string
		sig           *types.FunctionSignature
		args          []types.Value
		expectedScore int
	}{
		{
			name:          "nil signature",
			sig:           nil,
			args:          []types.Value{types.Int(1)},
			expectedScore: 0,
		},
		{
			name:          "exact match",
			sig:           types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			args:          []types.Value{types.Int(1)},
			expectedScore: 10,
		},
		{
			name:          "compatible types (int to float)",
			sig:           types.NewFunctionSignature("f", types.TypeFloat, types.Param("a", types.TypeFloat)),
			args:          []types.Value{types.Int(1)},
			expectedScore: 5,
		},
		{
			name:          "any type accepts anything",
			sig:           types.NewFunctionSignature("f", types.TypeAny, types.Param("a", types.TypeAny)),
			args:          []types.Value{types.String("hello")},
			expectedScore: 1,
		},
		{
			name:          "not enough args",
			sig:           types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt), types.Param("b", types.TypeInt)),
			args:          []types.Value{types.Int(1)},
			expectedScore: -1,
		},
		{
			name:          "too many args",
			sig:           types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			args:          []types.Value{types.Int(1), types.Int(2)},
			expectedScore: -1,
		},
		{
			name:          "incompatible type",
			sig:           types.NewFunctionSignature("f", types.TypeInt, types.Param("a", types.TypeInt)),
			args:          []types.Value{types.String("hello")},
			expectedScore: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := matchScore(tt.sig, tt.args)
			if score != tt.expectedScore {
				t.Errorf("matchScore() = %v, want %v", score, tt.expectedScore)
			}
		})
	}
}

func TestOverloadedFunctionCount(t *testing.T) {
	r := NewRegistry()

	// Register regular function
	r.RegisterBuiltIn("regular", func(args ...types.Value) (types.Value, error) {
		return types.Int(1), nil
	}, nil)

	// Register overloaded function with 2 overloads
	r.RegisterOverload(&Function{
		Name:      "overloaded",
		Signature: types.NewFunctionSignature("overloaded", types.TypeInt, types.Param("a", types.TypeInt)),
		BuiltIn:   func(args ...types.Value) (types.Value, error) { return types.Int(1), nil },
	})
	r.RegisterOverload(&Function{
		Name:      "overloaded",
		Signature: types.NewFunctionSignature("overloaded", types.TypeString, types.Param("a", types.TypeString)),
		BuiltIn:   func(args ...types.Value) (types.Value, error) { return types.String("a"), nil },
	})

	// Count should include all overloads
	if r.Count() != 3 {
		t.Errorf("Count() = %d, want 3", r.Count())
	}

	// CountUnique should count unique names
	if r.CountUnique() != 2 {
		t.Errorf("CountUnique() = %d, want 2", r.CountUnique())
	}
}

func TestUnregisterOverloaded(t *testing.T) {
	r := NewRegistry()

	r.RegisterOverload(&Function{
		Name:      "test",
		Signature: types.NewFunctionSignature("test", types.TypeInt, types.Param("a", types.TypeInt)),
		BuiltIn:   func(args ...types.Value) (types.Value, error) { return types.Int(1), nil },
	})
	r.RegisterOverload(&Function{
		Name:      "test",
		Signature: types.NewFunctionSignature("test", types.TypeString, types.Param("a", types.TypeString)),
		BuiltIn:   func(args ...types.Value) (types.Value, error) { return types.String("a"), nil },
	})

	if !r.Has("test") {
		t.Error("expected function to exist before unregister")
	}

	removed := r.Unregister("test")
	if !removed {
		t.Error("expected Unregister to return true")
	}

	if r.Has("test") {
		t.Error("expected function to be removed after unregister")
	}
}
