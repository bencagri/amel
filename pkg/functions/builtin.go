// Package functions provides function management for the AMEL DSL engine.
package functions

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bencagri/amel/internal/errors"
	"github.com/bencagri/amel/pkg/types"
)

// RegisterBuiltIns registers all built-in functions in the given registry.
func RegisterBuiltIns(r *Registry) error {
	builtins := []struct {
		name string
		fn   BuiltInFunc
		sig  *types.FunctionSignature
	}{
		// Aggregate functions
		{"count", builtinCount, types.NewVariadicSignature("count", types.TypeInt, types.Param("list", types.TypeAny))},
		{"sum", builtinSum, types.NewVariadicSignature("sum", types.TypeFloat, types.Param("values", types.TypeAny))},
		{"avg", builtinAvg, types.NewVariadicSignature("avg", types.TypeFloat, types.Param("values", types.TypeAny))},
		{"min", builtinMin, types.NewVariadicSignature("min", types.TypeAny, types.Param("values", types.TypeAny))},
		{"max", builtinMax, types.NewVariadicSignature("max", types.TypeAny, types.Param("values", types.TypeAny))},

		// Math functions
		{"abs", builtinAbs, types.NewFunctionSignature("abs", types.TypeFloat, types.Param("value", types.TypeAny))},
		{"ceil", builtinCeil, types.NewFunctionSignature("ceil", types.TypeInt, types.Param("value", types.TypeFloat))},
		{"floor", builtinFloor, types.NewFunctionSignature("floor", types.TypeInt, types.Param("value", types.TypeFloat))},
		{"round", builtinRound, types.NewFunctionSignature("round", types.TypeInt, types.Param("value", types.TypeFloat))},
		{"pow", builtinPow, types.NewFunctionSignature("pow", types.TypeFloat, types.Param("base", types.TypeAny), types.Param("exp", types.TypeAny))},
		{"sqrt", builtinSqrt, types.NewFunctionSignature("sqrt", types.TypeFloat, types.Param("value", types.TypeAny))},
		{"mod", builtinMod, types.NewFunctionSignature("mod", types.TypeInt, types.Param("a", types.TypeInt), types.Param("b", types.TypeInt))},

		// String functions
		{"len", builtinLen, types.NewFunctionSignature("len", types.TypeInt, types.Param("value", types.TypeAny))},
		{"lower", builtinLower, types.NewFunctionSignature("lower", types.TypeString, types.Param("str", types.TypeString))},
		{"upper", builtinUpper, types.NewFunctionSignature("upper", types.TypeString, types.Param("str", types.TypeString))},
		{"trim", builtinTrim, types.NewFunctionSignature("trim", types.TypeString, types.Param("str", types.TypeString))},
		{"contains", builtinContains, types.NewFunctionSignature("contains", types.TypeBool, types.Param("str", types.TypeString), types.Param("substr", types.TypeString))},
		{"startsWith", builtinStartsWith, types.NewFunctionSignature("startsWith", types.TypeBool, types.Param("str", types.TypeString), types.Param("prefix", types.TypeString))},
		{"endsWith", builtinEndsWith, types.NewFunctionSignature("endsWith", types.TypeBool, types.Param("str", types.TypeString), types.Param("suffix", types.TypeString))},
		{"substr", builtinSubstr, types.NewFunctionSignature("substr", types.TypeString, types.Param("str", types.TypeString), types.Param("start", types.TypeInt), types.Param("length", types.TypeInt))},
		{"replace", builtinReplace, types.NewFunctionSignature("replace", types.TypeString, types.Param("str", types.TypeString), types.Param("old", types.TypeString), types.Param("new", types.TypeString))},
		{"split", builtinSplit, types.NewFunctionSignature("split", types.TypeList, types.Param("str", types.TypeString), types.Param("sep", types.TypeString))},
		{"join", builtinJoin, types.NewFunctionSignature("join", types.TypeString, types.Param("list", types.TypeList), types.Param("sep", types.TypeString))},
		{"concat", builtinConcat, types.NewVariadicSignature("concat", types.TypeString, types.Param("strings", types.TypeString))},
		{"match", builtinMatch, types.NewFunctionSignature("match", types.TypeBool, types.Param("str", types.TypeString), types.Param("pattern", types.TypeString))},

		// Type conversion functions
		{"int", builtinInt, types.NewFunctionSignature("int", types.TypeInt, types.Param("value", types.TypeAny))},
		{"float", builtinFloat, types.NewFunctionSignature("float", types.TypeFloat, types.Param("value", types.TypeAny))},
		{"string", builtinString, types.NewFunctionSignature("string", types.TypeString, types.Param("value", types.TypeAny))},
		{"bool", builtinBool, types.NewFunctionSignature("bool", types.TypeBool, types.Param("value", types.TypeAny))},

		// List functions
		{"first", builtinFirst, types.NewFunctionSignature("first", types.TypeAny, types.Param("list", types.TypeList))},
		{"last", builtinLast, types.NewFunctionSignature("last", types.TypeAny, types.Param("list", types.TypeList))},
		{"at", builtinAt, types.NewFunctionSignature("at", types.TypeAny, types.Param("list", types.TypeList), types.Param("index", types.TypeInt))},
		{"reverse", builtinReverse, types.NewFunctionSignature("reverse", types.TypeList, types.Param("list", types.TypeList))},
		{"unique", builtinUnique, types.NewFunctionSignature("unique", types.TypeList, types.Param("list", types.TypeList))},
		{"flatten", builtinFlatten, types.NewFunctionSignature("flatten", types.TypeList, types.Param("list", types.TypeList))},
		{"slice", builtinSlice, types.NewFunctionSignature("slice", types.TypeList, types.Param("list", types.TypeList), types.Param("start", types.TypeInt), types.Param("end", types.TypeInt))},

		// Logical/utility functions
		{"coalesce", builtinCoalesce, types.NewVariadicSignature("coalesce", types.TypeAny, types.Param("values", types.TypeAny))},
		{"ifThenElse", builtinIfThenElse, types.NewFunctionSignature("ifThenElse", types.TypeAny, types.Param("condition", types.TypeBool), types.Param("then", types.TypeAny), types.Param("else", types.TypeAny))},
		{"isNull", builtinIsNull, types.NewFunctionSignature("isNull", types.TypeBool, types.Param("value", types.TypeAny))},
		{"isNotNull", builtinIsNotNull, types.NewFunctionSignature("isNotNull", types.TypeBool, types.Param("value", types.TypeAny))},
		{"isEmpty", builtinIsEmpty, types.NewFunctionSignature("isEmpty", types.TypeBool, types.Param("value", types.TypeAny))},
		{"typeOf", builtinTypeOf, types.NewFunctionSignature("typeOf", types.TypeString, types.Param("value", types.TypeAny))},

		// Additional list functions
		{"indexOf", builtinIndexOf, types.NewFunctionSignature("indexOf", types.TypeInt, types.Param("list", types.TypeList), types.Param("value", types.TypeAny))},
		{"sortAsc", builtinSortAsc, types.NewFunctionSignature("sortAsc", types.TypeList, types.Param("list", types.TypeList))},
		{"sortDesc", builtinSortDesc, types.NewFunctionSignature("sortDesc", types.TypeList, types.Param("list", types.TypeList))},
		{"all", builtinAll, types.NewFunctionSignature("all", types.TypeBool, types.Param("list", types.TypeList))},
		{"any", builtinAny, types.NewFunctionSignature("any", types.TypeBool, types.Param("list", types.TypeList))},

		// Additional numeric functions
		{"clamp", builtinClamp, types.NewFunctionSignature("clamp", types.TypeAny, types.Param("value", types.TypeAny), types.Param("min", types.TypeAny), types.Param("max", types.TypeAny))},
		{"between", builtinBetween, types.NewFunctionSignature("between", types.TypeBool, types.Param("value", types.TypeAny), types.Param("min", types.TypeAny), types.Param("max", types.TypeAny))},

		// Additional utility functions
		{"defaultVal", builtinDefaultVal, types.NewFunctionSignature("defaultVal", types.TypeAny, types.Param("value", types.TypeAny), types.Param("default", types.TypeAny))},
		{"format", builtinFormat, types.NewVariadicSignature("format", types.TypeString, types.Param("template", types.TypeString), types.Param("args", types.TypeAny))},

		// Additional string functions
		{"trimLeft", builtinTrimLeft, types.NewFunctionSignature("trimLeft", types.TypeString, types.Param("str", types.TypeString))},
		{"trimRight", builtinTrimRight, types.NewFunctionSignature("trimRight", types.TypeString, types.Param("str", types.TypeString))},
		{"padLeft", builtinPadLeft, types.NewFunctionSignature("padLeft", types.TypeString, types.Param("str", types.TypeString), types.Param("length", types.TypeInt), types.Param("pad", types.TypeString))},
		{"padRight", builtinPadRight, types.NewFunctionSignature("padRight", types.TypeString, types.Param("str", types.TypeString), types.Param("length", types.TypeInt), types.Param("pad", types.TypeString))},
		{"repeat", builtinRepeat, types.NewFunctionSignature("repeat", types.TypeString, types.Param("str", types.TypeString), types.Param("count", types.TypeInt))},
	}

	for _, b := range builtins {
		if err := r.RegisterBuiltIn(b.name, b.fn, b.sig); err != nil {
			return err
		}
	}

	return nil
}

