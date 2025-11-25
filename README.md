# GitHub Multi-Account Manager (ghmm)

A beautiful TUI tool to manage multiple GitHub accounts on a single machine with ease.

## Features

- 🎨 **Beautiful TUI** - Interactive terminal UI built with Bubble Tea
- 🔐 **Account Management** - Add, remove, and list GitHub accounts
- 🔑 **SSH Key Generation** - Automatic SSH key creation and management
- 📋 **Easy Copy** - One-click copy SSH public keys to clipboard
- 📂 **Directory Mapping** - Map directories to specific GitHub accounts
- 🐚 **Shell Support** - Auto-detects and configures zsh, bash, and fish
- 🚀 **Smart Clone** - Clone repos with the right account automatically
- 🔍 **Auto-detection** - Imports existing accounts from .gitconfig and .ssh/config

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/github-multi-account-manager.git
cd github-multi-account-manager

# Install
./install.sh

# Or build manually
go build -o bin/ghmm ./cmd/ghmm
go build -o bin/ghmm-cli ./cmd/ghmm-cli
```

## Quick Start

```bash
# First time? Run the setup wizard!
ghmm-cli setup

# That's it! The wizard will:
# 1. Ask for your account details
# 2. Generate SSH keys
# 3. Copy the key to clipboard
# 4. Wait for you to add it to GitHub
# 5. Test the connection
# 6. Repeat for multiple accounts
```

## Usage

### Interactive TUI
```bash
# Launch the beautiful TUI
ghmm

# TUI Features:
# - n: Add new account
# - g: Generate SSH key
# - t: Test connection
# - a: Apply configs
# - c: Copy SSH key
# - s: Auto-sync from existing setup
```

### CLI Commands
```bash
# Manual setup (if you prefer)
ghmm-cli add work user@work.com ~/code/work
ghmm-cli generate-key work
ghmm-cli test work
ghmm-cli set-default work
```

## How It Works

ghmm manages:
1. **SSH Config** (`~/.ssh/config`) - Creates host aliases for each account
2. **Git Config** (`~/.gitconfig`) - Sets up directory-based git configurations using includeIf
3. **Shell Config** - Adds smart clone helper to your shell config
4. **SSH Keys** - Manages keys in `~/.ssh/`

## Example

```bash
# Add your work account
ghmm-cli add work johndoe-work john@company.com ~/code/work

# Add personal account
ghmm-cli add personal johndoe john@personal.com ~/code/personal

# Set default account
ghmm-cli set-default work

# Launch TUI to see all accounts and apply configs
ghmm

# Now clone repos easily
cd ~/code/work
gclone company/repo  # Uses work account automatically

cd ~/code/personal
gclone johndoe/project  # Uses personal account automatically
```

## Requirements

- Go 1.21+
- macOS or Linux
- Git
- SSH

## License

MIT

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.
