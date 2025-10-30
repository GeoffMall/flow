
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

## ⚙️ Usage

### Picking Fields

Use the `-pick` flag to extract one or more fields from your data. You can use dot notation to access nested fields.

```bash
# Pick a single field
cat data.json | flow -pick user.name

# Pick multiple fields
cat data.json | flow -pick user.name -pick user.id

# Pick multiple fields by wildcard
flow -in data.yaml -pick items[*].name
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

# Pretty-print JSON with colors
flow -in input.json -color
```

## Roadmap

- [ ] Multiple input sources and/or folder support
- [ ] Advanced querying (e.g., filtering arrays)
- [ ] More supported formats (CSV, XML, Avro, Parquet)

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
