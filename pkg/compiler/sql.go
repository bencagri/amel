// Package compiler provides compilation targets for AMEL expressions.
package compiler

import (
	"fmt"
	"strings"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/ast"
)

// SQLDialect represents different SQL database dialects.
type SQLDialect int

const (
	DialectStandard SQLDialect = iota // Standard SQL
	DialectPostgres                   // PostgreSQL
	DialectMySQL                      // MySQL
	DialectSQLite                     // SQLite
)

// SQLCompiler compiles AMEL expressions to SQL WHERE clauses.
type SQLCompiler struct {
	dialect     SQLDialect
	fieldMapper func(string) string // Maps JSON paths to SQL column names
	paramStyle  ParamStyle
	params      []interface{}
	paramIndex  int
}

// ParamStyle represents how parameters are formatted in SQL.
type ParamStyle int

const (
	ParamQuestion ParamStyle = iota // ? placeholders (MySQL, SQLite)
	ParamDollar                     // $1, $2, ... (PostgreSQL)
	ParamNamed                      // :name (Oracle)
	ParamInline                     // Inline values (use with caution)
)

// SQLCompilerOption configures the SQL compiler.
type SQLCompilerOption func(*SQLCompiler)

// WithDialect sets the SQL dialect.
func WithDialect(dialect SQLDialect) SQLCompilerOption {
	return func(c *SQLCompiler) {
		c.dialect = dialect
		// Set default param style based on dialect
		switch dialect {
		case DialectPostgres:
			c.paramStyle = ParamDollar
		case DialectMySQL, DialectSQLite:
			c.paramStyle = ParamQuestion
		default:
			c.paramStyle = ParamQuestion
		}
	}
}

// WithParamStyle sets the parameter style.
func WithParamStyle(style ParamStyle) SQLCompilerOption {
	return func(c *SQLCompiler) {
		c.paramStyle = style
	}
}

// WithFieldMapper sets a custom function to map JSON paths to SQL column names.
// For example, "$.user.age" could be mapped to "users.age" or just "age".
func WithFieldMapper(mapper func(string) string) SQLCompilerOption {
	return func(c *SQLCompiler) {
		c.fieldMapper = mapper
	}
}

