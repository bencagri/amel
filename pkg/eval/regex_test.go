package eval

import (
	"testing"

	"github.com/bencagri/amel/pkg/parser"
)

func TestRegexExpression(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected bool
	}{
		{
			name:     "simple match",
			dsl:      `"hello" =~ "^hel"`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "simple no match",
			dsl:      `"hello" =~ "^world"`,
			payload:  nil,
			expected: false,
		},
		{
			name:     "negated match",
			dsl:      `"hello" !~ "^world"`,
			payload:  nil,
			expected: true,
		},
		{
			name:     "negated no match",
			dsl:      `"hello" !~ "^hel"`,
			payload:  nil,
			expected: false,
		},
		{
			name:     "match with json path",
			dsl:      `$.name =~ "^John"`,
			payload:  map[string]interface{}{"name": "John Doe"},
			expected: true,
		},
		{
			name:     "match email pattern",
			dsl:      `$.email =~ "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`,
			payload:  map[string]interface{}{"email": "test@example.com"},
			expected: true,
		},
		{
			name:     "invalid email pattern",
			dsl:      `$.email =~ "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`,
			payload:  map[string]interface{}{"email": "invalid-email"},
			expected: false,
		},
		{
			name:     "match phone number",
			dsl:      `$.phone =~ "^\\+?[0-9]{10,14}$"`,
			payload:  map[string]interface{}{"phone": "+12345678901"},
			expected: true,
		},
		{
			name:     "case insensitive match",
			dsl:      `$.status =~ "(?i)^active$"`,
			payload:  map[string]interface{}{"status": "ACTIVE"},
			expected: true,
		},
		{
			name:     "match with word boundary",
			dsl:      `$.text =~ "\\bworld\\b"`,
			payload:  map[string]interface{}{"text": "hello world foo"},
			expected: true,
		},
		{
			name:     "no match with word boundary",
			dsl:      `$.text =~ "\\bworld\\b"`,
			payload:  map[string]interface{}{"text": "helloworld"},
			expected: false,
		},
		{
			name:     "match digits only",
			dsl:      `$.code =~ "^[0-9]+$"`,
			payload:  map[string]interface{}{"code": "12345"},
			expected: true,
		},
		{
			name:     "no match digits only",
			dsl:      `$.code =~ "^[0-9]+$"`,
			payload:  map[string]interface{}{"code": "123abc"},
			expected: false,
		},
		{
			name:     "null value with match",
			dsl:      `$.missing =~ ".*"`,
			payload:  map[string]interface{}{"other": "value"},
			expected: false,
		},
		{
			name:     "null value with not match",
			dsl:      `$.missing !~ ".*"`,
			payload:  map[string]interface{}{"other": "value"},
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

func TestRegexExpressionErrors(t *testing.T) {
	tests := []struct {
		name    string
		dsl     string
		payload map[string]interface{}
		wantErr bool
	}{
		{
			name:    "invalid regex pattern",
			dsl:     `"hello" =~ "["`,
			payload: nil,
			wantErr: true,
		},
		{
			name:    "non-string left operand",
			dsl:     `123 =~ "^[0-9]+"`,
			payload: nil,
			wantErr: true,
		},
		{
			name:    "non-string pattern",
			dsl:     `"hello" =~ 123`,
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

func TestRegexExpressionWithExplanation(t *testing.T) {
	evaluator, err := New()
	if err != nil {
		t.Fatalf("failed to create evaluator: %v", err)
	}

	expr, err := parser.Parse(`$.name =~ "^John"`)
	if err != nil {
		t.Fatalf("failed to parse DSL: %v", err)
	}

	ctx, err := NewContext(map[string]interface{}{"name": "John Doe"})
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	result, explanation, err := evaluator.EvaluateWithExplanation(expr, ctx)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if result.Raw != true {
		t.Errorf("expected true, got %v", result.Raw)
	}

	if explanation == nil {
		t.Fatal("expected explanation, got nil")
	}

	if len(explanation.Children) != 2 {
		t.Errorf("expected 2 children in explanation, got %d", len(explanation.Children))
	}
}

func TestRegexInComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		dsl      string
		payload  map[string]interface{}
		expected bool
	}{
		{
			name:     "regex with AND",
			dsl:      `$.name =~ "^John" && $.age > 18`,
			payload:  map[string]interface{}{"name": "John Doe", "age": 25},
			expected: true,
		},
		{
			name:     "regex with OR",
			dsl:      `$.name =~ "^John" || $.name =~ "^Jane"`,
			payload:  map[string]interface{}{"name": "Jane Smith"},
			expected: true,
		},
		{
			name:     "regex with NOT",
			dsl:      `!($.email =~ "@spam\\.com$")`,
			payload:  map[string]interface{}{"email": "test@example.com"},
			expected: true,
		},
		{
			name:     "multiple regex conditions",
			dsl:      `$.email =~ "@example\\.com$" && $.phone =~ "^\\+1"`,
			payload:  map[string]interface{}{"email": "test@example.com", "phone": "+1234567890"},
			expected: true,
		},
		{
			name:     "regex with IN",
			dsl:      `$.status =~ "^(active|pending)$" && $.role IN ["admin", "user"]`,
			payload:  map[string]interface{}{"status": "active", "role": "admin"},
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

func TestLambdaParsing(t *testing.T) {
	tests := []struct {
		name    string
		dsl     string
		wantErr bool
	}{
		{
			name:    "simple lambda",
			dsl:     `x => x * 2`,
			wantErr: false,
		},
		{
			name:    "lambda with comparison",
			dsl:     `x => x > 5`,
			wantErr: false,
		},
		{
			name:    "lambda with complex expression",
			dsl:     `x => x * 2 + 1`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && expr == nil {
				t.Error("expected non-nil expression")
			}
		})
	}
}
