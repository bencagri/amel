// Package types provides type definitions and type checking for the AMEL DSL.
package types

import "fmt"

// Type represents the type of a value in the AMEL type system.
type Type int

const (
	TypeUnknown Type = iota
	TypeInt
	TypeFloat
	TypeString
	TypeBool
	TypeNull
	TypeList
	TypeAny
	TypeFunction
)

var typeNames = map[Type]string{
	TypeUnknown:  "unknown",
	TypeInt:      "int",
	TypeFloat:    "float",
	TypeString:   "string",
	TypeBool:     "bool",
	TypeNull:     "null",
	TypeList:     "list",
	TypeAny:      "any",
	TypeFunction: "function",
}

// String returns the string representation of a type.
func (t Type) String() string {
	if name, ok := typeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Type(%d)", t)
}

// ParseType parses a type name string into a Type.
func ParseType(name string) Type {
	for t, n := range typeNames {
		if n == name {
			return t
		}
	}
	return TypeUnknown
}

// IsNumeric returns true if the type is a numeric type (int or float).
func (t Type) IsNumeric() bool {
	return t == TypeInt || t == TypeFloat
}

// IsComparable returns true if the type can be compared with comparison operators.
func (t Type) IsComparable() bool {
	return t == TypeInt || t == TypeFloat || t == TypeString || t == TypeBool
}

// IsCompatible checks if two types are compatible for operations.
func (t Type) IsCompatible(other Type) bool {
	if t == TypeAny || other == TypeAny {
		return true
	}
	if t == other {
		return true
	}
	// int and float are compatible
	if t.IsNumeric() && other.IsNumeric() {
		return true
	}
	return false
}

// PromoteNumeric returns the promoted type when combining two numeric types.
// If either is float, result is float. Otherwise, int.
func PromoteNumeric(a, b Type) Type {
	if a == TypeFloat || b == TypeFloat {
		return TypeFloat
	}
	return TypeInt
}

// Value represents a typed value in the AMEL runtime.
type Value struct {
	Type Type
	Raw  interface{}
}

// NewValue creates a new Value with automatic type inference.
func NewValue(v interface{}) Value {
	if v == nil {
		return Value{Type: TypeNull, Raw: nil}
	}

	switch val := v.(type) {
	case int:
		return Value{Type: TypeInt, Raw: int64(val)}
	case int32:
		return Value{Type: TypeInt, Raw: int64(val)}
	case int64:
		return Value{Type: TypeInt, Raw: val}
	case float32:
		return Value{Type: TypeFloat, Raw: float64(val)}
	case float64:
		return Value{Type: TypeFloat, Raw: val}
	case string:
		return Value{Type: TypeString, Raw: val}
	case bool:
		return Value{Type: TypeBool, Raw: val}
	case []interface{}:
		return Value{Type: TypeList, Raw: val}
	case []Value:
		return Value{Type: TypeList, Raw: val}
	default:
		return Value{Type: TypeAny, Raw: val}
	}
}

// Int creates an integer Value.
func Int(v int64) Value {
	return Value{Type: TypeInt, Raw: v}
}

// Float creates a float Value.
func Float(v float64) Value {
	return Value{Type: TypeFloat, Raw: v}
}

// String creates a string Value.
func String(v string) Value {
	return Value{Type: TypeString, Raw: v}
}

// Bool creates a boolean Value.
func Bool(v bool) Value {
	return Value{Type: TypeBool, Raw: v}
}

// Null creates a null Value.
func Null() Value {
	return Value{Type: TypeNull, Raw: nil}
}

// List creates a list Value.
func List(elements ...Value) Value {
	return Value{Type: TypeList, Raw: elements}
}

// Any creates an any-typed Value.
func Any(v interface{}) Value {
	return Value{Type: TypeAny, Raw: v}
}

// IsNull returns true if the value is null.
func (v Value) IsNull() bool {
	return v.Type == TypeNull || v.Raw == nil
}

// IsTruthy returns the truthiness of a value.
func (v Value) IsTruthy() bool {
	if v.IsNull() {
		return false
	}

	switch v.Type {
	case TypeBool:
		return v.Raw.(bool)
	case TypeInt:
		return v.Raw.(int64) != 0
	case TypeFloat:
		return v.Raw.(float64) != 0
	case TypeString:
		return v.Raw.(string) != ""
	case TypeList:
		switch list := v.Raw.(type) {
		case []interface{}:
			return len(list) > 0
		case []Value:
			return len(list) > 0
		}
	}
	return v.Raw != nil
}

// AsInt converts the value to an int64.
func (v Value) AsInt() (int64, bool) {
	switch v.Type {
	case TypeInt:
		return v.Raw.(int64), true
	case TypeFloat:
		return int64(v.Raw.(float64)), true
	}
	return 0, false
}