// NewDefaultRegistry creates a registry with all built-in functions pre-registered.
func NewDefaultRegistry() (*Registry, error) {
	r := NewRegistry()
	if err := RegisterBuiltIns(r); err != nil {
		return nil, err
	}
	return r, nil
}

// ============================================================================
// Aggregate Functions
// ============================================================================

// builtinCount returns the count of elements.
// count(list) or count(a, b, c, ...)
func builtinCount(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Int(0), nil
	}

	// If single argument is a list, count its elements
	if len(args) == 1 && args[0].Type == types.TypeList {
		list, ok := args[0].AsList()
		if !ok {
			return types.Int(0), nil
		}
		return types.Int(int64(len(list))), nil
	}

	// Otherwise count the arguments
	return types.Int(int64(len(args))), nil
}

// builtinSum returns the sum of numeric values.
func builtinSum(args ...types.Value) (types.Value, error) {
	values := flattenToValues(args)

	var sum float64
	for _, v := range values {
		f, ok := v.AsFloat()
		if !ok {
			continue // Skip non-numeric values
		}
		sum += f
	}

	return types.Float(sum), nil
}

// builtinAvg returns the average of numeric values.
func builtinAvg(args ...types.Value) (types.Value, error) {
	values := flattenToValues(args)

	if len(values) == 0 {
		return types.Null(), nil
	}

	var sum float64
	var count int
	for _, v := range values {
		f, ok := v.AsFloat()
		if !ok {
			continue
		}
		sum += f
		count++
	}

	if count == 0 {
		return types.Null(), nil
	}

	return types.Float(sum / float64(count)), nil
}

