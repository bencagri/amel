// Package functions provides function management for the AMEL DSL engine.
package functions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/types"
	"github.com/dop251/goja"
)

// SandboxConfig defines configuration for the JavaScript sandbox.
type SandboxConfig struct {
	Timeout       time.Duration // Maximum execution time
	MemoryLimit   int64         // Maximum memory in bytes (informational, not enforced by goja)
	MaxStackDepth int           // Maximum call stack depth
	AllowedAPIs   []string      // List of allowed global APIs
}

// DefaultSandboxConfig returns the default sandbox configuration.
func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		Timeout:       100 * time.Millisecond,
		MemoryLimit:   10 * 1024 * 1024, // 10MB
		MaxStackDepth: 100,
		AllowedAPIs:   []string{"Math", "JSON", "Array", "Object", "String", "Number", "Boolean", "Date", "RegExp"},
	}
}

// Sandbox provides a secure JavaScript execution environment.
type Sandbox struct {
	config *SandboxConfig
	pool   *vmPool
}

// vmPool manages a pool of goja VMs for reuse.
type vmPool struct {
	mu   sync.Mutex
	vms  []*goja.Runtime
	max  int
	init func(*goja.Runtime)
}

// newVMPool creates a new VM pool.
func newVMPool(max int, initFn func(*goja.Runtime)) *vmPool {
	return &vmPool{
		vms:  make([]*goja.Runtime, 0, max),
		max:  max,
		init: initFn,
	}
}

// acquire gets a VM from the pool or creates a new one.
func (p *vmPool) acquire() *goja.Runtime {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.vms) > 0 {
		vm := p.vms[len(p.vms)-1]
		p.vms = p.vms[:len(p.vms)-1]
		return vm
	}

	vm := goja.New()
	if p.init != nil {
		p.init(vm)
	}
	return vm
}

// release returns a VM to the pool.
func (p *vmPool) release(vm *goja.Runtime) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.vms) < p.max {
		p.vms = append(p.vms, vm)
	}
	// If pool is full, let the VM be garbage collected
}

// NewSandbox creates a new JavaScript sandbox with the given configuration.
func NewSandbox(config *SandboxConfig) *Sandbox {
	if config == nil {
		config = DefaultSandboxConfig()
	}

	s := &Sandbox{
		config: config,
	}

	s.pool = newVMPool(10, s.initVM)
	return s
}

// initVM initializes a VM with sandbox restrictions.
func (s *Sandbox) initVM(vm *goja.Runtime) {
	// Remove dangerous globals
	restrictedGlobals := []string{
		"eval",
		"Function",
		"require",
		"module",
		"exports",
		"process",
		"global",
		"globalThis",
		"setTimeout",
		"setInterval",
		"setImmediate",
		"clearTimeout",
		"clearInterval",
		"clearImmediate",
		"XMLHttpRequest",
		"fetch",
		"WebSocket",
	}

	for _, name := range restrictedGlobals {
		_ = vm.Set(name, goja.Undefined())
	}

	// Add safe console implementation (limited logging)
	console := map[string]interface{}{
		"log":   func(args ...interface{}) {},
		"info":  func(args ...interface{}) {},
		"warn":  func(args ...interface{}) {},
		"error": func(args ...interface{}) {},
		"debug": func(args ...interface{}) {},
	}
	_ = vm.Set("console", console)

	// Set stack depth limit
	vm.SetMaxCallStackSize(s.config.MaxStackDepth)
}

// Execute runs a JavaScript function with the given arguments.
func (s *Sandbox) Execute(ctx context.Context, jsBody string, funcName string, args []types.Value) (types.Value, error) {
	vm := s.pool.acquire()
	defer s.pool.release(vm)

	// Re-initialize VM to ensure clean state
	s.initVM(vm)

	// Set up interrupt for timeout
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt("execution timeout")
		case <-time.After(s.config.Timeout):
			vm.Interrupt("execution timeout")
		case <-done:
			// Execution completed normally
		}
	}()

	// Convert arguments to JavaScript values
	jsArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		jsArgs[i] = s.valueToJS(vm, arg)
	}

	// Run the function definition
	_, err := vm.RunString(jsBody)
	if err != nil {
		return types.Null(), errors.Wrap(errors.ErrSandboxViolation, fmt.Sprintf("failed to compile JS function: %v", err), err)
	}

	// Get the function
	fn := vm.Get(funcName)
	if fn == nil || goja.IsUndefined(fn) || goja.IsNull(fn) {
		return types.Null(), errors.Newf(errors.ErrUndefinedFunction, "function '%s' not found in JS code", funcName)
	}

	callable, ok := goja.AssertFunction(fn)
	if !ok {
		return types.Null(), errors.Newf(errors.ErrInvalidSyntax, "'%s' is not a function", funcName)
	}

	// Call the function
	result, err := callable(goja.Undefined(), jsArgs...)
	if err != nil {
		if jsErr, ok := err.(*goja.InterruptedError); ok {
			return types.Null(), errors.Newf(errors.ErrTimeout, "JS execution interrupted: %v", jsErr.Value())
		}
		return types.Null(), errors.Wrap(errors.ErrSandboxViolation, fmt.Sprintf("JS execution failed: %v", err), err)
	}

	// Convert result back to AMEL Value
	return s.jsToValue(result), nil
}

