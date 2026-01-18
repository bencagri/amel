// Package optimizer provides AST optimization for the AMEL DSL.
package optimizer

import (
	"github.com/bencagri/amel/pkg/ast"
	"github.com/bencagri/amel/pkg/lexer"
	"github.com/bencagri/amel/pkg/types"
)

// Optimizer performs various optimizations on the AST.
type Optimizer struct {
	foldConstants bool
}

// Option is a function that configures the optimizer.
type Option func(*Optimizer)

// WithConstantFolding enables or disables constant folding.
func WithConstantFolding(enabled bool) Option {
	return func(o *Optimizer) {
		o.foldConstants = enabled
	}
}

// New creates a new Optimizer with the given options.
func New(opts ...Option) *Optimizer {
	o := &Optimizer{
		foldConstants: true, // enabled by default
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Optimize performs all enabled optimizations on the AST.
func (o *Optimizer) Optimize(expr ast.Expression) ast.Expression {
	if o.foldConstants {
		expr = o.foldConstant(expr)
	}
	return expr
}

// foldConstant recursively folds constant expressions.
func (o *Optimizer) foldConstant(expr ast.Expression) ast.Expression {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return o.foldBinaryExpression(e)

	case *ast.UnaryExpression:
		return o.foldUnaryExpression(e)

	case *ast.ListLiteral:
		return o.foldListLiteral(e)

	case *ast.GroupedExpression:
		return o.foldGroupedExpression(e)

	case *ast.FunctionCall:
		return o.foldFunctionCall(e)

	case *ast.IndexExpression:
		return o.foldIndexExpression(e)

	case *ast.InExpression:
		return o.foldInExpression(e)

	default:
		// Literals, identifiers, and JSONPath expressions cannot be folded
		return expr
	}
}

// foldBinaryExpression folds binary expressions with constant operands.
func (o *Optimizer) foldBinaryExpression(expr *ast.BinaryExpression) ast.Expression {
	// First, recursively fold children
	left := o.foldConstant(expr.Left)
	right := o.foldConstant(expr.Right)

	// Check if both operands are now literals
	leftLit := getLiteralValue(left)
	rightLit := getLiteralValue(right)

	if leftLit == nil || rightLit == nil {
		// Can't fold, return with optimized children
		return &ast.BinaryExpression{
			Token:    expr.Token,
			Left:     left,
			Operator: expr.Operator,
			Right:    right,
		}
	}

	// Try to evaluate the operation
	result := evaluateBinaryOp(expr.Operator, leftLit, rightLit)
	if result == nil {
		// Can't fold this operation
		return &ast.BinaryExpression{
			Token:    expr.Token,
			Left:     left,
			Operator: expr.Operator,
			Right:    right,
		}
	}

	return valueToLiteral(result, expr.Token)
}

// foldUnaryExpression folds unary expressions with constant operands.
func (o *Optimizer) foldUnaryExpression(expr *ast.UnaryExpression) ast.Expression {
	operand := o.foldConstant(expr.Operand)

	lit := getLiteralValue(operand)
	if lit == nil {
		return &ast.UnaryExpression{
			Token:    expr.Token,
			Operator: expr.Operator,
			Operand:  operand,
		}
	}

	result := evaluateUnaryOp(expr.Operator, lit)
	if result == nil {
		return &ast.UnaryExpression{
			Token:    expr.Token,
			Operator: expr.Operator,
			Operand:  operand,
		}
	}

	return valueToLiteral(result, expr.Token)
}

// foldListLiteral folds list elements.
func (o *Optimizer) foldListLiteral(expr *ast.ListLiteral) ast.Expression {
	elements := make([]ast.Expression, len(expr.Elements))
	for i, elem := range expr.Elements {
		elements[i] = o.foldConstant(elem)
	}
	return &ast.ListLiteral{
		Token:    expr.Token,
		Elements: elements,
	}
}

// foldGroupedExpression folds the inner expression.
func (o *Optimizer) foldGroupedExpression(expr *ast.GroupedExpression) ast.Expression {
	inner := o.foldConstant(expr.Expression)

	// If the inner expression is a literal, we can remove the grouping
	if isLiteral(inner) {
		return inner
	}

	return &ast.GroupedExpression{
		Token:      expr.Token,
		Expression: inner,
	}
}

// foldFunctionCall folds function arguments (but doesn't evaluate the function).
func (o *Optimizer) foldFunctionCall(expr *ast.FunctionCall) ast.Expression {
	args := make([]ast.Expression, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		args[i] = o.foldConstant(arg)
	}
	return &ast.FunctionCall{
		Token:     expr.Token,
		Name:      expr.Name,
		Arguments: args,
	}
}

// foldIndexExpression folds index expressions.
func (o *Optimizer) foldIndexExpression(expr *ast.IndexExpression) ast.Expression {
	left := o.foldConstant(expr.Left)
	index := o.foldConstant(expr.Index)

	// Try to evaluate constant list indexing
	listLit, listOk := left.(*ast.ListLiteral)
	indexLit, indexOk := index.(*ast.IntegerLiteral)

	if listOk && indexOk {
		idx := int(indexLit.Value)
		if idx >= 0 && idx < len(listLit.Elements) {
			return listLit.Elements[idx]
		}
	}

	return &ast.IndexExpression{
		Token: expr.Token,
		Left:  left,
		Index: index,
	}
}

// foldInExpression folds IN expressions.
func (o *Optimizer) foldInExpression(expr *ast.InExpression) ast.Expression {
	left := o.foldConstant(expr.Left)
	right := o.foldConstant(expr.Right)

	// Try to evaluate constant IN expressions
	leftVal := getLiteralValue(left)
	rightList, rightOk := right.(*ast.ListLiteral)

	if leftVal != nil && rightOk && areAllLiterals(rightList.Elements) {
		found := false
		for _, elem := range rightList.Elements {
			elemVal := getLiteralValue(elem)
			if elemVal != nil && valuesEqual(leftVal, elemVal) {
				found = true
				break
			}
		}

		result := found
		if expr.Negated {
			result = !result
		}

		return &ast.BooleanLiteral{
			Token: expr.Token,
			Value: result,
		}
	}

	return &ast.InExpression{
		Token:   expr.Token,
		Left:    left,
		Right:   right,
		Negated: expr.Negated,
	}
}

// getLiteralValue extracts the Go value from a literal expression.
func getLiteralValue(expr ast.Expression) interface{} {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value
	case *ast.FloatLiteral:
		return e.Value
	case *ast.StringLiteral:
		return e.Value
	case *ast.BooleanLiteral:
		return e.Value
	case *ast.NullLiteral:
		return nil
	default:
		return nil
	}
}