// builtinMin returns the minimum value.
func builtinMin(args ...types.Value) (types.Value, error) {
	values := flattenToValues(args)

	if len(values) == 0 {
		return types.Null(), nil
	}

	var minVal *types.Value
	for i := range values {
		v := values[i]
		if v.IsNull() {
			continue
		}
		if minVal == nil {
			minVal = &v
			continue
		}

		cmp, ok := v.Compare(*minVal)
		if ok && cmp < 0 {
			minVal = &v
		}
	}

	if minVal == nil {
		return types.Null(), nil
	}
	return *minVal, nil
}

// builtinMax returns the maximum value.
func builtinMax(args ...types.Value) (types.Value, error) {
	values := flattenToValues(args)

	if len(values) == 0 {
		return types.Null(), nil
	}

	var maxVal *types.Value
	for i := range values {
		v := values[i]
		if v.IsNull() {
			continue
		}
		if maxVal == nil {
			maxVal = &v
			continue
		}

		cmp, ok := v.Compare(*maxVal)
		if ok && cmp > 0 {
			maxVal = &v
		}
	}

	if maxVal == nil {
		return types.Null(), nil
	}
	return *maxVal, nil
}

// ============================================================================
// Math Functions
// ============================================================================

// builtinAbs returns the absolute value.
func builtinAbs(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	f, ok := args[0].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "abs requires a numeric value")
	}

	return types.Float(math.Abs(f)), nil
}