// NewSQLCompiler creates a new SQL compiler with the given options.
func NewSQLCompiler(opts ...SQLCompilerOption) *SQLCompiler {
	c := &SQLCompiler{
		dialect:    DialectStandard,
		paramStyle: ParamQuestion,
		fieldMapper: func(path string) string {
			// Default: convert $.user.name to user_name
			return defaultFieldMapper(path)
		},
		params:     make([]interface{}, 0),
		paramIndex: 0,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SQLResult contains the compiled SQL and parameters.
type SQLResult struct {
	SQL    string        // The WHERE clause (without "WHERE" keyword)
	Params []interface{} // The parameter values
}

// Compile compiles an AMEL expression to a SQL WHERE clause.
func (c *SQLCompiler) Compile(expr ast.Expression) (*SQLResult, error) {
	c.params = make([]interface{}, 0)
	c.paramIndex = 0

	sql, err := c.compile(expr)
	if err != nil {
		return nil, err
	}

	return &SQLResult{
		SQL:    sql,
		Params: c.params,
	}, nil
}

func (c *SQLCompiler) compile(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return c.compileParam(e.Value)

	case *ast.FloatLiteral:
		return c.compileParam(e.Value)

	case *ast.StringLiteral:
		return c.compileParam(e.Value)

	case *ast.BooleanLiteral:
		if c.dialect == DialectPostgres {
			return c.compileParam(e.Value)
		}
		// MySQL/SQLite use 1/0 for booleans
		if e.Value {
			return c.compileParam(1)
		}
		return c.compileParam(0)

	case *ast.NullLiteral:
		return "NULL", nil

	case *ast.Identifier:
		// Identifiers are treated as column names
		return c.escapeIdentifier(e.Value), nil

	case *ast.JSONPathExpression:
		return c.compileJSONPath(e)

	case *ast.BinaryExpression:
		return c.compileBinaryExpression(e)

	case *ast.UnaryExpression:
		return c.compileUnaryExpression(e)

	case *ast.InExpression:
		return c.compileInExpression(e)

	case *ast.RegexExpression:
		return c.compileRegexExpression(e)

	case *ast.ListLiteral:
		return c.compileListLiteral(e)

	case *ast.GroupedExpression:
		inner, err := c.compile(e.Expression)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil

	case *ast.FunctionCall:
		return c.compileFunctionCall(e)

	default:
		return "", errors.Newf(errors.ErrInvalidSyntax, "unsupported expression type for SQL: %T", expr)
	}
}

func (c *SQLCompiler) compileParam(value interface{}) (string, error) {
	if c.paramStyle == ParamInline {
		return c.inlineValue(value), nil
	}

	c.params = append(c.params, value)
	c.paramIndex++

	switch c.paramStyle {
	case ParamQuestion:
		return "?", nil
	case ParamDollar:
		return fmt.Sprintf("$%d", c.paramIndex), nil
	case ParamNamed:
		return fmt.Sprintf(":p%d", c.paramIndex), nil
	default:
		return "?", nil
	}
}

func (c *SQLCompiler) inlineValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return "'" + escaped + "'"
	case bool:
		if v {
			if c.dialect == DialectPostgres {
				return "TRUE"
			}
			return "1"
		}
		if c.dialect == DialectPostgres {
			return "FALSE"
		}
		return "0"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (c *SQLCompiler) compileJSONPath(jp *ast.JSONPathExpression) (string, error) {
	columnName := c.fieldMapper(jp.Path)
	return c.escapeIdentifier(columnName), nil
}

func (c *SQLCompiler) compileBinaryExpression(be *ast.BinaryExpression) (string, error) {
	left, err := c.compile(be.Left)
	if err != nil {
		return "", err
	}

	right, err := c.compile(be.Right)
	if err != nil {
		return "", err
	}

	// Handle NULL comparisons specially
	if isNullLiteral(be.Right) {
		switch be.Operator {
		case "==":
			return left + " IS NULL", nil
		case "!=":
			return left + " IS NOT NULL", nil
		}
	}

	if isNullLiteral(be.Left) {
		switch be.Operator {
		case "==":
			return right + " IS NULL", nil
		case "!=":
			return right + " IS NOT NULL", nil
		}
	}

	op := c.translateOperator(be.Operator)
	return fmt.Sprintf("(%s %s %s)", left, op, right), nil
}

func (c *SQLCompiler) compileUnaryExpression(ue *ast.UnaryExpression) (string, error) {
	operand, err := c.compile(ue.Operand)
	if err != nil {
		return "", err
	}

	switch ue.Operator {
	case "!", "NOT", "not":
		return fmt.Sprintf("NOT (%s)", operand), nil
	case "-":
		return fmt.Sprintf("-%s", operand), nil
	default:
		return "", errors.Newf(errors.ErrInvalidOperator, "unsupported unary operator for SQL: %s", ue.Operator)
	}
}

func (c *SQLCompiler) compileInExpression(ie *ast.InExpression) (string, error) {
	left, err := c.compile(ie.Left)
	if err != nil {
		return "", err
	}

	right, err := c.compile(ie.Right)
	if err != nil {
		return "", err
	}

	op := "IN"
	if ie.Negated {
		op = "NOT IN"
	}

	return fmt.Sprintf("%s %s %s", left, op, right), nil
}

func (c *SQLCompiler) compileRegexExpression(re *ast.RegexExpression) (string, error) {
	left, err := c.compile(re.Left)
	if err != nil {
		return "", err
	}

	// For regex, we need the pattern as a string
	pattern, ok := re.Pattern.(*ast.StringLiteral)
	if !ok {
		return "", errors.New(errors.ErrTypeMismatch, "regex pattern must be a string literal for SQL compilation")
	}

	switch c.dialect {
	case DialectPostgres:
		op := "~"
		if re.Negated {
			op = "!~"
		}
		param, err := c.compileParam(pattern.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, op, param), nil

	case DialectMySQL:
		op := "REGEXP"
		if re.Negated {
			op = "NOT REGEXP"
		}
		param, err := c.compileParam(pattern.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, op, param), nil

	case DialectSQLite:
		// SQLite requires a custom REGEXP function to be loaded
		op := "REGEXP"
		if re.Negated {
			return "", errors.New(errors.ErrInvalidOperator, "SQLite does not support NOT REGEXP natively")
		}
		param, err := c.compileParam(pattern.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, op, param), nil

	default:
		// Standard SQL doesn't have regex; use LIKE as fallback
		// Convert simple patterns
		likePattern := regexToLike(pattern.Value)
		op := "LIKE"
		if re.Negated {
			op = "NOT LIKE"
		}
		param, err := c.compileParam(likePattern)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, op, param), nil
	}
}

func (c *SQLCompiler) compileListLiteral(ll *ast.ListLiteral) (string, error) {
	if len(ll.Elements) == 0 {
		return "(NULL)", nil // Empty IN clause handling
	}

	parts := make([]string, len(ll.Elements))
	for i, elem := range ll.Elements {
		compiled, err := c.compile(elem)
		if err != nil {
			return "", err
		}
		parts[i] = compiled
	}

	return "(" + strings.Join(parts, ", ") + ")", nil
}

