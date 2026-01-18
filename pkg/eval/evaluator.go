// Package eval implements the AST evaluator for the AMEL DSL.
package eval

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/ast"
	"github.com/bencagri/amel/pkg/functions"
	"github.com/bencagri/amel/pkg/types"
	"github.com/tidwall/gjson"
)

// Higher-order function names that require special handling
var higherOrderFunctions = map[string]bool{
	"map":    true,
	"filter": true,
	"reduce": true,
	"find":   true,
	"some":   true,
	"every":  true,
}

// Evaluator evaluates AST expressions against a payload.
type Evaluator struct {
	functions *functions.Registry
	sandbox   *functions.Sandbox
	timeout   time.Duration
}

// EvalContext contains the context for evaluation.
type EvalContext struct {
	Payload     interface{}            // The JSON payload (map or raw JSON string)
	PayloadJSON string                 // The raw JSON string representation
	Variables   map[string]types.Value // Additional variables
	ctx         context.Context
}

// Explanation provides detailed information about an evaluation step.
type Explanation struct {
	Expression string         `json:"expression"`
	Result     types.Value    `json:"result"`
	Children   []*Explanation `json:"children,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

// Option is a function that configures the evaluator.
type Option func(*Evaluator)

// WithFunctions sets a custom function registry.
func WithFunctions(r *functions.Registry) Option {
	return func(e *Evaluator) {
		e.functions = r
	}
}

// WithTimeout sets the evaluation timeout.
func WithTimeout(d time.Duration) Option {
	return func(e *Evaluator) {
		e.timeout = d
	}
}

// WithSandbox sets a custom JavaScript sandbox for user-defined functions.
func WithSandbox(s *functions.Sandbox) Option {
	return func(e *Evaluator) {
		e.sandbox = s
	}
}

// New creates a new Evaluator with the given options.
func New(opts ...Option) (*Evaluator, error) {
	e := &Evaluator{
		timeout: 100 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(e)
	}

	// Create default function registry if not provided
	if e.functions == nil {
		r, err := functions.NewDefaultRegistry()
		if err != nil {
			return nil, err
		}
		e.functions = r
	}

	return e, nil
}

// NewContext creates a new evaluation context from a payload.
func NewContext(payload interface{}) (*EvalContext, error) {
	ctx := &EvalContext{
		Payload:   payload,
		Variables: make(map[string]types.Value),
		ctx:       context.Background(),
	}

	// Convert payload to JSON string for gjson
	switch p := payload.(type) {
	case string:
		ctx.PayloadJSON = p
	case []byte:
		ctx.PayloadJSON = string(p)
	case map[string]interface{}:
		// Use fmt for simple conversion
		jsonBytes, err := toJSON(p)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInvalidPath, "failed to convert payload to JSON", err)
		}
		ctx.PayloadJSON = string(jsonBytes)
	default:
		jsonBytes, err := toJSON(payload)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInvalidPath, "failed to convert payload to JSON", err)
		}
		ctx.PayloadJSON = string(jsonBytes)
	}

	return ctx, nil
}

// WithContext sets a Go context for the evaluation context.
func (ec *EvalContext) WithContext(ctx context.Context) *EvalContext {
	ec.ctx = ctx
	return ec
}

// SetVariable sets a variable in the evaluation context.
func (ec *EvalContext) SetVariable(name string, value types.Value) {
	ec.Variables[name] = value
}

// Evaluate evaluates an AST expression and returns the result.
func (e *Evaluator) Evaluate(expr ast.Expression, ctx *EvalContext) (types.Value, error) {
	// Always start with a fresh context to avoid reusing canceled contexts
	evalCtx := context.Background()

	// Create timeout context if timeout is set
	if e.timeout > 0 {
		var cancel context.CancelFunc
		evalCtx, cancel = context.WithTimeout(evalCtx, e.timeout)
		defer cancel()
	}

	ctx.ctx = evalCtx
	return e.eval(expr, ctx)
}

// EvaluateWithExplanation evaluates an expression and returns detailed explanation.
func (e *Evaluator) EvaluateWithExplanation(expr ast.Expression, ctx *EvalContext) (types.Value, *Explanation, error) {
	// Always start with a fresh context to avoid reusing canceled contexts
	evalCtx := context.Background()

	if e.timeout > 0 {
		var cancel context.CancelFunc
		evalCtx, cancel = context.WithTimeout(evalCtx, e.timeout)
		defer cancel()
	}

	ctx.ctx = evalCtx
	return e.evalWithExplanation(expr, ctx)
}

// EvaluateBool evaluates an expression and returns a boolean result.
func (e *Evaluator) EvaluateBool(expr ast.Expression, ctx *EvalContext) (bool, error) {
	result, err := e.Evaluate(expr, ctx)
	if err != nil {
		return false, err
	}
	return result.IsTruthy(), nil
}

// eval is the main evaluation dispatch function.
func (e *Evaluator) eval(node ast.Expression, ctx *EvalContext) (types.Value, error) {
	// Check for timeout
	select {
	case <-ctx.ctx.Done():
		return types.Null(), errors.New(errors.ErrTimeout, "evaluation timed out")
	default:
	}

	switch n := node.(type) {
	case *ast.IntegerLiteral:
		return types.Int(n.Value), nil

	case *ast.FloatLiteral:
		return types.Float(n.Value), nil

	case *ast.StringLiteral:
		return types.String(n.Value), nil

	case *ast.BooleanLiteral:
		return types.Bool(n.Value), nil

	case *ast.NullLiteral:
		return types.Null(), nil

	case *ast.ListLiteral:
		return e.evalListLiteral(n, ctx)

	case *ast.Identifier:
		return e.evalIdentifier(n, ctx)

	case *ast.JSONPathExpression:
		return e.evalJSONPath(n, ctx)

	case *ast.UnaryExpression:
		return e.evalUnaryExpression(n, ctx)

	case *ast.BinaryExpression:
		return e.evalBinaryExpression(n, ctx)

	case *ast.InExpression:
		return e.evalInExpression(n, ctx)

	case *ast.RegexExpression:
		return e.evalRegexExpression(n, ctx)

	case *ast.LambdaExpression:
		// Lambda expressions are not directly evaluated; they are used by higher-order functions
		return types.Null(), errors.New(errors.ErrInvalidSyntax, "lambda expressions cannot be evaluated directly")

	case *ast.FunctionCall:
		// Check if this is a higher-order function
		if higherOrderFunctions[n.Name] {
			return e.evalHigherOrderFunction(n, ctx)
		}
		return e.evalFunctionCall(n, ctx)

	case *ast.IndexExpression:
		return e.evalIndexExpression(n, ctx)

	case *ast.MemberExpression:
		return e.evalMemberExpression(n, ctx)

	case *ast.GroupedExpression:
		return e.eval(n.Expression, ctx)

	default:
		return types.Null(), errors.Newf(errors.ErrInvalidSyntax, "unknown expression type: %T", node)
	}
}

// evalWithExplanation evaluates and builds an explanation tree.
func (e *Evaluator) evalWithExplanation(node ast.Expression, ctx *EvalContext) (types.Value, *Explanation, error) {
	explanation := &Explanation{
		Expression: node.String(),
	}

	result, err := e.eval(node, ctx)
	if err != nil {
		explanation.Reason = fmt.Sprintf("Error: %s", err.Error())
		return result, explanation, err
	}

	explanation.Result = result

	// Add children explanations based on expression type
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		explanation.Reason = fmt.Sprintf("Integer literal: %d", n.Value)

	case *ast.FloatLiteral:
		explanation.Reason = fmt.Sprintf("Float literal: %v", n.Value)

	case *ast.StringLiteral:
		explanation.Reason = fmt.Sprintf("String literal: %q", n.Value)

	case *ast.BooleanLiteral:
		explanation.Reason = fmt.Sprintf("Boolean literal: %v", n.Value)

	case *ast.NullLiteral:
		explanation.Reason = "Null literal"

	case *ast.ListLiteral:
		children := make([]*Explanation, len(n.Elements))
		for i, elem := range n.Elements {
			_, childExp, _ := e.evalWithExplanation(elem, ctx)
			children[i] = childExp
		}
		explanation.Children = children
		explanation.Reason = fmt.Sprintf("List with %d elements", len(n.Elements))

	case *ast.Identifier:
		explanation.Reason = fmt.Sprintf("Identifier '%s' resolved to %v", n.Value, result.Raw)

	case *ast.JSONPathExpression:
		explanation.Reason = fmt.Sprintf("JSONPath '%s' resolved to %v", n.Path, result.Raw)

	case *ast.BinaryExpression:
		leftVal, leftExp, _ := e.evalWithExplanation(n.Left, ctx)
		rightVal, rightExp, _ := e.evalWithExplanation(n.Right, ctx)
		explanation.Children = []*Explanation{leftExp, rightExp}
		explanation.Reason = fmt.Sprintf("%v %s %v = %v", leftVal.Raw, n.Operator, rightVal.Raw, result.Raw)

	case *ast.UnaryExpression:
		operandVal, operandExp, _ := e.evalWithExplanation(n.Operand, ctx)
		explanation.Children = []*Explanation{operandExp}
		explanation.Reason = fmt.Sprintf("%s%v = %v", n.Operator, operandVal.Raw, result.Raw)

	case *ast.InExpression:
		leftVal, leftExp, _ := e.evalWithExplanation(n.Left, ctx)
		rightVal, rightExp, _ := e.evalWithExplanation(n.Right, ctx)
		explanation.Children = []*Explanation{leftExp, rightExp}
		op := "IN"
		if n.Negated {
			op = "NOT IN"
		}
		explanation.Reason = fmt.Sprintf("%v %s %v = %v", leftVal.Raw, op, rightVal.Raw, result.Raw)

	case *ast.RegexExpression:
		leftVal, leftExp, _ := e.evalWithExplanation(n.Left, ctx)
		patternVal, patternExp, _ := e.evalWithExplanation(n.Pattern, ctx)
		explanation.Children = []*Explanation{leftExp, patternExp}
		op := "=~"
		if n.Negated {
			op = "!~"
		}
		explanation.Reason = fmt.Sprintf("%v %s %v = %v", leftVal.Raw, op, patternVal.Raw, result.Raw)

	case *ast.FunctionCall:
		children := make([]*Explanation, len(n.Arguments))
		argVals := make([]interface{}, len(n.Arguments))
		for i, arg := range n.Arguments {
			argVal, argExp, _ := e.evalWithExplanation(arg, ctx)
			children[i] = argExp
			argVals[i] = argVal.Raw
		}
		explanation.Children = children
		explanation.Reason = fmt.Sprintf("Function %s(%v) = %v", n.Name, argVals, result.Raw)

	case *ast.IndexExpression:
		leftVal, leftExp, _ := e.evalWithExplanation(n.Left, ctx)
		indexVal, indexExp, _ := e.evalWithExplanation(n.Index, ctx)
		explanation.Children = []*Explanation{leftExp, indexExp}
		explanation.Reason = fmt.Sprintf("%v[%v] = %v", leftVal.Raw, indexVal.Raw, result.Raw)

	case *ast.MemberExpression:
		objVal, objExp, _ := e.evalWithExplanation(n.Object, ctx)
		explanation.Children = []*Explanation{objExp}
		explanation.Reason = fmt.Sprintf("%v.%s = %v", objVal.Raw, n.Property.Value, result.Raw)

	case *ast.GroupedExpression:
		innerVal, innerExp, _ := e.evalWithExplanation(n.Expression, ctx)
		explanation.Children = []*Explanation{innerExp}
		explanation.Reason = fmt.Sprintf("(%v) = %v", innerVal.Raw, result.Raw)

	default:
		explanation.Reason = fmt.Sprintf("Evaluated to %v", result.Raw)
	}

	return result, explanation, nil
}

// ============================================================================
// Literal and Identifier evaluation
// ============================================================================

func (e *Evaluator) evalListLiteral(list *ast.ListLiteral, ctx *EvalContext) (types.Value, error) {
	elements := make([]types.Value, len(list.Elements))
	for i, elem := range list.Elements {
		val, err := e.eval(elem, ctx)
		if err != nil {
			return types.Null(), err
		}
		elements[i] = val
	}
	return types.List(elements...), nil
}

func (e *Evaluator) evalIdentifier(ident *ast.Identifier, ctx *EvalContext) (types.Value, error) {
	// Check if it's a variable
	if val, ok := ctx.Variables[ident.Value]; ok {
		return val, nil
	}

	// Could also check for constants like "null", "true", "false" but those are handled as literals
	return types.Null(), errors.Newf(errors.ErrUndefinedVariable, "undefined variable: %s", ident.Value)
}

func (e *Evaluator) evalJSONPath(jp *ast.JSONPathExpression, ctx *EvalContext) (types.Value, error) {
	// Use gjson to resolve the path
	// Convert path from $.field to field (gjson doesn't need the $)
	path := jp.Path
	if len(path) > 1 && path[0] == '$' {
		if len(path) > 2 && path[1] == '.' {
			path = path[2:]
		} else {
			path = path[1:]
		}
	}

	// Handle root ($) by returning the entire payload
	if path == "" || path == "$" {
		return types.NewValue(ctx.Payload), nil
	}

	// Convert bracket notation to gjson dot notation
	// e.g., users[0].name -> users.0.name
	// e.g., data["key"] -> data.key
	path = convertToGjsonPath(path)

	result := gjson.Get(ctx.PayloadJSON, path)

	if !result.Exists() {
		return types.Null(), nil
	}

	return gjsonToValue(result), nil
}

// convertToGjsonPath converts JSONPath bracket notation to gjson dot notation.
// gjson uses dots for array indices: users.0.name instead of users[0].name
func convertToGjsonPath(path string) string {
	// Replace [N] with .N for numeric indices
	numericBracket := regexp.MustCompile(`\[(\d+)\]`)
	path = numericBracket.ReplaceAllString(path, ".$1")

	// Replace ["key"] or ['key'] with .key for string keys
	stringBracket := regexp.MustCompile(`\["([^"]+)"\]`)
	path = stringBracket.ReplaceAllString(path, ".$1")

	stringBracketSingle := regexp.MustCompile(`\['([^']+)'\]`)
	path = stringBracketSingle.ReplaceAllString(path, ".$1")

	// Clean up any leading dots
	if len(path) > 0 && path[0] == '.' {
		path = path[1:]
	}

	return path
}