// builtinCeil returns the ceiling of a number.
func builtinCeil(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	f, ok := args[0].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "ceil requires a numeric value")
	}

	return types.Int(int64(math.Ceil(f))), nil
}

// builtinFloor returns the floor of a number.
func builtinFloor(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	f, ok := args[0].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "floor requires a numeric value")
	}

	return types.Int(int64(math.Floor(f))), nil
}

// builtinRound returns the rounded value.
func builtinRound(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	f, ok := args[0].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "round requires a numeric value")
	}

	return types.Int(int64(math.Round(f))), nil
}

// builtinPow returns base raised to the power of exp.
func builtinPow(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "pow requires 2 arguments")
	}

	base, ok := args[0].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "pow base requires a numeric value")
	}

	exp, ok := args[1].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "pow exponent requires a numeric value")
	}

	return types.Float(math.Pow(base, exp)), nil
}

// builtinSqrt returns the square root.
func builtinSqrt(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	f, ok := args[0].AsFloat()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "sqrt requires a numeric value")
	}

	if f < 0 {
		return types.Null(), errors.New(errors.ErrInvalidOperator, "sqrt of negative number")
	}

	return types.Float(math.Sqrt(f)), nil
}

// builtinMod returns the modulo (remainder).
func builtinMod(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "mod requires 2 arguments")
	}

	a, ok := args[0].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "mod requires integer values")
	}

	b, ok := args[1].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "mod requires integer values")
	}

	if b == 0 {
		return types.Null(), errors.New(errors.ErrDivisionByZero, "division by zero")
	}

	return types.Int(a % b), nil
}

// ============================================================================
// String Functions
// ============================================================================

// builtinLen returns the length of a string (in runes/characters) or list.
func builtinLen(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Int(0), nil
	}

	switch args[0].Type {
	case types.TypeString:
		s, _ := args[0].AsString()
		// Use RuneCountInString for proper Unicode character count
		return types.Int(int64(utf8.RuneCountInString(s))), nil
	case types.TypeList:
		list, _ := args[0].AsList()
		return types.Int(int64(len(list))), nil
	default:
		return types.Int(0), nil
	}
}

// builtinLower returns the lowercase version of a string.
func builtinLower(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}

	s, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "lower requires a string value")
	}

	return types.String(strings.ToLower(s)), nil
}

// builtinUpper returns the uppercase version of a string.
func builtinUpper(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}

	s, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "upper requires a string value")
	}

	return types.String(strings.ToUpper(s)), nil
}

// builtinTrim removes leading and trailing whitespace.
func builtinTrim(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}

	s, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "trim requires a string value")
	}

	return types.String(strings.TrimSpace(s)), nil
}

// builtinContains checks if a string contains a substring.
func builtinContains(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Bool(false), nil
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "contains requires string values")
	}

	substr, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "contains requires string values")
	}

	return types.Bool(strings.Contains(str, substr)), nil
}

// builtinStartsWith checks if a string starts with a prefix.
func builtinStartsWith(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Bool(false), nil
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "startsWith requires string values")
	}

	prefix, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "startsWith requires string values")
	}

	return types.Bool(strings.HasPrefix(str, prefix)), nil
}

// builtinEndsWith checks if a string ends with a suffix.
func builtinEndsWith(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Bool(false), nil
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "endsWith requires string values")
	}

	suffix, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "endsWith requires string values")
	}

	return types.Bool(strings.HasSuffix(str, suffix)), nil
}