func (c *SQLCompiler) compileFunctionCall(fc *ast.FunctionCall) (string, error) {
	// Map AMEL functions to SQL functions
	switch strings.ToLower(fc.Name) {
	case "lower":
		return c.compileUnaryFunction("LOWER", fc)
	case "upper":
		return c.compileUnaryFunction("UPPER", fc)
	case "len", "length":
		return c.compileUnaryFunction(c.lengthFunction(), fc)
	case "trim":
		return c.compileUnaryFunction("TRIM", fc)
	case "abs":
		return c.compileUnaryFunction("ABS", fc)
	case "ceil", "ceiling":
		return c.compileUnaryFunction("CEIL", fc)
	case "floor":
		return c.compileUnaryFunction("FLOOR", fc)
	case "round":
		return c.compileUnaryFunction("ROUND", fc)
	case "coalesce":
		return c.compileVariadicFunction("COALESCE", fc)
	case "concat":
		return c.compileConcatFunction(fc)
	case "substr", "substring":
		return c.compileSubstrFunction(fc)
	case "isnull":
		if len(fc.Arguments) != 1 {
			return "", errors.New(errors.ErrArgumentCount, "isNull requires exactly 1 argument")
		}
		arg, err := c.compile(fc.Arguments[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s IS NULL)", arg), nil
	case "isnotnull":
		if len(fc.Arguments) != 1 {
			return "", errors.New(errors.ErrArgumentCount, "isNotNull requires exactly 1 argument")
		}
		arg, err := c.compile(fc.Arguments[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s IS NOT NULL)", arg), nil
	case "min":
		return c.compileVariadicFunction("MIN", fc)
	case "max":
		return c.compileVariadicFunction("MAX", fc)
	case "sum":
		return c.compileUnaryFunction("SUM", fc)
	case "avg":
		return c.compileUnaryFunction("AVG", fc)
	case "count":
		return c.compileUnaryFunction("COUNT", fc)
	case "contains":
		return c.compileContainsFunction(fc)
	case "startswith":
		return c.compileStartsWithFunction(fc)
	case "endswith":
		return c.compileEndsWithFunction(fc)
	default:
		return "", errors.Newf(errors.ErrUndefinedFunction, "unsupported function for SQL: %s", fc.Name)
	}
}

func (c *SQLCompiler) compileUnaryFunction(sqlFunc string, fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) != 1 {
		return "", errors.Newf(errors.ErrArgumentCount, "%s requires exactly 1 argument", fc.Name)
	}
	arg, err := c.compile(fc.Arguments[0])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", sqlFunc, arg), nil
}

func (c *SQLCompiler) compileVariadicFunction(sqlFunc string, fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) == 0 {
		return "", errors.Newf(errors.ErrArgumentCount, "%s requires at least 1 argument", fc.Name)
	}
	args := make([]string, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		compiled, err := c.compile(arg)
		if err != nil {
			return "", err
		}
		args[i] = compiled
	}
	return fmt.Sprintf("%s(%s)", sqlFunc, strings.Join(args, ", ")), nil
}

func (c *SQLCompiler) compileConcatFunction(fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) < 2 {
		return "", errors.New(errors.ErrArgumentCount, "concat requires at least 2 arguments")
	}

	args := make([]string, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		compiled, err := c.compile(arg)
		if err != nil {
			return "", err
		}
		args[i] = compiled
	}

	switch c.dialect {
	case DialectMySQL:
		return fmt.Sprintf("CONCAT(%s)", strings.Join(args, ", ")), nil
	case DialectPostgres, DialectSQLite:
		return "(" + strings.Join(args, " || ") + ")", nil
	default:
		return fmt.Sprintf("CONCAT(%s)", strings.Join(args, ", ")), nil
	}
}

func (c *SQLCompiler) compileSubstrFunction(fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) < 2 || len(fc.Arguments) > 3 {
		return "", errors.New(errors.ErrArgumentCount, "substr requires 2 or 3 arguments")
	}

	args := make([]string, len(fc.Arguments))
	for i, arg := range fc.Arguments {
		compiled, err := c.compile(arg)
		if err != nil {
			return "", err
		}
		args[i] = compiled
	}

	switch c.dialect {
	case DialectMySQL:
		return fmt.Sprintf("SUBSTRING(%s)", strings.Join(args, ", ")), nil
	default:
		return fmt.Sprintf("SUBSTR(%s)", strings.Join(args, ", ")), nil
	}
}

