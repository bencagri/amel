// Package functions provides function management for the AMEL DSL engine.
package functions

import (
	"fmt"
	"sync"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/types"
)

// BuiltInFunc is the signature for built-in Go functions.
type BuiltInFunc func(args ...types.Value) (types.Value, error)

// Function represents a callable function in the AMEL engine.
type Function struct {
	Name      string
	Signature *types.FunctionSignature
	BuiltIn   BuiltInFunc // For Go built-in functions
	JSBody    string      // For user-defined JS functions
	Pure      bool        // Whether the function has no side effects
}

// OverloadedFunction represents a function with multiple overloads.
type OverloadedFunction struct {
	Name      string
	Overloads []*Function
}

// IsBuiltIn returns true if this is a built-in Go function.
func (f *Function) IsBuiltIn() bool {
	return f.BuiltIn != nil
}

// IsJS returns true if this is a user-defined JavaScript function.
func (f *Function) IsJS() bool {
	return f.JSBody != ""
}

// Registry manages function registration and lookup.
type Registry struct {
	mu                  sync.RWMutex
	functions           map[string]*Function
	overloadedFunctions map[string]*OverloadedFunction
}

// NewRegistry creates a new function registry.
func NewRegistry() *Registry {
	return &Registry{
		functions:           make(map[string]*Function),
		overloadedFunctions: make(map[string]*OverloadedFunction),
	}
}

// Register adds a function to the registry.
func (r *Registry) Register(fn *Function) error {
	if fn == nil {
		return errors.New(errors.ErrInvalidSyntax, "cannot register nil function")
	}
	if fn.Name == "" {
		return errors.New(errors.ErrInvalidSyntax, "function name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[fn.Name]; exists {
		return errors.Newf(errors.ErrInvalidSyntax, "function '%s' is already registered", fn.Name)
	}

	r.functions[fn.Name] = fn
	return nil
}

// RegisterOverload adds a function overload to the registry.
// Multiple functions with the same name but different signatures can be registered.
func (r *Registry) RegisterOverload(fn *Function) error {
	if fn == nil {
		return errors.New(errors.ErrInvalidSyntax, "cannot register nil function")
	}
	if fn.Name == "" {
		return errors.New(errors.ErrInvalidSyntax, "function name cannot be empty")
	}
	if fn.Signature == nil {
		return errors.New(errors.ErrInvalidSyntax, "overloaded function must have a signature")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if there's already an overloaded function with this name
	if overloaded, exists := r.overloadedFunctions[fn.Name]; exists {
		// Check for duplicate signature
		for _, existing := range overloaded.Overloads {
			if signaturesMatch(existing.Signature, fn.Signature) {
				return errors.Newf(errors.ErrInvalidSyntax, "function '%s' with same signature already registered", fn.Name)
			}
		}
		overloaded.Overloads = append(overloaded.Overloads, fn)
		return nil
	}

	// Check if there's a non-overloaded function with this name
	if existing, exists := r.functions[fn.Name]; exists {
		// Convert to overloaded function
		if existing.Signature == nil {
			return errors.Newf(errors.ErrInvalidSyntax, "cannot add overload to function '%s' without signature", fn.Name)
		}
		r.overloadedFunctions[fn.Name] = &OverloadedFunction{
			Name:      fn.Name,
			Overloads: []*Function{existing, fn},
		}
		delete(r.functions, fn.Name)
		return nil
	}

	// Create new overloaded function
	r.overloadedFunctions[fn.Name] = &OverloadedFunction{
		Name:      fn.Name,
		Overloads: []*Function{fn},
	}
	return nil
}

// signaturesMatch checks if two function signatures match (same parameter types).
func signaturesMatch(a, b *types.FunctionSignature) bool {
	if a == nil || b == nil {
		return a == b
	}
	if len(a.Parameters) != len(b.Parameters) {
		return false
	}
	if a.Variadic != b.Variadic {
		return false
	}
	for i := range a.Parameters {
		if a.Parameters[i].Type != b.Parameters[i].Type {
			return false
		}
	}
	return true
}

// RegisterBuiltIn registers a built-in Go function.
func (r *Registry) RegisterBuiltIn(name string, fn BuiltInFunc, sig *types.FunctionSignature) error {
	return r.Register(&Function{
		Name:      name,
		Signature: sig,
		BuiltIn:   fn,
		Pure:      true, // Built-ins are assumed pure by default
	})
}

// Get retrieves a function by name.
// For overloaded functions, returns the first overload.
func (r *Registry) Get(name string) (*Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if fn, ok := r.functions[name]; ok {
		return fn, true
	}
	if overloaded, ok := r.overloadedFunctions[name]; ok && len(overloaded.Overloads) > 0 {
		return overloaded.Overloads[0], true
	}
	return nil, false
}

// GetOverloaded retrieves all overloads of a function by name.
func (r *Registry) GetOverloaded(name string) (*OverloadedFunction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if overloaded, ok := r.overloadedFunctions[name]; ok {
		return overloaded, true
	}
	// If it's a regular function, wrap it as overloaded
	if fn, ok := r.functions[name]; ok {
		return &OverloadedFunction{
			Name:      name,
			Overloads: []*Function{fn},
		}, true
	}
	return nil, false
}

// GetBestMatch retrieves the best matching overload for the given argument types.
func (r *Registry) GetBestMatch(name string, args []types.Value) (*Function, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check regular functions first
	if fn, ok := r.functions[name]; ok {
		return fn, true
	}

	// Check overloaded functions
	overloaded, ok := r.overloadedFunctions[name]
	if !ok || len(overloaded.Overloads) == 0 {
		return nil, false
	}

	// Find best match based on argument types
	var bestMatch *Function
	bestScore := -1

	for _, fn := range overloaded.Overloads {
		score := matchScore(fn.Signature, args)
		if score > bestScore {
			bestScore = score
			bestMatch = fn
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}

	// Return first overload as fallback
	return overloaded.Overloads[0], true
}

// matchScore calculates how well a function signature matches the given arguments.
// Higher score = better match. Returns -1 for incompatible.
func matchScore(sig *types.FunctionSignature, args []types.Value) int {
	if sig == nil {
		return 0 // No signature = generic match
	}

	// Check argument count
	minArgs := len(sig.Parameters)
	if sig.Variadic && minArgs > 0 {
		minArgs--
	}

	if len(args) < minArgs {
		return -1 // Not enough arguments
	}
	if !sig.Variadic && len(args) > len(sig.Parameters) {
		return -1 // Too many arguments
	}

	score := 0
	for i, arg := range args {
		var expectedType types.Type
		if i < len(sig.Parameters) {
			expectedType = sig.Parameters[i].Type
		} else if sig.Variadic && len(sig.Parameters) > 0 {
			expectedType = sig.Parameters[len(sig.Parameters)-1].Type
		} else {
			continue
		}

		if expectedType == types.TypeAny {
			score += 1 // Any type accepts anything
		} else if arg.Type == expectedType {
			score += 10 // Exact match
		} else if arg.Type.IsCompatible(expectedType) {
			score += 5 // Compatible type (e.g., int and float)
		} else {
			return -1 // Incompatible type
		}
	}

	return score
}

// Has checks if a function exists in the registry.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.functions[name]; ok {
		return true
	}
	_, ok := r.overloadedFunctions[name]
	return ok
}

// IsOverloaded checks if a function has multiple overloads.
func (r *Registry) IsOverloaded(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.overloadedFunctions[name]
	return ok
}

// Unregister removes a function from the registry.
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; exists {
		delete(r.functions, name)
		return true
	}
	if _, exists := r.overloadedFunctions[name]; exists {
		delete(r.overloadedFunctions, name)
		return true
	}
	return false
}