// ============================================================================
// Operator evaluation
// ============================================================================

func (e *Evaluator) evalUnaryExpression(expr *ast.UnaryExpression, ctx *EvalContext) (types.Value, error) {
	operand, err := e.eval(expr.Operand, ctx)
	if err != nil {
		return types.Null(), err
	}

	switch expr.Operator {
	case "!", "not", "NOT":
		return types.Bool(!operand.IsTruthy()), nil

	case "-":
		switch operand.Type {
		case types.TypeInt:
			v, _ := operand.AsInt()
			return types.Int(-v), nil
		case types.TypeFloat:
			v, _ := operand.AsFloat()
			return types.Float(-v), nil
		default:
			return types.Null(), errors.Newf(errors.ErrTypeMismatch,
				"cannot negate %s", operand.Type)
		}

	default:
		return types.Null(), errors.Newf(errors.ErrInvalidOperator,
			"unknown unary operator: %s", expr.Operator)
	}
}

func (e *Evaluator) evalBinaryExpression(expr *ast.BinaryExpression, ctx *EvalContext) (types.Value, error) {
	// Short-circuit evaluation for logical operators
	if expr.Operator == "&&" || expr.Operator == "and" || expr.Operator == "AND" {
		left, err := e.eval(expr.Left, ctx)
		if err != nil {
			return types.Null(), err
		}
		if !left.IsTruthy() {
			return types.Bool(false), nil
		}
		right, err := e.eval(expr.Right, ctx)
		if err != nil {
			return types.Null(), err
		}
		return types.Bool(right.IsTruthy()), nil
	}

	if expr.Operator == "||" || expr.Operator == "or" || expr.Operator == "OR" {
		left, err := e.eval(expr.Left, ctx)
		if err != nil {
			return types.Null(), err
		}
		if left.IsTruthy() {
			return types.Bool(true), nil
		}
		right, err := e.eval(expr.Right, ctx)
		if err != nil {
			return types.Null(), err
		}
		return types.Bool(right.IsTruthy()), nil
	}

	// Evaluate both sides for other operators
	left, err := e.eval(expr.Left, ctx)
	if err != nil {
		return types.Null(), err
	}

	right, err := e.eval(expr.Right, ctx)
	if err != nil {
		return types.Null(), err
	}

	switch expr.Operator {
	// Comparison operators
	case "==":
		return types.Bool(left.Equals(right)), nil

	case "!=":
		return types.Bool(!left.Equals(right)), nil

	case "<":
		cmp, ok := left.Compare(right)
		if !ok {
			return types.Null(), errors.Newf(errors.ErrTypeMismatch,
				"cannot compare %s and %s", left.Type, right.Type)
		}
		return types.Bool(cmp < 0), nil

	case ">":
		cmp, ok := left.Compare(right)
		if !ok {
			return types.Null(), errors.Newf(errors.ErrTypeMismatch,
				"cannot compare %s and %s", left.Type, right.Type)
		}
		return types.Bool(cmp > 0), nil

	case "<=":
		cmp, ok := left.Compare(right)
		if !ok {
			return types.Null(), errors.Newf(errors.ErrTypeMismatch,
				"cannot compare %s and %s", left.Type, right.Type)
		}
		return types.Bool(cmp <= 0), nil

	case ">=":
		cmp, ok := left.Compare(right)
		if !ok {
			return types.Null(), errors.Newf(errors.ErrTypeMismatch,
				"cannot compare %s and %s", left.Type, right.Type)
		}
		return types.Bool(cmp >= 0), nil

	// Arithmetic operators
	case "+":
		return e.evalAddition(left, right)

	case "-":
		return e.evalSubtraction(left, right)

	case "*":
		return e.evalMultiplication(left, right)

	case "/":
		return e.evalDivision(left, right)

	case "%":
		return e.evalModulo(left, right)

	default:
		return types.Null(), errors.Newf(errors.ErrInvalidOperator,
			"unknown binary operator: %s", expr.Operator)
	}
}