// isLiteral checks if an expression is a literal.
func isLiteral(expr ast.Expression) bool {
	switch expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.NullLiteral:
		return true
	default:
		return false
	}
}

// areAllLiterals checks if all expressions in a slice are literals.
func areAllLiterals(exprs []ast.Expression) bool {
	for _, expr := range exprs {
		if !isLiteral(expr) {
			return false
		}
	}
	return true
}

// valuesEqual compares two literal values for equality.
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle numeric comparisons with type promotion
	switch av := a.(type) {
	case int64:
		switch bv := b.(type) {
		case int64:
			return av == bv
		case float64:
			return float64(av) == bv
		}
	case float64:
		switch bv := b.(type) {
		case int64:
			return av == float64(bv)
		case float64:
			return av == bv
		}
	}

	return a == b
}

// evaluateBinaryOp evaluates a binary operation on two constant values.
func evaluateBinaryOp(op string, left, right interface{}) interface{} {
	switch op {
	case "+":
		return evalAdd(left, right)
	case "-":
		return evalSub(left, right)
	case "*":
		return evalMul(left, right)
	case "/":
		return evalDiv(left, right)
	case "%":
		return evalMod(left, right)
	case "==":
		return valuesEqual(left, right)
	case "!=":
		return !valuesEqual(left, right)
	case "<":
		return evalLT(left, right)
	case ">":
		return evalGT(left, right)
	case "<=":
		return evalLTE(left, right)
	case ">=":
		return evalGTE(left, right)
	case "&&", "AND":
		return evalAnd(left, right)
	case "||", "OR":
		return evalOr(left, right)
	default:
		return nil
	}
}

// evaluateUnaryOp evaluates a unary operation on a constant value.
func evaluateUnaryOp(op string, operand interface{}) interface{} {
	switch op {
	case "!":
		if b, ok := operand.(bool); ok {
			return !b
		}
	case "-":
		switch v := operand.(type) {
		case int64:
			return -v
		case float64:
			return -v
		}
	}
	return nil
}

// valueToLiteral converts a Go value to an AST literal.
func valueToLiteral(val interface{}, token lexer.Token) ast.Expression {
	switch v := val.(type) {
	case int64:
		return &ast.IntegerLiteral{Token: token, Value: v}
	case float64:
		return &ast.FloatLiteral{Token: token, Value: v}
	case string:
		return &ast.StringLiteral{Token: token, Value: v}
	case bool:
		return &ast.BooleanLiteral{Token: token, Value: v}
	case nil:
		return &ast.NullLiteral{Token: token}
	default:
		return nil
	}
}

// Arithmetic operations
func evalAdd(left, right interface{}) interface{} {
	switch lv := left.(type) {
	case int64:
		switch rv := right.(type) {
		case int64:
			return lv + rv
		case float64:
			return float64(lv) + rv
		}
	case float64:
		switch rv := right.(type) {
		case int64:
			return lv + float64(rv)
		case float64:
			return lv + rv
		}
	case string:
		if rv, ok := right.(string); ok {
			return lv + rv
		}
	}
	return nil
}

