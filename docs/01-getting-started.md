# Getting Started with AMEL

This guide will help you get up and running with AMEL (Adaptive Matching Expression Language) quickly.

## Installation

### Requirements

- Go 1.21 or higher

### Install via Go Modules

```bash
go get github.com/bencagri/amel
```

### Import in Your Project

```go
import (
    "github.com/bencagri/amel/pkg/engine"
)
```

## Basic Concepts

### The Engine

The `Engine` is the main entry point for AMEL. It handles:

- Parsing and compiling expressions
- Evaluating expressions against payloads
- Managing function registries
- Caching compiled expressions

```go
// Create a new engine with default settings
eng, err := engine.New()
if err != nil {
    log.Fatal(err)
}
```

### Expressions

AMEL expressions are strings that describe conditions or computations:

```go
// Comparison
"$.age >= 18"

// Logical operations
"$.age >= 18 && $.verified == true"

// Arithmetic
"$.price * $.quantity"

// Function calls
"upper($.name)"
```

### Payloads

Payloads are JSON-like data structures (Go maps) that expressions evaluate against:

```go
payload := map[string]interface{}{
    "user": map[string]interface{}{
        "name":     "John Doe",
        "age":      25,
        "email":    "john@example.com",
        "verified": true,
    },
    "order": map[string]interface{}{
        "total": 150.50,
        "items": 3,
    },
}
```

## Your First Expression

### Step 1: Create an Engine

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/bencagri/amel/pkg/engine"
)

func main() {
    eng, err := engine.New()
    if err != nil {
        log.Fatal(err)
    }
    
    // Engine is ready to use
}
```

### Step 2: Define Your Data

```go
payload := map[string]interface{}{
    "user": map[string]interface{}{
        "name": "Alice",
        "age":  30,
        "role": "admin",
    },
}
```

### Step 3: Evaluate Expressions

```go
// Direct evaluation (compiles and evaluates in one step)
result, err := eng.EvaluateDirect(`$.user.age >= 18`, payload)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Result: %v\n", result.Raw) // Output: true
```

## Compiled Expressions

For expressions that will be evaluated multiple times, compile once and reuse:

```go
// Compile the expression
compiled, err := eng.Compile(`$.user.role IN ["admin", "moderator"]`)
if err != nil {
    log.Fatal(err)
}

// Evaluate against different payloads
for _, user := range users {
    payload := map[string]interface{}{"user": user}
    result, err := eng.Evaluate(compiled, payload)
    if err != nil {
        log.Printf("Error: %v", err)
        continue
    }
    fmt.Printf("User %s: %v\n", user["name"], result.Raw)
}
```

## Engine Options

Configure the engine with various options:

```go
eng, err := engine.New(
    // Set execution timeout
    engine.WithTimeout(100 * time.Millisecond),
    
    // Enable explanation mode for debugging
    engine.WithExplainMode(true),
    
    // Enable expression caching
    engine.WithCaching(true),
    
    // Enable strict type checking
    engine.WithStrictTypes(true),
)
```

### Available Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithTimeout(d)` | Maximum execution time | 100ms |
| `WithExplainMode(bool)` | Enable evaluation explanations | false |
| `WithCaching(bool)` | Cache compiled expressions | false |
| `WithStrictTypes(bool)` | Enforce strict type checking | false |
| `WithOptimization(bool)` | Enable AST optimization | true |
| `WithSandboxConfig(cfg)` | Configure JS sandbox | default |

## Evaluation Methods

### EvaluateDirect

Compiles and evaluates in one step. Best for one-off evaluations:

```go
result, err := eng.EvaluateDirect(`$.price > 100`, payload)
```

### EvaluateDirectBool

Same as `EvaluateDirect` but returns a boolean directly:

```go
isExpensive, err := eng.EvaluateDirectBool(`$.price > 100`, payload)
if isExpensive {
    // Apply discount
}
```

### Compile + Evaluate

Compile once, evaluate many times:

```go
compiled, err := eng.Compile(`$.status == "active"`)
// ...later...
result, err := eng.Evaluate(compiled, payload)
```

### EvaluateBool

Get a boolean result from a compiled expression:

```go
isActive, err := eng.EvaluateBool(compiled, payload)
```

### EvaluateWithExplanation

Get detailed evaluation trace (requires `WithExplainMode(true)`):

