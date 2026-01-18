package compiler

import (
	"testing"

	"github.com/bencagri/amel/pkg/parser"
)

func TestSQLCompiler_Basic(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
		paramCount  int
	}{
		{
			name:        "simple comparison",
			dsl:         `$.age > 18`,
			expectedSQL: `("age" > ?)`,
			paramCount:  1,
		},
		{
			name:        "equality comparison",
			dsl:         `$.status == "active"`,
			expectedSQL: `("status" = ?)`,
			paramCount:  1,
		},
		{
			name:        "not equal comparison",
			dsl:         `$.status != "inactive"`,
			expectedSQL: `("status" <> ?)`,
			paramCount:  1,
		},
		{
			name:        "and expression",
			dsl:         `$.age > 18 && $.status == "active"`,
			expectedSQL: `(("age" > ?) AND ("status" = ?))`,
			paramCount:  2,
		},
		{
			name:        "or expression",
			dsl:         `$.role == "admin" || $.role == "superuser"`,
			expectedSQL: `(("role" = ?) OR ("role" = ?))`,
			paramCount:  2,
		},
		{
			name:        "in expression",
			dsl:         `$.status IN ["active", "pending"]`,
			expectedSQL: `"status" IN (?, ?)`,
			paramCount:  2,
		},
		{
			name:        "not in expression",
			dsl:         `$.status NOT IN ["deleted", "archived"]`,
			expectedSQL: `"status" NOT IN (?, ?)`,
			paramCount:  2,
		},
		{
			name:        "null comparison",
			dsl:         `$.deleted_at == null`,
			expectedSQL: `"deleted_at" IS NULL`,
			paramCount:  0,
		},
		{
			name:        "not null comparison",
			dsl:         `$.email != null`,
			expectedSQL: `"email" IS NOT NULL`,
			paramCount:  0,
		},
		{
			name:        "negation",
			dsl:         `!($.active == true)`,
			expectedSQL: `NOT (("active" = ?))`,
			paramCount:  1,
		},
		{
			name:        "complex nested expression",
			dsl:         `($.age >= 18 && $.age <= 65) || $.role == "admin"`,
			expectedSQL: `((("age" >= ?) AND ("age" <= ?)) OR ("role" = ?))`,
			paramCount:  3,
		},
		{
			name:        "less than or equal",
			dsl:         `$.price <= 100`,
			expectedSQL: `("price" <= ?)`,
			paramCount:  1,
		},
		{
			name:        "greater than or equal",
			dsl:         `$.quantity >= 5`,
			expectedSQL: `("quantity" >= ?)`,
			paramCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}

			if len(result.Params) != tt.paramCount {
				t.Errorf("expected %d params, got %d", tt.paramCount, len(result.Params))
			}
		})
	}
}

func TestSQLCompiler_PostgresDialect(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
		paramCount  int
	}{
		{
			name:        "simple comparison with postgres params",
			dsl:         `$.age > 18`,
			expectedSQL: `("age" > $1)`,
			paramCount:  1,
		},
		{
			name:        "multiple params",
			dsl:         `$.age > 18 && $.status == "active"`,
			expectedSQL: `(("age" > $1) AND ("status" = $2))`,
			paramCount:  2,
		},
		{
			name:        "in expression",
			dsl:         `$.status IN ["active", "pending", "review"]`,
			expectedSQL: `"status" IN ($1, $2, $3)`,
			paramCount:  3,
		},
		{
			name:        "boolean value",
			dsl:         `$.active == true`,
			expectedSQL: `("active" = $1)`,
			paramCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler(WithDialect(DialectPostgres))
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}

			if len(result.Params) != tt.paramCount {
				t.Errorf("expected %d params, got %d", tt.paramCount, len(result.Params))
			}
		})
	}
}

func TestSQLCompiler_MySQLDialect(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
	}{
		{
			name:        "identifier escaping",
			dsl:         `$.user_name == "john"`,
			expectedSQL: "(`user_name` = ?)",
		},
		{
			name:        "boolean as int",
			dsl:         `$.active == true`,
			expectedSQL: "(`active` = ?)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler(WithDialect(DialectMySQL))
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}
		})
	}
}

