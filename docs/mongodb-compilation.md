# MongoDB Compilation

AMEL expressions can be compiled to MongoDB query documents, allowing you to use the same expressions for both in-memory evaluation and MongoDB database queries.

## Table of Contents

- [Overview](#overview)
- [Basic Usage](#basic-usage)
- [Supported Operations](#supported-operations)
- [Function Mapping](#function-mapping)
- [Field Mapping](#field-mapping)
- [Output Formats](#output-formats)
- [Examples](#examples)
- [Limitations](#limitations)

---

## Overview

The MongoDB compiler transforms AMEL expressions into MongoDB query documents that can be used directly with the MongoDB driver.

```go
import (
    "github.com/bencagri/amel/pkg/compiler"
    "github.com/bencagri/amel/pkg/parser"
)

// Parse the AMEL expression
expr, _ := parser.Parse(`$.age > 18 && $.status == "active"`)

// Compile to MongoDB
mongoCompiler := compiler.NewMongoDBCompiler()
result, _ := mongoCompiler.Compile(expr)

fmt.Println(result.Query)
// Output: {"$and": [{"age": {"$gt": 18}}, {"status": "active"}]}
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
mongoCompiler := compiler.NewMongoDBCompiler()
result, err := mongoCompiler.Compile(expr)
if err != nil {
    log.Fatal(err)
}

// Use the result
fmt.Printf("%+v\n", result.Query)
// Output: map[price:map[$gt:100]]
```

### Using with MongoDB Driver

```go
import (
    "go.mongodb.org/mongo-driver/mongo"
)

// Compile the expression
expr, _ := parser.Parse(`$.status == "active" && $.age >= 18`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Use directly as a filter
cursor, err := collection.Find(ctx, result.Query)
```

### Convenience Function

```go
// Quick compilation
expr, _ := parser.Parse(`$.name == "John"`)
result, _ := compiler.CompileToMongoDB(expr)
```

---

## Supported Operations

### Comparison Operators

| AMEL | MongoDB |
|------|---------|
| `==` | Direct value or `$eq` |
| `!=` | `$ne` |
| `>` | `$gt` |
| `<` | `$lt` |
| `>=` | `$gte` |
| `<=` | `$lte` |

**Examples:**

```
$.age > 18
→ {"age": {"$gt": 18}}

$.status == "active"
→ {"status": "active"}

$.price != 0
→ {"price": {"$ne": 0}}

$.quantity >= 5
→ {"quantity": {"$gte": 5}}
```

### Logical Operators

| AMEL | MongoDB |
|------|---------|
| `&&` | `$and` |
| `\|\|` | `$or` |
| `!` | `$nor` or negated operator |

**Examples:**

```
$.age > 18 && $.status == "active"
→ {"$and": [{"age": {"$gt": 18}}, {"status": "active"}]}

$.role == "admin" || $.role == "superuser"
→ {"$or": [{"role": "admin"}, {"role": "superuser"}]}

!($.status == "active")
→ {"status": {"$ne": "active"}}

!($.age > 18 && $.status == "active")
→ {"$nor": [{"$and": [{"age": {"$gt": 18}}, {"status": "active"}]}]}
```

### Null Comparisons

| AMEL | MongoDB |
|------|---------|
| `== null` | `null` |
| `!= null` | `{"$ne": null}` |

**Examples:**

```
$.deleted_at == null
→ {"deleted_at": null}

$.email != null
→ {"email": {"$ne": null}}
```

### Boolean Comparisons

```
$.active == true
→ {"active": true}

$.archived == false
→ {"archived": false}
```

### IN Operator

| AMEL | MongoDB |
|------|---------|
| `IN [...]` | `$in` |
| `NOT IN [...]` | `$nin` |

**Examples:**

```
$.status IN ["active", "pending"]
→ {"status": {"$in": ["active", "pending"]}}

$.role NOT IN ["banned", "suspended"]
→ {"role": {"$nin": ["banned", "suspended"]}}

$.priority IN [1, 2, 3]
→ {"priority": {"$in": [1, 2, 3]}}
```

### Regex Operators

| AMEL | MongoDB |
|------|---------|
| `=~` | `$regex` |
| `!~` | `$not` with `$regex` |

**Examples:**

```
$.email =~ "@gmail.com$"
→ {"email": {"$regex": "@gmail.com$"}}

$.email !~ "@spam.com$"
→ {"email": {"$not": {"$regex": "@spam.com$"}}}
```

---

## Function Mapping

AMEL functions are mapped to MongoDB query operators where possible.

### Null Functions

| AMEL Function | MongoDB |
|---------------|---------|
| `isNull(x)` | `{x: null}` |
| `isNotNull(x)` | `{x: {"$ne": null}}` |

**Examples:**

```
isNull($.deleted_at)
→ {"deleted_at": null}

isNotNull($.email)
→ {"email": {"$ne": null}}
```

### String Functions

| AMEL Function | MongoDB |
|---------------|---------|
| `contains(x, y)` | `{x: {"$regex": "y"}}` |
| `startsWith(x, y)` | `{x: {"$regex": "^y"}}` |
| `endsWith(x, y)` | `{x: {"$regex": "y$"}}` |

**Examples:**

```
contains($.name, "john")
→ {"name": {"$regex": "john"}}

startsWith($.email, "admin")
→ {"email": {"$regex": "^admin"}}

endsWith($.email, ".com")
→ {"email": {"$regex": "\\.com$"}}
```

**Note:** Special regex characters in the pattern are escaped automatically.

### Existence Check

| AMEL Function | MongoDB |
|---------------|---------|
| `exists(x)` | `{x: {"$exists": true}}` |

**Example:**

```
exists($.metadata)
→ {"metadata": {"$exists": true}}
```

---

## Field Mapping

### Default Field Mapping

JSONPath expressions are converted to MongoDB field paths:

```
$.name                    → "name"
$.user.email              → "user.email"
$.items[0].price          → "items.0.price"
$.user.profile.settings   → "user.profile.settings"
```

### Custom Field Mapper

Map AMEL paths to MongoDB field names:

```go
mapper := func(path string) string {
    mappings := map[string]string{
        "$.user.firstName": "profile.first_name",
        "$.user.lastName":  "profile.last_name",
    }
    if mapped, ok := mappings[path]; ok {
        return mapped
    }
    return compiler.DefaultMongoFieldMapper(path)
}

mongoCompiler := compiler.NewMongoDBCompiler(
    compiler.WithMongoFieldMapper(mapper),
)
```

**Example:**

```go
expr, _ := parser.Parse(`$.user.firstName == "John"`)
result, _ := mongoCompiler.Compile(expr)
// Output: {"profile.first_name": "John"}
```

---

## Output Formats

### Query Map

The default output is a Go map that can be used directly with the MongoDB driver:

```go
result, _ := mongoCompiler.Compile(expr)
query := result.Query // map[string]interface{}

// Use with MongoDB driver
cursor, err := collection.Find(ctx, query)
```

### JSON String

Convert to JSON string:

```go
result, _ := mongoCompiler.Compile(expr)
jsonStr, err := result.ToJSON()
if err != nil {
    log.Fatal(err)
}
fmt.Println(jsonStr)
// {"$and":[{"age":{"$gt":18}},{"status":"active"}]}
```

### Pretty JSON

For debugging or logging:

```go
result, _ := mongoCompiler.Compile(expr)
prettyJSON, err := result.ToPrettyJSON()
if err != nil {
    log.Fatal(err)
}
fmt.Println(prettyJSON)
// {
//   "$and": [
//     {"age": {"$gt": 18}},
//     {"status": "active"}
//   ]
// }
```

---

## Examples

### Basic User Query

```go
expr, _ := parser.Parse(`$.email == "user@example.com"`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"email": "user@example.com"}
```

### Active Users Over 18

```go
expr, _ := parser.Parse(`$.age > 18 && $.active == true`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"$and": [{"age": {"$gt": 18}}, {"active": true}]}
```

### Multiple OR Conditions

```go
expr, _ := parser.Parse(`
    $.role == "admin" || 
    $.role == "manager" || 
    $.role == "superuser"
`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"$or": [
//   {"role": "admin"},
//   {"role": "manager"},
//   {"role": "superuser"}
// ]}
```

### Complex Nested Conditions

```go
expr, _ := parser.Parse(`
    ($.age >= 18 && $.age <= 65) || $.role == "admin"
`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"$or": [
//   {"$and": [
//     {"age": {"$gte": 18}},
//     {"age": {"$lte": 65}}
//   ]},
//   {"role": "admin"}
// ]}
```

### Status Filter with IN

```go
expr, _ := parser.Parse(`$.status IN ["pending", "approved", "processing"]`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"status": {"$in": ["pending", "approved", "processing"]}}
```

### Negated Conditions

```go
expr, _ := parser.Parse(`!($.status == "deleted")`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"status": {"$ne": "deleted"}}
```

### Email Domain Check

```go
expr, _ := parser.Parse(`$.email =~ "@company\\.com$"`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"email": {"$regex": "@company\\.com$"}}
```

### Nested Field Access

```go
expr, _ := parser.Parse(`$.user.profile.settings.theme == "dark"`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"user.profile.settings.theme": "dark"}
```

### Array Index Access

```go
expr, _ := parser.Parse(`$.items[0].name == "first"`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Query: {"items.0.name": "first"}
```

### Full Application Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bencagri/amel/pkg/compiler"
    "github.com/bencagri/amel/pkg/parser"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    collection := client.Database("mydb").Collection("users")

    // Define the filter expression
    filterExpr := `
        $.status == "active" &&
        $.age >= 18 &&
        $.role IN ["user", "admin"] &&
        isNotNull($.email)
    `

    // Parse and compile
    expr, err := parser.Parse(filterExpr)
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }

    mongoCompiler := compiler.NewMongoDBCompiler()
    result, err := mongoCompiler.Compile(expr)
    if err != nil {
        log.Fatalf("Compile error: %v", err)
    }

    // Debug: print the query
    prettyJSON, _ := result.ToPrettyJSON()
    fmt.Printf("MongoDB Query:\n%s\n\n", prettyJSON)

    // Execute query
    cursor, err := collection.Find(ctx, result.Query)
    if err != nil {
        log.Fatalf("Query error: %v", err)
    }
    defer cursor.Close(ctx)

    // Process results
    for cursor.Next(ctx) {
        var user struct {
            Name  string `bson:"name"`
            Email string `bson:"email"`
            Age   int    `bson:"age"`
        }
        if err := cursor.Decode(&user); err != nil {
            log.Printf("Decode error: %v", err)
            continue
        }
        fmt.Printf("User: %s (%s) - Age: %d\n", user.Name, user.Email, user.Age)
    }
}
```

### Combining with Aggregation

```go
// Compile the match stage filter
expr, _ := parser.Parse(`$.status == "active" && $.age >= 18`)
result, _ := compiler.NewMongoDBCompiler().Compile(expr)

// Build aggregation pipeline
pipeline := mongo.Pipeline{
    {{"$match", result.Query}},
    {{"$group", bson.D{
        {"_id", "$department"},
        {"count", bson.D{{"$sum", 1}}},
        {"avgAge", bson.D{{"$avg", "$age"}}},
    }}},
    {{"$sort", bson.D{{"count", -1}}}},
}

cursor, err := collection.Aggregate(ctx, pipeline)
```

---

## Limitations

### Unsupported Features

The following AMEL features cannot be compiled to MongoDB queries:

- Custom JavaScript functions
- Lambda expressions (`map`, `filter`, `reduce`)
- Arithmetic operations in conditions
- Most built-in functions (except null checks and string pattern functions)

### Workarounds

For unsupported features, consider:

1. **Use MongoDB aggregation pipeline** for complex transformations
2. **Filter in application code** after initial MongoDB query
3. **Combine approaches**: Use MongoDB for basic filtering, AMEL for complex evaluation

```go
// Basic filter via MongoDB
expr, _ := parser.Parse(`$.status == "active"`)
mongoResult, _ := compiler.NewMongoDBCompiler().Compile(expr)

cursor, _ := collection.Find(ctx, mongoResult.Query)

// Complex evaluation in application
eng, _ := engine.New()
complexExpr := `sum(map($.items, x => x.price)) > 100`

for cursor.Next(ctx) {
    var doc map[string]interface{}
    cursor.Decode(&doc)
    
    // Evaluate complex expression in memory
    matches, _ := eng.EvaluateDirectBool(complexExpr, doc)
    if matches {
        // Process matching document
    }
}
```

### Error Handling

```go
result, err := mongoCompiler.Compile(expr)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "unsupported"):
        log.Printf("Feature not supported in MongoDB: %v", err)
        // Fall back to in-memory evaluation
    default:
        log.Printf("Compilation error: %v", err)
    }
}
```

---

## See Also

- [SQL Compilation](./sql-compilation.md) - Compiling to SQL queries
- [Expression Syntax](./syntax.md) - AMEL syntax reference
- [Built-in Functions](./functions.md) - Function reference