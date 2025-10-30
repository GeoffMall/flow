# Gemini Code Assistant Context

## Project: `flow` - A Streaming JSON & YAML Processor

This document provides context for the Gemini code assistant to understand and assist with the development of `flow`.

### Project Overview

`flow` is a command-line tool written in Go for processing structured data like JSON and YAML. It is designed as a lightweight and fast alternative to tools like `jq`, with a focus on readability, composability, and streaming. The tool allows users to extract, transform, and filter data using simple command-line flags.

**Core Technologies:**

*   **Language:** Go (version 1.25.1)
*   **Dependencies:**
    *   `gopkg.in/yaml.v3`: For YAML parsing and serialization.
    *   `github.com/stretchr/testify`: For assertions in tests.

**Architecture:**

The application follows a modular architecture, with clear separation of concerns:

*   **`main.go`**: The entry point of the application, which calls the `app.Run()` function.
*   **`internal/app`**: Contains the core application logic, including input/output handling and orchestrating the processing pipeline.
*   **`internal/cli`**: Handles command-line flag parsing.
*   **`internal/parser`**: Responsible for parsing input streams (JSON or YAML).
*   **`internal/operation`**: Defines the data transformation operations (`pick`, `set`, `delete`).
*   **`internal/printer`**: Handles the output formatting (JSON or YAML, with options for color and compactness).

### Building and Running

**Building from source:**

```bash
go build -o flow
```

**Running the application:**

```bash
cat data.json | ./flow -pick user.name
./flow -in config.yaml -set server.port=8080
```

### Development Conventions

**Testing:**

The project has a `test` command in the `Makefile` that runs a suite of checks:

```bash
make test
```

This command executes the following:

*   `go test ./...`: Runs all unit tests.
*   `golangci-lint run`: Lints the codebase for style and errors.
*   `gosec --quiet -r`: A security scanner for Go code.
*   `govulncheck ./...`: Checks for known vulnerabilities in dependencies.

**Contribution Guidelines:**

While not explicitly stated, the presence of a comprehensive test suite and linting suggests that contributions should be well-tested and adhere to the existing code style.
