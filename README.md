# `flow` - A Streaming JSON & YAML Processor

> A lightweight, fast, Go-native CLI for extracting, transforming, and filtering structured data.

## Overview

`flow` is a command-line tool designed to make working with structured data (JSON and YAML) as **easy as piping**.  
This is an alternative to tools like [jq](https://stedolan.github.io/jq/), and emphasizes **readability**, **composability**, and **streaming** — making it ideal for DevOps, data wrangling, and quick debugging.

```bash
cat data.json | flow -pick user.name -pick user.id
flow -in config.yaml -set server.debug=false -delete server.secret
```

### Installation

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

### Key features:

- Streaming, no full in-memory parse required
- Simple flag-based syntax (no DSL to learn)
- Works with both JSON and YAML seamlessly
- Written in Go for speed and portability
- Friendly error messages
- Interoperable with stdin/stdout for easy piping

### ⚙️ Example Usage
```bash
# Pick two fields
cat data.json | flow -pick user.name -pick user.id

# Set and delete
flow -in data.yaml -set server.port=8080 -delete debug

# Convert YAML to JSON
flow -in config.yaml -to json

# Pretty-print JSON with colors
flow -in input.json -color

# Stream transformations
# Does not work yet, but coming soon:
# cat big.json | flow --pick data.items[*].name
```

### Roadmap

- [ ] Multiple input sources and/or folder support
- [ ] Advanced querying (e.g., filtering arrays)
- [ ] More supported formats (CSV, XML, Avro, Parquet)