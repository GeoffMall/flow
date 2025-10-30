# Agent Instructions for flow

## Build/Lint/Test Commands
- **Build**: `go build ./...`
- **Test all**: `go test ./...`
- **Test with coverage**: `go test ./... -race -covermode=atomic -coverprofile=coverage.out`
- **Test single package**: `go test ./path/to/package`
- **Test single function**: `go test -run TestFunctionName ./path/to/package`
- **Lint**: `golangci-lint run`
- **Security scan**: `gosec --quiet -r`
- **Vulnerability check**: `govulncheck ./...`
- **Format**: `gofmt -w . && goimports -w .`

## Code Style Guidelines
- **Go version**: 1.25.1
- **Testing**: Use testify (assert.Equal, assert.NoError, assert.Contains)
- **Imports**: Standard library → third-party → internal (grouped with blank lines)
- **Naming**: camelCase functions, PascalCase exported types/structs
- **Error handling**: Check errors immediately, use early returns
- **Cleanup**: Use defer for resource cleanup
- **Comments**: Document exported functions/types only
- **Formatting**: gofmt + goimports (enforced by golangci-lint)

## Testing Patterns
- Test functions: `Test_FunctionName_Description`
- Use `bytes.Buffer` for output testing
- Use `strings.NewReader` for input testing
- Table-driven tests for multiple test cases