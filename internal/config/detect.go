package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectFromGitconfig attempts to detect existing account configurations from .gitconfig
func DetectFromGitconfig() ([]Account, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	gitconfig := filepath.Join(home, ".gitconfig")
	data, err := os.ReadFile(gitconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read .gitconfig: %w", err)
	}

	return parseGitconfigForAccounts(string(data), home)
}

// parseGitconfigForAccounts extracts account info from gitconfig includeIf directives
func parseGitconfigForAccounts(content, home string) ([]Account, error) {
	var accounts []Account

	// Regex to match: [includeIf "gitdir:/path/to/dir/**"]
	includeIfRe := regexp.MustCompile(`\[includeIf\s+"gitdir:([^"]+)"\]`)
	pathRe := regexp.MustCompile(`path\s*=\s*(.+)`)

	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentDir string
	var currentPath string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for includeIf directive
		if matches := includeIfRe.FindStringSubmatch(line); matches != nil {
			gitdir := matches[1]
			// Remove trailing /** or **
			gitdir = strings.TrimSuffix(gitdir, "/**")
			gitdir = strings.TrimSuffix(gitdir, "**")
			gitdir = strings.TrimSuffix(gitdir, "/")
			currentDir = gitdir
		}

		// Check for path directive
		if currentDir != "" && strings.Contains(line, "path") {
			if matches := pathRe.FindStringSubmatch(line); matches != nil {
				currentPath = strings.TrimSpace(matches[1])
				// Expand ~
				if strings.HasPrefix(currentPath, "~/") {
					currentPath = filepath.Join(home, currentPath[2:])
				}

				// Try to parse the included config file
				if account, err := parseIncludedGitconfig(currentPath, currentDir); err == nil {
					accounts = append(accounts, account)
				}

				currentDir = ""
				currentPath = ""
			}
		}
	}

	return accounts, nil
}

// parseIncludedGitconfig reads a gitconfig-* file and extracts user info
func parseIncludedGitconfig(configPath, directory string) (Account, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Account{}, err
	}

	var name, email string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
			}
		}

		if strings.Contains(line, "email") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				email = strings.TrimSpace(parts[1])
			}
		}
	}

	if name == "" || email == "" {
		return Account{}, fmt.Errorf("incomplete account info in %s", configPath)
	}

	home, _ := os.UserHomeDir()

	// Extract account name from config filename
	// Handle formats:
	// - ~/.gitconfig-work -> work
	// - ~/SourceCode/work/.gitconfig -> work (from parent dir)
	baseName := filepath.Base(configPath)
	accountName := strings.TrimPrefix(baseName, ".gitconfig-")

	// If no prefix was trimmed (still ".gitconfig"), use parent directory name
	if accountName == ".gitconfig" || accountName == baseName {
		parentDir := filepath.Base(filepath.Dir(configPath))
		if parentDir != "" && parentDir != "." && parentDir != home {
			accountName = parentDir
		}
	}

	return Account{
		Name:       accountName,
		Username:   name,
		Email:      email,
		Directory:  directory,
		SSHKeyPath: filepath.Join(home, ".ssh", accountName+"_ssh"),
		HostAlias:  "github.com-" + accountName,
	}, nil
}

// ImportFromGitconfig imports detected accounts into the config
func (c *Config) ImportFromGitconfig() (int, error) {
	detected, err := DetectFromGitconfig()
	if err != nil {
		return 0, err
	}

	imported := 0
	for _, account := range detected {
		// Check if already exists
		exists := false
		for _, existing := range c.Accounts {
			if existing.Name == account.Name {
				exists = true
				break
			}
		}

		if !exists {
			c.Accounts = append(c.Accounts, account)
			imported++
		}
	}

	// Set first as default if none set
	if c.DefaultAccount == "" && len(c.Accounts) > 0 {
		c.DefaultAccount = c.Accounts[0].Name
	}

	if imported > 0 {
		if err := c.save(); err != nil {
			return imported, err
		}
	}

	return imported, nil
}
