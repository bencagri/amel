# SQL Compilation

AMEL expressions can be compiled to SQL WHERE clauses, enabling you to use the same expressions for both in-memory evaluation and database queries.

## Table of Contents

- [Overview](#overview)
- [Basic Usage](#basic-usage)
- [SQL Dialects](#sql-dialects)
- [Supported Operations](#supported-operations)
- [Function Mapping](#function-mapping)
- [Parameter Handling](#parameter-handling)
- [Field Mapping](#field-mapping)
- [Examples](#examples)

---

## Overview

The SQL compiler transforms AMEL expressions into SQL WHERE clauses with parameterized queries, protecting against SQL injection while maintaining the expressive power of AMEL.

```go
import (
    "github.com/bencagri/amel/pkg/compiler"
    "github.com/bencagri/amel/pkg/parser"
)

// Parse the AMEL expression
expr, _ := parser.Parse(`$.age > 18 && $.status == "active"`)

// Compile to SQL
sqlCompiler := compiler.NewSQLCompiler()
result, _ := sqlCompiler.Compile(expr)

fmt.Println(result.SQL)
// Output: (("age" > ?) AND ("status" = ?))

fmt.Println(result.Params)
// Output: [18, "active"]
```

---

## Basic Usage

### Simple Compilation

```go
import (
    "github.com/bencagri/amel/pkg/compiler"
    "github.com/bencagri/amel/pkg/parser"
)

// Parse expression
expr, err := parser.Parse(`$.price > 100`)
if err != nil {
    log.Fatal(err)
}

// Create compiler and compile
sqlCompiler := compiler.NewSQLCompiler()
result, err := sqlCompiler.Compile(expr)
if err != nil {
    log.Fatal(err)
}

// Use the result
fmt.Println(result.SQL)    // ("price" > ?)
fmt.Println(result.Params) // [100]
```

### Using with Database

```go
// Compile the expression
expr, _ := parser.Parse(`$.status == "active" && $.age >= 18`)
result, _ := compiler.NewSQLCompiler().Compile(expr)

// Build full query
query := fmt.Sprintf("SELECT * FROM users WHERE %s", result.SQL)

// Execute with parameters
rows, err := db.Query(query, result.Params...)
```

### Convenience Function

```go
// Quick compilation
expr, _ := parser.Parse(`$.name == "John"`)
result, _ := compiler.CompileToSQL(expr)
```

---

## SQL Dialects

AMEL supports multiple SQL dialects with appropriate syntax variations.

### Standard SQL (Default)

```go
sqlCompiler := compiler.NewSQLCompiler()
// or
sqlCompiler := compiler.NewSQLCompiler(compiler.WithDialect(compiler.DialectStandard))
```

**Features:**
- Double-quoted identifiers: `"column_name"`
- `?` placeholders for parameters
- Standard SQL functions

### PostgreSQL

```go
sqlCompiler := compiler.NewSQLCompiler(compiler.WithDialect(compiler.DialectPostgres))
```

**Features:**
- Double-quoted identifiers: `"column_name"`
- Numbered placeholders: `$1`, `$2`, etc.
- PostgreSQL-specific functions
- `~` for regex (native support)

**Example:**

```go
expr, _ := parser.Parse(`$.age > 18 && $.status == "active"`)
result, _ := compiler.NewSQLCompiler(
    compiler.WithDialect(compiler.DialectPostgres),
).Compile(expr)

fmt.Println(result.SQL)
// Output: (("age" > $1) AND ("status" = $2))
```

### MySQL

```go
sqlCompiler := compiler.NewSQLCompiler(compiler.WithDialect(compiler.DialectMySQL))
```

**Features:**
- Backtick identifiers: `` `column_name` ``
- `?` placeholders
- `REGEXP` for regex
- `CHAR_LENGTH` instead of `LENGTH`

**Example:**

```go
expr, _ := parser.Parse(`$.user_name == "john"`)
result, _ := compiler.NewSQLCompiler(
    compiler.WithDialect(compiler.DialectMySQL),
).Compile(expr)

fmt.Println(result.SQL)
// Output: (`user_name` = ?)
```

### SQLite

```go
sqlCompiler := compiler.NewSQLCompiler(compiler.WithDialect(compiler.DialectSQLite))
```

**Features:**
- Double-quoted identifiers
- `?` placeholders
- SQLite-compatible functions

---

## Supported Operations

### Comparison Operators

| AMEL | SQL |
|------|-----|
| `==` | `=` |
| `!=` | `<>` |
| `>` | `>` |
| `<` | `<` |
| `>=` | `>=` |
| `<=` | `<=` |

**Examples:**

```
$.age > 18           → ("age" > ?)
$.status == "active" → ("status" = ?)
$.price != 0         → ("price" <> ?)
```

### Logical Operators

| AMEL | SQL |
|------|-----|
| `&&` | `AND` |
| `\|\|` | `OR` |
| `!` | `NOT` |

**Examples:**

```
$.a && $.b           → (("a") AND ("b"))
$.a || $.b           → (("a") OR ("b"))
!$.active            → NOT ("active")
!($.a && $.b)        → NOT (("a") AND ("b"))
```

### Null Comparisons

| AMEL | SQL |
|------|-----|
| `== null` | `IS NULL` |
| `!= null` | `IS NOT NULL` |

**Examples:**

```
$.deleted_at == null → "deleted_at" IS NULL
$.email != null      → "email" IS NOT NULL
```

### IN Operator

| AMEL | SQL |
|------|-----|
| `IN [...]` | `IN (?, ?, ...)` |
| `NOT IN [...]` | `NOT IN (?, ?, ...)` |

**Examples:**

```
$.status IN ["active", "pending"]
→ "status" IN (?, ?)
→ Params: ["active", "pending"]

$.role NOT IN ["banned", "suspended"]
→ "role" NOT IN (?, ?)
→ Params: ["banned", "suspended"]
```

### Regex Operators

| AMEL | PostgreSQL | MySQL | Standard |
|------|------------|-------|----------|
| `=~` | `~` | `REGEXP` | Error |
| `!~` | `!~` | `NOT REGEXP` | Error |

**Note:** Regex is only supported in PostgreSQL and MySQL dialects.

**Examples:**

```go
// PostgreSQL
expr, _ := parser.Parse(`$.email =~ "@gmail\\.com$"`)
result, _ := compiler.NewSQLCompiler(
    compiler.WithDialect(compiler.DialectPostgres),
).Compile(expr)
// Output: "email" ~ $1

// MySQL
result, _ := compiler.NewSQLCompiler(
    compiler.WithDialect(compiler.DialectMySQL),
).Compile(expr)
// Output: `email` REGEXP ?
```

---

## Function Mapping

AMEL functions are mapped to SQL equivalents where possible.

### String Functions

| AMEL Function | SQL Function |
|---------------|--------------|
| `lower(x)` | `LOWER(x)` |
| `upper(x)` | `UPPER(x)` |
| `trim(x)` | `TRIM(x)` |
| `len(x)` | `LENGTH(x)` (MySQL: `CHAR_LENGTH(x)`) |
| `contains(x, y)` | `x LIKE '%y%'` |
| `startsWith(x, y)` | `x LIKE 'y%'` |
| `endsWith(x, y)` | `x LIKE '%y'` |

**Examples:**

```
lower($.name) == "john"
→ (LOWER("name") = ?)

len($.name) > 5
→ (LENGTH("name") > ?)

contains($.description, "important")
→ "description" LIKE ?
→ Params: ["%important%"]

startsWith($.email, "admin")
→ "email" LIKE ?
→ Params: ["admin%"]
```

### Math Functions

| AMEL Function | SQL Function |
|---------------|--------------|
| `abs(x)` | `ABS(x)` |
| `ceil(x)` | `CEIL(x)` |
| `floor(x)` | `FLOOR(x)` |
| `round(x)` | `ROUND(x)` |

### Null Functions

| AMEL Function | SQL Function |
|---------------|--------------|
| `isNull(x)` | `x IS NULL` |
| `isNotNull(x)` | `x IS NOT NULL` |
| `coalesce(x, y, ...)` | `COALESCE(x, y, ...)` |

**Examples:**

```
isNull($.deleted_at)
→ ("deleted_at" IS NULL)

isNotNull($.email)
→ ("email" IS NOT NULL)

coalesce($.nickname, $.name) == "John"
→ (COALESCE("nickname", "name") = ?)
```

---

## Parameter Handling

### Parameterized Queries (Default)

By default, values are replaced with placeholders:

```go
expr, _ := parser.Parse(`$.age > 18 && $.name == "John"`)
result, _ := compiler.NewSQLCompiler().Compile(expr)

fmt.Println(result.SQL)    // (("age" > ?) AND ("name" = ?))
fmt.Println(result.Params) // [18, "John"]
```

### Inline Parameters

For debugging or logging, you can inline the parameters:

```go
sqlCompiler := compiler.NewSQLCompiler(compiler.WithInlineParams(true))
result, _ := sqlCompiler.Compile(expr)

fmt.Println(result.SQL)
// (("age" > 18) AND ("name" = 'John'))
```

**Warning:** Never use inline parameters with user input in production!

### Parameter Value Types

Parameters maintain their Go types:

```go
result, _ := compiler.NewSQLCompiler().Compile(expr)

for i, param := range result.Params {
    fmt.Printf("Param %d: %v (type: %T)\n", i, param, param)
}
// Param 0: 18 (type: int64)
// Param 1: John (type: string)
// Param 2: true (type: bool)
```

---

## Field Mapping

### Default Field Mapping

JSONPath expressions are converted to column names:

```
$.name           → "name"
$.user.email     → "user.email"  (nested path)
$.items[0].price → "items.0.price"
```

### Custom Field Mapper

Map AMEL paths to database column names:

```go
mapper := func(path string) string {
    mappings := map[string]string{
        "$.user.firstName": "users.first_name",
        "$.user.lastName":  "users.last_name",
        "$.user.email":     "users.email_address",
    }
    if mapped, ok := mappings[path]; ok {
        return mapped
    }
    // Default: remove $. prefix
    return strings.TrimPrefix(path, "$.")
}

sqlCompiler := compiler.NewSQLCompiler(compiler.WithFieldMapper(mapper))
```

**Example:**

```go
expr, _ := parser.Parse(`$.user.firstName == "John"`)
result, _ := sqlCompiler.Compile(expr)
// Output: ("users.first_name" = ?)
```

### Table Prefixes

Add table prefixes to all fields:

```go
mapper := func(path string) string {
    field := strings.TrimPrefix(path, "$.")
    return "users." + field
}

sqlCompiler := compiler.NewSQLCompiler(compiler.WithFieldMapper(mapper))
```

---

## Examples

### User Authentication Query

```go
expr, _ := parser.Parse(`
    $.email == "user@example.com" && 
    $.active == true && 
    $.deleted_at == null
`)

result, _ := compiler.NewSQLCompiler(
    compiler.WithDialect(compiler.DialectPostgres),
).Compile(expr)

query := fmt.Sprintf("SELECT * FROM users WHERE %s", result.SQL)
// SELECT * FROM users WHERE 
//   (("email" = $1) AND ("active" = $2) AND "deleted_at" IS NULL)
```

### Age Range Query

```go
expr, _ := parser.Parse(`$.age >= 18 && $.age <= 65`)
result, _ := compiler.NewSQLCompiler().Compile(expr)

// SQL: (("age" >= ?) AND ("age" <= ?))
// Params: [18, 65]
```

### Status Filter with IN

```go
expr, _ := parser.Parse(`$.status IN ["pending", "approved", "processing"]`)
result, _ := compiler.NewSQLCompiler().Compile(expr)

// SQL: "status" IN (?, ?, ?)
// Params: ["pending", "approved", "processing"]
```

### Complex Business Rule

```go
expr, _ := parser.Parse(`
    ($.role == "admin" || $.role == "manager") &&
    $.department IN ["sales", "marketing"] &&
    $.active == true
`)

result, _ := compiler.NewSQLCompiler(
    compiler.WithDialect(compiler.DialectPostgres),
).Compile(expr)

// SQL: ((("role" = $1) OR ("role" = $2)) AND 
//       "department" IN ($3, $4) AND ("active" = $5))
```

### With String Functions

```go
expr, _ := parser.Parse(`
    lower($.email) == lower("User@Example.COM") &&
    len($.password) >= 8
`)

result, _ := compiler.NewSQLCompiler().Compile(expr)

// SQL: ((LOWER("email") = LOWER(?)) AND (LENGTH("password") >= ?))
```

### Full Application Example

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    "github.com/bencagri/amel/pkg/compiler"
    "github.com/bencagri/amel/pkg/parser"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "...")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Define the filter expression
    filterExpr := `
        $.status == "active" &&
        $.age >= 18 &&
        $.role IN ["user", "admin"]
    `
    
    // Parse and compile
    expr, err := parser.Parse(filterExpr)
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }
    
    sqlCompiler := compiler.NewSQLCompiler(
        compiler.WithDialect(compiler.DialectPostgres),
    )
    result, err := sqlCompiler.Compile(expr)
    if err != nil {
        log.Fatalf("Compile error: %v", err)
    }
    
    // Build and execute query
    query := fmt.Sprintf(`
        SELECT id, name, email 
        FROM users 
        WHERE %s 
        ORDER BY created_at DESC
    `, result.SQL)
    
    rows, err := db.Query(query, result.Params...)
    if err != nil {
        log.Fatalf("Query error: %v", err)
    }
    defer rows.Close()
    
    // Process results
    for rows.Next() {
        var id int
        var name, email string
        rows.Scan(&id, &name, &email)
        fmt.Printf("User: %d - %s (%s)\n", id, name, email)
    }
}
```

---

## Limitations

### Unsupported Features

The following AMEL features cannot be compiled to SQL:

- Custom JavaScript functions
- Lambda expressions (`map`, `filter`, `reduce`)
- Some built-in functions without SQL equivalents
- Complex arithmetic in some contexts

### Error Handling

```go
result, err := sqlCompiler.Compile(expr)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "unsupported"):
        log.Printf("Feature not supported in SQL: %v", err)
    default:
        log.Printf("Compilation error: %v", err)
    }
}
```

---

## See Also

- [MongoDB Compilation](./07-mongodb-compilation.md) - Compiling to MongoDB queries
- [Expression Syntax](./02-syntax.md) - AMEL syntax reference
- [Built-in Functions](./03-functions.md) - Function reference
