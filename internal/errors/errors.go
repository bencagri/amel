// Package errors provides error types and error handling for the AMEL DSL engine.
package errors

import (
	"fmt"
)

// ErrorCode represents different categories of errors in the AMEL engine.
type ErrorCode int

const (
	// Lexer errors (1xx)
	ErrUnexpectedCharacter ErrorCode = 100
	ErrUnterminatedString  ErrorCode = 101
	ErrInvalidNumber       ErrorCode = 102
	ErrInvalidEscape       ErrorCode = 103

	// Parser errors (2xx)
	ErrUnexpectedToken   ErrorCode = 200
	ErrMissingExpression ErrorCode = 201
	ErrUnmatchedParen    ErrorCode = 202
	ErrInvalidSyntax     ErrorCode = 203
	ErrUnexpectedEOF     ErrorCode = 204
	ErrInvalidJSONPath   ErrorCode = 205

	// Type errors (3xx)
	ErrTypeMismatch      ErrorCode = 300
	ErrUndefinedFunction ErrorCode = 301
	ErrArgumentCount     ErrorCode = 302
	ErrArgumentType      ErrorCode = 303
	ErrInvalidOperator   ErrorCode = 304
	ErrUndefinedVariable ErrorCode = 305

	// Runtime errors (4xx)
	ErrDivisionByZero   ErrorCode = 400
	ErrNullReference    ErrorCode = 401
	ErrIndexOutOfBounds ErrorCode = 402
	ErrTimeout          ErrorCode = 403
	ErrMemoryLimit      ErrorCode = 404
	ErrSandboxViolation ErrorCode = 405
	ErrFunctionPanic    ErrorCode = 406

	// JSONPath errors (5xx)
	ErrInvalidPath  ErrorCode = 500
	ErrPathNotFound ErrorCode = 501
)

// String returns the string representation of an error code.
func (c ErrorCode) String() string {
	switch c {
	case ErrUnexpectedCharacter:
		return "UnexpectedCharacter"
	case ErrUnterminatedString:
		return "UnterminatedString"
	case ErrInvalidNumber:
		return "InvalidNumber"
	case ErrInvalidEscape:
		return "InvalidEscape"
	case ErrUnexpectedToken:
		return "UnexpectedToken"
	case ErrMissingExpression:
		return "MissingExpression"
	case ErrUnmatchedParen:
		return "UnmatchedParen"
	case ErrInvalidSyntax:
		return "InvalidSyntax"
	case ErrUnexpectedEOF:
		return "UnexpectedEOF"
	case ErrInvalidJSONPath:
		return "InvalidJSONPath"
	case ErrTypeMismatch:
		return "TypeMismatch"
	case ErrUndefinedFunction:
		return "UndefinedFunction"
	case ErrArgumentCount:
		return "ArgumentCount"
	case ErrArgumentType:
		return "ArgumentType"
	case ErrInvalidOperator:
		return "InvalidOperator"
	case ErrUndefinedVariable:
		return "UndefinedVariable"
	case ErrDivisionByZero:
		return "DivisionByZero"
	case ErrNullReference:
		return "NullReference"
	case ErrIndexOutOfBounds:
		return "IndexOutOfBounds"
	case ErrTimeout:
		return "Timeout"
	case ErrMemoryLimit:
		return "MemoryLimit"
	case ErrSandboxViolation:
		return "SandboxViolation"
	case ErrFunctionPanic:
		return "FunctionPanic"
	case ErrInvalidPath:
		return "InvalidPath"
	case ErrPathNotFound:
		return "PathNotFound"
	default:
		return "Unknown"
	}
}

// Category returns the error category based on the error code.
func (c ErrorCode) Category() string {
	switch {
	case c >= 100 && c < 200:
		return "Lexer"
	case c >= 200 && c < 300:
		return "Parser"
	case c >= 300 && c < 400:
		return "Type"
	case c >= 400 && c < 500:
		return "Runtime"
	case c >= 500 && c < 600:
		return "JSONPath"
	default:
		return "Unknown"
	}
}

// Error represents an error in the AMEL engine with position information.
type Error struct {
	Code    ErrorCode
	Message string
	Line    int
	Column  int
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	pos := ""
	if e.Line > 0 {
		pos = fmt.Sprintf(" at line %d, column %d", e.Line, e.Column)
	}
	return fmt.Sprintf("%s Error [%d]%s: %s", e.Code.Category(), e.Code, pos, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error code.
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return false
}

// New creates a new Error with the given code and message.
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// NewAt creates a new Error with position information.
func NewAt(code ErrorCode, message string, line, column int) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Line:    line,
		Column:  column,
	}
}

// Wrap wraps an existing error with AMEL error information.
func Wrap(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WrapAt wraps an existing error with position information.
func WrapAt(code ErrorCode, message string, line, column int, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Line:    line,
		Column:  column,
		Cause:   cause,
	}
}

// Newf creates a new Error with a formatted message.
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewAtf creates a new Error with position information and a formatted message.
func NewAtf(code ErrorCode, line, column int, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Line:    line,
		Column:  column,
	}
}

// IsCode checks if an error has a specific error code.
func IsCode(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// IsCategory checks if an error belongs to a specific category.
func IsCategory(err error, category string) bool {
	if e, ok := err.(*Error); ok {
		return e.Code.Category() == category
	}
	return false
}
