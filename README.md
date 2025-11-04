
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
- [Examples](#-examples)
- [Alternatives](#alternatives)
- [Roadmap](#roadmap)
- [Contributing](#contributing)

## Installation

```bash
go install github.com/GeoffMall/flow@latest
```

or build from source:

```bash
git clone https://github.com/GeoffMall/flow.git
cd flow
go build -o flow
```

or download pre-built binaries from the releases page.

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
| **Convert YAML to JSON** | `yq -o json file.yaml` (requires yq) | `flow -in file.yaml -to json` |
| **Read from file** | `jq '.' < file.json` or `jq '.' file.json` | `flow -in file.json` |

**Key differences:**
- **Syntax**: `jq` uses a custom query language; `flow` uses simple CLI flags
- **Learning curve**: `jq` requires learning its DSL; `flow` is immediately intuitive
- **Formats**: `jq` is JSON-only (needs `yq` for YAML); `flow` handles both with auto-detection
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

`flow` can read from `stdin` or from a file using the `-in` flag. By default, `flow` outputs in the same format as the input. You can specify the output format using the `-to` flag.

```bash
# Read from a file and output as JSON
flow -in config.yaml -to json

# Disable colored output (colors are enabled by default)
flow -in input.json -no-color
```

## Alternatives

`flow` is part of a rich ecosystem of JSON/YAML processing tools. Here's how it compares to other popular options:

### Core JSON/YAML Processors

**[jq](https://jqlang.org/)** - The industry-standard JSON processor with a powerful filtering and transformation language. Written in C with no runtime dependencies, offering excellent performance and portability. Supports streaming with `--stream` flag for memory-efficient processing of large files.

**[yq (mikefarah)](https://github.com/mikefarah/yq)** - A multi-format processor supporting YAML, JSON, XML, TOML, and CSV with jq-like syntax. Written in Go, it preserves YAML comments and styling, making it ideal for editing configuration files in-place. Processes documents individually rather than true streaming.

**[dasel](https://github.com/TomWright/dasel)** - A unified selector tool supporting JSON, YAML, TOML, XML, and CSV with a single CSS-like selector syntax. Written in Go as a single binary alternative to jq/yq, with format conversion capabilities. Loads entire documents into memory.

### Interactive Viewers & Explorers

**[jless](https://jless.io/)** - A terminal-based JSON viewer with Vim-inspired navigation and expandable/collapsible nodes. Written in Rust, designed for exploring complete JSON documents interactively rather than streaming processing.

**[fx](https://fx.wtf/)** - An interactive JSON viewer with JavaScript-based data manipulation. Written in Go, supports streaming for line-delimited JSON, and handles YAML/TOML. Ideal for exploring logs and applying quick transformations.

**[jid](https://github.com/simeji/jid)** - An interactive JSON digger with auto-completion for incremental filtering. Written in Go, it helps explore unknown JSON structures interactively but doesn't support streaming.

### Specialized Tools

**[gron](https://github.com/tomnomnom/gron)** - Flattens JSON into discrete assignments, making it greppable with standard Unix tools. Written in Go with streaming support, it excels at finding paths to values in complex structures and can unflatten back to JSON.

**[miller (mlr)](https://github.com/johnkerl/miller)** - Like awk/sed/cut for structured data, processing CSV, TSV, and JSON. Written in Go (originally C), designed for streaming large files with format-aware operations that preserve headers and structure.

**[jp](https://github.com/therealklanni/jp)** - A Node.js-based JSON parser supporting Lodash `.get()` and JSONPath syntax. Can process line-delimited JSON streams, offering familiar JavaScript-style path expressions.

### JSON Generation & Conversion

**[jo](https://github.com/jpmens/jo)** - Creates JSON objects and arrays from shell command arguments. Written in C, designed for generating JSON in shell scripts rather than processing existing data.

**[jc](https://github.com/kellyjonbrazil/jc)** - Converts output from common CLI tools (ls, ps, dig, etc.) into JSON. Written in Python with an extensive parser library, enabling structured processing of traditionally unstructured command output.

**[jsonnet](https://jsonnet.org/)** - A data templating language that extends JSON with variables, functions, and imports. Written in C++ with bindings for multiple languages, designed for generating complex JSON configurations rather than processing streams.

### How `flow` Differs

`flow` emphasizes:
- **Streaming-first architecture** - True streaming with no full document parsing required
- **Simplicity** - Command-line flags instead of a custom query language
- **Format flexibility** - Auto-detection and conversion between JSON and YAML
- **Lightweight** - Single Go binary with minimal dependencies
- **Composability** - Pipeline-based operations (pick, set, delete) that work together naturally

## Roadmap

- [ ] Multiple input sources and/or folder support
- [ ] Advanced querying (e.g., filtering arrays)
- [ ] More supported formats (CSV, XML, Avro, Parquet)

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