func (e *Evaluator) evalAddition(left, right types.Value) (types.Value, error) {
	// String concatenation
	if left.Type == types.TypeString && right.Type == types.TypeString {
		l, _ := left.AsString()
		r, _ := right.AsString()
		return types.String(l + r), nil
	}

	// Numeric addition
	if left.Type.IsNumeric() && right.Type.IsNumeric() {
		if left.Type == types.TypeFloat || right.Type == types.TypeFloat {
			l, _ := left.AsFloat()
			r, _ := right.AsFloat()
			return types.Float(l + r), nil
		}
		l, _ := left.AsInt()
		r, _ := right.AsInt()
		return types.Int(l + r), nil
	}

	return types.Null(), errors.Newf(errors.ErrTypeMismatch,
		"cannot add %s and %s", left.Type, right.Type)
}

func (e *Evaluator) evalSubtraction(left, right types.Value) (types.Value, error) {
	if !left.Type.IsNumeric() || !right.Type.IsNumeric() {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"cannot subtract %s from %s", right.Type, left.Type)
	}

	if left.Type == types.TypeFloat || right.Type == types.TypeFloat {
		l, _ := left.AsFloat()
		r, _ := right.AsFloat()
		return types.Float(l - r), nil
	}

	l, _ := left.AsInt()
	r, _ := right.AsInt()
	return types.Int(l - r), nil
}

