#!/bin/sh
set -e

# flow installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/GeoffMall/flow/main/install.sh | sh
# Or: wget -qO- https://raw.githubusercontent.com/GeoffMall/flow/main/install.sh | sh

REPO="GeoffMall/flow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    error "Unsupported OS: $OS"
    ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    error "Unsupported architecture: $ARCH"
    ;;
esac

info "Detected platform: ${OS}/${ARCH}"

# Get latest version if not specified
if [ "$VERSION" = "latest" ]; then
  info "Fetching latest version..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    error "Failed to fetch latest version"
  fi
  info "Latest version: v${VERSION}"
fi

# Build download URL
BINARY="flow"
if [ "$OS" = "windows" ]; then
  BINARY="flow.exe"
fi

# Use tar.gz for Unix, zip for Windows (though both are available)
if [ "$OS" = "windows" ]; then
  EXT="zip"
else
  EXT="tar.gz"
fi

FILENAME="flow_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/$REPO/releases/download/v${VERSION}/${FILENAME}"

info "Downloading flow v${VERSION}..."
echo "  URL: $URL"

# Create temporary directory
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

cd "$TMPDIR"

# Download
if command -v curl > /dev/null 2>&1; then
  curl -fsSL "$URL" -o "$FILENAME"
elif command -v wget > /dev/null 2>&1; then
  wget -q "$URL" -O "$FILENAME"
else
  error "curl or wget is required"
fi

info "Extracting archive..."

# Extract
if [ "$EXT" = "zip" ]; then
  if command -v unzip > /dev/null 2>&1; then
    unzip -q "$FILENAME"
  else
    error "unzip is required to extract .zip files"
  fi
else
  tar xzf "$FILENAME"
fi

# Install
info "Installing to $INSTALL_DIR/$BINARY..."

if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY" "$INSTALL_DIR/$BINARY"
  chmod +x "$INSTALL_DIR/$BINARY"
else
  warn "Need sudo access to install to $INSTALL_DIR"
  sudo mv "$BINARY" "$INSTALL_DIR/$BINARY"
  sudo chmod +x "$INSTALL_DIR/$BINARY"
fi

info "Successfully installed flow v${VERSION}!"
echo ""
echo "Run 'flow --version' to verify installation."

# macOS-specific notice
if [ "$OS" = "darwin" ]; then
  echo ""
  warn "macOS users: If you see 'untrusted developer' warning, run:"
  echo "    xattr -d com.apple.quarantine $INSTALL_DIR/$BINARY"
fi

# Check if in PATH
if ! command -v flow > /dev/null 2>&1; then
  echo ""
  warn "$INSTALL_DIR is not in your PATH"
  echo "Add this to your ~/.bashrc or ~/.zshrc:"
  echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
fi
