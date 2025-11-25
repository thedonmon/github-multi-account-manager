# Quick Start - Testing the Bubbletea TUI

## Running Outside of the Session

### 1. Navigate to the project
```bash
cd ~/SourceCode/personal/github-multi-account-manager
```

### 2. Ensure you're on the right branch
```bash
git checkout bubbletea-conversion
```

### 3. Build the binary
```bash
go build -o bin/ghmm ./cmd/ghmm
```

### 4. Run the TUI
```bash
./bin/ghmm
```

That's it! The TUI will:
- Auto-import accounts from your existing `.gitconfig` (if configured)
- Display all your GitHub accounts in a table
- Show keyboard shortcuts at the bottom

## Keyboard Controls

- `q` or `Ctrl+C` - Quit
- `r` - Refresh account list
- `↑` `↓` - Navigate between accounts

## First Time Setup (if you don't have accounts)

If you don't have a `.gitconfig` with includeIf directives, you can add test accounts:

```bash
# Build the CLI helper (optional)
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

## What's Working vs Not Yet Implemented

### ✅ Working
- Display accounts in a table
- Auto-import from existing `.gitconfig`
- Keyboard navigation
- Config file management (YAML)
- Account add/remove via CLI

### ⏳ TODO
- Add accounts from TUI (modal dialog)
- SSH key generation button
- Test connection button
- Apply configurations (SSH, Git, Shell)
- Copy SSH key to clipboard
