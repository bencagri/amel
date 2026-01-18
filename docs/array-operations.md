# Array Operations

AMEL provides powerful array manipulation capabilities through lambda expressions and dedicated array functions. This guide covers all array operations in detail.

## Table of Contents

- [Overview](#overview)
- [Lambda Expressions](#lambda-expressions)
- [Map](#map)
- [Filter](#filter)
- [Reduce](#reduce)
- [Find](#find)
- [Some](#some)
- [Every](#every)
- [Other Array Functions](#other-array-functions)
- [Chaining Operations](#chaining-operations)
- [Performance Tips](#performance-tips)

---

## Overview

Array operations in AMEL allow you to transform, filter, and aggregate collections of data using a functional programming style.

```go
// Transform each element
result, _ := eng.EvaluateDirect(`map([1, 2, 3], x => x * 2)`, nil)
// Result: [2, 4, 6]

// Filter elements
result, _ := eng.EvaluateDirect(`filter([1, 2, 3, 4, 5], x => x > 2)`, nil)
// Result: [3, 4, 5]

// Aggregate to single value
result, _ := eng.EvaluateDirect(`reduce([1, 2, 3, 4], 0, (acc, x) => acc + x)`, nil)
// Result: 10
```

---

## Lambda Expressions

Lambda expressions define inline functions for array operations.

### Single Parameter Lambda

Used with `map`, `filter`, `find`, `some`, and `every`:

```
x => expression
```

**Examples:**

```
x => x * 2           // Double the value
x => x > 5           // Check if greater than 5
x => x + 1           // Increment by 1
x => x % 2 == 0      // Check if even
x => upper(x)        // Convert to uppercase
```

### Two Parameter Lambda

Used with `reduce`:

```
(accumulator, element) => expression
```

**Examples:**

```
(acc, x) => acc + x       // Sum
(acc, x) => acc * x       // Product
(sum, item) => sum + item.price    // Sum of property
(result, x) => result && x > 0     // All positive
```

### Lambda Scope

The lambda parameter shadows any payload variables with the same name:

```go
payload := map[string]interface{}{
    "x": 100,
    "numbers": []interface{}{1, 2, 3},
}

// Inside the lambda, x refers to the current element, not payload's x
result, _ := eng.EvaluateDirect(`map($.numbers, x => x * 2)`, payload)
// Result: [2, 4, 6]
```

---

## Map

Transforms each element in a list using a lambda function.

### Syntax

```
map(list, lambda) -> list
```

### Basic Examples

```
// Double each number
map([1, 2, 3], x => x * 2)
// Result: [2, 4, 6]

// Increment each number
map([1, 2, 3], x => x + 1)
// Result: [2, 3, 4]

// Square each number
map([1, 2, 3, 4], x => x * x)
// Result: [1, 4, 9, 16]
```

### With JSONPath

```go
payload := map[string]interface{}{
    "numbers": []interface{}{1, 2, 3, 4, 5},
}

result, _ := eng.EvaluateDirect(`map($.numbers, x => x * 2)`, payload)
// Result: [2, 4, 6, 8, 10]
```

### Type Conversion

```
// Convert numbers to booleans (is greater than 2)
map([1, 2, 3, 4], x => x > 2)
// Result: [false, false, true, true]

// Convert to strings (requires custom function)
map([1, 2, 3], x => string(x))
// Result: ["1", "2", "3"]
```

### Empty List

```
map([], x => x * 2)
// Result: []
```

---

## Filter

Selects elements that satisfy a predicate.

### Syntax

```
filter(list, predicate) -> list
```

### Basic Examples

```
// Numbers greater than 2
filter([1, 2, 3, 4, 5], x => x > 2)
// Result: [3, 4, 5]

// Even numbers
filter([1, 2, 3, 4, 5, 6], x => x % 2 == 0)
// Result: [2, 4, 6]

// Odd numbers
filter([1, 2, 3, 4, 5], x => x % 2 != 0)
// Result: [1, 3, 5]
```

### With JSONPath

```go
payload := map[string]interface{}{
    "ages": []interface{}{10, 15, 18, 21, 25, 30},
}

result, _ := eng.EvaluateDirect(`filter($.ages, x => x >= 18)`, payload)
// Result: [18, 21, 25, 30]
```

### Filter All / None

```
// All match
filter([1, 2, 3], x => x > 0)
// Result: [1, 2, 3]

// None match
filter([1, 2, 3], x => x > 10)
// Result: []

// Empty input
filter([], x => x > 0)
// Result: []
```

### Complex Predicates

```
// Multiple conditions
filter([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], x => x > 3 && x < 8)
// Result: [4, 5, 6, 7]

// Divisibility
filter([10, 15, 20, 25, 30], x => x % 5 == 0 && x % 10 == 0)
// Result: [10, 20, 30]
```

---

## Reduce

Aggregates a list to a single value using an accumulator.

### Syntax

```
reduce(list, initialValue, lambda) -> any
```

**Parameters:**

- `list`: The array to reduce
- `initialValue`: Starting value for the accumulator
- `lambda`: `(accumulator, element) => newAccumulator`

### Sum

```
// Sum of numbers
reduce([1, 2, 3, 4, 5], 0, (acc, x) => acc + x)
// Result: 15

// With initial value
reduce([1, 2, 3], 10, (acc, x) => acc + x)
// Result: 16 (10 + 1 + 2 + 3)
```

### Product

```
// Product of numbers
reduce([1, 2, 3, 4], 1, (acc, x) => acc * x)
// Result: 24 (1 * 1 * 2 * 3 * 4)
```

### With JSONPath

```go
payload := map[string]interface{}{
    "prices": []interface{}{10.0, 20.0, 30.0},
}

result, _ := eng.EvaluateDirect(`reduce($.prices, 0, (acc, x) => acc + x)`, payload)
// Result: 60
```

### Finding Maximum

```
// Maximum value
reduce([3, 1, 4, 1, 5, 9], 0, (max, x) => ifThenElse(x > max, x, max))
// Result: 9
```

### Edge Cases

```
// Empty list returns initial value
reduce([], 100, (acc, x) => acc + x)
// Result: 100

// Single element
reduce([5], 10, (acc, x) => acc + x)
// Result: 15
```

---

## Find

Returns the first element that satisfies a predicate.

### Syntax

```
find(list, predicate) -> any | null
```

### Basic Examples

```
// First number greater than 3
find([1, 2, 3, 4, 5], x => x > 3)
// Result: 4

// First even number
find([1, 3, 5, 6, 7], x => x % 2 == 0)
// Result: 6
```

### With JSONPath

```go
payload := map[string]interface{}{
    "scores": []interface{}{70, 85, 95, 100},
}

result, _ := eng.EvaluateDirect(`find($.scores, x => x >= 90)`, payload)
// Result: 95
```

### Not Found

```
// No match returns null
find([1, 2, 3], x => x > 10)
// Result: null

// Empty list returns null
find([], x => x > 0)
// Result: null
```

### Checking for Null Result

```go
result, _ := eng.EvaluateDirect(`
    coalesce(
        find($.items, x => x > 100),
        -1
    )
`, payload)
// Returns -1 if not found
```

---

## Some

Checks if at least one element satisfies a predicate.

### Syntax

```
some(list, predicate) -> bool
```

### Basic Examples

```
// Any number greater than 2?
some([1, 2, 3], x => x > 2)
// Result: true

// Any number greater than 10?
some([1, 2, 3], x => x > 10)
// Result: false
```

### With JSONPath

```go
payload := map[string]interface{}{
    "users": []interface{}{
        map[string]interface{}{"role": "user"},
        map[string]interface{}{"role": "admin"},
        map[string]interface{}{"role": "user"},
    },
}

// Note: Object property access in lambdas works with JSONPath payload
result, _ := eng.EvaluateDirect(`some($.users, x => x.role == "admin")`, payload)
// Result: true
```

### Short-Circuit Evaluation

`some` stops evaluating as soon as it finds a true result:

```
// Stops after finding first match
some([1, 2, 3, 4, 5], x => x == 2)
// Only evaluates x=1 and x=2, then returns true
```

### Edge Cases

```
// Empty list
some([], x => x > 0)
// Result: false

// All match
some([1, 2, 3], x => x > 0)
// Result: true

// None match
some([1, 2, 3], x => x > 10)
// Result: false
```

---

## Every

Checks if all elements satisfy a predicate.

### Syntax

```
every(list, predicate) -> bool
```

### Basic Examples

```
// All positive?
every([1, 2, 3], x => x > 0)
// Result: true

// All greater than 1?
every([1, 2, 3], x => x > 1)
// Result: false
```

### With JSONPath

```go
payload := map[string]interface{}{
    "items": []interface{}{
        map[string]interface{}{"inStock": true},
        map[string]interface{}{"inStock": true},
        map[string]interface{}{"inStock": false},
    },
}

result, _ := eng.EvaluateDirect(`every($.items, x => x.inStock == true)`, payload)
// Result: false
```

### Short-Circuit Evaluation

`every` stops evaluating as soon as it finds a false result:

```
// Stops after finding first non-match
every([5, 4, 3, 2, 1], x => x > 2)
// Evaluates until x=2, then returns false
```

### Edge Cases

```
// Empty list (vacuous truth)
every([], x => x > 0)
// Result: true

// All match
every([1, 2, 3], x => x > 0)
// Result: true

// None match
every([1, 2, 3], x => x > 10)
// Result: false
```

---

## Other Array Functions

### Basic Operations

```
// Length
len([1, 2, 3])                      // 3

// First element
first([1, 2, 3])                    // 1

// Last element
last([1, 2, 3])                     // 3

// Element at index
at([10, 20, 30], 1)                 // 20

// Index of element
indexOf([1, 2, 3], 2)               // 1 (returns -1 if not found)
```

### Transformation

```
// Reverse
reverse([1, 2, 3])                  // [3, 2, 1]

// Unique values
unique([1, 2, 2, 3, 3, 3])          // [1, 2, 3]

// Flatten nested arrays
flatten([[1, 2], [3, 4]])           // [1, 2, 3, 4]

// Slice (start, end)
slice([1, 2, 3, 4, 5], 1, 4)        // [2, 3, 4]
```

### Sorting

```
// Sort ascending
sortAsc([3, 1, 4, 1, 5])            // [1, 1, 3, 4, 5]

// Sort descending
sortDesc([3, 1, 4, 1, 5])           // [5, 4, 3, 1, 1]
```

### Aggregation

```
// Sum
sum([1, 2, 3, 4])                   // 10

// Average
avg([10, 20, 30])                   // 20.0

// Minimum
min([5, 2, 8, 1])                   // 1

// Maximum
max([5, 2, 8, 1])                   // 8

// Count
count([1, 2, 3])                    // 3
```

### Boolean Aggregation

```
// All true?
all([true, true, true])             // true
all([true, false, true])            // false

// Any true?
any([false, false, true])           // true
any([false, false, false])          // false
```

---

## Chaining Operations

Combine multiple array operations for complex transformations.

### Map then Filter

```
// Double then keep values > 5
filter(map([1, 2, 3, 4, 5], x => x * 2), y => y > 5)
// Step 1: map -> [2, 4, 6, 8, 10]
// Step 2: filter -> [6, 8, 10]
```

### Filter then Map

```
// Keep > 2 then double
map(filter([1, 2, 3, 4, 5], x => x > 2), y => y * 2)
// Step 1: filter -> [3, 4, 5]
// Step 2: map -> [6, 8, 10]
```

### Filter then Reduce

```
// Sum of even numbers
reduce(filter([1, 2, 3, 4, 5, 6], x => x % 2 == 0), 0, (acc, x) => acc + x)
// Step 1: filter -> [2, 4, 6]
// Step 2: reduce -> 12
```

### Complex Pipeline

```
// Get sum of squared even numbers
reduce(
    map(
        filter([1, 2, 3, 4, 5, 6], x => x % 2 == 0),
        y => y * y
    ),
    0,
    (acc, z) => acc + z
)
// Step 1: filter -> [2, 4, 6]
// Step 2: map squares -> [4, 16, 36]
// Step 3: reduce sum -> 56
```

### Practical Example: Order Total

```go
payload := map[string]interface{}{
    "items": []interface{}{
        map[string]interface{}{"name": "Widget", "price": 10.0, "quantity": 2, "taxable": true},
        map[string]interface{}{"name": "Gadget", "price": 25.0, "quantity": 1, "taxable": false},
        map[string]interface{}{"name": "Thing", "price": 5.0, "quantity": 4, "taxable": true},
    },
    "taxRate": 0.1,
}

// Calculate total including tax for taxable items
result, _ := eng.EvaluateDirect(`
    reduce(
        map($.items, item => item.price * item.quantity * 
            ifThenElse(item.taxable, 1 + $.taxRate, 1)),
        0,
        (total, subtotal) => total + subtotal
    )
`, payload)
```

---

## Performance Tips

### 1. Filter First

Reduce the dataset before expensive operations:

```
// Better: filter first
map(filter(largeArray, x => x.active), y => expensiveOperation(y))

// Worse: process all, then filter
filter(map(largeArray, x => expensiveOperation(x)), y => y.active)
```

### 2. Use Early Exit Functions

`some` and `every` stop processing when the result is determined:

```
// Stops at first match (efficient for large arrays)
some(largeArray, x => x == targetValue)

// Must check all elements
len(filter(largeArray, x => x == targetValue)) > 0
```

### 3. Avoid Redundant Operations

```
// Inefficient: two passes
len(filter(arr, x => x > 0)) > 0

// Efficient: one pass with early exit
some(arr, x => x > 0)
```

### 4. Use Built-in Functions

Built-in functions are optimized. Use them instead of reduce when possible:

```
// Use built-in sum
sum([1, 2, 3, 4, 5])

// Instead of reduce
reduce([1, 2, 3, 4, 5], 0, (acc, x) => acc + x)
```

### 5. Consider SQL/MongoDB Compilation

For database queries, compile expressions to native queries rather than fetching all data:

```go
// Efficient: database does the filtering
expr, _ := parser.Parse(`$.status == "active" && $.age >= 18`)
mongoQuery, _ := compiler.NewMongoDBCompiler().Compile(expr)
cursor, _ := collection.Find(ctx, mongoQuery.Query)

// Inefficient: fetch all, filter in app
cursor, _ := collection.Find(ctx, bson.M{})
// ...then filter with AMEL
```

---

## See Also

- [Built-in Functions](./functions.md) - Complete function reference
- [Expression Syntax](./syntax.md) - Lambda syntax details
- [Custom Functions](./custom-functions.md) - Writing array-processing functions