package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectExistingSetup detects accounts from both .gitconfig and .ssh/config
func DetectExistingSetup() ([]Account, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// First, get SSH config mappings (host alias -> SSH key path)
	sshMap := parseSSHConfig(filepath.Join(home, ".ssh", "config"))

	// Then get git config mappings (directory -> email/name)
	gitAccounts, err := parseGitconfigForAccounts(filepath.Join(home, ".gitconfig"), home)
	if err != nil {
		return nil, err
	}

	// Merge the information
	for i := range gitAccounts {
		// Try to find matching SSH config by looking for similar host alias or SSH key
		for hostAlias, sshKey := range sshMap {
			accountName := gitAccounts[i].Name
			username := strings.ToLower(gitAccounts[i].Username)
			hostSuffix := strings.TrimPrefix(hostAlias, "github.com-")
			sshKeyLower := strings.ToLower(sshKey)

			// Match strategies (in order of preference):
			// 1. GitHub username is in the host alias (e.g., "thedonmon" in "github.com-thedonmon")
			// 2. Host alias contains account name (e.g., "work" in "github.com-work")
			// 3. SSH key path contains account name or username
			// 4. Account name is abbreviated in host (e.g., "coldcart" -> "cc")
			if strings.Contains(strings.ToLower(hostAlias), username) ||
				strings.Contains(hostAlias, accountName) ||
				strings.Contains(accountName, hostSuffix) ||
				strings.Contains(sshKeyLower, accountName) ||
				strings.Contains(sshKeyLower, username) ||
				(len(accountName) > 2 && strings.HasPrefix(accountName, hostSuffix)) {
				gitAccounts[i].HostAlias = hostAlias
				gitAccounts[i].SSHKeyPath = sshKey
				break
			}
		}
	}

	return gitAccounts, nil
}

// parseSSHConfig extracts GitHub host aliases and their SSH keys from ~/.ssh/config
func parseSSHConfig(configPath string) map[string]string {
	result := make(map[string]string)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return result
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var currentHost string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Match: Host github.com-something
		if strings.HasPrefix(line, "Host ") {
			host := strings.TrimPrefix(line, "Host ")
			host = strings.TrimSpace(host)
			if strings.HasPrefix(host, "github.com-") {
				currentHost = host
			}
		}

		// Match: IdentityFile ~/.ssh/something_ssh
		if currentHost != "" && strings.Contains(line, "IdentityFile") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				identityFile := parts[1]
				// Expand ~
				if strings.HasPrefix(identityFile, "~/") {
					home, _ := os.UserHomeDir()
					identityFile = filepath.Join(home, identityFile[2:])
				}
				result[currentHost] = identityFile
				currentHost = ""
			}
		}
	}

	return result
}

// parseGitconfigForAccounts extracts account info from gitconfig includeIf directives
func parseGitconfigForAccounts(configPath, home string) ([]Account, error) {
	var accounts []Account

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .gitconfig: %w", err)
	}

	// Regex to match: [includeIf "gitdir:/path/to/dir/**"]
	includeIfRe := regexp.MustCompile(`\[includeIf\s+"gitdir:([^"]+)"\]`)
	pathRe := regexp.MustCompile(`path\s*=\s*(.+)`)

	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	var currentDir string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for includeIf directive
		if matches := includeIfRe.FindStringSubmatch(line); matches != nil {
			gitdir := matches[1]
			// Remove trailing /** or **
			gitdir = strings.TrimSuffix(gitdir, "/**")
			gitdir = strings.TrimSuffix(gitdir, "**")
			gitdir = strings.TrimSuffix(gitdir, "/")
			// Expand ~
			if strings.HasPrefix(gitdir, "~/") {
				gitdir = filepath.Join(home, gitdir[2:])
			}
			currentDir = gitdir
		}

		// Check for path directive
		if currentDir != "" && strings.Contains(line, "path") {
			if matches := pathRe.FindStringSubmatch(line); matches != nil {
				currentPath := strings.TrimSpace(matches[1])
				// Expand ~
				if strings.HasPrefix(currentPath, "~/") {
					currentPath = filepath.Join(home, currentPath[2:])
				}

				// Try to parse the included config file
				if account, err := parseIncludedGitconfig(currentPath, currentDir, home); err == nil {
					accounts = append(accounts, account)
				}

				currentDir = ""
			}
		}
	}

	return accounts, nil
}

// parseIncludedGitconfig reads a gitconfig-* file and extracts user info
func parseIncludedGitconfig(configPath, directory, home string) (Account, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Account{}, err
	}

	var name, email string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "name") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
			}
		}

		if strings.Contains(line, "email") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				email = strings.TrimSpace(parts[1])
			}
		}
	}

	if name == "" || email == "" {
		return Account{}, fmt.Errorf("incomplete account info in %s", configPath)
	}

	// Extract account name from config filename or directory
	accountName := extractAccountName(configPath, home)

	// Try to extract GitHub username from email if it's a noreply address
	// Format: 12345+username@users.noreply.github.com
	githubUsername := name
	if strings.Contains(email, "@users.noreply.github.com") {
		parts := strings.Split(email, "@")
		if len(parts) > 0 {
			userPart := parts[0]
			if plusIdx := strings.Index(userPart, "+"); plusIdx != -1 {
				githubUsername = userPart[plusIdx+1:]
			}
		}
	}

	return Account{
		Name:       accountName,
		Username:   githubUsername,
		Email:      email,
		Directory:  directory,
		SSHKeyPath: "", // Will be filled from SSH config
		HostAlias:  "", // Will be filled from SSH config
	}, nil
}

// extractAccountName determines account name from config path
func extractAccountName(configPath, home string) string {
	baseName := filepath.Base(configPath)

	// Format: ~/.gitconfig-work -> work
	if strings.HasPrefix(baseName, ".gitconfig-") {
		return strings.TrimPrefix(baseName, ".gitconfig-")
	}

	// Format: ~/SourceCode/work/.gitconfig -> work (from parent dir)
	if baseName == ".gitconfig" {
		parentDir := filepath.Base(filepath.Dir(configPath))
		if parentDir != "" && parentDir != "." && parentDir != filepath.Base(home) {
			return parentDir
		}
	}

	return baseName
}

// ImportFromExistingSetup imports detected accounts into the config
func (c *Config) ImportFromExistingSetup() (int, error) {
	detected, err := DetectExistingSetup()
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
			// Import account even without SSH config - we can generate it later
			// Set default SSH key path if not found
			if account.SSHKeyPath == "" {
				home, _ := os.UserHomeDir()
				account.SSHKeyPath = filepath.Join(home, ".ssh", account.Name+"_ssh")
			}
			if account.HostAlias == "" {
				account.HostAlias = "github.com-" + account.Name
			}

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
