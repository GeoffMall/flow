# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`flow` is a streaming JSON & YAML processor CLI tool written in Go. It's designed as a lightweight alternative to `jq`, emphasizing readability, composability, and streaming (no full in-memory parsing required). Users can extract, transform, and filter structured data using simple command-line flags.

## Commands

### Building and Running
```bash
# Build from source
go build -o flow

# Run with stdin
echo '{"user":{"name":"Alice"}}' | ./flow -pick user.name

# Run with file input
./flow -in data.json -pick user.name

# Convert formats
./flow -in config.yaml -to json -color
```

### Testing
```bash
# Run all tests (includes linting, security checks, vulnerability scanning)
make test

# Run unit tests only
go test ./...

# Test specific package
go test ./internal/operation

# Test specific function
go test -run TestPick_WildcardExpansion ./internal/operation

# Test with coverage
go test ./... -race -covermode=atomic -coverprofile=coverage.out
```

### Code Quality
```bash
# Lint
golangci-lint run

# Security scan
gosec --quiet -r

# Vulnerability check
govulncheck ./...

# Format code
gofmt -w . && goimports -w .
```

## Architecture

### Core Data Flow
```
stdin/file → Parser (auto-detect) → ForEach(doc) → Pipeline.Apply(doc) → Printer → stdout/file
```

The application processes data in a streaming fashion without loading entire documents into memory.

### Key Packages

**`internal/app/app.go`**: Application orchestration layer
- `run()` function coordinates the entire processing pipeline
- Builds pipeline from CLI flags in order: Pick → Set → Delete
- Uses `parser.ForEach()` callback pattern for streaming
- Handles input/output file descriptors (stdin/stdout by default)

**`internal/parser/`**: Streaming input parsing
- **Auto-detection**: Peeks at first bytes to determine JSON vs YAML format
  - `{` or `[` → JSON
  - `---` or `%YAML` → YAML
  - `:` before `,` or `}` → likely YAML
- **Streaming behavior**:
  - JSON: Supports concatenated documents; if root is array, streams each element separately
  - YAML: Streams documents separated by `---`
- **Normalization**: Converts YAML's `map[any]any` to JSON-compatible `map[string]any`
- Uses `bufio.Reader` with 64KB buffer for efficient peeking
- Only one document in memory at a time

**`internal/operation/`**: Data transformation operations
- **Operation interface**: `Apply(v any) (any, error)` - all transformations implement this
- **Path syntax**: Supports dot notation (`user.name`), array indexing (`items[0]`), wildcards (`items[*].name`)
- **segment struct**: Internal representation of path steps with optional array indices
- **Pick**: Extracts specified fields, preserves nested structure, expands wildcards
- **Set**: Parses values as JSON-ish (tries JSON first, falls back to string), creates intermediate maps/arrays as needed
- **Delete**: Removes fields at paths, handles array element shifting
- **Pipeline**: Chains operations sequentially, returns `StepError` with context on failure

**`internal/printer/`**: Output formatting
- Format conversion (JSON ↔ YAML)
- Pretty-printing with 2-space indentation
- Compact/minified output support
- **color.go**: State machine-based ANSI colorizer for JSON (keys=blue, strings=green, numbers=orange, bools/null=purple)

**`internal/cli/flag.go`**: CLI flag parsing
- Implements custom `multiStringFlag` type for repeatable flags (`-pick`, `-set`, `-delete`)
- Uses standard library `flag` package
- `Flags` struct holds all parsed arguments

### Path Parsing Details

Path strings are parsed into `segment` structs representing each step:
- `user.name` → `[{field: "user"}, {field: "name"}]`
- `items[0].name` → `[{field: "items", index: 0}, {field: "name"}]`
- `items[*]` → Expands to all concrete indices at evaluation time

Wildcards are resolved during operation application by calling `expandWildcards()`, which traverses the data structure and generates concrete paths for each matching element.

### Testing Patterns

Tests use `github.com/stretchr/testify` for assertions:
```go
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.Contains(t, output, substring)
```

Test helper pattern in `app_test.go`:
```go
runTest(input, opts) → (output, error)
assertRunSucceeds(input, opts, expected)
assertRunFails(input, opts)
```

Use table-driven tests for comprehensive coverage:
```go
tests := []struct{
    name     string
    input    string
    expected string
}{...}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

## Adding New Features

### Adding a New Operation
1. Create new file in `internal/operation/` (e.g., `filter.go`)
2. Implement the `Operation` interface with `Apply()` and `Description()` methods
3. Add corresponding flag in `internal/cli/flag.go` to the `Flags` struct
4. Update flag parsing in `cli/flag.go` to populate the new flag
5. Add operation construction logic to `buildPipeline()` in `internal/app/app.go`
6. Write tests following the existing pattern (e.g., `filter_test.go`)

### Extending Output Formats
1. Add encoder logic in `internal/printer/printer.go` → `Write()` method
2. Add format validation in `internal/cli/flag.go`
3. Update auto-detection if input format detection is also needed

### Modifying Parser Behavior
- Format detection heuristics: `internal/parser/parser.go` → `detectFormat()` and helper functions
- Streaming logic: Modify `forEachJSON()` or `forEachYAML()` in `parser.go`
- Data normalization: Update `normalizeYAML()` or `normalizeJSON()`

## Important Implementation Notes

- **Streaming First**: Never buffer entire input; process documents one at a time via callbacks
- **Type System**: All data internally represented as `any` with Go's JSON types (`map[string]any`, `[]any`, primitives)
- **Error Context**: Operations wrap errors with context (which operation failed, which path)
- **Wildcard Expansion**: Happens at operation application time, not parse time
- **Format Agnostic Operations**: Operations work on normalized Go values, not format-specific structures
- **Security**: File path operations marked with `#nosec G304` where intentional; all code scanned by gosec