// List returns all registered function names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	names := make([]string, 0, len(r.functions)+len(r.overloadedFunctions))
	for name := range r.functions {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	for name := range r.overloadedFunctions {
		if !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	return names
}

// ListSignatures returns all registered function signatures.
func (r *Registry) ListSignatures() []*types.FunctionSignature {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sigs := make([]*types.FunctionSignature, 0, len(r.functions)+len(r.overloadedFunctions)*2)
	for _, fn := range r.functions {
		if fn.Signature != nil {
			sigs = append(sigs, fn.Signature)
		}
	}
	for _, overloaded := range r.overloadedFunctions {
		for _, fn := range overloaded.Overloads {
			if fn.Signature != nil {
				sigs = append(sigs, fn.Signature)
			}
		}
	}
	return sigs
}

// ListOverloads returns all overloads for a function name.
func (r *Registry) ListOverloads(name string) []*Function {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if overloaded, ok := r.overloadedFunctions[name]; ok {
		result := make([]*Function, len(overloaded.Overloads))
		copy(result, overloaded.Overloads)
		return result
	}
	if fn, ok := r.functions[name]; ok {
		return []*Function{fn}
	}
	return nil
}

// Call invokes a function by name with the given arguments.
// For overloaded functions, it selects the best matching overload.
func (r *Registry) Call(name string, args ...types.Value) (types.Value, error) {
	fn, ok := r.GetBestMatch(name, args)
	if !ok {
		return types.Null(), errors.Newf(errors.ErrUndefinedFunction, "undefined function '%s'", name)
	}

	// Validate arguments against signature
	if fn.Signature != nil {
		if err := fn.Signature.ValidateArgs(args); err != nil {
			return types.Null(), errors.Wrap(errors.ErrArgumentType, err.Error(), err)
		}
	}

	// Call the function
	if fn.IsBuiltIn() {
		result, err := fn.BuiltIn(args...)
		if err != nil {
			return types.Null(), errors.Wrap(errors.ErrFunctionPanic, fmt.Sprintf("function '%s' failed: %v", name, err), err)
		}
		return result, nil
	}

	// JS functions need to be executed via sandbox (handled elsewhere)
	return types.Null(), errors.Newf(errors.ErrInvalidSyntax, "JS function '%s' must be called via sandbox", name)
}

// Clone creates a copy of the registry.
func (r *Registry) Clone() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := NewRegistry()
	for name, fn := range r.functions {
		clone.functions[name] = fn
	}
	return clone
}

// Merge adds all functions from another registry.
// Existing functions with the same name will be overwritten.
func (r *Registry) Merge(other *Registry) {
	if other == nil {
		return
	}

	other.mu.RLock()
	defer other.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	for name, fn := range other.functions {
		r.functions[name] = fn
	}
	for name, overloaded := range other.overloadedFunctions {
		r.overloadedFunctions[name] = overloaded
	}
}

// Clear removes all functions from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.functions = make(map[string]*Function)
	r.overloadedFunctions = make(map[string]*OverloadedFunction)
}

// Count returns the number of registered functions (including overloads).
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := len(r.functions)
	for _, overloaded := range r.overloadedFunctions {
		count += len(overloaded.Overloads)
	}
	return count
}

// CountUnique returns the number of unique function names.
func (r *Registry) CountUnique() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.functions) + len(r.overloadedFunctions)
}
