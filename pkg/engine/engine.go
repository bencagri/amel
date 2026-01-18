// Package engine provides the main AMEL engine facade.
package engine

import (
	"time"

	"github.com/bencagri/amel/pkg/ast"
	"github.com/bencagri/amel/pkg/eval"
	"github.com/bencagri/amel/pkg/functions"
	"github.com/bencagri/amel/pkg/optimizer"
	"github.com/bencagri/amel/pkg/parser"
	"github.com/bencagri/amel/pkg/types"
)

// Engine is the main AMEL DSL engine.
type Engine struct {
	evaluator       *eval.Evaluator
	functions       *functions.Registry
	sandbox         *functions.Sandbox
	optimizer       *optimizer.Optimizer
	timeout         time.Duration
	explainMode     bool
	strictTypes     bool
	caching         bool
	optimizeEnabled bool
	cache           map[string]*CompiledExpression
}

// CompiledExpression represents a pre-parsed expression ready for evaluation.
type CompiledExpression struct {
	AST       ast.Expression
	Optimized ast.Expression
	Source    string
}

// Result represents the result of an evaluation.
type Result struct {
	Value       types.Value
	Explanation *eval.Explanation
}

// Option is a function that configures the engine.
type Option func(*Engine)

// WithTimeout sets the evaluation timeout.
func WithTimeout(d time.Duration) Option {
	return func(e *Engine) {
		e.timeout = d
	}
}

// WithExplainMode enables explanation generation.
func WithExplainMode(enabled bool) Option {
	return func(e *Engine) {
		e.explainMode = enabled
	}
}

// WithStrictTypes enables strict type checking.
func WithStrictTypes(enabled bool) Option {
	return func(e *Engine) {
		e.strictTypes = enabled
	}
}

// WithCaching enables expression caching.
func WithCaching(enabled bool) Option {
	return func(e *Engine) {
		e.caching = enabled
		if enabled && e.cache == nil {
			e.cache = make(map[string]*CompiledExpression)
		}
	}
}

// WithOptimization enables AST optimization (constant folding).
func WithOptimization(enabled bool) Option {
	return func(e *Engine) {
		e.optimizeEnabled = enabled
	}
}

// WithFunctions sets a custom function registry.
func WithFunctions(r *functions.Registry) Option {
	return func(e *Engine) {
		e.functions = r
	}
}

// WithSandbox sets a custom JavaScript sandbox.
func WithSandbox(s *functions.Sandbox) Option {
	return func(e *Engine) {
		e.sandbox = s
	}
}

// WithSandboxConfig sets sandbox configuration.
func WithSandboxConfig(config *functions.SandboxConfig) Option {
	return func(e *Engine) {
		e.sandbox = functions.NewSandbox(config)
	}
}

// New creates a new AMEL engine with the given options.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{
		timeout:         100 * time.Millisecond,
		optimizeEnabled: true, // enabled by default
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

	// Create default sandbox if not provided
	if e.sandbox == nil {
		e.sandbox = functions.NewSandbox(&functions.SandboxConfig{
			Timeout:       e.timeout,
			MemoryLimit:   10 * 1024 * 1024, // 10MB
			MaxStackDepth: 100,
		})
	}

	// Create optimizer if optimization is enabled
	if e.optimizeEnabled {
		e.optimizer = optimizer.New(optimizer.WithConstantFolding(true))
	}

	// Create evaluator with sandbox support
	evaluator, err := eval.New(
		eval.WithFunctions(e.functions),
		eval.WithTimeout(e.timeout),
		eval.WithSandbox(e.sandbox),
	)
	if err != nil {
		return nil, err
	}
	e.evaluator = evaluator

	return e, nil
}

// Compile parses a DSL expression and returns a compiled expression.
func (e *Engine) Compile(dsl string) (*CompiledExpression, error) {
	// Check cache
	if e.caching {
		if cached, ok := e.cache[dsl]; ok {
			return cached, nil
		}
	}

	// Parse the expression
	expr, err := parser.Parse(dsl)
	if err != nil {
		return nil, err
	}

	// Optimize the AST if optimizer is available
	var optimized ast.Expression
	if e.optimizer != nil {
		optimized = e.optimizer.Optimize(expr)
	} else {
		optimized = expr
	}

	compiled := &CompiledExpression{
		AST:       expr,
		Optimized: optimized,
		Source:    dsl,
	}

	// Store in cache
	if e.caching {
		e.cache[dsl] = compiled
	}

	return compiled, nil
}

// Evaluate evaluates a compiled expression against a payload.
func (e *Engine) Evaluate(expr *CompiledExpression, payload interface{}) (types.Value, error) {
	ctx, err := eval.NewContext(payload)
	if err != nil {
		return types.Null(), err
	}

	// Use optimized AST if available
	astToEval := expr.Optimized
	if astToEval == nil {
		astToEval = expr.AST
	}

	return e.evaluator.Evaluate(astToEval, ctx)
}