func (e *Evaluator) evalMultiplication(left, right types.Value) (types.Value, error) {
	if !left.Type.IsNumeric() || !right.Type.IsNumeric() {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"cannot multiply %s and %s", left.Type, right.Type)
	}

	if left.Type == types.TypeFloat || right.Type == types.TypeFloat {
		l, _ := left.AsFloat()
		r, _ := right.AsFloat()
		return types.Float(l * r), nil
	}

	l, _ := left.AsInt()
	r, _ := right.AsInt()
	return types.Int(l * r), nil
}

func (e *Evaluator) evalDivision(left, right types.Value) (types.Value, error) {
	if !left.Type.IsNumeric() || !right.Type.IsNumeric() {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"cannot divide %s by %s", left.Type, right.Type)
	}

	// Check for division by zero
	r, _ := right.AsFloat()
	if r == 0 {
		return types.Null(), errors.New(errors.ErrDivisionByZero, "division by zero")
	}

	l, _ := left.AsFloat()
	return types.Float(l / r), nil
}

func (e *Evaluator) evalModulo(left, right types.Value) (types.Value, error) {
	if left.Type != types.TypeInt || right.Type != types.TypeInt {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"modulo requires integers, got %s and %s", left.Type, right.Type)
	}

	r, _ := right.AsInt()
	if r == 0 {
		return types.Null(), errors.New(errors.ErrDivisionByZero, "modulo by zero")
	}

	l, _ := left.AsInt()
	return types.Int(l % r), nil
}

