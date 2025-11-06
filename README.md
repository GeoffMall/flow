
# `flow` - A Streaming JSON & YAML Processor

> A lightweight, fast, Go-native CLI for extracting, transforming, and filtering structured data.

`flow` is a command-line tool designed to make working with structured data (JSON and YAML) as **easy as piping**.  
This is an alternative to tools like [jq](https://stedolan.github.io/jq/), and emphasizes **readability**, **composability**, and **streaming** — making it ideal for DevOps, data wrangling, and quick debugging.

## Table of Contents

- [Installation](#installation)
- [Key Features](#key-features)
- [Usage](#-usage)
  - [Picking Fields](#picking-fields)
  - [Setting Fields](#setting-fields)
  - [Deleting Fields](#deleting-fields)
  - [Input and Output](#input-and-output)
- [Alternatives](#alternatives)
- [Roadmap](#roadmap)
- [Contributing](#contributing)

## Installation

### macOS & Linux

#### Homebrew (Recommended)
```bash
brew install GeoffMall/tap/flow
```

#### Install Script
```bash
curl -fsSL https://raw.githubusercontent.com/GeoffMall/flow/main/install.sh | sh
```

#### Go Install
```bash
go install github.com/GeoffMall/flow@latest
```

#### Manual Download
Download the latest release for your platform from the [releases page](https://github.com/GeoffMall/flow/releases).

**macOS users:** If you see an "untrusted developer" warning, run:
```bash
xattr -d com.apple.quarantine /usr/local/bin/flow
```

### Windows

#### Scoop (Recommended)
```powershell
scoop bucket add flow https://github.com/GeoffMall/scoop-bucket
scoop install flow
```

#### Manual Download
1. Download `flow_*_windows_amd64.zip` from the [releases page](https://github.com/GeoffMall/flow/releases)
2. Extract the zip file
3. Add the `flow.exe` to your PATH

### Build from Source

```bash
git clone https://github.com/GeoffMall/flow.git
cd flow
make build
# Or for all platforms:
make build-all
```

## Key Features

- Streaming, no full in-memory parse required
- Simple flag-based syntax (no DSL to learn)
- Works with both JSON and YAML seamlessly
- Written in Go for speed and portability
- Friendly error messages
- Interoperable with stdin/stdout for easy piping

## Comparison with jq

Here's how `flow` compares to `jq` for common data extraction tasks:

| Task | jq | flow |
|------|-----|------|
| **Extract a field** | `jq '.user.name'` | `flow -pick user.name` |
| **Extract nested field** | `jq '.server.config.port'` | `flow -pick server.config.port` |
| **Multiple fields** | `jq '{name: .user.name, id: .user.id}'` | `flow -pick user.name -pick user.id` |
| **Array element** | `jq '.items[0]'` | `flow -pick items[0]` |
| **All array items** | `jq '.items[]'` | `flow -pick items[*]` |
| **Nested array fields** | `jq '.items[].name'` | `flow -pick items[*].name` |
| **Convert YAML to JSON** | `yq -o json file.yaml` (requires yq) | `flow -in file.yaml -to json` (auto-detected from .yaml extension) |
| **Read from file** | `jq '.' < file.json` or `jq '.' file.json` | `flow -in file.json` |

**Key differences:**
- **Syntax**: `jq` uses a custom query language; `flow` uses simple CLI flags
- **Learning curve**: `jq` requires learning its DSL; `flow` is immediately intuitive
- **Formats**: `jq` is JSON-only (needs `yq` for YAML); `flow` handles both (defaults to JSON, detects YAML from .yaml/.yml extensions)
- **Streaming**: Both support streaming, but `flow` does it by default without special flags
- **Output**: `flow` now outputs values just like `jq` (e.g., `-pick user.name` returns just `"alice"`, not `{"user": {"name": "alice"}}`)

## ⚙️ Usage

### Picking Fields

Use the `-pick` flag to extract one or more fields from your data. You can use dot notation to access nested fields.

**By default, `flow` outputs values like `jq`** - extracting just the value without preserving the full path structure:

```bash
# Pick a single field - outputs just the value
echo '{"user":{"name":"alice","age":30}}' | flow -pick user.name
# Output: "alice"

# Pick multiple fields - outputs flattened object
echo '{"user":{"name":"alice","id":7}}' | flow -pick user.name -pick user.id
# Output: {"name": "alice", "id": 7}

# Pick with wildcard - outputs array of values
echo '{"items":[{"name":"a"},{"name":"b"}]}' | flow -pick 'items[*].name'
# Output: ["a", "b"]
```

**Backward compatibility:** Use `--preserve-hierarchy` to maintain the full path structure (legacy behavior):

```bash
echo '{"user":{"name":"alice"}}' | flow -pick user.name --preserve-hierarchy
# Output: {"user": {"name": "alice"}}
```

### Setting Fields

Use the `-set` flag to add or modify fields. You can set multiple fields in a single command.

```bash
# Set a field
flow -in config.yaml -set server.port=8080

# Set multiple fields
flow -in config.yaml -set server.port=8080 -set server.host=localhost
```

### Deleting Fields

Use the `-delete` flag to remove fields.

```bash
# Delete a field
flow -in config.yaml -delete server.secret
```

### Input and Output

`flow` can read from `stdin` or from a file using the `-in` flag.

**Input Format:**
- JSON is the default format
- YAML is automatically detected for files with `.yaml` or `.yml` extensions
- Use `-from yaml` to explicitly specify YAML input (required when piping YAML from stdin)

**Output Format:**
- Defaults to JSON
- Use `-to yaml` to output as YAML

```bash
# Read YAML file (auto-detected from extension)
flow -in config.yaml -pick server.port

# Read YAML from stdin (requires -from flag)
cat config.yaml | flow -from yaml -pick server.port

# Convert YAML to JSON
flow -in config.yaml -to json

# Convert JSON to YAML
flow -in data.json -to yaml

# Disable colored output (colors are enabled by default)
flow -in input.json -no-color
```

## Alternatives

If you're exploring other tools for JSON/YAML processing:

- **[jq](https://jqlang.org/)** - Industry-standard JSON processor with powerful query language
- **[yq](https://github.com/mikefarah/yq)** - Multi-format processor (YAML, JSON, XML, TOML, CSV)
- **[fx](https://fx.wtf/)** - Interactive JSON viewer with JavaScript-based manipulation
- **[gron](https://github.com/tomnomnom/gron)** - Makes JSON greppable by flattening into assignments
- **[dasel](https://github.com/TomWright/dasel)** - Unified selector for JSON, YAML, TOML, XML, and CSV
- **[jless](https://jless.io/)** - Terminal-based JSON viewer with Vim-style navigation
- **[miller](https://github.com/johnkerl/miller)** - Like awk/sed for CSV, TSV, and JSON

**Why `flow`?**
- **Streaming-first** - No full document parsing required
- **Simple syntax** - CLI flags instead of a custom query language
- **Format flexible** - Auto-detects and converts JSON/YAML seamlessly
- **Single binary** - Just one Go executable, no dependencies

## Roadmap

- [ ] Multiple input sources and/or folder support
- [ ] Advanced querying (e.g., filtering arrays)
- [ ] More supported formats (CSV, XML, Avro, Parquet)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:
- Setting up your development environment
- Creating feature branches from `dev`
- PR title requirements
- Code quality standards