// builtinSubstr extracts a substring.
func builtinSubstr(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "substr requires 3 arguments (str, start, length)")
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "substr requires a string value")
	}

	start, ok := args[1].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "substr start requires an integer")
	}

	length, ok := args[2].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "substr length requires an integer")
	}

	runes := []rune(str)
	strLen := int64(len(runes))

	// Handle negative start (from end)
	if start < 0 {
		start = strLen + start
	}

	// Bounds checking
	if start < 0 {
		start = 0
	}
	if start >= strLen {
		return types.String(""), nil
	}

	end := start + length
	if end > strLen {
		end = strLen
	}
	if end < start {
		return types.String(""), nil
	}

	return types.String(string(runes[start:end])), nil
}

// builtinReplace replaces all occurrences of old with new.
func builtinReplace(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "replace requires 3 arguments")
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "replace requires string values")
	}

	old, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "replace requires string values")
	}

	newStr, ok := args[2].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "replace requires string values")
	}

	return types.String(strings.ReplaceAll(str, old, newStr)), nil
}

// builtinSplit splits a string by separator.
func builtinSplit(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "split requires 2 arguments")
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "split requires a string value")
	}

	sep, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "split separator requires a string")
	}

	parts := strings.Split(str, sep)
	result := make([]types.Value, len(parts))
	for i, p := range parts {
		result[i] = types.String(p)
	}

	return types.List(result...), nil
}

// builtinJoin joins a list of strings with a separator.
func builtinJoin(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "join requires 2 arguments")
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "join requires a list value")
	}

	sep, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "join separator requires a string")
	}

	strs := make([]string, 0, len(list))
	for _, v := range list {
		if s, ok := v.AsString(); ok {
			strs = append(strs, s)
		} else {
			strs = append(strs, fmt.Sprintf("%v", v.Raw))
		}
	}

	return types.String(strings.Join(strs, sep)), nil
}

// builtinConcat concatenates strings.
func builtinConcat(args ...types.Value) (types.Value, error) {
	var sb strings.Builder
	for _, arg := range args {
		if s, ok := arg.AsString(); ok {
			sb.WriteString(s)
		} else {
			sb.WriteString(fmt.Sprintf("%v", arg.Raw))
		}
	}
	return types.String(sb.String()), nil
}

// builtinMatch checks if a string matches a regular expression.
func builtinMatch(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Bool(false), nil
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "match requires a string value")
	}

	pattern, ok := args[1].AsString()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "match pattern requires a string")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return types.Null(), errors.Wrap(errors.ErrInvalidSyntax, "invalid regex pattern", err)
	}

	return types.Bool(re.MatchString(str)), nil
}

// ============================================================================
// Type Conversion Functions
// ============================================================================

// builtinInt converts a value to an integer.
func builtinInt(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Int(0), nil
	}

	switch args[0].Type {
	case types.TypeInt:
		return args[0], nil
	case types.TypeFloat:
		f, _ := args[0].AsFloat()
		return types.Int(int64(f)), nil
	case types.TypeString:
		s, _ := args[0].AsString()
		var i int64
		_, err := fmt.Sscanf(s, "%d", &i)
		if err != nil {
			return types.Null(), errors.Wrap(errors.ErrTypeMismatch, "cannot convert string to int", err)
		}
		return types.Int(i), nil
	case types.TypeBool:
		b, _ := args[0].AsBool()
		if b {
			return types.Int(1), nil
		}
		return types.Int(0), nil
	default:
		return types.Null(), errors.New(errors.ErrTypeMismatch, "cannot convert to int")
	}
}

// builtinFloat converts a value to a float.
func builtinFloat(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Float(0), nil
	}

	switch args[0].Type {
	case types.TypeInt:
		i, _ := args[0].AsInt()
		return types.Float(float64(i)), nil
	case types.TypeFloat:
		return args[0], nil
	case types.TypeString:
		s, _ := args[0].AsString()
		var f float64
		_, err := fmt.Sscanf(s, "%f", &f)
		if err != nil {
			return types.Null(), errors.Wrap(errors.ErrTypeMismatch, "cannot convert string to float", err)
		}
		return types.Float(f), nil
	case types.TypeBool:
		b, _ := args[0].AsBool()
		if b {
			return types.Float(1), nil
		}
		return types.Float(0), nil
	default:
		return types.Null(), errors.New(errors.ErrTypeMismatch, "cannot convert to float")
	}
}

