#!/bin/bash

# GitHub Multi-Account Manager Installation Script

set -e

echo "🚀 Installing GitHub Multi-Account Manager..."

# Check for Python 3
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is required but not installed."
    echo "Please install Python 3.8 or higher and try again."
    exit 1
fi

# Check Python version
PYTHON_VERSION=$(python3 -c 'import sys; print(".".join(map(str, sys.version_info[:2])))')
REQUIRED_VERSION="3.8"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$PYTHON_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then 
    echo "❌ Python $REQUIRED_VERSION or higher is required (found $PYTHON_VERSION)"
    exit 1
fi

echo "✅ Python $PYTHON_VERSION detected"

# Install using pip
echo "📦 Installing ghmm..."
pip3 install -e .

echo ""
echo "✨ Installation complete!"
echo ""
echo "📚 Quick Start:"
echo "  - Run 'ghmm' to launch the interactive TUI"
echo "  - Run 'ghmm add-account' to add your first account"
echo "  - Run 'ghmm --help' for all commands"
echo ""
echo "🔗 Documentation: https://github.com/thedonmon/github-multi-account-manager"