// ExecuteExpression runs a JavaScript expression and returns the result.
func (s *Sandbox) ExecuteExpression(ctx context.Context, expr string) (types.Value, error) {
	vm := s.pool.acquire()
	defer s.pool.release(vm)

	// Re-initialize VM
	s.initVM(vm)

	// Set up interrupt
	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt("execution timeout")
		case <-time.After(s.config.Timeout):
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()

	result, err := vm.RunString(expr)
	if err != nil {
		if jsErr, ok := err.(*goja.InterruptedError); ok {
			return types.Null(), errors.Newf(errors.ErrTimeout, "JS execution interrupted: %v", jsErr.Value())
		}
		return types.Null(), errors.Wrap(errors.ErrSandboxViolation, fmt.Sprintf("JS expression failed: %v", err), err)
	}

	return s.jsToValue(result), nil
}

// valueToJS converts an AMEL Value to a goja Value.
func (s *Sandbox) valueToJS(vm *goja.Runtime, v types.Value) goja.Value {
	if v.IsNull() {
		return goja.Null()
	}

	switch v.Type {
	case types.TypeInt:
		return vm.ToValue(v.Raw.(int64))
	case types.TypeFloat:
		return vm.ToValue(v.Raw.(float64))
	case types.TypeString:
		return vm.ToValue(v.Raw.(string))
	case types.TypeBool:
		return vm.ToValue(v.Raw.(bool))
	case types.TypeList:
		list, ok := v.AsList()
		if !ok {
			return goja.Null()
		}
		arr := make([]interface{}, len(list))
		for i, elem := range list {
			arr[i] = s.valueToJS(vm, elem).Export()
		}
		return vm.ToValue(arr)
	default:
		return vm.ToValue(v.Raw)
	}
}

// jsToValue converts a goja Value to an AMEL Value.
func (s *Sandbox) jsToValue(v goja.Value) types.Value {
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return types.Null()
	}

	exported := v.Export()
	if exported == nil {
		return types.Null()
	}

	switch val := exported.(type) {
	case int64:
		return types.Int(val)
	case int:
		return types.Int(int64(val))
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return types.Int(int64(val))
		}
		return types.Float(val)
	case string:
		return types.String(val)
	case bool:
		return types.Bool(val)
	case []interface{}:
		elements := make([]types.Value, len(val))
		for i, elem := range val {
			elements[i] = types.NewValue(elem)
		}
		return types.List(elements...)
	case map[string]interface{}:
		// Return as Any type for objects
		return types.Any(val)
	default:
		return types.NewValue(exported)
	}
}

// SetTimeout updates the execution timeout.
func (s *Sandbox) SetTimeout(d time.Duration) {
	s.config.Timeout = d
}

// SetMemoryLimit updates the memory limit (informational).
func (s *Sandbox) SetMemoryLimit(bytes int64) {
	s.config.MemoryLimit = bytes
}

// SetMaxStackDepth updates the maximum call stack depth.
func (s *Sandbox) SetMaxStackDepth(depth int) {
	s.config.MaxStackDepth = depth
}

// Config returns the current sandbox configuration.
func (s *Sandbox) Config() *SandboxConfig {
	return s.config
}