// EvaluateWithExplanation evaluates an expression and returns detailed explanation.
// Note: Uses the original AST (not optimized) for better explanation accuracy.
func (e *Engine) EvaluateWithExplanation(expr *CompiledExpression, payload interface{}) (types.Value, *eval.Explanation, error) {
	ctx, err := eval.NewContext(payload)
	if err != nil {
		return types.Null(), nil, err
	}

	// Use original AST for explanations to show the full expression tree
	return e.evaluator.EvaluateWithExplanation(expr.AST, ctx)
}

// EvaluateBool evaluates a compiled expression and returns a boolean result.
func (e *Engine) EvaluateBool(expr *CompiledExpression, payload interface{}) (bool, error) {
	ctx, err := eval.NewContext(payload)
	if err != nil {
		return false, err
	}

	// Use optimized AST if available
	astToEval := expr.Optimized
	if astToEval == nil {
		astToEval = expr.AST
	}

	return e.evaluator.EvaluateBool(astToEval, ctx)
}

// EvaluateDirect compiles and evaluates an expression in one step.
func (e *Engine) EvaluateDirect(dsl string, payload interface{}) (types.Value, error) {
	compiled, err := e.Compile(dsl)
	if err != nil {
		return types.Null(), err
	}

	return e.Evaluate(compiled, payload)
}

// EvaluateDirectBool compiles and evaluates an expression, returning a boolean.
func (e *Engine) EvaluateDirectBool(dsl string, payload interface{}) (bool, error) {
	compiled, err := e.Compile(dsl)
	if err != nil {
		return false, err
	}

	return e.EvaluateBool(compiled, payload)
}

// RegisterFunction registers a user-defined JavaScript function.
// The source should be in the format: function name(params): returnType { body }
func (e *Engine) RegisterFunction(source string) error {
	return e.functions.RegisterJSFunction(source, e.sandbox)
}

// RegisterBuiltIn registers a built-in Go function.
func (e *Engine) RegisterBuiltIn(name string, fn functions.BuiltInFunc, sig *types.FunctionSignature) error {
	return e.functions.RegisterBuiltIn(name, fn, sig)
}

// ClearCache clears the expression cache.
func (e *Engine) ClearCache() {
	if e.cache != nil {
		e.cache = make(map[string]*CompiledExpression)
	}
}

// GetFunctionRegistry returns the function registry.
func (e *Engine) GetFunctionRegistry() *functions.Registry {
	return e.functions
}

// GetSandbox returns the JavaScript sandbox.
func (e *Engine) GetSandbox() *functions.Sandbox {
	return e.sandbox
}

// GetOptimizer returns the AST optimizer.
func (e *Engine) GetOptimizer() *optimizer.Optimizer {
	return e.optimizer
}

// ListFunctions returns all registered function names.
func (e *Engine) ListFunctions() []string {
	return e.functions.List()
}

// ============================================================================
// Input/Output structures for JSON API
// ============================================================================

// EvalRequest represents an evaluation request.
type EvalRequest struct {
	Payload   interface{} `json:"payload"`
	DSL       string      `json:"dsl"`
	Functions []string    `json:"functions,omitempty"`
}

// EvalResponse represents an evaluation response.
type EvalResponse struct {
	Result      interface{}       `json:"result"`
	Type        string            `json:"type"`
	Explanation *eval.Explanation `json:"explanation,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// EvaluateRequest evaluates a request and returns a response.
func (e *Engine) EvaluateRequest(req *EvalRequest) *EvalResponse {
	resp := &EvalResponse{}

	// Register any custom functions
	for _, fnSrc := range req.Functions {
		if err := e.RegisterFunction(fnSrc); err != nil {
			resp.Error = err.Error()
			return resp
		}
	}

	// Compile the expression
	compiled, err := e.Compile(req.DSL)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	// Evaluate
	if e.explainMode {
		value, explanation, err := e.EvaluateWithExplanation(compiled, req.Payload)
		if err != nil {
			resp.Error = err.Error()
			return resp
		}
		resp.Result = value.Raw
		resp.Type = value.Type.String()
		resp.Explanation = explanation
	} else {
		value, err := e.Evaluate(compiled, req.Payload)
		if err != nil {
			resp.Error = err.Error()
			return resp
		}
		resp.Result = value.Raw
		resp.Type = value.Type.String()
	}

	return resp
}

// ============================================================================
// Convenience functions
// ============================================================================

// Eval is a convenience function for one-off evaluation.
func Eval(dsl string, payload interface{}) (types.Value, error) {
	engine, err := New()
	if err != nil {
		return types.Null(), err
	}
	return engine.EvaluateDirect(dsl, payload)
}

// EvalBool is a convenience function for one-off boolean evaluation.
func EvalBool(dsl string, payload interface{}) (bool, error) {
	engine, err := New()
	if err != nil {
		return false, err
	}
	return engine.EvaluateDirectBool(dsl, payload)
}

// MustEval is like Eval but panics on error.
func MustEval(dsl string, payload interface{}) types.Value {
	result, err := Eval(dsl, payload)
	if err != nil {
		panic(err)
	}
	return result
}

// MustEvalBool is like EvalBool but panics on error.
func MustEvalBool(dsl string, payload interface{}) bool {
	result, err := EvalBool(dsl, payload)
	if err != nil {
		panic(err)
	}
	return result
}
