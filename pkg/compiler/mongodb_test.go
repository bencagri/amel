package compiler

import (
	"encoding/json"
	"testing"

	"github.com/bencagri/amel/pkg/parser"
)

func TestMongoDBCompiler_Basic(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "simple equality",
			dsl:  `$.name == "John"`,
			expectedQuery: map[string]interface{}{
				"name": "John",
			},
		},
		{
			name: "not equal",
			dsl:  `$.status != "deleted"`,
			expectedQuery: map[string]interface{}{
				"status": map[string]interface{}{"$ne": "deleted"},
			},
		},
		{
			name: "greater than",
			dsl:  `$.age > 18`,
			expectedQuery: map[string]interface{}{
				"age": map[string]interface{}{"$gt": int64(18)},
			},
		},
		{
			name: "less than",
			dsl:  `$.price < 100`,
			expectedQuery: map[string]interface{}{
				"price": map[string]interface{}{"$lt": int64(100)},
			},
		},
		{
			name: "greater than or equal",
			dsl:  `$.quantity >= 5`,
			expectedQuery: map[string]interface{}{
				"quantity": map[string]interface{}{"$gte": int64(5)},
			},
		},
		{
			name: "less than or equal",
			dsl:  `$.score <= 100`,
			expectedQuery: map[string]interface{}{
				"score": map[string]interface{}{"$lte": int64(100)},
			},
		},
		{
			name: "null comparison",
			dsl:  `$.deleted_at == null`,
			expectedQuery: map[string]interface{}{
				"deleted_at": nil,
			},
		},
		{
			name: "not null comparison",
			dsl:  `$.email != null`,
			expectedQuery: map[string]interface{}{
				"email": map[string]interface{}{"$ne": nil},
			},
		},
		{
			name: "boolean true",
			dsl:  `$.active == true`,
			expectedQuery: map[string]interface{}{
				"active": true,
			},
		},
		{
			name: "boolean false",
			dsl:  `$.archived == false`,
			expectedQuery: map[string]interface{}{
				"archived": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_LogicalOperators(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "and expression",
			dsl:  `$.age > 18 && $.status == "active"`,
			expectedQuery: map[string]interface{}{
				"$and": []interface{}{
					map[string]interface{}{"age": map[string]interface{}{"$gt": int64(18)}},
					map[string]interface{}{"status": "active"},
				},
			},
		},
		{
			name: "or expression",
			dsl:  `$.role == "admin" || $.role == "superuser"`,
			expectedQuery: map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{"role": "admin"},
					map[string]interface{}{"role": "superuser"},
				},
			},
		},
		{
			name: "complex and/or",
			dsl:  `($.age >= 18 && $.age <= 65) || $.role == "admin"`,
			expectedQuery: map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{
						"$and": []interface{}{
							map[string]interface{}{"age": map[string]interface{}{"$gte": int64(18)}},
							map[string]interface{}{"age": map[string]interface{}{"$lte": int64(65)}},
						},
					},
					map[string]interface{}{"role": "admin"},
				},
			},
		},
		{
			name: "triple and",
			dsl:  `$.a == 1 && $.b == 2 && $.c == 3`,
			expectedQuery: map[string]interface{}{
				"$and": []interface{}{
					map[string]interface{}{"a": int64(1)},
					map[string]interface{}{"b": int64(2)},
					map[string]interface{}{"c": int64(3)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_InExpression(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "in expression",
			dsl:  `$.status IN ["active", "pending"]`,
			expectedQuery: map[string]interface{}{
				"status": map[string]interface{}{
					"$in": []interface{}{"active", "pending"},
				},
			},
		},
		{
			name: "not in expression",
			dsl:  `$.status NOT IN ["deleted", "archived"]`,
			expectedQuery: map[string]interface{}{
				"status": map[string]interface{}{
					"$nin": []interface{}{"deleted", "archived"},
				},
			},
		},
		{
			name: "in with integers",
			dsl:  `$.priority IN [1, 2, 3]`,
			expectedQuery: map[string]interface{}{
				"priority": map[string]interface{}{
					"$in": []interface{}{int64(1), int64(2), int64(3)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_RegexExpression(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "regex match",
			dsl:  `$.email =~ "@gmail.com$"`,
			expectedQuery: map[string]interface{}{
				"email": map[string]interface{}{
					"$regex": "@gmail.com$",
				},
			},
		},
		{
			name: "regex not match",
			dsl:  `$.email !~ "@spam.com$"`,
			expectedQuery: map[string]interface{}{
				"email": map[string]interface{}{
					"$not": map[string]interface{}{
						"$regex": "@spam.com$",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_Functions(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "isNull function",
			dsl:  `isNull($.deleted_at)`,
			expectedQuery: map[string]interface{}{
				"deleted_at": nil,
			},
		},
		{
			name: "isNotNull function",
			dsl:  `isNotNull($.email)`,
			expectedQuery: map[string]interface{}{
				"email": map[string]interface{}{"$ne": nil},
			},
		},
		{
			name: "contains function",
			dsl:  `contains($.name, "john")`,
			expectedQuery: map[string]interface{}{
				"name": map[string]interface{}{"$regex": "john"},
			},
		},
		{
			name: "startsWith function",
			dsl:  `startsWith($.email, "admin")`,
			expectedQuery: map[string]interface{}{
				"email": map[string]interface{}{"$regex": "^admin"},
			},
		},
		{
			name: "endsWith function",
			dsl:  `endsWith($.email, ".com")`,
			expectedQuery: map[string]interface{}{
				"email": map[string]interface{}{"$regex": "\\.com$"},
			},
		},
		{
			name: "exists function",
			dsl:  `exists($.metadata)`,
			expectedQuery: map[string]interface{}{
				"metadata": map[string]interface{}{"$exists": true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_Negation(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "negated equality",
			dsl:  `!($.status == "active")`,
			expectedQuery: map[string]interface{}{
				"status": map[string]interface{}{"$ne": "active"},
			},
		},
		{
			name: "negated complex expression",
			dsl:  `!($.age > 18 && $.status == "active")`,
			expectedQuery: map[string]interface{}{
				"$nor": []interface{}{
					map[string]interface{}{
						"$and": []interface{}{
							map[string]interface{}{"age": map[string]interface{}{"$gt": int64(18)}},
							map[string]interface{}{"status": "active"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_NestedFields(t *testing.T) {
	tests := []struct {
		name          string
		dsl           string
		expectedQuery map[string]interface{}
	}{
		{
			name: "nested field access",
			dsl:  `$.user.name == "John"`,
			expectedQuery: map[string]interface{}{
				"user.name": "John",
			},
		},
		{
			name: "deeply nested field",
			dsl:  `$.user.profile.settings.theme == "dark"`,
			expectedQuery: map[string]interface{}{
				"user.profile.settings.theme": "dark",
			},
		},
		{
			name: "array index access",
			dsl:  `$.items[0].name == "first"`,
			expectedQuery: map[string]interface{}{
				"items.0.name": "first",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			assertJSONEqual(t, tt.expectedQuery, result.Query)
		})
	}
}

func TestMongoDBCompiler_CustomFieldMapper(t *testing.T) {
	mapper := func(path string) string {
		// Custom mapping: $.user.firstName -> users.first_name
		if path == "$.user.firstName" {
			return "users.first_name"
		}
		return defaultMongoFieldMapper(path)
	}

	expr, err := parser.Parse(`$.user.firstName == "John"`)
	if err != nil {
		t.Fatalf("failed to parse DSL: %v", err)
	}

	compiler := NewMongoDBCompiler(WithMongoFieldMapper(mapper))
	result, err := compiler.Compile(expr)
	if err != nil {
		t.Fatalf("failed to compile: %v", err)
	}

	expected := map[string]interface{}{
		"users.first_name": "John",
	}
	assertJSONEqual(t, expected, result.Query)
}

func TestMongoDBCompiler_ToJSON(t *testing.T) {
	expr, err := parser.Parse(`$.age > 18 && $.status == "active"`)
	if err != nil {
		t.Fatalf("failed to parse DSL: %v", err)
	}

	compiler := NewMongoDBCompiler()
	result, err := compiler.Compile(expr)
	if err != nil {
		t.Fatalf("failed to compile: %v", err)
	}

	jsonStr, err := result.ToJSON()
	if err != nil {
		t.Fatalf("failed to convert to JSON: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Verify structure
	if _, ok := parsed["$and"]; !ok {
		t.Error("expected $and in output")
	}
}

func TestMongoDBCompiler_ToPrettyJSON(t *testing.T) {
	expr, err := parser.Parse(`$.name == "John"`)
	if err != nil {
		t.Fatalf("failed to parse DSL: %v", err)
	}

	compiler := NewMongoDBCompiler()
	result, err := compiler.Compile(expr)
	if err != nil {
		t.Fatalf("failed to compile: %v", err)
	}

	prettyJSON, err := result.ToPrettyJSON()
	if err != nil {
		t.Fatalf("failed to convert to pretty JSON: %v", err)
	}

	// Should contain newlines (formatted)
	if len(prettyJSON) == 0 {
		t.Error("expected non-empty pretty JSON")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(prettyJSON), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
}

func TestMongoDBCompiler_Errors(t *testing.T) {
	tests := []struct {
		name    string
		dsl     string
		wantErr bool
	}{
		{
			name:    "unsupported function",
			dsl:     `customFunc($.name)`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				if tt.wantErr {
					return // Parse error is acceptable
				}
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			_, err = compiler.Compile(expr)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestDefaultMongoFieldMapper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$.user.name", "user.name"},
		{"$.data[0].value", "data.0.value"},
		{"$.items[0][1].name", "items.0.1.name"},
		{"$.simple", "simple"},
		{"$", ""},
		{"$.a.b.c.d", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := defaultMongoFieldMapper(tt.input)
			if result != tt.expected {
				t.Errorf("defaultMongoFieldMapper(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompileToMongoDB_Convenience(t *testing.T) {
	expr, err := parser.Parse(`$.age > 18`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	result, err := CompileToMongoDB(expr)
	if err != nil {
		t.Fatalf("failed to compile: %v", err)
	}

	expected := map[string]interface{}{
		"age": map[string]interface{}{"$gt": int64(18)},
	}
	assertJSONEqual(t, expected, result.Query)
}

func TestMongoDBCompiler_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name string
		dsl  string
	}{
		{
			name: "access control expression",
			dsl:  `$.user.role == "admin" || ($.user.role == "manager" && $.user.department == "sales")`,
		},
		{
			name: "age range with status",
			dsl:  `$.age >= 18 && $.age <= 65 && $.status IN ["active", "pending"]`,
		},
		{
			name: "multiple or conditions",
			dsl:  `$.type == "A" || $.type == "B" || $.type == "C"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewMongoDBCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			// Verify it produces valid JSON
			_, err = result.ToJSON()
			if err != nil {
				t.Fatalf("failed to convert to JSON: %v", err)
			}
		})
	}
}

// Helper function to compare JSON structures
func assertJSONEqual(t *testing.T, expected, actual map[string]interface{}) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected: %v", err)
	}

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("failed to marshal actual: %v", err)
	}

	// Re-unmarshal for consistent comparison
	var expectedNorm, actualNorm interface{}
	json.Unmarshal(expectedJSON, &expectedNorm)
	json.Unmarshal(actualJSON, &actualNorm)

	expectedNormJSON, _ := json.Marshal(expectedNorm)
	actualNormJSON, _ := json.Marshal(actualNorm)

	if string(expectedNormJSON) != string(actualNormJSON) {
		t.Errorf("JSON mismatch:\nexpected: %s\nactual:   %s", string(expectedNormJSON), string(actualNormJSON))
	}
}