// ParseJSFunction parses a JavaScript function definition and extracts metadata.
// Expected format: function name(param1, param2): returnType { body }
// The return type annotation is optional and uses a colon prefix.
func ParseJSFunction(source string) (name string, params []string, returnType types.Type, body string, err error) {
	// This is a simplified parser. For production, consider using a proper JS parser.
	// Format: function name(params) { body } or function name(params): type { body }

	// Find "function" keyword
	const funcKeyword = "function "
	idx := 0
	for ; idx < len(source) && (source[idx] == ' ' || source[idx] == '\t' || source[idx] == '\n'); idx++ {
	}

	if idx+len(funcKeyword) > len(source) || source[idx:idx+len(funcKeyword)] != funcKeyword {
		return "", nil, types.TypeAny, "", errors.New(errors.ErrInvalidSyntax, "JS function must start with 'function' keyword")
	}
	idx += len(funcKeyword)

	// Skip whitespace
	for ; idx < len(source) && (source[idx] == ' ' || source[idx] == '\t'); idx++ {
	}

	// Extract function name
	nameStart := idx
	for ; idx < len(source) && isIdentChar(source[idx]); idx++ {
	}
	name = source[nameStart:idx]
	if name == "" {
		return "", nil, types.TypeAny, "", errors.New(errors.ErrInvalidSyntax, "JS function must have a name")
	}

	// Skip whitespace
	for ; idx < len(source) && (source[idx] == ' ' || source[idx] == '\t'); idx++ {
	}

	// Expect '('
	if idx >= len(source) || source[idx] != '(' {
		return "", nil, types.TypeAny, "", errors.New(errors.ErrInvalidSyntax, "expected '(' after function name")
	}
	idx++

	// Extract parameters
	params = []string{}
	for idx < len(source) && source[idx] != ')' {
		// Skip whitespace and commas
		for idx < len(source) && (source[idx] == ' ' || source[idx] == '\t' || source[idx] == ',' || source[idx] == '\n') {
			idx++
		}
		if idx < len(source) && source[idx] == ')' {
			break
		}

		// Extract parameter name
		paramStart := idx
		for idx < len(source) && isIdentChar(source[idx]) {
			idx++
		}
		if idx > paramStart {
			params = append(params, source[paramStart:idx])
		}

		// Skip optional type annotation (param: type)
		for idx < len(source) && (source[idx] == ' ' || source[idx] == '\t') {
			idx++
		}
		if idx < len(source) && source[idx] == ':' {
			idx++
			for idx < len(source) && source[idx] != ',' && source[idx] != ')' {
				idx++
			}
		}
	}

	if idx >= len(source) || source[idx] != ')' {
		return "", nil, types.TypeAny, "", errors.New(errors.ErrInvalidSyntax, "expected ')' after parameters")
	}
	idx++

	// Skip whitespace
	for idx < len(source) && (source[idx] == ' ' || source[idx] == '\t') {
		idx++
	}

	// Check for return type annotation
	returnType = types.TypeAny
	if idx < len(source) && source[idx] == ':' {
		idx++
		// Skip whitespace
		for idx < len(source) && (source[idx] == ' ' || source[idx] == '\t') {
			idx++
		}
		// Extract return type
		typeStart := idx
		for idx < len(source) && isIdentChar(source[idx]) {
			idx++
		}
		typeName := source[typeStart:idx]
		returnType = types.ParseType(typeName)
	}

	// Skip whitespace
	for idx < len(source) && (source[idx] == ' ' || source[idx] == '\t' || source[idx] == '\n') {
		idx++
	}

	// Expect '{'
	if idx >= len(source) || source[idx] != '{' {
		return "", nil, types.TypeAny, "", errors.New(errors.ErrInvalidSyntax, "expected '{' for function body")
	}

	// Find matching '}' for the body
	braceCount := 1
	bodyStart := idx
	idx++
	for idx < len(source) && braceCount > 0 {
		if source[idx] == '{' {
			braceCount++
		} else if source[idx] == '}' {
			braceCount--
		} else if source[idx] == '"' || source[idx] == '\'' || source[idx] == '`' {
			// Skip string literals
			quote := source[idx]
			idx++
			for idx < len(source) && source[idx] != quote {
				if source[idx] == '\\' {
					idx++
				}
				idx++
			}
		}
		idx++
	}

	if braceCount != 0 {
		return "", nil, types.TypeAny, "", errors.New(errors.ErrInvalidSyntax, "unmatched braces in function body")
	}

	body = source[bodyStart:idx]

	// Reconstruct full function for execution
	body = source // Return the full source as body for execution

	return name, params, returnType, body, nil
}

// isIdentChar returns true if the character can be part of an identifier.
func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '$'
}

// RegisterJSFunction parses a JS function source and registers it in the registry.
func (r *Registry) RegisterJSFunction(source string, sandbox *Sandbox) error {
	name, params, returnType, body, err := ParseJSFunction(source)
	if err != nil {
		return err
	}

	// Build parameter definitions
	paramDefs := make([]types.ParameterDef, len(params))
	for i, p := range params {
		paramDefs[i] = types.ParameterDef{
			Name: p,
			Type: types.TypeAny, // JS is dynamically typed
		}
	}

	fn := &Function{
		Name: name,
		Signature: &types.FunctionSignature{
			Name:       name,
			Parameters: paramDefs,
			ReturnType: returnType,
		},
		JSBody: body,
		Pure:   false, // Assume JS functions may have side effects
	}

	return r.Register(fn)
}

// CallJS invokes a JavaScript function through the sandbox.
func (r *Registry) CallJS(ctx context.Context, sandbox *Sandbox, name string, args []types.Value) (types.Value, error) {
	fn, ok := r.Get(name)
	if !ok {
		return types.Null(), errors.Newf(errors.ErrUndefinedFunction, "undefined function '%s'", name)
	}

	if !fn.IsJS() {
		return types.Null(), errors.Newf(errors.ErrInvalidSyntax, "function '%s' is not a JS function", name)
	}

	// Validate arguments
	if fn.Signature != nil {
		if err := fn.Signature.ValidateArgs(args); err != nil {
			return types.Null(), errors.Wrap(errors.ErrArgumentType, err.Error(), err)
		}
	}

	return sandbox.Execute(ctx, fn.JSBody, name, args)
}
