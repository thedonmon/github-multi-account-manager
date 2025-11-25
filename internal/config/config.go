package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Account represents a GitHub account configuration
type Account struct {
	Name       string `yaml:"name"`
	Username   string `yaml:"username"`
	Email      string `yaml:"email"`
	Directory  string `yaml:"directory"`
	SSHKeyPath string `yaml:"ssh_key_path"`
	HostAlias  string `yaml:"host_alias"`
}

// Config manages ghmm configuration
type Config struct {
	Accounts       []Account `yaml:"accounts"`
	DefaultAccount string    `yaml:"default_account,omitempty"`
	configDir      string
	configFile     string
}

// New creates a new Config instance
func New(configDir string) (*Config, error) {
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(home, ".ghmm")
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	c := &Config{
		configDir:  configDir,
		configFile: filepath.Join(configDir, "config.yaml"),
	}

	if err := c.load(); err != nil {
		return nil, err
	}

	return c, nil
}

// load reads the configuration from file
func (c *Config) load() error {
	data, err := os.ReadFile(c.configFile)
	if os.IsNotExist(err) {
		// Initialize with empty config
		c.Accounts = []Account{}
		c.DefaultAccount = ""
		return c.save()
	}
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// save writes the configuration to file
func (c *Config) save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddAccount adds a new GitHub account
func (c *Config) AddAccount(name, username, email, directory string) error {
	// Check if account already exists
	for _, acc := range c.Accounts {
		if acc.Name == name {
			return fmt.Errorf("account '%s' already exists", name)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Expand ~ in directory path
	if directory[:2] == "~/" {
		directory = filepath.Join(home, directory[2:])
	}

	account := Account{
		Name:       name,
		Username:   username,
		Email:      email,
		Directory:  directory,
		SSHKeyPath: filepath.Join(home, ".ssh", name+"_ssh"),
		HostAlias:  "github.com-" + name,
	}

	c.Accounts = append(c.Accounts, account)

	// Set as default if it's the first account
	if len(c.Accounts) == 1 {
		c.DefaultAccount = name
	}

	return c.save()
}

// RemoveAccount removes a GitHub account
func (c *Config) RemoveAccount(name string) error {
	originalLen := len(c.Accounts)
	newAccounts := make([]Account, 0, len(c.Accounts))

	for _, acc := range c.Accounts {
		if acc.Name != name {
			newAccounts = append(newAccounts, acc)
		}
	}

	if len(newAccounts) == originalLen {
		return fmt.Errorf("account '%s' not found", name)
	}

	c.Accounts = newAccounts

	// Reset default if we removed it
	if c.DefaultAccount == name {
		if len(c.Accounts) > 0 {
			c.DefaultAccount = c.Accounts[0].Name
		} else {
			c.DefaultAccount = ""
		}
	}

	return c.save()
}

// GetAccount retrieves an account by name
func (c *Config) GetAccount(name string) (*Account, error) {
	for _, acc := range c.Accounts {
		if acc.Name == name {
			return &acc, nil
		}
	}
	return nil, fmt.Errorf("account '%s' not found", name)
}

// ListAccounts returns all accounts
func (c *Config) ListAccounts() []Account {
	return c.Accounts
}

// SetDefaultAccount sets the default account
func (c *Config) SetDefaultAccount(name string) error {
	for _, acc := range c.Accounts {
		if acc.Name == name {
			c.DefaultAccount = name
			return c.save()
		}
	}
	return fmt.Errorf("account '%s' not found", name)
}

// GetDefaultAccount returns the default account name
func (c *Config) GetDefaultAccount() string {
	return c.DefaultAccount
}

// ConfigFile returns the path to the config file
func (c *Config) ConfigFile() string {
	return c.configFile
}