func (e *Evaluator) evalRegexExpression(re *ast.RegexExpression, ctx *EvalContext) (types.Value, error) {
	// Evaluate the left side (string to match)
	leftVal, err := e.eval(re.Left, ctx)
	if err != nil {
		return types.Null(), err
	}

	// Evaluate the pattern
	patternVal, err := e.eval(re.Pattern, ctx)
	if err != nil {
		return types.Null(), err
	}

	// Left must be a string
	leftStr, ok := leftVal.AsString()
	if !ok {
		// If left is null, return false (no match)
		if leftVal.IsNull() {
			return types.Bool(re.Negated), nil
		}
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "regex match requires string, got %s", leftVal.Type)
	}

	// Pattern must be a string
	patternStr, ok := patternVal.AsString()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "regex pattern must be string, got %s", patternVal.Type)
	}

	// Compile and match the regex
	re2, err := regexp.Compile(patternStr)
	if err != nil {
		return types.Null(), errors.Newf(errors.ErrInvalidSyntax, "invalid regex pattern: %v", err)
	}

	matched := re2.MatchString(leftStr)
	if re.Negated {
		matched = !matched
	}

	return types.Bool(matched), nil
}

func (e *Evaluator) evalInExpression(inExpr *ast.InExpression, ctx *EvalContext) (types.Value, error) {
	left, err := e.eval(inExpr.Left, ctx)
	if err != nil {
		return types.Null(), err
	}

	right, err := e.eval(inExpr.Right, ctx)
	if err != nil {
		return types.Null(), err
	}

	// Right must be a list
	list, ok := right.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"IN operator requires a list on the right side, got %s", right.Type)
	}

	// Check if left is in the list
	found := false
	for _, elem := range list {
		if left.Equals(elem) {
			found = true
			break
		}
	}

	if inExpr.Negated {
		return types.Bool(!found), nil
	}
	return types.Bool(found), nil
}

// ============================================================================
// Higher-order function evaluation (map, filter, reduce, find, some, every)
// ============================================================================