// builtinString converts a value to a string.
func builtinString(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}

	switch args[0].Type {
	case types.TypeString:
		return args[0], nil
	case types.TypeNull:
		return types.String("null"), nil
	default:
		return types.String(fmt.Sprintf("%v", args[0].Raw)), nil
	}
}

// builtinBool converts a value to a boolean.
func builtinBool(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Bool(false), nil
	}

	return types.Bool(args[0].IsTruthy()), nil
}

// ============================================================================
// List Functions
// ============================================================================

// builtinFirst returns the first element of a list.
func builtinFirst(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	list, ok := args[0].AsList()
	if !ok || len(list) == 0 {
		return types.Null(), nil
	}

	return list[0], nil
}

// builtinLast returns the last element of a list.
func builtinLast(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Null(), nil
	}

	list, ok := args[0].AsList()
	if !ok || len(list) == 0 {
		return types.Null(), nil
	}

	return list[len(list)-1], nil
}

// builtinAt returns the element at a specific index.
func builtinAt(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "at requires 2 arguments")
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "at requires a list value")
	}

	idx, ok := args[1].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "at index requires an integer")
	}

	// Handle negative index
	if idx < 0 {
		idx = int64(len(list)) + idx
	}

	if idx < 0 || idx >= int64(len(list)) {
		return types.Null(), errors.New(errors.ErrIndexOutOfBounds, "index out of bounds")
	}

	return list[idx], nil
}

// builtinReverse reverses a list.
func builtinReverse(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.List(), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "reverse requires a list value")
	}

	result := make([]types.Value, len(list))
	for i, v := range list {
		result[len(list)-1-i] = v
	}

	return types.List(result...), nil
}

// builtinUnique returns unique elements from a list.
func builtinUnique(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.List(), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "unique requires a list value")
	}

	var result []types.Value
	seen := make(map[string]bool)

	for _, v := range list {
		key := fmt.Sprintf("%v:%v", v.Type, v.Raw)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	return types.List(result...), nil
}

// builtinFlatten flattens nested lists.
func builtinFlatten(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.List(), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "flatten requires a list value")
	}

	var result []types.Value
	flattenRecursive(list, &result)

	return types.List(result...), nil
}

func flattenRecursive(list []types.Value, result *[]types.Value) {
	for _, v := range list {
		if nested, ok := v.AsList(); ok {
			flattenRecursive(nested, result)
		} else {
			*result = append(*result, v)
		}
	}
}

// builtinSlice returns a slice of a list.
func builtinSlice(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "slice requires 3 arguments")
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "slice requires a list value")
	}

	start, ok := args[1].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "slice start requires an integer")
	}

	end, ok := args[2].AsInt()
	if !ok {
		return types.Null(), errors.New(errors.ErrTypeMismatch, "slice end requires an integer")
	}

	listLen := int64(len(list))

	// Handle negative indices
	if start < 0 {
		start = listLen + start
	}
	if end < 0 {
		end = listLen + end
	}

	// Bounds checking
	if start < 0 {
		start = 0
	}
	if end > listLen {
		end = listLen
	}
	if start >= end {
		return types.List(), nil
	}

	return types.List(list[start:end]...), nil
}

// ============================================================================
// Logical/Utility Functions
// ============================================================================

// builtinCoalesce returns the first non-null value.
func builtinCoalesce(args ...types.Value) (types.Value, error) {
	for _, arg := range args {
		if !arg.IsNull() {
			return arg, nil
		}
	}
	return types.Null(), nil
}

// builtinIfThenElse returns then if condition is true, else otherwise.
func builtinIfThenElse(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "ifThenElse requires 3 arguments")
	}

	condition, ok := args[0].AsBool()
	if !ok {
		condition = args[0].IsTruthy()
	}

	if condition {
		return args[1], nil
	}
	return args[2], nil
}

// builtinIsNull checks if a value is null.
func builtinIsNull(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Bool(true), nil
	}
	return types.Bool(args[0].IsNull()), nil
}