func evalSub(left, right interface{}) interface{} {
	switch lv := left.(type) {
	case int64:
		switch rv := right.(type) {
		case int64:
			return lv - rv
		case float64:
			return float64(lv) - rv
		}
	case float64:
		switch rv := right.(type) {
		case int64:
			return lv - float64(rv)
		case float64:
			return lv - rv
		}
	}
	return nil
}

func evalMul(left, right interface{}) interface{} {
	switch lv := left.(type) {
	case int64:
		switch rv := right.(type) {
		case int64:
			return lv * rv
		case float64:
			return float64(lv) * rv
		}
	case float64:
		switch rv := right.(type) {
		case int64:
			return lv * float64(rv)
		case float64:
			return lv * rv
		}
	}
	return nil
}

func evalDiv(left, right interface{}) interface{} {
	switch lv := left.(type) {
	case int64:
		switch rv := right.(type) {
		case int64:
			if rv == 0 {
				return nil // Division by zero
			}
			return lv / rv
		case float64:
			if rv == 0 {
				return nil
			}
			return float64(lv) / rv
		}
	case float64:
		switch rv := right.(type) {
		case int64:
			if rv == 0 {
				return nil
			}
			return lv / float64(rv)
		case float64:
			if rv == 0 {
				return nil
			}
			return lv / rv
		}
	}
	return nil
}

func evalMod(left, right interface{}) interface{} {
	if lv, lok := left.(int64); lok {
		if rv, rok := right.(int64); rok {
			if rv == 0 {
				return nil
			}
			return lv % rv
		}
	}
	return nil
}

// Comparison operations
func evalLT(left, right interface{}) interface{} {
	cmp, ok := compare(left, right)
	if !ok {
		return nil
	}
	return cmp < 0
}

func evalGT(left, right interface{}) interface{} {
	cmp, ok := compare(left, right)
	if !ok {
		return nil
	}
	return cmp > 0
}

func evalLTE(left, right interface{}) interface{} {
	cmp, ok := compare(left, right)
	if !ok {
		return nil
	}
	return cmp <= 0
}

func evalGTE(left, right interface{}) interface{} {
	cmp, ok := compare(left, right)
	if !ok {
		return nil
	}
	return cmp >= 0
}

// compare returns -1, 0, or 1 and whether comparison was successful.
func compare(left, right interface{}) (int, bool) {
	switch lv := left.(type) {
	case int64:
		switch rv := right.(type) {
		case int64:
			if lv < rv {
				return -1, true
			} else if lv > rv {
				return 1, true
			}
			return 0, true
		case float64:
			lf := float64(lv)
			if lf < rv {
				return -1, true
			} else if lf > rv {
				return 1, true
			}
			return 0, true
		}
	case float64:
		switch rv := right.(type) {
		case int64:
			rf := float64(rv)
			if lv < rf {
				return -1, true
			} else if lv > rf {
				return 1, true
			}
			return 0, true
		case float64:
			if lv < rv {
				return -1, true
			} else if lv > rv {
				return 1, true
			}
			return 0, true
		}
	case string:
		if rv, ok := right.(string); ok {
			if lv < rv {
				return -1, true
			} else if lv > rv {
				return 1, true
			}
			return 0, true
		}
	}
	return 0, false
}

// Logical operations
func evalAnd(left, right interface{}) interface{} {
	lb, lok := left.(bool)
	rb, rok := right.(bool)
	if lok && rok {
		return lb && rb
	}
	// Short-circuit: if left is false, result is false
	if lok && !lb {
		return false
	}
	return nil
}

func evalOr(left, right interface{}) interface{} {
	lb, lok := left.(bool)
	rb, rok := right.(bool)
	if lok && rok {
		return lb || rb
	}
	// Short-circuit: if left is true, result is true
	if lok && lb {
		return true
	}
	return nil
}

// OptimizeForCache is a helper that returns an optimized expression suitable for caching.
// It performs all safe optimizations that don't depend on runtime context.
func OptimizeForCache(expr ast.Expression) ast.Expression {
	opt := New()
	return opt.Optimize(expr)
}

// Stats holds statistics about optimizations performed.
type Stats struct {
	ConstantsFolded  int
	ExpressionsTotal int
}

// OptimizeWithStats performs optimization and returns statistics.
func (o *Optimizer) OptimizeWithStats(expr ast.Expression) (ast.Expression, *Stats) {
	stats := &Stats{}
	result := o.optimizeWithStats(expr, stats)
	return result, stats
}