func (e *Evaluator) evalHigherOrderFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	switch call.Name {
	case "map":
		return e.evalMapFunction(call, ctx)
	case "filter":
		return e.evalFilterFunction(call, ctx)
	case "reduce":
		return e.evalReduceFunction(call, ctx)
	case "find":
		return e.evalFindFunction(call, ctx)
	case "some":
		return e.evalSomeFunction(call, ctx)
	case "every":
		return e.evalEveryFunction(call, ctx)
	default:
		return types.Null(), errors.Newf(errors.ErrUndefinedFunction, "unknown higher-order function: %s", call.Name)
	}
}

// evalMapFunction implements: map(list, x => expr) or map(list, "expr", "x")
func (e *Evaluator) evalMapFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	if len(call.Arguments) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "map() requires at least 2 arguments: list and lambda")
	}

	// Evaluate the list
	listVal, err := e.eval(call.Arguments[0], ctx)
	if err != nil {
		return types.Null(), err
	}

	list, ok := listVal.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "map() first argument must be a list, got %s", listVal.Type)
	}

	// Get the lambda or string expression
	lambda, paramName, err := e.extractLambda(call.Arguments[1], call.Arguments, 2)
	if err != nil {
		return types.Null(), err
	}

	// Apply the lambda to each element
	result := make([]types.Value, len(list))
	for i, elem := range list {
		// Set the variable in context
		ctx.SetVariable(paramName, elem)
		val, err := e.eval(lambda, ctx)
		if err != nil {
			return types.Null(), errors.Newf(errors.ErrFunctionPanic, "map() failed at index %d: %v", i, err)
		}
		result[i] = val
	}

	return types.List(result...), nil
}

// evalFilterFunction implements: filter(list, x => expr) or filter(list, "expr", "x")
func (e *Evaluator) evalFilterFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	if len(call.Arguments) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "filter() requires at least 2 arguments: list and lambda")
	}

	// Evaluate the list
	listVal, err := e.eval(call.Arguments[0], ctx)
	if err != nil {
		return types.Null(), err
	}

	list, ok := listVal.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "filter() first argument must be a list, got %s", listVal.Type)
	}

	// Get the lambda or string expression
	lambda, paramName, err := e.extractLambda(call.Arguments[1], call.Arguments, 2)
	if err != nil {
		return types.Null(), err
	}

	// Filter the list
	result := make([]types.Value, 0)
	for i, elem := range list {
		ctx.SetVariable(paramName, elem)
		val, err := e.eval(lambda, ctx)
		if err != nil {
			return types.Null(), errors.Newf(errors.ErrFunctionPanic, "filter() failed at index %d: %v", i, err)
		}
		if val.IsTruthy() {
			result = append(result, elem)
		}
	}

	return types.List(result...), nil
}

// evalReduceFunction implements: reduce(list, initial, (acc, x) => expr) or reduce(list, initial, "expr", "acc", "x")
func (e *Evaluator) evalReduceFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	if len(call.Arguments) < 3 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "reduce() requires at least 3 arguments: list, initial value, and lambda")
	}

	// Evaluate the list
	listVal, err := e.eval(call.Arguments[0], ctx)
	if err != nil {
		return types.Null(), err
	}

	list, ok := listVal.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "reduce() first argument must be a list, got %s", listVal.Type)
	}

	// Evaluate the initial value
	accumulator, err := e.eval(call.Arguments[1], ctx)
	if err != nil {
		return types.Null(), err
	}

	// Get the lambda - for reduce we need acc and x parameters
	lambda, accName, elemName, err := e.extractReduceLambda(call.Arguments[2], call.Arguments, 3)
	if err != nil {
		return types.Null(), err
	}

	// Reduce the list
	for i, elem := range list {
		ctx.SetVariable(accName, accumulator)
		ctx.SetVariable(elemName, elem)
		val, err := e.eval(lambda, ctx)
		if err != nil {
			return types.Null(), errors.Newf(errors.ErrFunctionPanic, "reduce() failed at index %d: %v", i, err)
		}
		accumulator = val
	}

	return accumulator, nil
}

// evalFindFunction implements: find(list, x => expr) - returns first matching element or null
func (e *Evaluator) evalFindFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	if len(call.Arguments) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "find() requires at least 2 arguments: list and lambda")
	}

	// Evaluate the list
	listVal, err := e.eval(call.Arguments[0], ctx)
	if err != nil {
		return types.Null(), err
	}

	list, ok := listVal.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "find() first argument must be a list, got %s", listVal.Type)
	}

	// Get the lambda or string expression
	lambda, paramName, err := e.extractLambda(call.Arguments[1], call.Arguments, 2)
	if err != nil {
		return types.Null(), err
	}

	// Find the first matching element
	for i, elem := range list {
		ctx.SetVariable(paramName, elem)
		val, err := e.eval(lambda, ctx)
		if err != nil {
			return types.Null(), errors.Newf(errors.ErrFunctionPanic, "find() failed at index %d: %v", i, err)
		}
		if val.IsTruthy() {
			return elem, nil
		}
	}

	return types.Null(), nil
}