// builtinIsNotNull checks if a value is not null.
func builtinIsNotNull(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Bool(false), nil
	}
	return types.Bool(!args[0].IsNull()), nil
}

// builtinIsEmpty checks if a value is empty (null, empty string, or empty list).
func builtinIsEmpty(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Bool(true), nil
	}

	v := args[0]
	if v.IsNull() {
		return types.Bool(true), nil
	}

	switch v.Type {
	case types.TypeString:
		s, _ := v.AsString()
		return types.Bool(s == ""), nil
	case types.TypeList:
		list, _ := v.AsList()
		return types.Bool(len(list) == 0), nil
	default:
		return types.Bool(false), nil
	}
}

// builtinTypeOf returns the type name of a value.
func builtinTypeOf(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String("null"), nil
	}
	return types.String(args[0].Type.String()), nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// flattenToValues flattens arguments to a single slice of values.
// If a single list argument is passed, it extracts its elements.
func flattenToValues(args []types.Value) []types.Value {
	if len(args) == 1 && args[0].Type == types.TypeList {
		list, ok := args[0].AsList()
		if ok {
			return list
		}
	}
	return args
}

// ============================================================================
// Additional List Functions
// ============================================================================

// builtinIndexOf returns the index of a value in a list, or -1 if not found.
func builtinIndexOf(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Int(-1), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Int(-1), nil
	}

	target := args[1]
	for i, elem := range list {
		if elem.Equals(target) {
			return types.Int(int64(i)), nil
		}
	}
	return types.Int(-1), nil
}

// builtinSortAsc sorts a list in ascending order.
func builtinSortAsc(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.List(), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.List(), nil
	}

	// Create a copy to avoid modifying the original
	sorted := make([]types.Value, len(list))
	copy(sorted, list)

	// Simple bubble sort (can be optimized for production)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			cmp, ok := sorted[j].Compare(sorted[j+1])
			if ok && cmp > 0 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return types.List(sorted...), nil
}

// builtinSortDesc sorts a list in descending order.
func builtinSortDesc(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.List(), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.List(), nil
	}

	// Create a copy to avoid modifying the original
	sorted := make([]types.Value, len(list))
	copy(sorted, list)

	// Simple bubble sort (can be optimized for production)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			cmp, ok := sorted[j].Compare(sorted[j+1])
			if ok && cmp < 0 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return types.List(sorted...), nil
}

// builtinAll returns true if all elements in the list are truthy.
func builtinAll(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Bool(true), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Bool(args[0].IsTruthy()), nil
	}

	for _, elem := range list {
		if !elem.IsTruthy() {
			return types.Bool(false), nil
		}
	}
	return types.Bool(true), nil
}

// builtinAny returns true if any element in the list is truthy.
func builtinAny(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.Bool(false), nil
	}

	list, ok := args[0].AsList()
	if !ok {
		return types.Bool(args[0].IsTruthy()), nil
	}

	for _, elem := range list {
		if elem.IsTruthy() {
			return types.Bool(true), nil
		}
	}
	return types.Bool(false), nil
}

// ============================================================================
// Additional Numeric Functions
// ============================================================================

// builtinClamp constrains a value between min and max.
func builtinClamp(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.Null(), errors.New(errors.ErrArgumentCount, "clamp requires 3 arguments: value, min, max")
	}

	value := args[0]
	minVal := args[1]
	maxVal := args[2]

	// Compare value with min
	cmpMin, okMin := value.Compare(minVal)
	if okMin && cmpMin < 0 {
		return minVal, nil
	}

	// Compare value with max
	cmpMax, okMax := value.Compare(maxVal)
	if okMax && cmpMax > 0 {
		return maxVal, nil
	}

	return value, nil
}

// builtinBetween checks if a value is between min and max (inclusive).
func builtinBetween(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.Bool(false), errors.New(errors.ErrArgumentCount, "between requires 3 arguments: value, min, max")
	}

	value := args[0]
	minVal := args[1]
	maxVal := args[2]

	cmpMin, okMin := value.Compare(minVal)
	cmpMax, okMax := value.Compare(maxVal)

	if okMin && okMax {
		return types.Bool(cmpMin >= 0 && cmpMax <= 0), nil
	}

	return types.Bool(false), nil
}

