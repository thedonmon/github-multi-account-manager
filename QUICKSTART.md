# Quick Start - Testing the Bubbletea TUI

## 🚀 The Fastest Way to Get Started

### For New Users (Clean Machine)

```bash
# 1. Build the binaries
cd ~/SourceCode/personal/github-multi-account-manager
go build -o bin/ghmm ./cmd/ghmm
go build -o bin/ghmm-cli ./cmd/ghmm-cli

# 2. Run the setup wizard - it does EVERYTHING for you!
./bin/ghmm-cli setup

# Follow the prompts to:
# - Add account details
# - Generate SSH keys (automatic!)
# - Get key copied to clipboard
# - Add to GitHub
# - Test connection
# - Add more accounts
```

**That's it!** You'll have a fully working multi-account setup in under 5 minutes! 🎉

### For Existing Users (Has .gitconfig)

```bash
# 1. Build and run the TUI
cd ~/SourceCode/personal/github-multi-account-manager
go build -o bin/ghmm ./cmd/ghmm
./bin/ghmm

# The TUI will:
# - Auto-import accounts from your existing `.gitconfig`
# - Display all your GitHub accounts in a table
# - Show keyboard shortcuts at the bottom
```

## Keyboard Controls

- `q` or `Ctrl+C` - Quit
- `n` - Add new account (interactive form)
- `s` - Auto-sync from existing .gitconfig/.ssh/config
- `r` - Refresh account list
- `a` - Apply configurations to SSH, Git, and Shell
- `c` - Copy SSH key to clipboard
- `enter` - Show account details
- `↑` `↓` - Navigate between accounts

## First Time Setup (if you don't have accounts)

If you don't have a `.gitconfig` with includeIf directives, you can:

**Option 1: Use the TUI (Recommended)**
```bash
# Launch TUI
./bin/ghmm

# The TUI will show:
# - Press 'n' to add an account interactively
# - Press 's' to auto-sync from existing setup
```

**Option 2: Use the CLI**
```bash
# Build the CLI helper
go build -o bin/ghmm-cli ./cmd/ghmm-cli

# Add accounts
./bin/ghmm-cli add work johndoe-work john@work.com ~/code/work
./bin/ghmm-cli add personal johndoe john@personal.com ~/code/personal

# List to verify
./bin/ghmm-cli list

# Launch TUI
./bin/ghmm
```

## What You Should See

```
┌─────────────────────────────────────────────────────────────────┐
│ 📦 GitHub Multi-Account Manager                                │
│                                                                 │
│ ┌──────────────────────────────────────────────────────────┐  │
│ │ Name           Username      Email            Directory  │  │
│ │ ⭐ work         johndoe-work  john@work.com    ~/code/w... │  │
│ │   personal     johndoe       john@personal... ~/code/p... │  │
│ └──────────────────────────────────────────────────────────┘  │
│                                                                 │
│ Ready                                                           │
│ q: quit • r: refresh                                           │
└─────────────────────────────────────────────────────────────────┘
```

## Cleanup (Remove Test Data)

```bash
# Remove the test config
rm -rf ~/.ghmm

# Or remove specific accounts
./bin/ghmm-cli remove work
./bin/ghmm-cli remove personal
```

## Performance Comparison

Want to see how much faster Go is?

```bash
# Go version startup
time ./bin/ghmm  # Press 'q' immediately

# Python version startup (switch branches first)
git checkout main
time python -m ghmm.cli tui  # Press 'q' immediately
```

The Go version should be ~20-30x faster to start!

## Troubleshooting

### "No accounts configured"
- The TUI tried to auto-import from `.gitconfig` but found nothing
- Add accounts manually using `ghmm-cli` (see above)

### Build errors
```bash
# Make sure you have Go installed
go version

# Clean and rebuild
rm -rf bin/
go mod tidy
go build -o bin/ghmm ./cmd/ghmm
```

### Can't find the binary
```bash
# Use the full path
/Users/don/SourceCode/personal/github-multi-account-manager/bin/ghmm
```

## What's Working

### ✅ Fully Implemented
- Display accounts in a beautiful table
- Auto-import from existing `.gitconfig` and `.ssh/config`
- Keyboard navigation
- Config file management (YAML)
- Account add/remove via CLI and TUI
- Interactive account addition form in TUI
- Apply configurations (SSH, Git, Shell)
- Copy SSH key to clipboard
- Show detailed account information
- Auto-sync from existing setup

### 🚀 Future Enhancements
- SSH key generation from TUI
- Test SSH connection to GitHub
- Account editing
