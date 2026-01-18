// Package compiler provides compilation targets for AMEL expressions.
package compiler

import (
	"encoding/json"
	"strings"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/ast"
)

// MongoDBCompiler compiles AMEL expressions to MongoDB query documents.
type MongoDBCompiler struct {
	fieldMapper func(string) string // Maps JSON paths to MongoDB field names
}

// MongoDBCompilerOption configures the MongoDB compiler.
type MongoDBCompilerOption func(*MongoDBCompiler)

// WithMongoFieldMapper sets a custom function to map JSON paths to MongoDB field names.
func WithMongoFieldMapper(mapper func(string) string) MongoDBCompilerOption {
	return func(c *MongoDBCompiler) {
		c.fieldMapper = mapper
	}
}

// NewMongoDBCompiler creates a new MongoDB compiler with the given options.
func NewMongoDBCompiler(opts ...MongoDBCompilerOption) *MongoDBCompiler {
	c := &MongoDBCompiler{
		fieldMapper: func(path string) string {
			return defaultMongoFieldMapper(path)
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// MongoDBResult contains the compiled MongoDB query.
type MongoDBResult struct {
	Query map[string]interface{} // The MongoDB query document
}

// ToJSON returns the query as a JSON string.
func (r *MongoDBResult) ToJSON() (string, error) {
	data, err := json.Marshal(r.Query)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToPrettyJSON returns the query as a formatted JSON string.
func (r *MongoDBResult) ToPrettyJSON() (string, error) {
	data, err := json.MarshalIndent(r.Query, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Compile compiles an AMEL expression to a MongoDB query document.
func (c *MongoDBCompiler) Compile(expr ast.Expression) (*MongoDBResult, error) {
	query, err := c.compile(expr)
	if err != nil {
		return nil, err
	}

	return &MongoDBResult{
		Query: query,
	}, nil
}

func (c *MongoDBCompiler) compile(expr ast.Expression) (map[string]interface{}, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return c.compileBinaryExpression(e)

	case *ast.UnaryExpression:
		return c.compileUnaryExpression(e)

	case *ast.InExpression:
		return c.compileInExpression(e)

	case *ast.RegexExpression:
		return c.compileRegexExpression(e)

	case *ast.GroupedExpression:
		return c.compile(e.Expression)

	case *ast.FunctionCall:
		return c.compileFunctionCall(e)

	case *ast.BooleanLiteral:
		// A bare boolean literal as the entire expression
		return map[string]interface{}{"$expr": e.Value}, nil

	default:
		return nil, errors.Newf(errors.ErrInvalidSyntax, "unsupported expression type for MongoDB: %T", expr)
	}
}

func (c *MongoDBCompiler) compileBinaryExpression(be *ast.BinaryExpression) (map[string]interface{}, error) {
	// Handle logical operators (AND, OR)
	switch be.Operator {
	case "&&", "AND", "and":
		return c.compileLogicalExpression("$and", be)
	case "||", "OR", "or":
		return c.compileLogicalExpression("$or", be)
	}

	// Handle comparison operators
	field, err := c.extractField(be.Left)
	if err != nil {
		// It might be a field on the right side
		field, err = c.extractField(be.Right)
		if err != nil {
			// Both sides might be expressions - use $expr
			return c.compileExpressionOperator(be)
		}
		// Field is on the right, value on the left - swap the comparison
		return c.compileComparisonSwapped(field, be.Operator, be.Left)
	}

	// Normal case: field on left, value on right
	return c.compileComparison(field, be.Operator, be.Right)
}

func (c *MongoDBCompiler) compileLogicalExpression(op string, be *ast.BinaryExpression) (map[string]interface{}, error) {
	left, err := c.compile(be.Left)
	if err != nil {
		return nil, err
	}

	right, err := c.compile(be.Right)
	if err != nil {
		return nil, err
	}

	// Flatten nested $and/$or operations
	conditions := make([]interface{}, 0)

	if leftConditions, ok := left[op]; ok {
		if arr, ok := leftConditions.([]interface{}); ok {
			conditions = append(conditions, arr...)
		} else {
			conditions = append(conditions, left)
		}
	} else {
		conditions = append(conditions, left)
	}

	if rightConditions, ok := right[op]; ok {
		if arr, ok := rightConditions.([]interface{}); ok {
			conditions = append(conditions, arr...)
		} else {
			conditions = append(conditions, right)
		}
	} else {
		conditions = append(conditions, right)
	}

	return map[string]interface{}{op: conditions}, nil
}

func (c *MongoDBCompiler) compileComparison(field, operator string, valueExpr ast.Expression) (map[string]interface{}, error) {
	value, err := c.extractValue(valueExpr)
	if err != nil {
		return nil, err
	}

	// Handle null comparisons
	if value == nil {
		switch operator {
		case "==":
			return map[string]interface{}{field: nil}, nil
		case "!=":
			return map[string]interface{}{field: map[string]interface{}{"$ne": nil}}, nil
		}
	}

	switch operator {
	case "==":
		return map[string]interface{}{field: value}, nil
	case "!=":
		return map[string]interface{}{field: map[string]interface{}{"$ne": value}}, nil
	case "<":
		return map[string]interface{}{field: map[string]interface{}{"$lt": value}}, nil
	case ">":
		return map[string]interface{}{field: map[string]interface{}{"$gt": value}}, nil
	case "<=":
		return map[string]interface{}{field: map[string]interface{}{"$lte": value}}, nil
	case ">=":
		return map[string]interface{}{field: map[string]interface{}{"$gte": value}}, nil
	case "+", "-", "*", "/", "%":
		// Arithmetic operations need $expr
		return c.compileArithmeticExpression(field, operator, value)
	default:
		return nil, errors.Newf(errors.ErrInvalidOperator, "unsupported operator for MongoDB: %s", operator)
	}
}

func (c *MongoDBCompiler) compileComparisonSwapped(field, operator string, valueExpr ast.Expression) (map[string]interface{}, error) {
	// When field is on right side, we need to swap the operator
	swappedOp := operator
	switch operator {
	case "<":
		swappedOp = ">"
	case ">":
		swappedOp = "<"
	case "<=":
		swappedOp = ">="
	case ">=":
		swappedOp = "<="
	}
	return c.compileComparison(field, swappedOp, valueExpr)
}

func (c *MongoDBCompiler) compileArithmeticExpression(field, operator string, value interface{}) (map[string]interface{}, error) {
	var mongoOp string
	switch operator {
	case "+":
		mongoOp = "$add"
	case "-":
		mongoOp = "$subtract"
	case "*":
		mongoOp = "$multiply"
	case "/":
		mongoOp = "$divide"
	case "%":
		mongoOp = "$mod"
	}

	return map[string]interface{}{
		"$expr": map[string]interface{}{
			mongoOp: []interface{}{"$" + field, value},
		},
	}, nil
}

func (c *MongoDBCompiler) compileExpressionOperator(be *ast.BinaryExpression) (map[string]interface{}, error) {
	left, err := c.compileToAggregationExpr(be.Left)
	if err != nil {
		return nil, err
	}

	right, err := c.compileToAggregationExpr(be.Right)
	if err != nil {
		return nil, err
	}

	var mongoOp string
	switch be.Operator {
	case "==":
		mongoOp = "$eq"
	case "!=":
		mongoOp = "$ne"
	case "<":
		mongoOp = "$lt"
	case ">":
		mongoOp = "$gt"
	case "<=":
		mongoOp = "$lte"
	case ">=":
		mongoOp = "$gte"
	case "+":
		mongoOp = "$add"
	case "-":
		mongoOp = "$subtract"
	case "*":
		mongoOp = "$multiply"
	case "/":
		mongoOp = "$divide"
	case "%":
		mongoOp = "$mod"
	default:
		return nil, errors.Newf(errors.ErrInvalidOperator, "unsupported operator for MongoDB $expr: %s", be.Operator)
	}

	return map[string]interface{}{
		"$expr": map[string]interface{}{
			mongoOp: []interface{}{left, right},
		},
	}, nil
}

func (c *MongoDBCompiler) compileToAggregationExpr(expr ast.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value, nil
	case *ast.FloatLiteral:
		return e.Value, nil
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.BooleanLiteral:
		return e.Value, nil
	case *ast.NullLiteral:
		return nil, nil
	case *ast.Identifier:
		return "$" + e.Value, nil
	case *ast.JSONPathExpression:
		field := c.fieldMapper(e.Path)
		return "$" + field, nil
	case *ast.BinaryExpression:
		left, err := c.compileToAggregationExpr(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := c.compileToAggregationExpr(e.Right)
		if err != nil {
			return nil, err
		}

		var mongoOp string
		switch e.Operator {
		case "+":
			mongoOp = "$add"
		case "-":
			mongoOp = "$subtract"
		case "*":
			mongoOp = "$multiply"
		case "/":
			mongoOp = "$divide"
		case "%":
			mongoOp = "$mod"
		default:
			return nil, errors.Newf(errors.ErrInvalidOperator, "unsupported arithmetic operator: %s", e.Operator)
		}
		return map[string]interface{}{mongoOp: []interface{}{left, right}}, nil
	default:
		return nil, errors.Newf(errors.ErrInvalidSyntax, "cannot convert to aggregation expression: %T", expr)
	}
}

func (c *MongoDBCompiler) compileUnaryExpression(ue *ast.UnaryExpression) (map[string]interface{}, error) {
	switch ue.Operator {
	case "!", "NOT", "not":
		inner, err := c.compile(ue.Operand)
		if err != nil {
			return nil, err
		}

		// Handle special cases for negation
		if len(inner) == 1 {
			for field, condition := range inner {
				// If it's a simple equality, convert to $ne
				if !strings.HasPrefix(field, "$") {
					switch cond := condition.(type) {
					case map[string]interface{}:
						// Already has an operator, wrap in $not
						return map[string]interface{}{field: map[string]interface{}{"$not": cond}}, nil
					default:
						// Simple equality, convert to $ne
						return map[string]interface{}{field: map[string]interface{}{"$ne": cond}}, nil
					}
				}
			}
		}

		// General case: use $nor
		return map[string]interface{}{"$nor": []interface{}{inner}}, nil

	case "-":
		// Unary minus - use $expr with $multiply
		operand, err := c.compileToAggregationExpr(ue.Operand)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"$expr": map[string]interface{}{
				"$multiply": []interface{}{operand, -1},
			},
		}, nil

	default:
		return nil, errors.Newf(errors.ErrInvalidOperator, "unsupported unary operator for MongoDB: %s", ue.Operator)
	}
}

func (c *MongoDBCompiler) compileInExpression(ie *ast.InExpression) (map[string]interface{}, error) {
	field, err := c.extractField(ie.Left)
	if err != nil {
		return nil, err
	}

	values, err := c.extractListValues(ie.Right)
	if err != nil {
		return nil, err
	}

	if ie.Negated {
		return map[string]interface{}{field: map[string]interface{}{"$nin": values}}, nil
	}
	return map[string]interface{}{field: map[string]interface{}{"$in": values}}, nil
}

func (c *MongoDBCompiler) compileRegexExpression(re *ast.RegexExpression) (map[string]interface{}, error) {
	field, err := c.extractField(re.Left)
	if err != nil {
		return nil, err
	}

	pattern, ok := re.Pattern.(*ast.StringLiteral)
	if !ok {
		return nil, errors.New(errors.ErrTypeMismatch, "regex pattern must be a string literal")
	}

	regexDoc := map[string]interface{}{
		"$regex": pattern.Value,
	}

	if re.Negated {
		return map[string]interface{}{field: map[string]interface{}{"$not": regexDoc}}, nil
	}
	return map[string]interface{}{field: regexDoc}, nil
}

func (c *MongoDBCompiler) compileFunctionCall(fc *ast.FunctionCall) (map[string]interface{}, error) {
	switch strings.ToLower(fc.Name) {
	case "isnull":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "isNull requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{field: nil}, nil

	case "isnotnull":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "isNotNull requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{field: map[string]interface{}{"$ne": nil}}, nil

	case "contains":
		if len(fc.Arguments) != 2 {
			return nil, errors.New(errors.ErrArgumentCount, "contains requires exactly 2 arguments")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		substr, ok := fc.Arguments[1].(*ast.StringLiteral)
		if !ok {
			return nil, errors.New(errors.ErrTypeMismatch, "contains second argument must be a string literal")
		}
		// Escape regex special characters
		escaped := escapeRegexPattern(substr.Value)
		return map[string]interface{}{field: map[string]interface{}{"$regex": escaped}}, nil

	case "startswith":
		if len(fc.Arguments) != 2 {
			return nil, errors.New(errors.ErrArgumentCount, "startsWith requires exactly 2 arguments")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		prefix, ok := fc.Arguments[1].(*ast.StringLiteral)
		if !ok {
			return nil, errors.New(errors.ErrTypeMismatch, "startsWith second argument must be a string literal")
		}
		escaped := escapeRegexPattern(prefix.Value)
		return map[string]interface{}{field: map[string]interface{}{"$regex": "^" + escaped}}, nil

	case "endswith":
		if len(fc.Arguments) != 2 {
			return nil, errors.New(errors.ErrArgumentCount, "endsWith requires exactly 2 arguments")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		suffix, ok := fc.Arguments[1].(*ast.StringLiteral)
		if !ok {
			return nil, errors.New(errors.ErrTypeMismatch, "endsWith second argument must be a string literal")
		}
		escaped := escapeRegexPattern(suffix.Value)
		return map[string]interface{}{field: map[string]interface{}{"$regex": escaped + "$"}}, nil

	case "len", "length":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "len requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		// Return an expression that can be used in comparisons
		return map[string]interface{}{
			"$expr": map[string]interface{}{
				"$strLenCP": "$" + field,
			},
		}, nil

	case "lower":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "lower requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"$expr": map[string]interface{}{
				"$toLower": "$" + field,
			},
		}, nil

	case "upper":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "upper requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"$expr": map[string]interface{}{
				"$toUpper": "$" + field,
			},
		}, nil

	case "abs":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "abs requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"$expr": map[string]interface{}{
				"$abs": "$" + field,
			},
		}, nil

	case "exists":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "exists requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{field: map[string]interface{}{"$exists": true}}, nil

	case "typeof", "type":
		if len(fc.Arguments) != 1 {
			return nil, errors.New(errors.ErrArgumentCount, "typeOf requires exactly 1 argument")
		}
		field, err := c.extractField(fc.Arguments[0])
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"$expr": map[string]interface{}{
				"$type": "$" + field,
			},
		}, nil

	default:
		return nil, errors.Newf(errors.ErrUndefinedFunction, "unsupported function for MongoDB: %s", fc.Name)
	}
}