// evalSomeFunction implements: some(list, x => expr) - returns true if any element matches
func (e *Evaluator) evalSomeFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	if len(call.Arguments) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "some() requires at least 2 arguments: list and lambda")
	}

	// Evaluate the list
	listVal, err := e.eval(call.Arguments[0], ctx)
	if err != nil {
		return types.Null(), err
	}

	list, ok := listVal.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "some() first argument must be a list, got %s", listVal.Type)
	}

	// Get the lambda or string expression
	lambda, paramName, err := e.extractLambda(call.Arguments[1], call.Arguments, 2)
	if err != nil {
		return types.Null(), err
	}

	// Check if any element matches
	for i, elem := range list {
		ctx.SetVariable(paramName, elem)
		val, err := e.eval(lambda, ctx)
		if err != nil {
			return types.Null(), errors.Newf(errors.ErrFunctionPanic, "some() failed at index %d: %v", i, err)
		}
		if val.IsTruthy() {
			return types.Bool(true), nil
		}
	}

	return types.Bool(false), nil
}

// evalEveryFunction implements: every(list, x => expr) - returns true if all elements match
func (e *Evaluator) evalEveryFunction(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	if len(call.Arguments) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "every() requires at least 2 arguments: list and lambda")
	}

	// Evaluate the list
	listVal, err := e.eval(call.Arguments[0], ctx)
	if err != nil {
		return types.Null(), err
	}

	list, ok := listVal.AsList()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch, "every() first argument must be a list, got %s", listVal.Type)
	}

	// Empty list returns true for every()
	if len(list) == 0 {
		return types.Bool(true), nil
	}

	// Get the lambda or string expression
	lambda, paramName, err := e.extractLambda(call.Arguments[1], call.Arguments, 2)
	if err != nil {
		return types.Null(), err
	}

	// Check if all elements match
	for i, elem := range list {
		ctx.SetVariable(paramName, elem)
		val, err := e.eval(lambda, ctx)
		if err != nil {
			return types.Null(), errors.Newf(errors.ErrFunctionPanic, "every() failed at index %d: %v", i, err)
		}
		if !val.IsTruthy() {
			return types.Bool(false), nil
		}
	}

	return types.Bool(true), nil
}

// extractLambda extracts the lambda expression and parameter name from a function argument
// It supports both lambda syntax (x => expr) and string syntax ("expr", "x")
func (e *Evaluator) extractLambda(arg ast.Expression, allArgs []ast.Expression, nextIdx int) (ast.Expression, string, error) {
	// Check if it's a lambda expression
	if lambda, ok := arg.(*ast.LambdaExpression); ok {
		if len(lambda.Parameters) != 1 {
			return nil, "", errors.New(errors.ErrArgumentCount, "lambda must have exactly 1 parameter")
		}
		return lambda.Body, lambda.Parameters[0].Value, nil
	}

	// Otherwise, it should be an expression to evaluate with variable "x" as default
	// Or we can have an optional parameter name as the next argument
	paramName := "x" // default parameter name

	// Check if next argument is a string literal for parameter name
	if nextIdx < len(allArgs) {
		if strLit, ok := allArgs[nextIdx].(*ast.StringLiteral); ok {
			paramName = strLit.Value
		}
	}

	return arg, paramName, nil
}

// extractReduceLambda extracts lambda for reduce function which needs two parameters (acc, x)
func (e *Evaluator) extractReduceLambda(arg ast.Expression, allArgs []ast.Expression, nextIdx int) (ast.Expression, string, string, error) {
	// Check if it's a lambda expression
	if lambda, ok := arg.(*ast.LambdaExpression); ok {
		if len(lambda.Parameters) != 2 {
			return nil, "", "", errors.New(errors.ErrArgumentCount, "reduce lambda must have exactly 2 parameters (accumulator, element)")
		}
		return lambda.Body, lambda.Parameters[0].Value, lambda.Parameters[1].Value, nil
	}

	// Otherwise use default parameter names or get from additional arguments
	accName := "acc"
	elemName := "x"

	// Check for custom parameter names
	if nextIdx < len(allArgs) {
		if strLit, ok := allArgs[nextIdx].(*ast.StringLiteral); ok {
			accName = strLit.Value
		}
	}
	if nextIdx+1 < len(allArgs) {
		if strLit, ok := allArgs[nextIdx+1].(*ast.StringLiteral); ok {
			elemName = strLit.Value
		}
	}

	return arg, accName, elemName, nil
}

