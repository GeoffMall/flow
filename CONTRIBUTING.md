# Contributing to Flow

Thank you for your interest in contributing to `flow`! This document provides guidelines and instructions for contributing.

## Branching Strategy

We use a two-branch workflow:

- **`main`**: Production-ready code. Only accepts PRs from `dev`.
- **`dev`**: Integration branch for testing features. Only accepts PRs from feature branches.
- **`feature/*`**: Feature branches created from `dev`.

## Development Workflow

### 1. Set Up Your Development Environment

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/flow.git
cd flow

# Add upstream remote
git remote add upstream https://github.com/GeoffMall/flow.git

# Install dependencies and verify tests pass
go mod download
make test
```

### 2. Create a Feature Branch

**IMPORTANT: Always branch from `dev`, not `main`!**

```bash
# Fetch latest changes
git fetch upstream

# Checkout dev and update
git checkout dev
git pull upstream dev

# Create your feature branch from dev
git checkout -b feature/your-feature-name
```

**Feature branch naming conventions:**
- `feature/add-csv-support` - for new features
- `fix/parser-crash` - for bug fixes
- `docs/update-readme` - for documentation
- `refactor/cleanup-parser` - for refactoring
- `test/add-edge-cases` - for test improvements

### 3. Make Your Changes

```bash
# Make changes to the code
# ...

# Run tests frequently
make test

# Format code before committing
gofmt -w . && goimports -w .

# Commit with descriptive messages
git add .
git commit -m "feat: add CSV format support"
```

**Commit message format:**
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat: add new feature` - New features
- `fix: resolve bug` - Bug fixes
- `docs: update documentation` - Documentation changes
- `test: add tests` - Test additions
- `refactor: restructure code` - Code refactoring
- `chore: update dependencies` - Maintenance tasks
- `perf: improve performance` - Performance improvements

### 4. Push and Create Pull Request

```bash
# Push your feature branch to your fork
git push origin feature/your-feature-name
```

Then on GitHub:
1. Go to https://github.com/GeoffMall/flow
2. Click **"Compare & pull request"**
3. **Set base branch to `dev`** (not `main`!)
4. **Title your PR using Conventional Commits format:**
   - ✅ `feat: add CSV format support`
   - ✅ `fix: resolve parser crash on empty input`
   - ✅ `docs: update installation instructions`
   - ❌ `Add CSV support` (missing type prefix)
   - ❌ `feat Add CSV` (missing colon)
5. Provide a clear description of your changes
6. Click **"Create pull request"**

**PR Title Requirements:**
- Must follow format: `type: description`
- Valid types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `perf`, `ci`, `build`, `style`
- Examples:
  - `feat: add streaming CSV parser`
  - `fix: handle empty YAML documents`
  - `docs: add CSV format examples to README`

### 5. Code Review Process

- Your PR will be reviewed by maintainers
- Address any feedback by pushing new commits to your branch
- Once approved, your PR will be **squash merged** into `dev` with the PR title as the commit message
- Your feature branch will be automatically deleted after merge

### 6. Keeping Your Branch Updated

If `dev` has been updated while you're working:

```bash
# Fetch latest changes
git fetch upstream

# Rebase your feature branch on latest dev
git checkout feature/your-feature-name
git rebase upstream/dev

# If you've already pushed, you'll need to force push
git push origin feature/your-feature-name --force-with-lease
```

## Code Quality Standards

### Running Tests

```bash
# Run all tests (includes linting, security checks, vulnerability scanning)
make test

# Run unit tests only
go test ./...

# Run tests with coverage
go test ./... -race -covermode=atomic -coverprofile=coverage.out
```

### Code Quality Checks

```bash
# Lint your code
golangci-lint run

# Security scan
gosec --quiet ./...

# Vulnerability check
govulncheck ./...

# Format code
gofmt -w . && goimports -w .
```

**All CI checks must pass before merge:**
- ✅ Unit tests
- ✅ Linting (golangci-lint)
- ✅ Security scan (gosec)
- ✅ Vulnerability check (govulncheck)

## Adding New Features

### Adding a New Data Format

See the [CLAUDE.md](./CLAUDE.md#adding-new-features) file for detailed instructions on adding new formats like CSV, Parquet, or Avro.

### Adding a New Operation

1. Create new file in `internal/operation/` (e.g., `filter.go`)
2. Implement the `Operation` interface
3. Add corresponding CLI flag in `internal/cli/flag.go`
4. Update `buildPipeline()` in `internal/runner/app.go`
5. Write comprehensive tests
6. Update documentation

## Architecture Guidelines

- **Streaming First**: Never buffer entire input; process documents one at a time
- **Format Agnostic**: Operations work on normalized Go values, not format-specific structures
- **Error Context**: Wrap errors with context about which operation/path failed
- **Type System**: Use `map[string]any`, `[]any`, and primitives for all internal data

## Getting Help

- Check existing [Issues](https://github.com/GeoffMall/flow/issues)
- Read the [README.md](./README.md) and [CLAUDE.md](./CLAUDE.md)
- Ask questions by opening a new issue

## Release Process

Releases are automated and happen when `dev` is merged into `main`:
1. Features are merged into `dev` via squash merge
2. Periodically, `dev` is merged into `main` (preserving all commits)
3. On merge to `main`, semantic-release creates a new version
4. GoReleaser builds binaries and updates Homebrew/Scoop

**Contributors don't need to worry about releases - maintainers handle this!**

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help create a welcoming environment for all contributors

## Questions?

Feel free to open an issue or discussion if you have any questions!
