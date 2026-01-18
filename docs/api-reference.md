# API Reference

Complete Go API reference for the AMEL engine and its components.

## Table of Contents

- [Engine Package](#engine-package)
- [Parser Package](#parser-package)
- [Compiler Package](#compiler-package)
- [Types Package](#types-package)
- [Functions Package](#functions-package)
- [Evaluator Package](#evaluator-package)

---

## Engine Package

```go
import "github.com/bencagri/amel/pkg/engine"
```

The engine package provides the main facade for AMEL functionality.

### Creating an Engine

#### New

Creates a new AMEL engine with optional configuration.

```go
func New(opts ...Option) (*Engine, error)
```

**Example:**

```go
eng, err := engine.New()
if err != nil {
    log.Fatal(err)
}
```

**With Options:**

```go
eng, err := engine.New(
    engine.WithTimeout(100 * time.Millisecond),
    engine.WithCaching(true),
    engine.WithExplainMode(true),
)
```

---

### Engine Methods

#### Compile

Parses and compiles an AMEL expression.

```go
func (e *Engine) Compile(dsl string) (*CompiledExpression, error)
```

**Example:**

```go
compiled, err := eng.Compile(`$.age >= 18 && $.verified == true`)
if err != nil {
    log.Fatal(err)
}
```

---

#### Evaluate

Evaluates a compiled expression against a payload.

```go
func (e *Engine) Evaluate(expr *CompiledExpression, payload interface{}) (types.Value, error)
```

**Example:**

```go
payload := map[string]interface{}{"age": 25, "verified": true}
result, err := eng.Evaluate(compiled, payload)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Raw) // true
```

---

#### EvaluateBool

Evaluates a compiled expression and returns a boolean result.

```go
func (e *Engine) EvaluateBool(expr *CompiledExpression, payload interface{}) (bool, error)
```

**Example:**

```go
isAdult, err := eng.EvaluateBool(compiled, payload)
if isAdult {
    fmt.Println("User is an adult")
}
```

---

#### EvaluateDirect

Compiles and evaluates an expression in one step.

```go
func (e *Engine) EvaluateDirect(dsl string, payload interface{}) (types.Value, error)
```

**Example:**

```go
result, err := eng.EvaluateDirect(`$.price * 1.1`, payload)
```

---

#### EvaluateDirectBool

Compiles and evaluates an expression, returning a boolean.

```go
func (e *Engine) EvaluateDirectBool(dsl string, payload interface{}) (bool, error)
```

**Example:**

```go
canPurchase, err := eng.EvaluateDirectBool(`$.balance >= $.price`, payload)
```

---

#### EvaluateWithExplanation

Evaluates with detailed explanation trace.

```go
func (e *Engine) EvaluateWithExplanation(
    expr *CompiledExpression, 
    payload interface{},
) (types.Value, *eval.Explanation, error)
```

**Example:**

```go
result, explanation, err := eng.EvaluateWithExplanation(compiled, payload)
if err == nil {
    fmt.Printf("Expression: %s\n", explanation.Expression)
    fmt.Printf("Result: %v\n", explanation.Result.Raw)
    fmt.Printf("Reason: %s\n", explanation.Reason)
}
```

---

#### RegisterFunction

Registers a JavaScript function.

```go
func (e *Engine) RegisterFunction(source string) error
```

**Example:**

```go
err := eng.RegisterFunction(`
    function double(x) {
        return x * 2;
    }
`)
```

---

#### RegisterBuiltIn

Registers a Go built-in function.

```go
func (e *Engine) RegisterBuiltIn(
    name string, 
    fn func(args ...types.Value) (types.Value, error),
    sig *types.FunctionSignature,
) error
```

**Example:**

```go
eng.RegisterBuiltIn(
    "customMax",
    func(args ...types.Value) (types.Value, error) {
        a, _ := args[0].AsFloat()
        b, _ := args[1].AsFloat()
        if a > b {
            return types.Float(a), nil
        }
        return types.Float(b), nil
    },
    types.NewFunctionSignature("customMax", types.TypeFloat,
        types.Param("a", types.TypeFloat),
        types.Param("b", types.TypeFloat),
    ),
)
```

---

#### GetRegistry

Returns the function registry.

```go
func (e *Engine) GetRegistry() *functions.Registry
```

---

#### GetSandbox

Returns the JavaScript sandbox.

```go
func (e *Engine) GetSandbox() *functions.Sandbox
```

---

#### GetOptimizer

Returns the AST optimizer (nil if disabled).

```go
func (e *Engine) GetOptimizer() *optimizer.Optimizer
```

---

### Engine Options

#### WithTimeout

Sets the maximum execution timeout.

```go
func WithTimeout(d time.Duration) Option
```

**Default:** 100ms

---

#### WithCaching

Enables/disables expression caching.

```go
func WithCaching(enabled bool) Option
```

**Default:** false

---

#### WithExplainMode

Enables/disables explanation generation.

```go
func WithExplainMode(enabled bool) Option
```

**Default:** false

---

#### WithStrictTypes

Enables/disables strict type checking.

```go
func WithStrictTypes(enabled bool) Option
```

**Default:** false

---

#### WithOptimization

Enables/disables AST optimization.

```go
func WithOptimization(enabled bool) Option
```

**Default:** true

---

#### WithSandboxConfig

Configures the JavaScript sandbox.

```go
func WithSandboxConfig(config *functions.SandboxConfig) Option
```

**Example:**

```go
config := &functions.SandboxConfig{
    Timeout:       200 * time.Millisecond,
    MemoryLimit:   5 * 1024 * 1024,
    MaxStackDepth: 50,
}
eng, _ := engine.New(engine.WithSandboxConfig(config))
```

---

#### WithSandbox

Uses a pre-configured sandbox instance.

```go
func WithSandbox(sandbox *functions.Sandbox) Option
```

---

### CompiledExpression

```go
type CompiledExpression struct {
    AST    ast.Expression  // Parsed AST
    Source string          // Original source
}
```

---

### Convenience Functions

#### Eval

Quick evaluation without creating an engine.

```go
func Eval(dsl string, payload interface{}) (types.Value, error)
```

---

#### EvalBool

Quick boolean evaluation.

```go
func EvalBool(dsl string, payload interface{}) (bool, error)
```

---

#### MustEval

Evaluation that panics on error.

```go
func MustEval(dsl string, payload interface{}) types.Value
```

---

#### MustEvalBool

Boolean evaluation that panics on error.

```go
func MustEvalBool(dsl string, payload interface{}) bool
```

---

## Parser Package

```go
import "github.com/bencagri/amel/pkg/parser"
```

### Parse

Parses an AMEL expression string into an AST.

```go
func Parse(input string) (ast.Expression, error)
```

**Example:**

```go
expr, err := parser.Parse(`$.age >= 18 && $.active == true`)
if err != nil {
    log.Fatal(err)
}
```

---

## Compiler Package

```go
import "github.com/bencagri/amel/pkg/compiler"
```

### SQL Compiler

#### NewSQLCompiler

Creates a new SQL compiler.

```go
func NewSQLCompiler(opts ...SQLOption) *SQLCompiler
```

---

#### Compile

Compiles an AST to SQL.

```go
func (c *SQLCompiler) Compile(expr ast.Expression) (*SQLResult, error)
```

---

#### SQLResult

```go
type SQLResult struct {
    SQL    string        // SQL WHERE clause
    Params []interface{} // Parameter values
}
```

---

#### SQL Options

```go
func WithDialect(dialect SQLDialect) SQLOption
func WithFieldMapper(mapper func(string) string) SQLOption
func WithInlineParams(inline bool) SQLOption
```

---

#### SQL Dialects

```go
const (
    DialectStandard SQLDialect = iota
    DialectPostgres
    DialectMySQL
    DialectSQLite
)
```

---

#### CompileToSQL

Convenience function for quick compilation.

```go
func CompileToSQL(expr ast.Expression) (*SQLResult, error)
```

---

### MongoDB Compiler

#### NewMongoDBCompiler

Creates a new MongoDB compiler.

```go
func NewMongoDBCompiler(opts ...MongoOption) *MongoDBCompiler
```

---

#### Compile

Compiles an AST to MongoDB query.

```go
func (c *MongoDBCompiler) Compile(expr ast.Expression) (*MongoDBResult, error)
```

---

#### MongoDBResult

```go
type MongoDBResult struct {
    Query map[string]interface{} // MongoDB query document
}

func (r *MongoDBResult) ToJSON() (string, error)
func (r *MongoDBResult) ToPrettyJSON() (string, error)
```

---

#### MongoDB Options

```go
func WithMongoFieldMapper(mapper func(string) string) MongoOption
```

---

#### CompileToMongoDB

Convenience function for quick compilation.

```go
func CompileToMongoDB(expr ast.Expression) (*MongoDBResult, error)
```

---

## Types Package

```go
import "github.com/bencagri/amel/pkg/types"
```

### Type Constants

```go
const (
    TypeUnknown Type = iota
    TypeInt
    TypeFloat
    TypeString
    TypeBool
    TypeNull
    TypeList
    TypeAny
)
```

---

### Value

Represents a typed value in AMEL.

```go
type Value struct {
    Type Type
    Raw  interface{}
}
```

#### Value Constructors

```go
func Int(v int64) Value
func Float(v float64) Value
func String(v string) Value
func Bool(v bool) Value
func Null() Value
func List(values ...Value) Value
func Any(v interface{}) Value
```

---

#### Value Methods

```go
func (v Value) AsInt() (int64, bool)
func (v Value) AsFloat() (float64, bool)
func (v Value) AsString() (string, bool)
func (v Value) AsBool() (bool, bool)
func (v Value) AsList() ([]Value, bool)
func (v Value) IsTruthy() bool
func (v Value) IsNull() bool
```

---

### FunctionSignature

```go
type FunctionSignature struct {
    Name       string
    Parameters []ParameterDef
    ReturnType Type
    Variadic   bool
}
```

---

#### NewFunctionSignature

Creates a function signature.

```go
func NewFunctionSignature(name string, returnType Type, params ...ParameterDef) *FunctionSignature
```

---

#### NewVariadicSignature

Creates a variadic function signature.

```go
func NewVariadicSignature(name string, returnType Type, params ...ParameterDef) *FunctionSignature
```

---

#### Param

Creates a parameter definition.

```go
func Param(name string, t Type) ParameterDef
```

**Example:**

```go
sig := types.NewFunctionSignature("add", types.TypeInt,
    types.Param("a", types.TypeInt),
    types.Param("b", types.TypeInt),
)
```

---

## Functions Package

```go
import "github.com/bencagri/amel/pkg/functions"
```

### Registry

Manages function definitions.

#### NewRegistry

Creates a new empty registry.

```go
func NewRegistry() *Registry
```

---

#### NewDefaultRegistry

Creates a registry with all built-in functions.

```go
func NewDefaultRegistry() (*Registry, error)
```

---

#### Registry Methods

```go
func (r *Registry) Register(name string, fn *Function) error
func (r *Registry) RegisterBuiltIn(name string, fn BuiltInFunc, sig *types.FunctionSignature) error
func (r *Registry) RegisterOverload(fn *Function) error
func (r *Registry) Get(name string) (*Function, bool)
func (r *Registry) GetBestMatch(name string, args []types.Value) (*Function, bool)
func (r *Registry) Has(name string) bool
func (r *Registry) IsOverloaded(name string) bool
func (r *Registry) ListOverloads(name string) []*Function
func (r *Registry) Unregister(name string) bool
func (r *Registry) List() []string
func (r *Registry) Count() int
func (r *Registry) CountUnique() int
func (r *Registry) Call(name string, args ...types.Value) (types.Value, error)
```

---

### Function

```go
type Function struct {
    Name      string
    Signature *types.FunctionSignature
    BuiltIn   func(args ...types.Value) (types.Value, error)
    JSBody    string
}

func (f *Function) IsJS() bool
func (f *Function) IsBuiltIn() bool
```

---

### Sandbox

Secure JavaScript execution environment.

#### NewSandbox

Creates a new sandbox.

```go
func NewSandbox(config *SandboxConfig) *Sandbox
```

---

#### SandboxConfig

```go
type SandboxConfig struct {
    Timeout       time.Duration
    MemoryLimit   int64
    MaxStackDepth int
}
```

**Defaults:**

- Timeout: 100ms
- MemoryLimit: 10MB
- MaxStackDepth: 100

---

#### Sandbox Methods

```go
func (s *Sandbox) Execute(ctx context.Context, jsBody, funcName string, args []types.Value) (types.Value, error)
func (s *Sandbox) ExecuteExpression(ctx context.Context, expression string) (types.Value, error)
func (s *Sandbox) SetTimeout(d time.Duration)
func (s *Sandbox) SetMemoryLimit(bytes int64)
func (s *Sandbox) SetMaxStackDepth(depth int)
func (s *Sandbox) Config() *SandboxConfig
```

---

#### ParseJSFunction

Parses a JavaScript function definition.

```go
func ParseJSFunction(source string) (name string, params []string, returnType types.Type, body string, err error)
```

---

## Evaluator Package

```go
import "github.com/bencagri/amel/pkg/eval"
```

### Evaluator

#### New

Creates a new evaluator.

```go
func New() (*Evaluator, error)
```

---

#### Evaluator Methods

```go
func (e *Evaluator) Evaluate(expr ast.Expression, ctx *Context) (types.Value, error)
func (e *Evaluator) EvaluateBool(expr ast.Expression, ctx *Context) (bool, error)
func (e *Evaluator) EvaluateWithExplanation(expr ast.Expression, ctx *Context) (types.Value, *Explanation, error)
```

---

### Context

Evaluation context with payload and functions.

```go
func NewContext(payload interface{}) (*Context, error)
func NewContextWithRegistry(payload interface{}, registry *functions.Registry) (*Context, error)
```

---

### Explanation

```go
type Explanation struct {
    Expression string
    Result     types.Value
    Children   []*Explanation
    Reason     string
}
```

---

## Error Handling

### Error Types

```go
type ErrorCode int

const (
    // Lexer errors (1xx)
    ErrUnexpectedCharacter ErrorCode = 100
    ErrUnterminatedString  ErrorCode = 101
    ErrInvalidNumber       ErrorCode = 102

    // Parser errors (2xx)
    ErrUnexpectedToken     ErrorCode = 200
    ErrMissingExpression   ErrorCode = 201
    ErrUnmatchedParen      ErrorCode = 202
    ErrInvalidSyntax       ErrorCode = 203

    // Type errors (3xx)
    ErrTypeMismatch        ErrorCode = 300
    ErrUndefinedFunction   ErrorCode = 301
    ErrArgumentCount       ErrorCode = 302
    ErrArgumentType        ErrorCode = 303
    ErrInvalidOperator     ErrorCode = 304

    // Runtime errors (4xx)
    ErrDivisionByZero      ErrorCode = 400
    ErrNullReference       ErrorCode = 401
    ErrIndexOutOfBounds    ErrorCode = 402
    ErrTimeout             ErrorCode = 403
    ErrMemoryLimit         ErrorCode = 404
    ErrSandboxViolation    ErrorCode = 405

    // JSONPath errors (5xx)
    ErrInvalidPath         ErrorCode = 500
    ErrPathNotFound        ErrorCode = 501
)
```

---

### Error Structure

```go
type Error struct {
    Code    ErrorCode
    Message string
    Line    int
    Column  int
    Cause   error
}

func (e *Error) Error() string
func (e *Error) Unwrap() error
```

---

## See Also

- [Getting Started](./getting-started.md) - Quick introduction
- [Expression Syntax](./syntax.md) - Language reference
- [Built-in Functions](./functions.md) - Function documentation
- [Custom Functions](./custom-functions.md) - Extending AMEL