// ============================================================================
// Additional Utility Functions
// ============================================================================

// builtinDefaultVal returns the value if not null, otherwise returns the default.
func builtinDefaultVal(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.Null(), nil
	}

	if args[0].IsNull() {
		return args[1], nil
	}
	return args[0], nil
}

// builtinFormat formats a string with placeholders.
// format("Hello, %s! You are %d years old.", "John", 30)
func builtinFormat(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}

	template, ok := args[0].AsString()
	if !ok {
		return types.String(""), nil
	}

	// Simple placeholder replacement
	result := template
	for i := 1; i < len(args); i++ {
		placeholder := fmt.Sprintf("{%d}", i-1)
		val := args[i]
		var replacement string
		switch val.Type {
		case types.TypeString:
			replacement = val.Raw.(string)
		case types.TypeInt:
			replacement = fmt.Sprintf("%d", val.Raw)
		case types.TypeFloat:
			replacement = fmt.Sprintf("%v", val.Raw)
		case types.TypeBool:
			replacement = fmt.Sprintf("%v", val.Raw)
		case types.TypeNull:
			replacement = "null"
		default:
			replacement = fmt.Sprintf("%v", val.Raw)
		}
		result = strings.Replace(result, placeholder, replacement, 1)
	}

	return types.String(result), nil
}

// ============================================================================
// Additional String Functions
// ============================================================================

// builtinTrimLeft removes leading whitespace from a string.
func builtinTrimLeft(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}
	str, ok := args[0].AsString()
	if !ok {
		return types.String(""), nil
	}
	return types.String(strings.TrimLeft(str, " \t\n\r")), nil
}

// builtinTrimRight removes trailing whitespace from a string.
func builtinTrimRight(args ...types.Value) (types.Value, error) {
	if len(args) == 0 {
		return types.String(""), nil
	}
	str, ok := args[0].AsString()
	if !ok {
		return types.String(""), nil
	}
	return types.String(strings.TrimRight(str, " \t\n\r")), nil
}

// builtinPadLeft pads a string on the left to a specified length.
func builtinPadLeft(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.String(""), errors.New(errors.ErrArgumentCount, "padLeft requires 3 arguments: str, length, pad")
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.String(""), nil
	}

	length, ok := args[1].AsInt()
	if !ok {
		return types.String(str), nil
	}

	pad, ok := args[2].AsString()
	if !ok || len(pad) == 0 {
		pad = " "
	}

	for int64(len(str)) < length {
		str = pad + str
	}

	// Trim if padding went over
	if int64(len(str)) > length {
		str = str[len(str)-int(length):]
	}

	return types.String(str), nil
}

// builtinPadRight pads a string on the right to a specified length.
func builtinPadRight(args ...types.Value) (types.Value, error) {
	if len(args) < 3 {
		return types.String(""), errors.New(errors.ErrArgumentCount, "padRight requires 3 arguments: str, length, pad")
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.String(""), nil
	}

	length, ok := args[1].AsInt()
	if !ok {
		return types.String(str), nil
	}

	pad, ok := args[2].AsString()
	if !ok || len(pad) == 0 {
		pad = " "
	}

	for int64(len(str)) < length {
		str = str + pad
	}

	// Trim if padding went over
	if int64(len(str)) > length {
		str = str[:length]
	}

	return types.String(str), nil
}

// builtinRepeat repeats a string a specified number of times.
func builtinRepeat(args ...types.Value) (types.Value, error) {
	if len(args) < 2 {
		return types.String(""), nil
	}

	str, ok := args[0].AsString()
	if !ok {
		return types.String(""), nil
	}

	count, ok := args[1].AsInt()
	if !ok || count < 0 {
		return types.String(""), nil
	}

	if count > 10000 {
		count = 10000 // Safety limit
	}

	return types.String(strings.Repeat(str, int(count))), nil
}
