# Testing the Bubbletea Version

## Quick Start

### 1. Build the binaries

```bash
cd /Users/don/SourceCode/personal/github-multi-account-manager
go build -o bin/ghmm ./cmd/ghmm
go build -o bin/ghmm-cli ./cmd/ghmm-cli
```

### 2. Add some test accounts

```bash
# Add a work account
./bin/ghmm-cli add work johndoe-work john@company.com ~/code/work

# Add a personal account
./bin/ghmm-cli add personal johndoe john@gmail.com ~/code/personal

# List accounts
./bin/ghmm-cli list
```

### 3. Launch the TUI

```bash
./bin/ghmm
```

## TUI Keyboard Shortcuts

- `q` or `Ctrl+C` - Quit
- `r` - Refresh the account list
- Arrow keys - Navigate the table

## What You'll See

The TUI displays a table with:
- **Name** - Account name (⭐ indicates default)
- **Username** - GitHub username
- **Email** - Git commit email
- **Directory** - Directory mapped to this account
- **Status** - Account status (SSH key presence, etc.)

## Testing Features

### Test the Config Manager

```bash
# View the config file directly
cat ~/.ghmm/config.yaml

# Add multiple accounts
./bin/ghmm-cli add test1 user1 user1@test.com ~/test1
./bin/ghmm-cli add test2 user2 user2@test.com ~/test2
./bin/ghmm-cli list

# Remove an account
./bin/ghmm-cli remove test1
./bin/ghmm-cli list
```

### Check Generated Files

The Go version creates the same files as Python:

```bash
# Config file
cat ~/.ghmm/config.yaml

# SSH config (if you run apply-config - not yet implemented in Go)
# cat ~/.ssh/config

# Git config (if you run apply-config - not yet implemented in Go)
# cat ~/.gitconfig
```

## Comparing with Python Version

### Python version:
```bash
# Switch back to main branch to test
git checkout main
python -m ghmm.cli tui
```

### Go version:
```bash
git checkout bubbletea-conversion
./bin/ghmm
```

## What's Working

✅ Config management (YAML read/write)
✅ Account CRUD operations
✅ Basic TUI with table view
✅ Keyboard navigation
✅ Account listing

## What's Not Yet Implemented

⏳ SSH key generation from TUI
⏳ Modal dialogs for adding accounts
⏳ Apply configuration (SSH, Git, Shell)
⏳ Test connection button
⏳ Copy SSH key to clipboard

## Performance Test

```bash
# Measure startup time
time ./bin/ghmm  # Press q immediately

# Compare with Python
time python -m ghmm.cli tui  # Press q immediately
```

You should see the Go version starts ~20-30x faster!

## Cleanup

```bash
# Remove test accounts
./bin/ghmm-cli remove work
./bin/ghmm-cli remove personal

# Or delete the config entirely
rm -rf ~/.ghmm
```
