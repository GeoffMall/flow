.PHONY: test build build-all install clean release-test

# Version info (can be overridden)
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -s -w \
	-X github.com/GeoffMall/flow/internal/version.Version=$(VERSION) \
	-X github.com/GeoffMall/flow/internal/version.GitCommit=$(COMMIT) \
	-X github.com/GeoffMall/flow/internal/version.BuildDate=$(DATE)

# Test target (linting, security, vulnerabilities)
test:
	go test ./...
	golangci-lint run
	gosec --quiet -r
	govulncheck ./...

# Build for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o flow

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/flow-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/flow-linux-arm64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/flow-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/flow-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/flow-windows-amd64.exe
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/flow-windows-arm64.exe
	@echo "Build complete! Binaries in dist/"

# Install to /usr/local/bin
install: build
	install -m 755 flow /usr/local/bin/flow

# Clean build artifacts
clean:
	rm -f flow flow.exe
	rm -rf dist/

# Test release locally (requires goreleaser)
release-test:
	goreleaser release --snapshot --clean --skip=publish