func (c *SQLCompiler) compileContainsFunction(fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) != 2 {
		return "", errors.New(errors.ErrArgumentCount, "contains requires exactly 2 arguments")
	}

	haystack, err := c.compile(fc.Arguments[0])
	if err != nil {
		return "", err
	}

	needle, ok := fc.Arguments[1].(*ast.StringLiteral)
	if !ok {
		return "", errors.New(errors.ErrTypeMismatch, "contains second argument must be a string literal")
	}

	pattern := "%" + escapeLikePattern(needle.Value) + "%"
	param, err := c.compileParam(pattern)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s LIKE %s", haystack, param), nil
}

func (c *SQLCompiler) compileStartsWithFunction(fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) != 2 {
		return "", errors.New(errors.ErrArgumentCount, "startsWith requires exactly 2 arguments")
	}

	str, err := c.compile(fc.Arguments[0])
	if err != nil {
		return "", err
	}

	prefix, ok := fc.Arguments[1].(*ast.StringLiteral)
	if !ok {
		return "", errors.New(errors.ErrTypeMismatch, "startsWith second argument must be a string literal")
	}

	pattern := escapeLikePattern(prefix.Value) + "%"
	param, err := c.compileParam(pattern)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s LIKE %s", str, param), nil
}

func (c *SQLCompiler) compileEndsWithFunction(fc *ast.FunctionCall) (string, error) {
	if len(fc.Arguments) != 2 {
		return "", errors.New(errors.ErrArgumentCount, "endsWith requires exactly 2 arguments")
	}

	str, err := c.compile(fc.Arguments[0])
	if err != nil {
		return "", err
	}

	suffix, ok := fc.Arguments[1].(*ast.StringLiteral)
	if !ok {
		return "", errors.New(errors.ErrTypeMismatch, "endsWith second argument must be a string literal")
	}

	pattern := "%" + escapeLikePattern(suffix.Value)
	param, err := c.compileParam(pattern)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s LIKE %s", str, param), nil
}

func (c *SQLCompiler) translateOperator(op string) string {
	switch op {
	case "==":
		return "="
	case "!=":
		return "<>"
	case "&&", "AND", "and":
		return "AND"
	case "||", "OR", "or":
		return "OR"
	case "!":
		return "NOT"
	default:
		return op
	}
}

func (c *SQLCompiler) escapeIdentifier(name string) string {
	switch c.dialect {
	case DialectPostgres:
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	case DialectMySQL:
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	default:
		// Standard SQL uses double quotes
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	}
}

func (c *SQLCompiler) lengthFunction() string {
	switch c.dialect {
	case DialectMySQL:
		return "CHAR_LENGTH"
	default:
		return "LENGTH"
	}
}

// Helper functions

func defaultFieldMapper(path string) string {
	// Convert $.user.name to user_name
	// Convert $.data[0].value to data_0_value
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, "$")

	// Replace . with _
	path = strings.ReplaceAll(path, ".", "_")

	// Replace [n] with _n_
	path = strings.ReplaceAll(path, "[", "_")
	path = strings.ReplaceAll(path, "]", "")

	// Clean up multiple underscores
	for strings.Contains(path, "__") {
		path = strings.ReplaceAll(path, "__", "_")
	}

	// Remove leading/trailing underscores
	path = strings.Trim(path, "_")

	return path
}

func isNullLiteral(expr ast.Expression) bool {
	_, ok := expr.(*ast.NullLiteral)
	return ok
}

func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// regexToLike attempts to convert simple regex patterns to LIKE patterns.
// This is a best-effort conversion for basic patterns.
func regexToLike(pattern string) string {
	// Handle common simple patterns
	result := pattern

	// ^pattern$ -> pattern (exact match)
	if strings.HasPrefix(result, "^") && strings.HasSuffix(result, "$") {
		return strings.TrimSuffix(strings.TrimPrefix(result, "^"), "$")
	}

	// ^pattern -> pattern% (starts with)
	if strings.HasPrefix(result, "^") {
		return strings.TrimPrefix(result, "^") + "%"
	}

	// pattern$ -> %pattern (ends with)
	if strings.HasSuffix(result, "$") {
		return "%" + strings.TrimSuffix(result, "$")
	}

	// .* -> % (any characters)
	result = strings.ReplaceAll(result, ".*", "%")

	// . -> _ (single character)
	result = strings.ReplaceAll(result, ".", "_")

	// Default: wrap with % for contains behavior
	if !strings.HasPrefix(result, "%") && !strings.HasSuffix(result, "%") {
		result = "%" + result + "%"
	}

	return result
}

// CompileToSQL is a convenience function that compiles an AMEL expression to SQL.
func CompileToSQL(expr ast.Expression, opts ...SQLCompilerOption) (*SQLResult, error) {
	compiler := NewSQLCompiler(opts...)
	return compiler.Compile(expr)
}