func (c *MongoDBCompiler) extractField(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.JSONPathExpression:
		return c.fieldMapper(e.Path), nil
	case *ast.Identifier:
		return e.Value, nil
	default:
		return "", errors.Newf(errors.ErrInvalidSyntax, "expected field reference, got %T", expr)
	}
}

func (c *MongoDBCompiler) extractValue(expr ast.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value, nil
	case *ast.FloatLiteral:
		return e.Value, nil
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.BooleanLiteral:
		return e.Value, nil
	case *ast.NullLiteral:
		return nil, nil
	case *ast.JSONPathExpression:
		// Field-to-field comparison
		return "$" + c.fieldMapper(e.Path), nil
	case *ast.Identifier:
		return "$" + e.Value, nil
	default:
		return nil, errors.Newf(errors.ErrInvalidSyntax, "expected value, got %T", expr)
	}
}

func (c *MongoDBCompiler) extractListValues(expr ast.Expression) ([]interface{}, error) {
	list, ok := expr.(*ast.ListLiteral)
	if !ok {
		return nil, errors.New(errors.ErrTypeMismatch, "expected list for IN expression")
	}

	values := make([]interface{}, len(list.Elements))
	for i, elem := range list.Elements {
		val, err := c.extractValue(elem)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}
	return values, nil
}

// Helper functions

func defaultMongoFieldMapper(path string) string {
	// Convert $.user.name to user.name (MongoDB dot notation)
	// Convert $.data[0].value to data.0.value
	path = strings.TrimPrefix(path, "$.")
	path = strings.TrimPrefix(path, "$")

	// Replace [n] with .n
	path = strings.ReplaceAll(path, "[", ".")
	path = strings.ReplaceAll(path, "]", "")

	// Clean up multiple dots
	for strings.Contains(path, "..") {
		path = strings.ReplaceAll(path, "..", ".")
	}

	// Remove leading/trailing dots
	path = strings.Trim(path, ".")

	return path
}

func escapeRegexPattern(s string) string {
	// Escape regex special characters for MongoDB regex
	special := []string{"\\", ".", "+", "*", "?", "(", ")", "[", "]", "{", "}", "^", "$", "|"}
	result := s
	for _, char := range special {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// CompileToMongoDB is a convenience function that compiles an AMEL expression to MongoDB.
func CompileToMongoDB(expr ast.Expression, opts ...MongoDBCompilerOption) (*MongoDBResult, error) {
	compiler := NewMongoDBCompiler(opts...)
	return compiler.Compile(expr)
}