```go
eng, _ := engine.New(engine.WithExplainMode(true))
compiled, _ := eng.Compile(`$.age >= 18 && $.verified`)

result, explanation, err := eng.EvaluateWithExplanation(compiled, payload)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Result: %v\n", result.Raw)
fmt.Printf("Expression: %s\n", explanation.Expression)
fmt.Printf("Reason: %s\n", explanation.Reason)
for _, child := range explanation.Children {
    fmt.Printf("  - %s: %v (%s)\n", child.Expression, child.Result.Raw, child.Reason)
}
```

## Working with Results

The `Value` type provides methods to safely extract typed values:

```go
result, _ := eng.EvaluateDirect(`$.count * 2`, payload)

// Check the type
fmt.Println(result.Type) // types.TypeInt

// Extract as specific types
if intVal, ok := result.AsInt(); ok {
    fmt.Printf("Integer: %d\n", intVal)
}

if floatVal, ok := result.AsFloat(); ok {
    fmt.Printf("Float: %f\n", floatVal)
}

if strVal, ok := result.AsString(); ok {
    fmt.Printf("String: %s\n", strVal)
}

if boolVal, ok := result.AsBool(); ok {
    fmt.Printf("Boolean: %v\n", boolVal)
}

if listVal, ok := result.AsList(); ok {
    fmt.Printf("List length: %d\n", len(listVal))
}

// Check truthiness
if result.IsTruthy() {
    fmt.Println("Result is truthy")
}
```

## Error Handling

AMEL provides detailed error information:

```go
result, err := eng.EvaluateDirect(`$.invalid.path`, payload)
if err != nil {
    // Check error type
    switch e := err.(type) {
    case *errors.Error:
        fmt.Printf("Error Code: %d\n", e.Code)
        fmt.Printf("Message: %s\n", e.Message)
        fmt.Printf("Line: %d, Column: %d\n", e.Line, e.Column)
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Complete Example

Here's a complete example demonstrating common use cases:

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/bencagri/amel/pkg/engine"
)

func main() {
    // Create engine with options
    eng, err := engine.New(
        engine.WithTimeout(100*time.Millisecond),
        engine.WithCaching(true),
        engine.WithExplainMode(true),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Sample payload
    payload := map[string]interface{}{
        "user": map[string]interface{}{
            "name":     "John Doe",
            "age":      25,
            "email":    "john@example.com",
            "verified": true,
            "roles":    []interface{}{"user", "admin"},
        },
        "order": map[string]interface{}{
            "total":    299.99,
            "currency": "USD",
            "items":    []interface{}{
                map[string]interface{}{"name": "Widget", "price": 99.99},
                map[string]interface{}{"name": "Gadget", "price": 200.00},
            },
        },
    }

    // Example 1: Simple condition
    isAdult, _ := eng.EvaluateDirectBool(`$.user.age >= 18`, payload)
    fmt.Printf("Is adult: %v\n", isAdult)

    // Example 2: Complex condition
    canCheckout, _ := eng.EvaluateDirectBool(
        `$.user.verified == true && $.order.total > 0 && "admin" IN $.user.roles`,
        payload,
    )
    fmt.Printf("Can checkout: %v\n", canCheckout)

    // Example 3: String manipulation
    result, _ := eng.EvaluateDirect(`upper($.user.name)`, payload)
    fmt.Printf("Uppercase name: %v\n", result.Raw)

    // Example 4: Arithmetic
    result, _ = eng.EvaluateDirect(`$.order.total * 1.1`, payload) // Add 10% tax
    fmt.Printf("Total with tax: %v\n", result.Raw)

    // Example 5: With explanation
    compiled, _ := eng.Compile(`$.user.age >= 21 || "admin" IN $.user.roles`)
    _, explanation, _ := eng.EvaluateWithExplanation(compiled, payload)
    fmt.Printf("\nExpression: %s\n", explanation.Expression)
    fmt.Printf("Result: %v\n", explanation.Result.Raw)
    fmt.Printf("Reason: %s\n", explanation.Reason)
}
```

## Next Steps

- Learn the full [Expression Syntax](./02-syntax.md)
- Explore [Built-in Functions](./03-functions.md)
- Create [Custom Functions](./04-custom-functions.md)
- Compile to [SQL](./06-sql-compilation.md) or [MongoDB](./07-mongodb-compilation.md)