// ============================================================================
// Function and member evaluation
// ============================================================================

func (e *Evaluator) evalFunctionCall(call *ast.FunctionCall, ctx *EvalContext) (types.Value, error) {
	// Evaluate arguments
	args := make([]types.Value, len(call.Arguments))
	for i, arg := range call.Arguments {
		val, err := e.eval(arg, ctx)
		if err != nil {
			return types.Null(), err
		}
		args[i] = val
	}

	// Check if this is a JS function that needs the sandbox
	fn, ok := e.functions.Get(call.Name)
	if ok && fn.IsJS() {
		if e.sandbox == nil {
			return types.Null(), errors.Newf(errors.ErrSandboxViolation,
				"cannot execute JS function '%s': sandbox not configured", call.Name)
		}
		return e.functions.CallJS(ctx.ctx, e.sandbox, call.Name, args)
	}

	// Call the built-in function
	return e.functions.Call(call.Name, args...)
}

func (e *Evaluator) evalIndexExpression(expr *ast.IndexExpression, ctx *EvalContext) (types.Value, error) {
	left, err := e.eval(expr.Left, ctx)
	if err != nil {
		return types.Null(), err
	}

	index, err := e.eval(expr.Index, ctx)
	if err != nil {
		return types.Null(), err
	}

	if left.Type != types.TypeList {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"cannot index %s", left.Type)
	}

	list, _ := left.AsList()
	idx, ok := index.AsInt()
	if !ok {
		return types.Null(), errors.Newf(errors.ErrTypeMismatch,
			"index must be an integer, got %s", index.Type)
	}

	// Handle negative indices
	if idx < 0 {
		idx = int64(len(list)) + idx
	}

	if idx < 0 || idx >= int64(len(list)) {
		return types.Null(), errors.New(errors.ErrIndexOutOfBounds, "index out of bounds")
	}

	return list[idx], nil
}

func (e *Evaluator) evalMemberExpression(expr *ast.MemberExpression, ctx *EvalContext) (types.Value, error) {
	// For member expressions on identifiers, we can convert to variable lookup
	// or treat as JSONPath-like access

	object, err := e.eval(expr.Object, ctx)
	if err != nil {
		return types.Null(), err
	}

	// If object is a map-like structure, access the property
	if object.Type == types.TypeAny {
		if m, ok := object.Raw.(map[string]interface{}); ok {
			if val, exists := m[expr.Property.Value]; exists {
				return types.NewValue(val), nil
			}
		}
	}

	return types.Null(), nil
}

// ============================================================================
// Helper functions
// ============================================================================

// gjsonToValue converts a gjson.Result to a types.Value.
func gjsonToValue(result gjson.Result) types.Value {
	switch result.Type {
	case gjson.Null:
		return types.Null()
	case gjson.False:
		return types.Bool(false)
	case gjson.True:
		return types.Bool(true)
	case gjson.Number:
		// Check if it's an integer
		if result.Num == float64(int64(result.Num)) {
			return types.Int(int64(result.Num))
		}
		return types.Float(result.Num)
	case gjson.String:
		return types.String(result.Str)
	case gjson.JSON:
		if result.IsArray() {
			arr := result.Array()
			elements := make([]types.Value, len(arr))
			for i, elem := range arr {
				elements[i] = gjsonToValue(elem)
			}
			return types.List(elements...)
		}
		// For objects, return as any
		return types.Any(result.Value())
	default:
		return types.Any(result.Value())
	}
}

// toJSON converts a value to JSON bytes.
func toJSON(v interface{}) ([]byte, error) {
	// Simple JSON encoder without external dependencies
	return marshalJSON(v)
}

// marshalJSON is a simple JSON marshaler.
func marshalJSON(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	switch val := v.(type) {
	case bool:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case int:
		return []byte(fmt.Sprintf("%d", val)), nil
	case int64:
		return []byte(fmt.Sprintf("%d", val)), nil
	case float64:
		return []byte(fmt.Sprintf("%v", val)), nil
	case string:
		return []byte(fmt.Sprintf("%q", val)), nil
	case []interface{}:
		result := []byte("[")
		for i, elem := range val {
			if i > 0 {
				result = append(result, ',')
			}
			b, err := marshalJSON(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, b...)
		}
		result = append(result, ']')
		return result, nil
	case map[string]interface{}:
		result := []byte("{")
		first := true
		for k, elem := range val {
			if !first {
				result = append(result, ',')
			}
			first = false
			result = append(result, fmt.Sprintf("%q:", k)...)
			b, err := marshalJSON(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, b...)
		}
		result = append(result, '}')
		return result, nil
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}
