# Expression Syntax

This document provides a complete reference for AMEL's expression syntax.

## Table of Contents

- [Literals](#literals)
- [Variables and JSONPath](#variables-and-jsonpath)
- [Operators](#operators)
- [Function Calls](#function-calls)
- [Lambda Expressions](#lambda-expressions)
- [Lists](#lists)
- [Operator Precedence](#operator-precedence)
- [Reserved Keywords](#reserved-keywords)
- [Comments](#comments)

## Literals

AMEL supports the following literal types:

### Integer Literals

64-bit signed integers:

```
42
-17
0
1000000
```

### Float Literals

64-bit floating-point numbers:

```
3.14
-2.5
0.0
100.0
```

### String Literals

UTF-8 strings enclosed in double or single quotes:

```
"hello world"
'single quotes work too'
"with \"escaped\" quotes"
"line\nbreak"
```

#### Escape Sequences

| Sequence | Description |
|----------|-------------|
| `\\` | Backslash |
| `\"` | Double quote |
| `\'` | Single quote |
| `\n` | Newline |
| `\r` | Carriage return |
| `\t` | Tab |

### Boolean Literals

```
true
false
```

### Null Literal

```
null
```

## Variables and JSONPath

Access data from the payload using JSONPath expressions starting with `$`:

### Basic Property Access

```
$.name              // Root-level property
$.user.name         // Nested property
$.user.address.city // Deeply nested property
```

### Array Access

```
$.items[0]          // First element
$.items[1].name     // Property of second element
$.matrix[0][1]      // Nested array access
```

### Complex Paths

```
$.users[0].addresses[0].street
$.data.results[5].metadata.tags[0]
```

### Handling Missing Paths

When a JSONPath points to a non-existent property, it returns `null`:

```
$.missing.property   // Returns null if path doesn't exist
```

Use `isNull()` or `coalesce()` to handle missing values:

```
isNull($.optional)
coalesce($.nickname, $.name, "Anonymous")
```

## Operators

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal to | `$.age == 18` |
| `!=` | Not equal to | `$.status != "deleted"` |
| `>` | Greater than | `$.price > 100` |
| `<` | Less than | `$.quantity < 10` |
| `>=` | Greater than or equal | `$.age >= 21` |
| `<=` | Less than or equal | `$.score <= 100` |

```
$.user.age >= 18
$.status == "active"
$.price != null
$.count > 0
```

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `&&` | Logical AND | `$.a && $.b` |
| `\|\|` | Logical OR | `$.a \|\| $.b` |
| `!` | Logical NOT | `!$.active` |

```
$.age >= 18 && $.verified == true
$.role == "admin" || $.role == "moderator"
!$.deleted
!($.status == "inactive")
```

**Short-circuit evaluation:** AMEL uses short-circuit evaluation for logical operators:

- `&&` stops evaluating if the left side is false
- `||` stops evaluating if the left side is true

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `$.a + $.b` |
| `-` | Subtraction | `$.a - $.b` |
| `*` | Multiplication | `$.a * $.b` |
| `/` | Division | `$.a / $.b` |
| `%` | Modulo | `$.a % $.b` |

```
$.price * $.quantity
$.total - $.discount
$.score / $.max * 100
$.index % 2 == 0
```

**String concatenation:** The `+` operator also concatenates strings:

```
"Hello, " + $.name + "!"
```

### Unary Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `-` | Numeric negation | `-$.value` |
| `!` | Logical negation | `!$.active` |

### Membership Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `IN` | Value in list | `$.role IN ["a", "b"]` |
| `NOT IN` | Value not in list | `$.status NOT IN ["x", "y"]` |

```
$.user.role IN ["admin", "moderator", "editor"]
$.status NOT IN ["deleted", "archived", "suspended"]
"premium" IN $.user.tags
```

### Regex Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=~` | Regex match | `$.email =~ "@gmail\\.com$"` |
| `!~` | Regex not match | `$.name !~ "^test"` |

```
// Email validation
$.email =~ "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"

// Phone number pattern
$.phone =~ "^\\+?[0-9]{10,14}$"

// Case insensitive match
$.status =~ "(?i)^active$"

// Word boundary
$.text =~ "\\bword\\b"

// Negated match
$.email !~ "@spam\\.com$"
```

### Index Operator

Access list elements by index:

```
$.items[0]           // First element
$.items[1]           // Second element
$.items[-1]          // Last element (negative indexing)
[1, 2, 3][0]         // Direct list access
```

## Function Calls

Call built-in or user-defined functions:

```
functionName(arg1, arg2, ...)
```

### Examples

```
// String functions
upper($.name)
lower($.email)
trim($.input)
len($.description)
contains($.text, "keyword")
startsWith($.url, "https://")
endsWith($.file, ".pdf")

// Math functions
abs($.balance)
max($.a, $.b, $.c)
min($.x, $.y)
round($.price)
floor($.value)
ceil($.score)

// Type functions
int($.stringNumber)
float($.value)
string($.count)
bool($.flag)

// Null handling
coalesce($.nickname, $.name, "Unknown")
isNull($.optional)
isNotNull($.required)

// Conditional
ifThenElse($.age >= 18, "adult", "minor")
```

See [Built-in Functions](./functions.md) for the complete list.

## Lambda Expressions

Lambda expressions are used with array operations:

### Single Parameter Lambda

```
x => x * 2
x => x > 5
x => x + 1
```

### Two Parameter Lambda (for reduce)

```
(acc, x) => acc + x
(sum, item) => sum + item.price
```

### Usage with Array Functions

```
// Map: transform each element
map([1, 2, 3], x => x * 2)                    // [2, 4, 6]

// Filter: select matching elements
filter([1, 2, 3, 4, 5], x => x > 2)          // [3, 4, 5]

// Reduce: aggregate to single value
reduce([1, 2, 3, 4], 0, (acc, x) => acc + x) // 10

// Find: first matching element
find([1, 2, 3, 4], x => x > 2)               // 3

// Some: any element matches
some([1, 2, 3], x => x > 2)                  // true

// Every: all elements match
every([1, 2, 3], x => x > 0)                 // true
```

## Lists

### List Literals

```
[]                           // Empty list
[1, 2, 3]                   // Integer list
["a", "b", "c"]             // String list
[true, false, true]         // Boolean list
[1, "mixed", true]          // Mixed types
[[1, 2], [3, 4]]           // Nested lists
```

### List Access

```
[10, 20, 30][0]             // 10
[10, 20, 30][1]             // 20
[10, 20, 30][-1]            // 30 (last element)
```

### List Operations

```
// Length
len([1, 2, 3])              // 3

// Membership
5 IN [1, 2, 3, 4, 5]        // true
"x" NOT IN ["a", "b", "c"]  // true

// Aggregation
sum([1, 2, 3, 4])           // 10
avg([10, 20, 30])           // 20
min([5, 2, 8, 1])           // 1
max([5, 2, 8, 1])           // 8

// Element access
first([1, 2, 3])            // 1
last([1, 2, 3])             // 3
at([10, 20, 30], 1)         // 20

// Transformation
reverse([1, 2, 3])          // [3, 2, 1]
unique([1, 2, 2, 3, 3])     // [1, 2, 3]
flatten([[1, 2], [3, 4]])   // [1, 2, 3, 4]
slice([1, 2, 3, 4], 1, 3)   // [2, 3]
```

## Operator Precedence

From lowest to highest precedence:

| Precedence | Operators | Associativity | Description |
|------------|-----------|---------------|-------------|
| 1 | `\|\|` | Left | Logical OR |
| 2 | `&&` | Left | Logical AND |
| 3 | `!` | Right | Logical NOT |
| 4 | `==` `!=` `>` `<` `>=` `<=` `IN` `NOT IN` `=~` `!~` | Left | Comparison |
| 5 | `+` `-` | Left | Addition, Subtraction |
| 6 | `*` `/` `%` | Left | Multiplication, Division, Modulo |
| 7 | Unary `-` | Right | Negation |
| 8 | `[]` `()` | - | Index, Function call, Grouping |

### Examples

```
// Parsed as: (a && b) || c
a && b || c

// Parsed as: a && (b || c)
a && (b || c)

// Parsed as: (a > b) && (c < d)
a > b && c < d

// Parsed as: a + (b * c)
a + b * c

// Parsed as: (a + b) * c
(a + b) * c

// Parsed as: !(a == b)
!a == b

// Parsed as: (!a) == b (use parentheses)
(!a) == b
```

## Reserved Keywords

The following words are reserved and cannot be used as identifiers:

```
true
false
null
IN
NOT
AND
OR
```

**Note:** `AND` and `OR` are reserved but `&&` and `||` are the actual operators.

## Comments

AMEL supports both single-line and block comments:

### Single-line Comments

```
// This is a comment
$.age >= 18  // Check if adult
```

### Block Comments

```
/* This is a
   multi-line comment */
$.user.name
```

## Grammar Reference (EBNF)

For completeness, here's the formal grammar:

```ebnf
Expression     = LogicalOr ;

LogicalOr      = LogicalAnd { "||" LogicalAnd } ;
LogicalAnd     = LogicalNot { "&&" LogicalNot } ;
LogicalNot     = "!" LogicalNot | Comparison ;

Comparison     = Arithmetic [ CompOp Arithmetic ] ;
CompOp         = "==" | "!=" | ">" | "<" | ">=" | "<=" 
               | "IN" | "NOT IN" | "=~" | "!~" ;

Arithmetic     = Term { ("+" | "-") Term } ;
Term           = Factor { ("*" | "/" | "%") Factor } ;
Factor         = Unary ;
Unary          = "-" Unary | Primary ;

Primary        = Literal
               | Identifier
               | JSONPath
               | FunctionCall
               | ListLiteral
               | LambdaExpression
               | "(" Expression ")" ;

Literal        = Integer | Float | String | Boolean | Null ;
JSONPath       = "$" { "." Identifier | "[" (Integer | String) "]" } ;
FunctionCall   = Identifier "(" [ ArgList ] ")" ;
ArgList        = Expression { "," Expression } ;
ListLiteral    = "[" [ Expression { "," Expression } ] "]" ;
LambdaExpression = Identifier "=>" Expression
                 | "(" Identifier "," Identifier ")" "=>" Expression ;
```

## Examples

### Access Control

```
$.user.role IN ["admin", "moderator"] ||
($.user.verified == true && $.user.reputation >= 1000)
```

### E-commerce Rules

```
$.order.total >= 100 && $.customer.tier == "gold" ||
$.order.total >= 500
```

### Data Validation

```
isNotNull($.email) &&
$.email =~ "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$" &&
len($.password) >= 8
```

### Feature Flags

```
$.feature.enabled == true &&
$.user.id % 100 < $.feature.rolloutPercent
```

### Complex Calculations

```
sum(map($.items, x => x.price * x.quantity)) * 
(1 - coalesce($.discount, 0) / 100)
```
