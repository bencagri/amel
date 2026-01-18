# Custom Functions

AMEL allows you to extend its functionality by writing custom functions in JavaScript. These functions run in a secure sandbox with configurable timeouts and resource limits.

## Table of Contents

- [Overview](#overview)
- [Registering Functions](#registering-functions)
- [Function Syntax](#function-syntax)
- [Type Annotations](#type-annotations)
- [Working with Arguments](#working-with-arguments)
- [Returning Values](#returning-values)
- [Sandbox Security](#sandbox-security)
- [Best Practices](#best-practices)
- [Examples](#examples)

---

## Overview

Custom functions in AMEL are:

- Written in JavaScript
- Executed in a secure sandbox (using goja)
- Type-safe with optional return type annotations
- Callable from AMEL expressions like built-in functions

```go
// Register a custom function
eng.RegisterFunction(`function double(x) { return x * 2; }`)

// Use it in expressions
result, _ := eng.EvaluateDirect(`double(5)`, nil)  // 10
```

---

## Registering Functions

### Basic Registration

Use `RegisterFunction` to add a JavaScript function:

```go
eng, _ := engine.New()

// Register a simple function
err := eng.RegisterFunction(`
    function greet(name) {
        return "Hello, " + name + "!";
    }
`)
if err != nil {
    log.Fatal(err)
}

// Use the function
result, _ := eng.EvaluateDirect(`greet("World")`, nil)
fmt.Println(result.Raw)  // "Hello, World!"
```

### Multiple Functions

Register multiple functions separately:

```go
eng.RegisterFunction(`function add(a, b) { return a + b; }`)
eng.RegisterFunction(`function subtract(a, b) { return a - b; }`)
eng.RegisterFunction(`function multiply(a, b) { return a * b; }`)
```

### Registering Go Built-in Functions

For performance-critical functions, register Go functions directly:

```go
import "github.com/bencagri/amel/pkg/types"

eng.RegisterBuiltIn(
    "customMax",
    func(args ...types.Value) (types.Value, error) {
        if len(args) < 2 {
            return types.Null(), fmt.Errorf("customMax requires at least 2 arguments")
        }
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

## Function Syntax

### Basic Function

```javascript
function functionName(param1, param2) {
    // function body
    return result;
}
```

### With Return Type Annotation

```javascript
function functionName(param1, param2): returnType {
    // function body
    return result;
}
```

### Arrow Functions (Not Supported at Registration)

Arrow functions cannot be registered directly, but you can use them inside regular functions:

```javascript
function sumArray(arr) {
    return arr.reduce((a, b) => a + b, 0);
}
```

---

## Type Annotations

You can optionally specify return types for better documentation and validation:

### Supported Types

| Annotation | Description |
|------------|-------------|
| `: int` | 64-bit integer |
| `: float` | 64-bit floating-point |
| `: string` | UTF-8 string |
| `: bool` | Boolean |
| `: list` | Array/list |
| `: any` | Any type (default) |

### Examples

```javascript
// Returns an integer
function square(x): int {
    return x * x;
}

// Returns a float
function divide(a, b): float {
    return a / b;
}

// Returns a string
function formatName(first, last): string {
    return first + " " + last;
}

// Returns a boolean
function isAdult(age): bool {
    return age >= 18;
}

// Returns a list
function range(n): list {
    return Array.from({length: n}, (_, i) => i);
}
```

---

## Working with Arguments

### Accessing Arguments

Arguments are passed directly to your function:

```javascript
function process(name, age, active) {
    // name: string
    // age: number
    // active: boolean
    return name + " is " + age + " years old";
}
```

### Handling Arrays

```javascript
function sumAll(numbers) {
    return numbers.reduce((sum, n) => sum + n, 0);
}
```

Usage:

```
sumAll([1, 2, 3, 4, 5])  // 15
sumAll($.prices)         // sum of prices
```

### Handling Objects

```javascript
function fullName(user) {
    return user.firstName + " " + user.lastName;
}
```

Usage:

```
fullName($.user)  // Passes the user object
```

### Optional Arguments

```javascript
function greet(name, greeting) {
    if (greeting === undefined || greeting === null) {
        greeting = "Hello";
    }
    return greeting + ", " + name + "!";
}
```

### Variadic Arguments

```javascript
function maxOf() {
    var args = Array.prototype.slice.call(arguments);
    return Math.max.apply(null, args);
}
```

---

## Returning Values

### Simple Values

```javascript
function getNumber() { return 42; }
function getString() { return "hello"; }
function getBool() { return true; }
function getNull() { return null; }
```

### Arrays

```javascript
function getList() {
    return [1, 2, 3];
}

function double(arr) {
    return arr.map(x => x * 2);
}
```

### Objects

Note: Returning objects converts them to maps in Go:

```javascript
function createPerson(name, age) {
    return {
        name: name,
        age: age,
        isAdult: age >= 18
    };
}
```

### Computed Values

```javascript
function calculateTax(amount, rate) {
    var tax = amount * (rate / 100);
    var total = amount + tax;
    return {
        tax: tax,
        total: total
    };
}
```

---

## Sandbox Security

Custom functions run in a secure sandbox with the following restrictions:

### Default Constraints

| Constraint | Default Value | Description |
|------------|---------------|-------------|
| Timeout | 100ms | Maximum execution time |
| Memory Limit | 10MB | Maximum memory usage |
| Stack Depth | 100 | Maximum call stack depth |
| Loop Iterations | 10,000 | Maximum loop iterations |

### Configuring the Sandbox

```go
import "github.com/bencagri/amel/pkg/functions"

config := &functions.SandboxConfig{
    Timeout:       200 * time.Millisecond,
    MemoryLimit:   5 * 1024 * 1024,  // 5MB
    MaxStackDepth: 50,
}

eng, _ := engine.New(engine.WithSandboxConfig(config))
```

### Restricted APIs

The following JavaScript APIs are **NOT available** in the sandbox:

- `eval`, `Function` constructor
- `require`, `import`
- `setTimeout`, `setInterval`
- `fetch`, `XMLHttpRequest`
- `process`, `__dirname`, `__filename`
- `fs`, `path`, `os`, `child_process`

### Allowed APIs

The following JavaScript APIs **ARE available**:

| API | Description |
|-----|-------------|
| `Math.*` | All math methods (sin, cos, sqrt, etc.) |
| `String.*` | All string methods |
| `Array.*` | All array methods |
| `Object.keys`, `Object.values`, `Object.entries` | Object utilities |
| `JSON.parse`, `JSON.stringify` | JSON handling |
| `Number.*`, `parseInt`, `parseFloat` | Number utilities |
| `Boolean` | Boolean conversion |
| `Date` | Date object (read-only, no timers) |
| `console.log` | Logging (for debugging) |

### Timeout Handling

Functions that exceed the timeout are terminated:

```go
// This will timeout
eng.RegisterFunction(`
    function infiniteLoop() {
        while(true) {}
        return 1;
    }
`)

_, err := eng.EvaluateDirect(`infiniteLoop()`, nil)
// err: "timeout exceeded"
```

---

## Best Practices

### 1. Keep Functions Simple

```javascript
// Good: Single responsibility
function calculateDiscount(price, percent) {
    return price * (percent / 100);
}

// Avoid: Multiple responsibilities
function processOrder(order) {
    // Too complex - break into smaller functions
}
```

### 2. Handle Edge Cases

```javascript
function safeDivide(a, b) {
    if (b === 0) {
        return null;  // or throw an error
    }
    return a / b;
}
```

### 3. Use Type Annotations

```javascript
// Clear intent with type annotation
function isEligible(age): bool {
    return age >= 18 && age <= 65;
}
```

### 4. Document Complex Logic

```javascript
function calculateCompoundInterest(principal, rate, time) {
    // Formula: A = P(1 + r/n)^(nt)
    // Assuming annual compounding (n = 1)
    var r = rate / 100;
    return principal * Math.pow(1 + r, time);
}
```

### 5. Avoid Global State

```javascript
// Bad: Uses global state
var counter = 0;
function increment() {
    counter++;
    return counter;
}

// Good: Pure function
function add(a, b) {
    return a + b;
}
```

### 6. Use Built-in Functions When Possible

```go
// Use built-in for better performance
result, _ := eng.EvaluateDirect(`max($.a, $.b)`, payload)

// Instead of custom JavaScript
eng.RegisterFunction(`
    function customMax(a, b) {
        return a > b ? a : b;
    }
`)
```

---

## Examples

### E-commerce Discount Calculator

```javascript
function calculateDiscount(total, tier) {
    var discounts = {
        "gold": 0.20,
        "silver": 0.10,
        "bronze": 0.05
    };
    var rate = discounts[tier] || 0;
    return total * rate;
}
```

Usage:

```
calculateDiscount($.order.total, $.customer.tier)
```

### Data Validation

```javascript
function isValidEmail(email): bool {
    var pattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return pattern.test(email);
}
```

Usage:

```
isValidEmail($.user.email)
```

### String Formatting

```javascript
function formatCurrency(amount, currency): string {
    var symbols = {
        "USD": "$",
        "EUR": "€",
        "GBP": "£"
    };
    var symbol = symbols[currency] || currency + " ";
    return symbol + amount.toFixed(2);
}
```

Usage:

```
formatCurrency($.price, $.currency)
```

### Array Processing

```javascript
function getTopScores(scores, n): list {
    return scores
        .slice()
        .sort(function(a, b) { return b - a; })
        .slice(0, n);
}
```

Usage:

```
getTopScores($.scores, 5)
```

### Complex Business Logic

```javascript
function evaluateRisk(user): string {
    var score = 0;
    
    // Age factor
    if (user.age >= 25 && user.age <= 55) {
        score += 20;
    }
    
    // Verification factor
    if (user.verified) {
        score += 30;
    }
    
    // History factor
    if (user.orderCount > 10) {
        score += 25;
    } else if (user.orderCount > 5) {
        score += 15;
    }
    
    // Reputation factor
    if (user.reputation > 100) {
        score += 25;
    }
    
    // Determine risk level
    if (score >= 75) return "low";
    if (score >= 50) return "medium";
    return "high";
}
```

Usage:

```
evaluateRisk($.user) == "low"
```

### Working with Dates

```javascript
function daysSince(dateString): int {
    var past = new Date(dateString);
    var now = new Date();
    var diff = now - past;
    return Math.floor(diff / (1000 * 60 * 60 * 24));
}
```

Usage:

```
daysSince($.user.registeredAt) > 30
```

### Combining with Built-in Functions

```javascript
function processItems(items) {
    return items
        .filter(function(item) { return item.active; })
        .map(function(item) { return item.value * 1.1; });
}
```

Usage in expression:

```
sum(processItems($.items)) > 1000
```

---

## Debugging Custom Functions

### Using console.log

```javascript
function debug(value) {
    console.log("Debug value:", value);
    return value * 2;
}
```

### Testing Functions

Test your functions in isolation before using in expressions:

```go
eng.RegisterFunction(`
    function myFunc(x) {
        return x * 2;
    }
`)

// Test with simple input
result, err := eng.EvaluateDirect(`myFunc(5)`, nil)
if err != nil {
    log.Printf("Function error: %v", err)
}
fmt.Printf("Result: %v\n", result.Raw)
```

---

## See Also

- [Built-in Functions](./functions.md) - All built-in functions
- [Expression Syntax](./syntax.md) - Language syntax reference
- [API Reference](./api-reference.md) - Complete Go API