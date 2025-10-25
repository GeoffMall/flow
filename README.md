# `flow` - A Streaming JSON & YAML Processor

> A lightweight, fast, Go-native CLI for extracting, transforming, and filtering structured data.

## 🧭 Overview

`flow` is a command-line tool designed to make working with structured data (JSON and YAML) as **easy as piping**.  
Unlike tools like [jq](https://stedolan.github.io/jq/), `flow` emphasizes **readability**, **composability**, and **streaming** — making it ideal for DevOps, data wrangling, and quick debugging.

```bash
cat data.json | flow -pick user.name -pick user.id
flow -in config.yaml -set server.debug=false -delete server.secret
```

### Key features:

- Streaming, no full in-memory parse required
- Simple flag-based syntax (no DSL to learn)
- Works with both JSON and YAML seamlessly
- Written in Go for speed and portability
- Friendly error messages


### Design Philosophy

- Minimal cognitive load: flags over DSL. flow --pick user.name is easier than jq '.user.name'.
- Predictable transformations: operations are composable and executed in deterministic order.
- Stream-first: never fully load large files; process as they come.
- Interoperable: works with stdin/stdout by default, supports files optionally.
- Speed: Go streaming parser is optimized for performance.
- Extensible: each new operation lives in its own file with a shared interface.


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
