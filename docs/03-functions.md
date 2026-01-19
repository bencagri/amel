# Built-in Functions

AMEL provides over 60 built-in functions for common operations. This document describes all available functions organized by category.

## Table of Contents

- [String Functions](#string-functions)
- [Math Functions](#math-functions)
- [List/Array Functions](#listarray-functions)
- [Type Conversion Functions](#type-conversion-functions)
- [Null Handling Functions](#null-handling-functions)
- [Conditional Functions](#conditional-functions)
- [Aggregate Functions](#aggregate-functions)
- [Array Operation Functions](#array-operation-functions)

---

## String Functions

### len

Returns the length of a string or list.

```
len(value) -> int
```

**Examples:**

```
len("hello")           // 5
len("")                // 0
len([1, 2, 3])         // 3
len($.user.name)       // length of name
```

---

### lower

Converts a string to lowercase.

```
lower(string) -> string
```

**Examples:**

```
lower("HELLO")         // "hello"
lower("John Doe")      // "john doe"
lower($.user.name)     // lowercase name
```

---

### upper

Converts a string to uppercase.

```
upper(string) -> string
```

**Examples:**

```
upper("hello")         // "HELLO"
upper("John Doe")      // "JOHN DOE"
upper($.status)        // uppercase status
```

---

### trim

Removes leading and trailing whitespace.

```
trim(string) -> string
```

**Examples:**

```
trim("  hello  ")      // "hello"
trim("\t\ntext\n")     // "text"
trim($.input)          // trimmed input
```

---

### trimLeft

Removes leading whitespace only.

```
trimLeft(string) -> string
```

**Examples:**

```
trimLeft("  hello  ")  // "hello  "
trimLeft("\t\ntext")   // "text"
```

---

### trimRight

Removes trailing whitespace only.

```
trimRight(string) -> string
```

**Examples:**

```
trimRight("  hello  ") // "  hello"
trimRight("text\n\t")  // "text"
```

---

### contains

Checks if a string contains a substring.

```
contains(string, substring) -> bool
```

**Examples:**

```
contains("hello world", "world")     // true
contains("hello", "xyz")             // false
contains($.description, "important") // check for keyword
```

---

### startsWith

Checks if a string starts with a prefix.

```
startsWith(string, prefix) -> bool
```

**Examples:**

```
startsWith("hello", "hel")           // true
startsWith("hello", "world")         // false
startsWith($.url, "https://")        // check protocol
```

---

### endsWith

Checks if a string ends with a suffix.

```
endsWith(string, suffix) -> bool
```

**Examples:**

```
endsWith("hello", "lo")              // true
endsWith("document.pdf", ".pdf")     // true
endsWith($.filename, ".jpg")         // check extension
```

---

### substr

Extracts a substring from a string.

```
substr(string, start, length) -> string
```

**Parameters:**

- `string`: The source string
- `start`: Starting index (0-based)
- `length`: Number of characters to extract

**Examples:**

```
substr("hello world", 0, 5)          // "hello"
substr("hello world", 6, 5)          // "world"
substr($.name, 0, 1)                 // first character
```

---

### replace

Replaces occurrences of a substring.

```
replace(string, old, new) -> string
```

**Examples:**

```
replace("hello world", "world", "there")  // "hello there"
replace("aaa", "a", "b")                  // "bbb"
replace($.text, "\n", " ")                // replace newlines
```

---

### split

Splits a string into a list by delimiter.

```
split(string, delimiter) -> list
```

**Examples:**

```
split("a,b,c", ",")                  // ["a", "b", "c"]
split("hello world", " ")            // ["hello", "world"]
split($.tags, ";")                   // split tags
```

---

### join

Joins a list of strings with a delimiter.

```
join(list, delimiter) -> string
```

**Examples:**

```
join(["a", "b", "c"], ",")           // "a,b,c"
join(["hello", "world"], " ")        // "hello world"
join($.words, "-")                   // join with hyphens
```

---

### concat

Concatenates multiple strings.

```
concat(string1, string2, ...) -> string
```

**Examples:**

```
concat("hello", " ", "world")        // "hello world"
concat($.firstName, " ", $.lastName) // full name
```

---

### padLeft

Pads a string on the left to reach a target length.

```
padLeft(string, length, padChar) -> string
```

**Examples:**

```
padLeft("42", 5, "0")                // "00042"
padLeft("x", 3, "-")                 // "--x"
padLeft(string($.id), 6, "0")        // zero-padded ID
```

---

### padRight

Pads a string on the right to reach a target length.

```
padRight(string, length, padChar) -> string
```

**Examples:**

```
padRight("42", 5, "0")               // "42000"
padRight("x", 3, "-")                // "x--"
```

---

### repeat

Repeats a string a specified number of times.

```
repeat(string, count) -> string
```

**Examples:**

```
repeat("ab", 3)                      // "ababab"
repeat("-", 10)                      // "----------"
repeat("*", $.level)                 // stars based on level
```

---

### match

Tests if a string matches a regular expression pattern.

```
match(string, pattern) -> bool
```

**Examples:**

```
match("test@example.com", "^[\\w.]+@[\\w.]+$")  // true
match("12345", "^\\d+$")                         // true
match($.email, "^[\\w.]+@company\\.com$")        // domain check
```

---

### format

Formats a string with placeholders (sprintf-style).

```
format(template, arg1, arg2, ...) -> string
```

**Examples:**

```
format("Hello, %s!", "World")        // "Hello, World!"
format("Value: %d", 42)              // "Value: 42"
format("%s has %d items", $.name, $.count)
```

---

## Math Functions

### abs

Returns the absolute value of a number.

```
abs(number) -> number
```

**Examples:**

```
abs(-5)                              // 5
abs(5)                               // 5
abs($.balance)                       // absolute balance
```

---

### ceil

Rounds a number up to the nearest integer.

```
ceil(float) -> int
```

**Examples:**

```
ceil(4.1)                            // 5
ceil(4.9)                            // 5
ceil(-4.1)                           // -4
ceil($.score)                        // round up score
```

---

### floor

Rounds a number down to the nearest integer.

```
floor(float) -> int
```

**Examples:**

```
floor(4.1)                           // 4
floor(4.9)                           // 4
floor(-4.1)                          // -5
floor($.price)                       // round down price
```

---

### round

Rounds a number to the nearest integer.

```
round(float) -> int
```

**Examples:**

```
round(4.4)                           // 4
round(4.5)                           // 5
round(4.6)                           // 5
round($.average)                     // round average
```

---

### pow

Raises a number to a power.

```
pow(base, exponent) -> number
```

**Examples:**

```
pow(2, 3)                            // 8
pow(10, 2)                           // 100
pow($.base, $.exp)                   // dynamic power
```

---

### sqrt

Returns the square root of a number.

```
sqrt(number) -> float
```

**Examples:**

```
sqrt(16)                             // 4
sqrt(2)                              // 1.414...
sqrt($.area)                         // side from area
```

**Note:** Returns an error for negative numbers.

---

### mod

Returns the remainder of division (modulo operation).

```
mod(a, b) -> int
```

**Examples:**

```
mod(10, 3)                           // 1
mod(15, 5)                           // 0
mod($.id, 100)                       // last two digits
```

**Note:** Returns an error if `b` is zero.

---

### clamp

Constrains a value within a range.

```
clamp(value, min, max) -> number
```

**Examples:**

```
clamp(5, 0, 10)                      // 5
clamp(-5, 0, 10)                     // 0
clamp(15, 0, 10)                     // 10
clamp($.score, 0, 100)               // constrain score
```

---

### between

Checks if a value is within a range (inclusive).

```
between(value, min, max) -> bool
```

**Examples:**

```
between(5, 0, 10)                    // true
between(15, 0, 10)                   // false
between($.age, 18, 65)               // age check
```

---

## List/Array Functions

### first

Returns the first element of a list.

```
first(list) -> any
```

**Examples:**

```
first([1, 2, 3])                     // 1
first(["a", "b", "c"])               // "a"
first($.items)                       // first item
```

**Note:** Returns `null` for empty lists.

---

### last

Returns the last element of a list.

```
last(list) -> any
```

**Examples:**

```
last([1, 2, 3])                      // 3
last(["a", "b", "c"])                // "c"
last($.items)                        // last item
```

**Note:** Returns `null` for empty lists.

---

### at

Returns the element at a specific index.

```
at(list, index) -> any
```

**Examples:**

```
at([10, 20, 30], 0)                  // 10
at([10, 20, 30], 1)                  // 20
at($.items, 2)                       // third item
```

**Note:** Returns an error for out-of-bounds index.

---

### indexOf

Returns the index of the first occurrence of a value.

```
indexOf(list, value) -> int
```

**Examples:**

```
indexOf([1, 2, 3], 2)                // 1
indexOf(["a", "b", "c"], "b")        // 1
indexOf([1, 2, 3], 5)                // -1 (not found)
```

---

### reverse

Reverses a list.

```
reverse(list) -> list
```

**Examples:**

```
reverse([1, 2, 3])                   // [3, 2, 1]
reverse(["a", "b", "c"])             // ["c", "b", "a"]
reverse($.items)                     // reversed items
```

---

### unique

Removes duplicate values from a list.

```
unique(list) -> list
```

**Examples:**

```
unique([1, 2, 2, 3, 3, 3])           // [1, 2, 3]
unique(["a", "b", "a"])              // ["a", "b"]
unique($.tags)                       // deduplicated tags
```

---

### flatten

Flattens a nested list by one level.

```
flatten(list) -> list
```

**Examples:**

```
flatten([[1, 2], [3, 4]])            // [1, 2, 3, 4]
flatten([[1], [2, 3], [4]])          // [1, 2, 3, 4]
flatten($.nestedItems)               // flattened items
```

---

### slice

Extracts a portion of a list.

```
slice(list, start, end) -> list
```

**Parameters:**

- `list`: The source list
- `start`: Starting index (inclusive, 0-based)
- `end`: Ending index (exclusive)

**Examples:**

```
slice([1, 2, 3, 4, 5], 1, 4)         // [2, 3, 4]
slice([1, 2, 3], 0, 2)               // [1, 2]
slice($.items, 0, 10)                // first 10 items
```

---

### sortAsc

Sorts a list in ascending order.

```
sortAsc(list) -> list
```

**Examples:**

```
sortAsc([3, 1, 4, 1, 5])             // [1, 1, 3, 4, 5]
sortAsc(["c", "a", "b"])             // ["a", "b", "c"]
sortAsc($.scores)                    // sorted scores
```

---

### sortDesc

Sorts a list in descending order.

```
sortDesc(list) -> list
```

**Examples:**

```
sortDesc([3, 1, 4, 1, 5])            // [5, 4, 3, 1, 1]
sortDesc(["c", "a", "b"])            // ["c", "b", "a"]
sortDesc($.scores)                   // sorted descending
```

---

## Type Conversion Functions

### int

Converts a value to an integer.

```
int(value) -> int
```

**Examples:**

```
int("42")                            // 42
int(3.7)                             // 3
int(true)                            // 1
int(false)                           // 0
int($.stringNumber)                  // parse string
```

---

### float

Converts a value to a float.

```
float(value) -> float
```

**Examples:**

```
float("3.14")                        // 3.14
float(42)                            // 42.0
float($.value)                       // convert to float
```

---

### string

Converts a value to a string.

```
string(value) -> string
```

**Examples:**

```
string(42)                           // "42"
string(3.14)                         // "3.14"
string(true)                         // "true"
string(null)                         // "null"
string($.count)                      // number to string
```

---

### bool

Converts a value to a boolean.

```
bool(value) -> bool
```

**Examples:**

```
bool(1)                              // true
bool(0)                              // false
bool("true")                         // true
bool("")                             // false
bool($.flag)                         // convert to bool
```

---

### typeOf

Returns the type of a value as a string.

```
typeOf(value) -> string
```

**Examples:**

```
typeOf(42)                           // "int"
typeOf(3.14)                         // "float"
typeOf("hello")                      // "string"
typeOf(true)                         // "bool"
typeOf(null)                         // "null"
typeOf([1, 2, 3])                    // "list"
typeOf($.value)                      // dynamic type check
```

---

## Null Handling Functions

### isNull

Checks if a value is null.

```
isNull(value) -> bool
```

**Examples:**

```
isNull(null)                         // true
isNull("")                           // false
isNull(0)                            // false
isNull($.optional)                   // check if missing
```

---

### isNotNull

Checks if a value is not null.

```
isNotNull(value) -> bool
```

**Examples:**

```
isNotNull("hello")                   // true
isNotNull(null)                      // false
isNotNull($.required)                // check if present
```

---

### isEmpty

Checks if a value is empty (null, empty string, or empty list).

```
isEmpty(value) -> bool
```

**Examples:**

```
isEmpty(null)                        // true
isEmpty("")                          // true
isEmpty([])                          // true
isEmpty("hello")                     // false
isEmpty([1, 2])                      // false
isEmpty($.field)                     // check if empty
```

---

### coalesce

Returns the first non-null value from arguments.

```
coalesce(value1, value2, ...) -> any
```

**Examples:**

```
coalesce(null, "default")            // "default"
coalesce(null, null, "fallback")     // "fallback"
coalesce("value", "default")         // "value"
coalesce($.nickname, $.name, "Anonymous")
```

---

### defaultVal

Returns a default value if the input is null.

```
defaultVal(value, default) -> any
```

**Examples:**

```
defaultVal(null, 0)                  // 0
defaultVal(5, 0)                     // 5
defaultVal($.count, 0)               // default to 0
```

---

### exists

Checks if a value exists (is not null and not undefined).

```
exists(value) -> bool
```

**Examples:**

```
exists($.user.email)                 // true if email exists
exists($.optional)                   // check presence
```

---

## Conditional Functions

### ifThenElse

Returns one of two values based on a condition.

```
ifThenElse(condition, thenValue, elseValue) -> any
```

**Examples:**

```
ifThenElse(true, "yes", "no")        // "yes"
ifThenElse(false, "yes", "no")       // "no"
ifThenElse($.age >= 18, "adult", "minor")
ifThenElse($.score > 90, "A", ifThenElse($.score > 80, "B", "C"))
```

---

## Aggregate Functions

### count

Returns the number of elements in a list.

```
count(list) -> int
```

**Examples:**

```
count([1, 2, 3])                     // 3
count([])                            // 0
count($.items)                       // number of items
```

---

### sum

Returns the sum of all numbers in a list.

```
sum(list) -> number
```

**Examples:**

```
sum([1, 2, 3, 4])                    // 10
sum([1.5, 2.5, 3.0])                 // 7.0
sum($.prices)                        // total price
```

---

### avg

Returns the average of all numbers in a list.

```
avg(list) -> float
```

**Examples:**

```
avg([10, 20, 30])                    // 20.0
avg([1, 2, 3, 4])                    // 2.5
avg($.scores)                        // average score
```

**Note:** Returns `null` for empty lists.

---

### min

Returns the minimum value from a list or arguments.

```
min(list) -> any
min(a, b, ...) -> any
```

**Examples:**

```
min([5, 2, 8, 1])                    // 1
min(5, 2, 8, 1)                      // 1
min($.values)                        // minimum value
min($.a, $.b, $.c)                   // min of multiple
```

---

### max

Returns the maximum value from a list or arguments.

```
max(list) -> any
max(a, b, ...) -> any
```

**Examples:**

```
max([5, 2, 8, 1])                    // 8
max(5, 2, 8, 1)                      // 8
max($.values)                        // maximum value
max($.a, $.b, $.c)                   // max of multiple
```

---

### all

Checks if all boolean values in a list are true.

```
all(list) -> bool
```

**Examples:**

```
all([true, true, true])              // true
all([true, false, true])             // false
all($.checks)                        // all passed
```

---

### any

Checks if any boolean value in a list is true.

```
any(list) -> bool
```

**Examples:**

```
any([false, false, true])            // true
any([false, false, false])           // false
any($.warnings)                      // any warnings
```

---

## Array Operation Functions

These functions use lambda expressions to process arrays.

### map

Transforms each element in a list using a lambda.

```
map(list, lambda) -> list
```

**Examples:**

```
map([1, 2, 3], x => x * 2)           // [2, 4, 6]
map([1, 2, 3], x => x + 1)           // [2, 3, 4]
map(["a", "b"], x => upper(x))       // ["A", "B"]
map($.items, x => x.price)           // extract prices
```

---

### filter

Selects elements that satisfy a predicate.

```
filter(list, lambda) -> list
```

**Examples:**

```
filter([1, 2, 3, 4, 5], x => x > 2)  // [3, 4, 5]
filter([1, 2, 3, 4], x => x % 2 == 0) // [2, 4]
filter($.users, x => x.active)       // active users
```

---

### reduce

Aggregates a list to a single value.

```
reduce(list, initialValue, lambda) -> any
```

**Parameters:**

- `list`: The list to reduce
- `initialValue`: Starting accumulator value
- `lambda`: `(accumulator, element) => newAccumulator`

**Examples:**

```
reduce([1, 2, 3, 4], 0, (acc, x) => acc + x)  // 10 (sum)
reduce([1, 2, 3, 4], 1, (acc, x) => acc * x)  // 24 (product)
reduce($.items, 0, (sum, x) => sum + x.price) // total price
```

---

### find

Returns the first element that satisfies a predicate.

```
find(list, lambda) -> any
```

**Examples:**

```
find([1, 2, 3, 4], x => x > 2)       // 3
find($.users, x => x.name == "John") // find John
find([1, 2, 3], x => x > 10)         // null (not found)
```

**Note:** Returns `null` if no element matches.

---

### some

Checks if any element satisfies a predicate.

```
some(list, lambda) -> bool
```

**Examples:**

```
some([1, 2, 3], x => x > 2)          // true
some([1, 2, 3], x => x > 10)         // false
some($.users, x => x.role == "admin") // any admin?
```

---

### every

Checks if all elements satisfy a predicate.

```
every(list, lambda) -> bool
```

**Examples:**

```
every([1, 2, 3], x => x > 0)         // true
every([1, 2, 3], x => x > 1)         // false
every($.items, x => x.inStock)       // all in stock?
```

---

## Function Overloading

Some functions support multiple signatures (overloading). The appropriate version is selected based on argument types:

```
// min with list
min([1, 2, 3])                       // 1

// min with varargs
min(1, 2, 3)                         // 1

// Both work correctly
```

Functions with overloading:

- `min` - list or varargs
- `max` - list or varargs
- `sum` - list or single value
- `count` - list or single value
- `len` - string or list

---

## Error Handling

Functions may return errors for invalid inputs:

| Function | Error Condition |
|----------|-----------------|
| `sqrt` | Negative number |
| `mod` | Division by zero |
| `at` | Index out of bounds |
| `int` | Invalid string format |
| `float` | Invalid string format |
| `match` | Invalid regex pattern |

**Example:**

```go
result, err := eng.EvaluateDirect(`sqrt(-1)`, nil)
if err != nil {
    // Handle error: cannot compute square root of negative number
}
```

---

## See Also

- [Expression Syntax](./02-syntax.md) - Language syntax reference
- [Custom Functions](./03-custom-functions.md) - Creating JavaScript functions
- [Array Operations](./04-array-operations.md) - Detailed array operation guide
