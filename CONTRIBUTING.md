# Contributing to AMEL

Thank you for your interest in contributing to AMEL (Adaptive Matching Expression Language)! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)
- [License](#license)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Please:

- Be respectful and considerate in all interactions
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept responsibility for your mistakes and learn from them

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/amel.git
   cd amel
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/bencagri/amel.git
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## How to Contribute

### Types of Contributions

We welcome the following types of contributions:

- 🐛 **Bug fixes** - Fix issues and improve stability
- ✨ **New features** - Add new functionality
- 📚 **Documentation** - Improve or add documentation
- 🧪 **Tests** - Add or improve test coverage
- ⚡ **Performance** - Optimize existing code
- 🔧 **Refactoring** - Improve code quality without changing behavior

### Before You Start

1. **Check existing issues** to see if your contribution is already being worked on
2. **Open an issue** to discuss significant changes before implementing
3. **Review the documentation** to understand the project architecture

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/amel.git
cd amel

# Install dependencies
go mod download

# Verify setup
go test ./...
```

### Project Structure

```
amel/
├── pkg/
│   ├── ast/          # Abstract Syntax Tree definitions
│   ├── lexer/        # Tokenizer
│   ├── parser/       # Expression parser
│   ├── eval/         # Expression evaluator
│   ├── functions/    # Function registry and sandbox
│   ├── compiler/     # SQL/MongoDB compilers
│   ├── types/        # Type definitions
│   ├── optimizer/    # AST optimizer
│   └── engine/       # Main engine facade
├── docs/             # Documentation
├── internal/         # Internal packages
└── examples/         # Example code
```

## Coding Standards

### Go Style

- Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Keep functions small and focused
- Write descriptive variable and function names

### Code Formatting

```bash
# Format code
go fmt ./...

# Run linter
golint ./...

# Run vet
go vet ./...
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase, short | `lexer`, `ast` |
| Interfaces | PascalCase, often `-er` suffix | `Evaluator`, `Parser` |
| Structs | PascalCase | `BinaryExpression` |
| Functions | PascalCase (exported), camelCase (unexported) | `Evaluate`, `parseExpression` |
| Constants | PascalCase or ALL_CAPS | `TypeInt`, `MAX_STACK_DEPTH` |
| Variables | camelCase | `tokenList`, `currentIndex` |

### Documentation

- Add comments to all exported functions, types, and constants
- Use complete sentences starting with the item name
- Include examples in documentation where helpful

```go
// Evaluate evaluates the given expression against the provided context.
// It returns a Value containing the result and any error that occurred.
func (e *Evaluator) Evaluate(expr ast.Expression, ctx *Context) (types.Value, error) {
    // ...
}
```

## Testing Guidelines

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./pkg/eval/...

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

1. **Test file naming**: `*_test.go` in the same package
2. **Test function naming**: `TestFunctionName_Scenario`
3. **Use table-driven tests** for multiple cases
4. **Include both positive and negative test cases**

```go
func TestEvaluator_Comparison(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected interface{}
    }{
        {"greater than true", "5 > 3", true},
        {"greater than false", "3 > 5", false},
        {"equal", "5 == 5", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := evaluate(tt.input)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if result != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### Test Coverage Requirements

| Component | Minimum Coverage |
|-----------|------------------|
| Lexer | 95% |
| Parser | 90% |
| Evaluator | 85% |
| Functions | 85% |
| Engine | 80% |

## Pull Request Process

### Before Submitting

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all tests** and ensure they pass:
   ```bash
   go test ./...
   ```

3. **Run linters**:
   ```bash
   go fmt ./...
   go vet ./...
   ```

4. **Update documentation** if needed

5. **Write clear commit messages**:
   ```
   Short summary (50 chars or less)

   More detailed explanation if necessary. Wrap at 72 characters.
   Explain the problem this commit solves and why.

   - Bullet points are okay
   - Use present tense: "Add feature" not "Added feature"
   ```

### Submitting a Pull Request

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Description of what and why
   - Link to related issue(s)
   - Screenshots if applicable

3. **PR Title Format**:
   - `feat: Add new feature` - New feature
   - `fix: Fix bug description` - Bug fix
   - `docs: Update documentation` - Documentation only
   - `test: Add tests for X` - Test additions
   - `refactor: Refactor X` - Code refactoring
   - `perf: Improve performance of X` - Performance improvement

### Review Process

1. A maintainer will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR

## Reporting Issues

### Bug Reports

When reporting a bug, please include:

1. **AMEL version** you're using
2. **Go version**: `go version`
3. **Operating system**
4. **Steps to reproduce** the bug
5. **Expected behavior**
6. **Actual behavior**
7. **Code sample** demonstrating the issue

```markdown
**Version**: v1.0.0
**Go Version**: go1.21.0
**OS**: macOS 14.0

**Description**:
Brief description of the bug.

**Steps to Reproduce**:
1. Step one
2. Step two
3. Step three

**Expected Behavior**:
What you expected to happen.

**Actual Behavior**:
What actually happened.

**Code Sample**:
```go
// Minimal code to reproduce
```
```

### Feature Requests

When requesting a feature, please include:

1. **Problem description** - What problem does this solve?
2. **Proposed solution** - How do you envision it working?
3. **Alternatives considered** - What other approaches did you consider?
4. **Use cases** - Who would benefit from this feature?

## Adding New Built-in Functions

If you want to add a new built-in function:

1. **Add the implementation** in `pkg/functions/builtin.go`
2. **Register the function** in `NewDefaultRegistry()`
3. **Add tests** in `pkg/functions/builtin_test.go`
4. **Document the function** in `docs/functions.md`

```go
// Example: Adding a new function
func builtinNewFunc(args ...types.Value) (types.Value, error) {
    if len(args) != 1 {
        return types.Null(), fmt.Errorf("newFunc requires 1 argument")
    }
    // Implementation
    return result, nil
}

// In NewDefaultRegistry():
r.RegisterBuiltIn("newFunc", builtinNewFunc, 
    types.NewFunctionSignature("newFunc", types.TypeString,
        types.Param("input", types.TypeString),
    ))
```

## Adding New Operators

To add a new operator:

1. **Add token type** in `pkg/lexer/token.go`
2. **Update lexer** in `pkg/lexer/lexer.go`
3. **Update parser** if needed in `pkg/parser/parser.go`
4. **Add evaluation logic** in `pkg/eval/evaluator.go`
5. **Add compiler support** in `pkg/compiler/` if applicable
6. **Add comprehensive tests**
7. **Update documentation**

## Questions?

If you have questions about contributing, please:

1. Check the [documentation](./docs/)
2. Search existing [issues](https://github.com/bencagri/amel/issues)
3. Open a new issue with the `question` label

## License

By contributing to AMEL, you agree that your contributions will be licensed under the same license as the project (AGPL-3.0). For proprietary use, a commercial license is available.

---

Thank you for contributing to AMEL! 🎉