func TestSQLCompiler_Functions(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
		dialect     SQLDialect
	}{
		{
			name:        "lower function",
			dsl:         `lower($.name) == "john"`,
			expectedSQL: `(LOWER("name") = ?)`,
			dialect:     DialectStandard,
		},
		{
			name:        "upper function",
			dsl:         `upper($.code) == "ABC"`,
			expectedSQL: `(UPPER("code") = ?)`,
			dialect:     DialectStandard,
		},
		{
			name:        "trim function",
			dsl:         `trim($.input) == "test"`,
			expectedSQL: `(TRIM("input") = ?)`,
			dialect:     DialectStandard,
		},
		{
			name:        "length function standard",
			dsl:         `len($.name) > 5`,
			expectedSQL: `(LENGTH("name") > ?)`,
			dialect:     DialectStandard,
		},
		{
			name:        "length function mysql",
			dsl:         `len($.name) > 5`,
			expectedSQL: "(CHAR_LENGTH(`name`) > ?)",
			dialect:     DialectMySQL,
		},
		{
			name:        "abs function",
			dsl:         `abs($.balance) > 100`,
			expectedSQL: `(ABS("balance") > ?)`,
			dialect:     DialectStandard,
		},
		{
			name:        "coalesce function",
			dsl:         `coalesce($.nickname, $.name) == "John"`,
			expectedSQL: `(COALESCE("nickname", "name") = ?)`,
			dialect:     DialectStandard,
		},
		{
			name:        "isNull function",
			dsl:         `isNull($.deleted_at)`,
			expectedSQL: `("deleted_at" IS NULL)`,
			dialect:     DialectStandard,
		},
		{
			name:        "isNotNull function",
			dsl:         `isNotNull($.email)`,
			expectedSQL: `("email" IS NOT NULL)`,
			dialect:     DialectStandard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler(WithDialect(tt.dialect))
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}
		})
	}
}

func TestSQLCompiler_StringFunctions(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
		paramCount  int
	}{
		{
			name:        "contains function",
			dsl:         `contains($.name, "john")`,
			expectedSQL: `"name" LIKE ?`,
			paramCount:  1,
		},
		{
			name:        "startsWith function",
			dsl:         `startsWith($.email, "admin")`,
			expectedSQL: `"email" LIKE ?`,
			paramCount:  1,
		},
		{
			name:        "endsWith function",
			dsl:         `endsWith($.email, ".com")`,
			expectedSQL: `"email" LIKE ?`,
			paramCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}

			if len(result.Params) != tt.paramCount {
				t.Errorf("expected %d params, got %d", tt.paramCount, len(result.Params))
			}
		})
	}
}

func TestSQLCompiler_ParamValues(t *testing.T) {
	tests := []struct {
		name           string
		dsl            string
		expectedParams []interface{}
	}{
		{
			name:           "integer param",
			dsl:            `$.age > 18`,
			expectedParams: []interface{}{int64(18)},
		},
		{
			name:           "string param",
			dsl:            `$.name == "John"`,
			expectedParams: []interface{}{"John"},
		},
		{
			name:           "multiple params",
			dsl:            `$.age >= 18 && $.status == "active"`,
			expectedParams: []interface{}{int64(18), "active"},
		},
		{
			name:           "float param",
			dsl:            `$.price < 99.99`,
			expectedParams: []interface{}{99.99},
		},
		{
			name:           "boolean param",
			dsl:            `$.active == true`,
			expectedParams: []interface{}{1}, // MySQL/SQLite style
		},
		{
			name:           "in expression params",
			dsl:            `$.status IN ["a", "b", "c"]`,
			expectedParams: []interface{}{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if len(result.Params) != len(tt.expectedParams) {
				t.Fatalf("expected %d params, got %d", len(tt.expectedParams), len(result.Params))
			}

			for i, expected := range tt.expectedParams {
				if result.Params[i] != expected {
					t.Errorf("param %d: expected %v (%T), got %v (%T)",
						i, expected, expected, result.Params[i], result.Params[i])
				}
			}
		})
	}
}

func TestSQLCompiler_InlineParams(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
	}{
		{
			name:        "inline integer",
			dsl:         `$.age > 18`,
			expectedSQL: `("age" > 18)`,
		},
		{
			name:        "inline string",
			dsl:         `$.name == "John"`,
			expectedSQL: `("name" = 'John')`,
		},
		{
			name:        "inline string with quote",
			dsl:         `$.name == "O'Brien"`,
			expectedSQL: `("name" = 'O''Brien')`,
		},
		{
			name:        "inline boolean standard",
			dsl:         `$.active == true`,
			expectedSQL: `("active" = 1)`,
		},
		{
			name:        "inline null",
			dsl:         `$.deleted_at == null`,
			expectedSQL: `"deleted_at" IS NULL`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler(WithParamStyle(ParamInline))
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}

			if len(result.Params) != 0 {
				t.Errorf("expected 0 params with inline style, got %d", len(result.Params))
			}
		})
	}
}

