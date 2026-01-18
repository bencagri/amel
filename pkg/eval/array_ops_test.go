package eval

import (
	"testing"

	"github.com/bencagri/amel/pkg/parser"
)

func TestMapFunction(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected interface{}
	}{
		{
			name:     "map with lambda - double values",
			dsl:      `map([1, 2, 3], x => x * 2)`,
			payload:  nil,
			expected: []interface{}{int64(2), int64(4), int64(6)},
		},
		{
			name:     "map with lambda - increment",
			dsl:      `map([1, 2, 3], x => x + 1)`,
			payload:  nil,
			expected: []interface{}{int64(2), int64(3), int64(4)},
		},
		{
			name:     "map with json path",
			dsl:      `map($.numbers, x => x * 2)`,
			payload:  map[string]interface{}{"numbers": []interface{}{1, 2, 3}},
			expected: []interface{}{int64(2), int64(4), int64(6)},
		},
		{
			name:     "map empty list",
			dsl:      `map([], x => x * 2)`,
			payload:  nil,
			expected: []interface{}{},
		},
		{
			name:     "map with comparison",
			dsl:      `map([1, 2, 3], x => x > 1)`,
			payload:  nil,
			expected: []interface{}{false, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.Evaluate(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			list, ok := result.AsList()
			if !ok {
				t.Fatalf("expected list result, got %s", result.Type)
			}

			expectedList := tt.expected.([]interface{})
			if len(list) != len(expectedList) {
				t.Fatalf("expected %d elements, got %d", len(expectedList), len(list))
			}

			for i, v := range list {
				if v.Raw != expectedList[i] {
					t.Errorf("element %d: expected %v, got %v", i, expectedList[i], v.Raw)
				}
			}
		})
	}
}

func TestFilterFunction(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected []interface{}
	}{
		{
			name:     "filter greater than 2",
			dsl:      `filter([1, 2, 3, 4, 5], x => x > 2)`,
			payload:  nil,
			expected: []interface{}{int64(3), int64(4), int64(5)},
		},
		{
			name:     "filter even numbers",
			dsl:      `filter([1, 2, 3, 4, 5, 6], x => x % 2 == 0)`,
			payload:  nil,
			expected: []interface{}{int64(2), int64(4), int64(6)},
		},
		{
			name:     "filter with json path",
			dsl:      `filter($.ages, x => x >= 18)`,
			payload:  map[string]interface{}{"ages": []interface{}{10, 15, 18, 21, 25}},
			expected: []interface{}{int64(18), int64(21), int64(25)},
		},
		{
			name:     "filter empty list",
			dsl:      `filter([], x => x > 0)`,
			payload:  nil,
			expected: []interface{}{},
		},
		{
			name:     "filter none match",
			dsl:      `filter([1, 2, 3], x => x > 10)`,
			payload:  nil,
			expected: []interface{}{},
		},
		{
			name:     "filter all match",
			dsl:      `filter([1, 2, 3], x => x > 0)`,
			payload:  nil,
			expected: []interface{}{int64(1), int64(2), int64(3)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.Evaluate(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			list, ok := result.AsList()
			if !ok {
				t.Fatalf("expected list result, got %s", result.Type)
			}

			if len(list) != len(tt.expected) {
				t.Fatalf("expected %d elements, got %d", len(tt.expected), len(list))
			}

			for i, v := range list {
				if v.Raw != tt.expected[i] {
					t.Errorf("element %d: expected %v, got %v", i, tt.expected[i], v.Raw)
				}
			}
		})
	}
}

func TestReduceFunction(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected interface{}
	}{
		{
			name:     "reduce sum",
			dsl:      `reduce([1, 2, 3, 4, 5], 0, (acc, x) => acc + x)`,
			payload:  nil,
			expected: int64(15),
		},
		{
			name:     "reduce product",
			dsl:      `reduce([1, 2, 3, 4], 1, (acc, x) => acc * x)`,
			payload:  nil,
			expected: int64(24),
		},
		{
			name:     "reduce with json path",
			dsl:      `reduce($.prices, 0, (acc, x) => acc + x)`,
			payload:  map[string]interface{}{"prices": []interface{}{10, 20, 30}},
			expected: int64(60),
		},
		{
			name:     "reduce empty list returns initial",
			dsl:      `reduce([], 100, (acc, x) => acc + x)`,
			payload:  nil,
			expected: int64(100),
		},
		{
			name:     "reduce single element",
			dsl:      `reduce([5], 10, (acc, x) => acc + x)`,
			payload:  nil,
			expected: int64(15),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.Evaluate(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			if result.Raw != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result.Raw)
			}
		})
	}
}