func (o *Optimizer) optimizeWithStats(expr ast.Expression, stats *Stats) ast.Expression {
	stats.ExpressionsTotal++

	if !o.foldConstants {
		return expr
	}

	switch e := expr.(type) {
	case *ast.BinaryExpression:
		left := o.optimizeWithStats(e.Left, stats)
		right := o.optimizeWithStats(e.Right, stats)

		leftLit := getLiteralValue(left)
		rightLit := getLiteralValue(right)

		if leftLit != nil && rightLit != nil {
			result := evaluateBinaryOp(e.Operator, leftLit, rightLit)
			if result != nil {
				stats.ConstantsFolded++
				return valueToLiteral(result, e.Token)
			}
		}

		return &ast.BinaryExpression{
			Token:    e.Token,
			Left:     left,
			Operator: e.Operator,
			Right:    right,
		}

	case *ast.UnaryExpression:
		operand := o.optimizeWithStats(e.Operand, stats)

		lit := getLiteralValue(operand)
		if lit != nil {
			result := evaluateUnaryOp(e.Operator, lit)
			if result != nil {
				stats.ConstantsFolded++
				return valueToLiteral(result, e.Token)
			}
		}

		return &ast.UnaryExpression{
			Token:    e.Token,
			Operator: e.Operator,
			Operand:  operand,
		}

	case *ast.ListLiteral:
		elements := make([]ast.Expression, len(e.Elements))
		for i, elem := range e.Elements {
			elements[i] = o.optimizeWithStats(elem, stats)
		}
		return &ast.ListLiteral{
			Token:    e.Token,
			Elements: elements,
		}

	case *ast.GroupedExpression:
		inner := o.optimizeWithStats(e.Expression, stats)
		if isLiteral(inner) {
			stats.ConstantsFolded++
			return inner
		}
		return &ast.GroupedExpression{
			Token:      e.Token,
			Expression: inner,
		}

	case *ast.FunctionCall:
		args := make([]ast.Expression, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = o.optimizeWithStats(arg, stats)
		}
		return &ast.FunctionCall{
			Token:     e.Token,
			Name:      e.Name,
			Arguments: args,
		}

	case *ast.IndexExpression:
		left := o.optimizeWithStats(e.Left, stats)
		index := o.optimizeWithStats(e.Index, stats)

		listLit, listOk := left.(*ast.ListLiteral)
		indexLit, indexOk := index.(*ast.IntegerLiteral)

		if listOk && indexOk {
			idx := int(indexLit.Value)
			if idx >= 0 && idx < len(listLit.Elements) {
				stats.ConstantsFolded++
				return listLit.Elements[idx]
			}
		}

		return &ast.IndexExpression{
			Token: e.Token,
			Left:  left,
			Index: index,
		}

	case *ast.InExpression:
		left := o.optimizeWithStats(e.Left, stats)
		right := o.optimizeWithStats(e.Right, stats)

		leftVal := getLiteralValue(left)
		rightList, rightOk := right.(*ast.ListLiteral)

		if leftVal != nil && rightOk && areAllLiterals(rightList.Elements) {
			found := false
			for _, elem := range rightList.Elements {
				elemVal := getLiteralValue(elem)
				if elemVal != nil && valuesEqual(leftVal, elemVal) {
					found = true
					break
				}
			}

			result := found
			if e.Negated {
				result = !result
			}

			stats.ConstantsFolded++
			return &ast.BooleanLiteral{
				Token: e.Token,
				Value: result,
			}
		}

		return &ast.InExpression{
			Token:   e.Token,
			Left:    left,
			Right:   right,
			Negated: e.Negated,
		}

	default:
		return expr
	}
}

// IsConstant checks if an expression is entirely constant (can be evaluated at compile time).
func IsConstant(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral,
		*ast.BooleanLiteral, *ast.NullLiteral:
		return true

	case *ast.ListLiteral:
		return areAllLiterals(e.Elements)

	case *ast.BinaryExpression:
		return IsConstant(e.Left) && IsConstant(e.Right)

	case *ast.UnaryExpression:
		return IsConstant(e.Operand)

	case *ast.GroupedExpression:
		return IsConstant(e.Expression)

	case *ast.InExpression:
		return IsConstant(e.Left) && IsConstant(e.Right)

	default:
		// Identifiers, JSONPath, function calls, etc. are not constant
		return false
	}
}

// EvaluateConstant evaluates a constant expression and returns the result.
// Returns nil if the expression cannot be evaluated as a constant.
func EvaluateConstant(expr ast.Expression) *types.Value {
	opt := New()
	optimized := opt.Optimize(expr)

	val := getLiteralValue(optimized)
	if val == nil && optimized != nil {
		// Check for null literal
		if _, ok := optimized.(*ast.NullLiteral); ok {
			result := types.Null()
			return &result
		}
		return nil
	}

	result := types.NewValue(val)
	return &result
}
