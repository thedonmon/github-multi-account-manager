# GitHub Multi-Account Manager (ghmm)

A beautiful TUI tool to manage multiple GitHub accounts on a single machine with ease.

## Features

- 🔐 **Account Management** - Add, remove, and list GitHub accounts
- 🔑 **SSH Key Generation** - Automatic SSH key creation and management
- 📋 **Easy Copy** - One-click copy SSH public keys to clipboard
- 📂 **Directory Mapping** - Map directories to specific GitHub accounts
- 🐚 **Shell Support** - Auto-detects and configures zsh, bash, and fish
- 🚀 **Smart Clone** - Clone repos with the right account automatically
- ✅ **Connection Testing** - Verify SSH connections to GitHub

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/github-multi-account-manager.git
cd github-multi-account-manager

# Install
pip install -e .

# Or use the install script
./install.sh
```

## Usage

```bash
# Launch the TUI
ghmm

# Or use CLI commands
ghmm add-account
ghmm list-accounts
ghmm map-directory
ghmm generate-key
```

## How It Works

ghmm manages:
1. **SSH Config** (`~/.ssh/config`) - Creates host aliases for each account
2. **Git Config** (`~/.gitconfig`) - Sets up directory-based git configurations
3. **Shell Config** - Adds smart clone helper to your shell config
4. **SSH Keys** - Generates and stores keys in `~/.ssh/`

## Example

```bash
# Add your work account
ghmm add-account
> Account name: work
> GitHub username: johndoe-work
> Email: john@company.com
> Directory: ~/code/work

# Add personal account
ghmm add-account
> Account name: personal
> GitHub username: johndoe
> Email: john@personal.com
> Directory: ~/code/personal

# Now clone repos easily
cd ~/code/work
gclone company/repo  # Uses work account automatically

cd ~/code/personal
gclone johndoe/project  # Uses personal account automatically
```

## Requirements

- Python 3.8+
- macOS or Linux
- Git
- SSH

## License

MIT

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.