func TestFindFunction(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected interface{}
		isNull   bool
	}{
		{
			name:     "find first greater than 3",
			dsl:      `find([1, 2, 3, 4, 5], x => x > 3)`,
			payload:  nil,
			expected: int64(4),
		},
		{
			name:     "find first even",
			dsl:      `find([1, 3, 5, 6, 7], x => x % 2 == 0)`,
			payload:  nil,
			expected: int64(6),
		},
		{
			name:     "find with json path",
			dsl:      `find($.scores, x => x >= 90)`,
			payload:  map[string]interface{}{"scores": []interface{}{70, 85, 95, 100}},
			expected: int64(95),
		},
		{
			name:     "find no match returns null",
			dsl:      `find([1, 2, 3], x => x > 10)`,
			payload:  nil,
			expected: nil,
			isNull:   true,
		},
		{
			name:     "find in empty list returns null",
			dsl:      `find([], x => x > 0)`,
			payload:  nil,
			expected: nil,
			isNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.Evaluate(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			if tt.isNull {
				if !result.IsNull() {
					t.Errorf("expected null, got %v", result.Raw)
				}
			} else {
				if result.Raw != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, result.Raw)
				}
			}
		})
	}
}

func TestSomeFunction(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected bool
	}{
		{
			name:     "some match exists",
			dsl:      `some([1, 2, 3, 4, 5], x => x > 3)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "some no match",
			dsl:      `some([1, 2, 3], x => x > 10)`,
			payload:  nil,
			expected: false,
		},
		{
			name:     "some with json path",
			dsl:      `some($.ages, x => x >= 18)`,
			payload:  map[string]interface{}{"ages": []interface{}{10, 15, 18}},
			expected: true,
		},
		{
			name:     "some empty list",
			dsl:      `some([], x => x > 0)`,
			payload:  nil,
			expected: false,
		},
		{
			name:     "some first element matches",
			dsl:      `some([10, 1, 2], x => x > 5)`,
			payload:  nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.EvaluateBool(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEveryFunction(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected bool
	}{
		{
			name:     "every all match",
			dsl:      `every([2, 4, 6, 8], x => x % 2 == 0)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "every not all match",
			dsl:      `every([2, 4, 5, 8], x => x % 2 == 0)`,
			payload:  nil,
			expected: false,
		},
		{
			name:     "every with json path",
			dsl:      `every($.ages, x => x >= 18)`,
			payload:  map[string]interface{}{"ages": []interface{}{18, 21, 25, 30}},
			expected: true,
		},
		{
			name:     "every empty list returns true",
			dsl:      `every([], x => x > 0)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "every single element matches",
			dsl:      `every([5], x => x > 0)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "every single element does not match",
			dsl:      `every([5], x => x > 10)`,
			payload:  nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.EvaluateBool(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestArrayOpsErrors(t *testing.T) {
	tests := []struct {
		name    string
		dsl     string
		payload map[string]interface{}
		wantErr bool
	}{
		{
			name:    "map with non-list",
			dsl:     `map(123, x => x * 2)`,
			payload: nil,
			wantErr: true,
		},
		{
			name:    "filter with non-list",
			dsl:     `filter("hello", x => x > 0)`,
			payload: nil,
			wantErr: true,
		},
		{
			name:    "reduce with non-list",
			dsl:     `reduce(123, 0, (acc, x) => acc + x)`,
			payload: nil,
			wantErr: true,
		},
		{
			name:    "map missing arguments",
			dsl:     `map([1, 2, 3])`,
			payload: nil,
			wantErr: true,
		},
		{
			name:    "reduce missing initial value",
			dsl:     `reduce([1, 2, 3], (acc, x) => acc + x)`,
			payload: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				// Parse errors are acceptable for some tests
				if tt.wantErr {
					return
				}
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			_, err = evaluator.Evaluate(expr, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got error: %v", tt.wantErr, err)
			}
		})
	}
}

func TestArrayOpsWithComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected interface{}
	}{
		{
			name:     "chained map and filter",
			dsl:      `filter(map([1, 2, 3, 4], x => x * 2), x => x > 4)`,
			payload:  nil,
			expected: []interface{}{int64(6), int64(8)},
		},
		{
			name:     "map with arithmetic in body",
			dsl:      `map([1, 2, 3], x => x * x + 1)`,
			payload:  nil,
			expected: []interface{}{int64(2), int64(5), int64(10)},
		},
		{
			name:     "filter with AND condition",
			dsl:      `filter([1, 2, 3, 4, 5, 6], x => x > 2 && x < 6)`,
			payload:  nil,
			expected: []interface{}{int64(3), int64(4), int64(5)},
		},
		{
			name:     "reduce to find max",
			dsl:      `reduce([3, 1, 4, 1, 5, 9, 2, 6], 0, (acc, x) => ifThenElse(x > acc, x, acc))`,
			payload:  nil,
			expected: int64(9),
		},
		{
			name:     "some with complex condition",
			dsl:      `some([10, 20, 30], x => x > 15 && x < 25)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "every with json path variable access",
			dsl:      `every($.items, x => x > $.threshold)`,
			payload:  map[string]interface{}{"items": []interface{}{10, 20, 30}, "threshold": 5},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.Evaluate(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			switch expected := tt.expected.(type) {
			case []interface{}:
				list, ok := result.AsList()
				if !ok {
					t.Fatalf("expected list result, got %s", result.Type)
				}
				if len(list) != len(expected) {
					t.Fatalf("expected %d elements, got %d", len(expected), len(list))
				}
				for i, v := range list {
					if v.Raw != expected[i] {
						t.Errorf("element %d: expected %v, got %v", i, expected[i], v.Raw)
					}
				}
			case bool:
				if result.Raw != expected {
					t.Errorf("expected %v, got %v", expected, result.Raw)
				}
			default:
				if result.Raw != expected {
					t.Errorf("expected %v, got %v", expected, result.Raw)
				}
			}
		})
	}
}

func TestArrayOpsWithDefaultParameter(t *testing.T) {
	// Test using the default "x" parameter when not using lambda syntax
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected interface{}
	}{
		{
			name:     "map with expression using default x",
			dsl:      `map([1, 2, 3], x * 2)`,
			payload:  nil,
			expected: []interface{}{int64(2), int64(4), int64(6)},
		},
		{
			name:     "filter with expression using default x",
			dsl:      `filter([1, 2, 3, 4, 5], x > 2)`,
			payload:  nil,
			expected: []interface{}{int64(3), int64(4), int64(5)},
		},
		{
			name:     "some with expression using default x",
			dsl:      `some([1, 2, 3], x > 2)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "every with expression using default x",
			dsl:      `every([2, 4, 6], x % 2 == 0)`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "find with expression using default x",
			dsl:      `find([1, 2, 3, 4], x > 2)`,
			payload:  nil,
			expected: int64(3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator, err := New()
			if err != nil {
				t.Fatalf("failed to create evaluator: %v", err)
			}

			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			ctx, err := NewContext(tt.payload)
			if err != nil {
				t.Fatalf("failed to create context: %v", err)
			}

			result, err := evaluator.Evaluate(expr, ctx)
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}

			switch expected := tt.expected.(type) {
			case []interface{}:
				list, ok := result.AsList()
				if !ok {
					t.Fatalf("expected list result, got %s", result.Type)
				}
				if len(list) != len(expected) {
					t.Fatalf("expected %d elements, got %d", len(expected), len(list))
				}
				for i, v := range list {
					if v.Raw != expected[i] {
						t.Errorf("element %d: expected %v, got %v", i, expected[i], v.Raw)
					}
				}
			case bool:
				if result.Raw != expected {
					t.Errorf("expected %v, got %v", expected, result.Raw)
				}
			default:
				if result.Raw != expected {
					t.Errorf("expected %v, got %v", expected, result.Raw)
				}
			}
		})
	}
}
