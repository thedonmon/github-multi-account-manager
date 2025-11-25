#!/bin/bash

# GitHub Multi-Account Manager Installation Script

set -e

echo "🚀 Installing GitHub Multi-Account Manager..."

# Check for Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is required but not installed."
    echo "Please install Go 1.21 or higher and try again."
    echo "Visit: https://go.dev/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

echo "✅ Go $GO_VERSION detected"

# Build binaries
echo "📦 Building ghmm..."
mkdir -p bin
go build -o bin/ghmm ./cmd/ghmm
go build -o bin/ghmm-cli ./cmd/ghmm-cli

# Optional: Move to PATH
read -p "Install to /usr/local/bin? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    sudo mv bin/ghmm /usr/local/bin/
    sudo mv bin/ghmm-cli /usr/local/bin/
    echo "✅ Installed to /usr/local/bin/"
else
    echo "📂 Binaries are in ./bin/"
    echo "   Add to PATH or run directly: ./bin/ghmm"
fi

echo ""
echo "✨ Installation complete!"
echo ""
echo "📚 Quick Start:"
echo "  - Run 'ghmm' to launch the interactive TUI"
echo "  - Run 'ghmm-cli add <name> <username> <email> <directory>' to add accounts"
echo "  - Run 'ghmm-cli --help' for all commands"
echo ""
echo "🔗 Documentation: https://github.com/thedonmon/github-multi-account-manager"