// AsFloat converts the value to a float64.
func (v Value) AsFloat() (float64, bool) {
	switch v.Type {
	case TypeInt:
		return float64(v.Raw.(int64)), true
	case TypeFloat:
		return v.Raw.(float64), true
	}
	return 0, false
}

// AsString converts the value to a string.
func (v Value) AsString() (string, bool) {
	if v.Type == TypeString {
		return v.Raw.(string), true
	}
	return "", false
}

// AsBool converts the value to a bool.
func (v Value) AsBool() (bool, bool) {
	if v.Type == TypeBool {
		return v.Raw.(bool), true
	}
	return false, false
}

// AsList converts the value to a list of Values.
func (v Value) AsList() ([]Value, bool) {
	if v.Type != TypeList {
		return nil, false
	}

	switch list := v.Raw.(type) {
	case []Value:
		return list, true
	case []interface{}:
		result := make([]Value, len(list))
		for i, elem := range list {
			result[i] = NewValue(elem)
		}
		return result, true
	}
	return nil, false
}

// Equals checks if two values are equal.
func (v Value) Equals(other Value) bool {
	// Handle null comparison
	if v.IsNull() && other.IsNull() {
		return true
	}
	if v.IsNull() || other.IsNull() {
		return false
	}

	// Handle numeric comparison with type promotion
	if v.Type.IsNumeric() && other.Type.IsNumeric() {
		vf, _ := v.AsFloat()
		of, _ := other.AsFloat()
		return vf == of
	}

	// Same type comparison
	if v.Type != other.Type {
		return false
	}

	switch v.Type {
	case TypeString:
		return v.Raw.(string) == other.Raw.(string)
	case TypeBool:
		return v.Raw.(bool) == other.Raw.(bool)
	case TypeList:
		vList, _ := v.AsList()
		oList, _ := other.AsList()
		if len(vList) != len(oList) {
			return false
		}
		for i := range vList {
			if !vList[i].Equals(oList[i]) {
				return false
			}
		}
		return true
	}

	return v.Raw == other.Raw
}

// Compare compares two values and returns:
// -1 if v < other
//
//	0 if v == other
//	1 if v > other
func (v Value) Compare(other Value) (int, bool) {
	// Numeric comparison
	if v.Type.IsNumeric() && other.Type.IsNumeric() {
		vf, _ := v.AsFloat()
		of, _ := other.AsFloat()
		if vf < of {
			return -1, true
		} else if vf > of {
			return 1, true
		}
		return 0, true
	}

	// String comparison
	if v.Type == TypeString && other.Type == TypeString {
		vs := v.Raw.(string)
		os := other.Raw.(string)
		if vs < os {
			return -1, true
		} else if vs > os {
			return 1, true
		}
		return 0, true
	}

	return 0, false
}

// ParameterDef defines a function parameter.
type ParameterDef struct {
	Name string
	Type Type
}

// FunctionSignature defines the signature of a function.
type FunctionSignature struct {
	Name       string
	Parameters []ParameterDef
	ReturnType Type
	Variadic   bool // if true, last parameter can accept multiple values
}

// NewFunctionSignature creates a new function signature.
func NewFunctionSignature(name string, returnType Type, params ...ParameterDef) *FunctionSignature {
	return &FunctionSignature{
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
	}
}

// NewVariadicSignature creates a new variadic function signature.
func NewVariadicSignature(name string, returnType Type, params ...ParameterDef) *FunctionSignature {
	return &FunctionSignature{
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Variadic:   true,
	}
}

// Param creates a parameter definition.
func Param(name string, typ Type) ParameterDef {
	return ParameterDef{Name: name, Type: typ}
}

// ValidateArgs validates that the given arguments match the function signature.
func (sig *FunctionSignature) ValidateArgs(args []Value) error {
	minArgs := len(sig.Parameters)
	if sig.Variadic && minArgs > 0 {
		minArgs-- // variadic functions need at least (params - 1) args
	}

	if len(args) < minArgs {
		return fmt.Errorf("function %s requires at least %d arguments, got %d",
			sig.Name, minArgs, len(args))
	}

	if !sig.Variadic && len(args) > len(sig.Parameters) {
		return fmt.Errorf("function %s accepts at most %d arguments, got %d",
			sig.Name, len(sig.Parameters), len(args))
	}

	// Check argument types
	for i, arg := range args {
		var expectedType Type
		if i < len(sig.Parameters) {
			expectedType = sig.Parameters[i].Type
		} else if sig.Variadic && len(sig.Parameters) > 0 {
			// For variadic functions, use the last parameter's type
			expectedType = sig.Parameters[len(sig.Parameters)-1].Type
		}

		if expectedType != TypeAny && !arg.Type.IsCompatible(expectedType) {
			return fmt.Errorf("function %s argument %d: expected %s, got %s",
				sig.Name, i+1, expectedType, arg.Type)
		}
	}

	return nil
}