func TestSQLCompiler_CustomFieldMapper(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		mapper      func(string) string
		expectedSQL string
	}{
		{
			name: "table prefix mapping",
			dsl:  `$.user.name == "John"`,
			mapper: func(path string) string {
				// Map $.user.name to users.name
				if path == "$.user.name" {
					return "users.name"
				}
				return defaultFieldMapper(path)
			},
			expectedSQL: `("users.name" = ?)`,
		},
		{
			name: "snake_case to column",
			dsl:  `$.firstName == "John"`,
			mapper: func(path string) string {
				if path == "$.firstName" {
					return "first_name"
				}
				return defaultFieldMapper(path)
			},
			expectedSQL: `("first_name" = ?)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler(WithFieldMapper(tt.mapper))
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}
		})
	}
}

func TestSQLCompiler_RegexExpression(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		dialect     SQLDialect
		expectedSQL string
		shouldErr   bool
	}{
		{
			name:        "postgres regex match",
			dsl:         `$.email =~ "@gmail.com$"`,
			dialect:     DialectPostgres,
			expectedSQL: `"email" ~ $1`,
		},
		{
			name:        "postgres regex not match",
			dsl:         `$.email !~ "@spam.com$"`,
			dialect:     DialectPostgres,
			expectedSQL: `"email" !~ $1`,
		},
		{
			name:        "mysql regex match",
			dsl:         `$.email =~ "@gmail.com$"`,
			dialect:     DialectMySQL,
			expectedSQL: "`email` REGEXP ?",
		},
		{
			name:        "standard sql uses like",
			dsl:         `$.name =~ "^John"`,
			dialect:     DialectStandard,
			expectedSQL: `"name" LIKE ?`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler(WithDialect(tt.dialect))
			result, err := compiler.Compile(expr)

			if tt.shouldErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}
		})
	}
}

func TestSQLCompiler_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name        string
		dsl         string
		expectedSQL string
		paramCount  int
	}{
		{
			name:        "access control expression",
			dsl:         `$.user.role == "admin" || ($.user.role == "manager" && $.user.department == $.resource.department)`,
			expectedSQL: `(("user_role" = ?) OR (("user_role" = ?) AND ("user_department" = "resource_department")))`,
			paramCount:  2,
		},
		{
			name:        "age range check",
			dsl:         `$.age >= 18 && $.age <= 65 && $.status IN ["active", "pending"]`,
			expectedSQL: `((("age" >= ?) AND ("age" <= ?)) AND "status" IN (?, ?))`,
			paramCount:  4,
		},
		{
			name:        "triple or condition",
			dsl:         `$.type == "A" || $.type == "B" || $.type == "C"`,
			expectedSQL: `((("type" = ?) OR ("type" = ?)) OR ("type" = ?))`,
			paramCount:  3,
		},
		{
			name:        "negated complex expression",
			dsl:         `!($.deleted == true || $.archived == true)`,
			expectedSQL: `NOT ((("deleted" = ?) OR ("archived" = ?)))`,
			paramCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.Parse(tt.dsl)
			if err != nil {
				t.Fatalf("failed to parse DSL: %v", err)
			}

			compiler := NewSQLCompiler()
			result, err := compiler.Compile(expr)
			if err != nil {
				t.Fatalf("failed to compile: %v", err)
			}

			if result.SQL != tt.expectedSQL {
				t.Errorf("expected SQL: %s, got: %s", tt.expectedSQL, result.SQL)
			}

			if len(result.Params) != tt.paramCount {
				t.Errorf("expected %d params, got %d", tt.paramCount, len(result.Params))
			}
		})
	}
}

func TestSQLCompiler_Errors(t *testing.T) {
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

			compiler := NewSQLCompiler()
			_, err = compiler.Compile(expr)

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestDefaultFieldMapper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"$.user.name", "user_name"},
		{"$.data[0].value", "data_0_value"},
		{"$.items[0][1].name", "items_0_1_name"},
		{"$.simple", "simple"},
		{"$", ""},
		{"$.a.b.c.d", "a_b_c_d"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := defaultFieldMapper(tt.input)
			if result != tt.expected {
				t.Errorf("defaultFieldMapper(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompileToSQL_Convenience(t *testing.T) {
	expr, err := parser.Parse(`$.age > 18 && $.active == true`)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	result, err := CompileToSQL(expr, WithDialect(DialectPostgres))
	if err != nil {
		t.Fatalf("failed to compile: %v", err)
	}

	expectedSQL := `(("age" > $1) AND ("active" = $2))`
	if result.SQL != expectedSQL {
		t.Errorf("expected SQL: %s, got: %s", expectedSQL, result.SQL)
	}

	if len(result.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(result.Params))
	}
}
