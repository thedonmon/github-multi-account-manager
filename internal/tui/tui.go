package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donbowman/github-multi-account-manager/internal/config"
	"github.com/donbowman/github-multi-account-manager/internal/git"
	"github.com/donbowman/github-multi-account-manager/internal/shell"
	"github.com/donbowman/github-multi-account-manager/internal/ssh"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("green")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("red")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("cyan"))

	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)

type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Quit        key.Binding
	Refresh     key.Binding
	Apply       key.Binding
	CopyKey     key.Binding
	ShowDetails key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Apply: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply configs"),
	),
	CopyKey: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy SSH key"),
	),
	ShowDetails: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("enter", "details"),
	),
}

type model struct {
	table          table.Model
	config         *config.Config
	sshManager     *ssh.Manager
	gitManager     *git.Manager
	shellManager   *shell.Manager
	statusMsg      string
	errorMsg       string
	showingDetails bool
	detailsText    string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If showing details, any key dismisses
		if m.showingDetails {
			m.showingDetails = false
			m.detailsText = ""
			return m, nil
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Refresh):
			m = m.refreshTable()
			m.statusMsg = "✓ Refreshed"
			m.errorMsg = ""

		case key.Matches(msg, keys.Apply):
			m = m.applyConfigs()

		case key.Matches(msg, keys.CopyKey):
			m = m.copySSHKey()

		case key.Matches(msg, keys.ShowDetails):
			m = m.showDetails()
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.showingDetails {
		return m.renderDetails()
	}

	title := titleStyle.Render("📦 GitHub Multi-Account Manager")

	var status string
	if m.errorMsg != "" {
		status = errorStyle.Render(m.errorMsg)
	} else if m.statusMsg != "" {
		status = statusStyle.Render(m.statusMsg)
	}

	help := helpStyle.Render(
		"q:quit • r:refresh • a:apply configs • c:copy SSH key • enter:details",
	)

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n",
		title,
		baseStyle.Render(m.table.View()),
		status,
		help,
	)
}

func (m model) renderDetails() string {
	title := titleStyle.Render("Account Details")
	content := baseStyle.Render(m.detailsText)
	help := helpStyle.Render("Press any key to return")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n", title, content, help)
}

func (m model) refreshTable() model {
	accounts := m.config.ListAccounts()
	defaultAcc := m.config.GetDefaultAccount()

	rows := []table.Row{}
	for _, acc := range accounts {
		name := acc.Name
		if name == defaultAcc {
			name = "⭐ " + name
		}

		// Check if SSH key exists
		status := "⚠️ No key"
		if _, err := os.Stat(acc.SSHKeyPath); err == nil {
			status = "✅ Ready"
		}

		rows = append(rows, table.Row{
			name,
			acc.Username,
			acc.Email,
			acc.Directory,
			status,
		})
	}

	m.table.SetRows(rows)
	return m
}

func (m model) applyConfigs() model {
	accounts := m.config.ListAccounts()
	defaultAcc := m.config.GetDefaultAccount()

	// Update SSH config
	if err := m.sshManager.UpdateSSHConfig(accounts); err != nil {
		m.errorMsg = fmt.Sprintf("❌ SSH config failed: %v", err)
		m.statusMsg = ""
		return m
	}

	// Update Git config
	if err := m.gitManager.UpdateGitconfig(accounts); err != nil {
		m.errorMsg = fmt.Sprintf("❌ Git config failed: %v", err)
		m.statusMsg = ""
		return m
	}

	// Create individual gitconfigs
	for _, acc := range accounts {
		if err := m.gitManager.CreateAccountGitconfig(acc); err != nil {
			m.errorMsg = fmt.Sprintf("❌ Failed to create gitconfig for %s: %v", acc.Name, err)
			m.statusMsg = ""
			return m
		}
	}

	// Update shell config
	if err := m.shellManager.UpdateShellConfig(accounts, defaultAcc); err != nil {
		m.errorMsg = fmt.Sprintf("❌ Shell config failed: %v", err)
		m.statusMsg = ""
		return m
	}

	m.statusMsg = fmt.Sprintf("✓ Configs applied! Reload shell: %s", m.shellManager.GetReloadCommand())
	m.errorMsg = ""
	return m
}

func (m model) copySSHKey() model {
	if m.table.Cursor() < 0 {
		m.errorMsg = "❌ No account selected"
		m.statusMsg = ""
		return m
	}

	accounts := m.config.ListAccounts()
	if m.table.Cursor() >= len(accounts) {
		return m
	}

	account := accounts[m.table.Cursor()]
	pubKey, err := m.sshManager.GetPublicKey(account.SSHKeyPath)
	if err != nil {
		m.errorMsg = fmt.Sprintf("❌ No SSH key found for %s", account.Name)
		m.statusMsg = ""
		return m
	}

	if err := clipboard.WriteAll(pubKey); err != nil {
		m.errorMsg = "❌ Failed to copy to clipboard"
		m.statusMsg = ""
		return m
	}

	m.statusMsg = fmt.Sprintf("✓ SSH key for %s copied to clipboard!", account.Name)
	m.errorMsg = ""
	return m
}

func (m model) showDetails() model {
	if m.table.Cursor() < 0 {
		return m
	}

	accounts := m.config.ListAccounts()
	if m.table.Cursor() >= len(accounts) {
		return m
	}

	account := accounts[m.table.Cursor()]

	var details strings.Builder
	details.WriteString(fmt.Sprintf("Account: %s\n\n", infoStyle.Render(account.Name)))
	details.WriteString(fmt.Sprintf("Username:     %s\n", account.Username))
	details.WriteString(fmt.Sprintf("Email:        %s\n", account.Email))
	details.WriteString(fmt.Sprintf("Directory:    %s\n", account.Directory))
	details.WriteString(fmt.Sprintf("SSH Key:      %s\n", account.SSHKeyPath))
	details.WriteString(fmt.Sprintf("Host Alias:   %s\n\n", account.HostAlias))

	// Check SSH key status
	if _, err := os.Stat(account.SSHKeyPath); err == nil {
		pubKey, err := m.sshManager.GetPublicKey(account.SSHKeyPath)
		if err == nil {
			details.WriteString(statusStyle.Render("✓ SSH Key exists\n\n"))
			details.WriteString("Public Key:\n")
			details.WriteString(infoStyle.Render(pubKey))
			details.WriteString("\n\nPress 'c' in main view to copy to clipboard")
		}
	} else {
		details.WriteString(errorStyle.Render("⚠ No SSH key found\n"))
		details.WriteString("Generate with: ssh-keygen -t ed25519 -C \"" + account.Email + "\" -f " + account.SSHKeyPath)
	}

	m.detailsText = details.String()
	m.showingDetails = true
	return m
}

// Run starts the TUI application
func Run() error {
	// Initialize managers
	cfg, err := config.New("")
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Auto-import from .gitconfig if no accounts configured
	if len(cfg.ListAccounts()) == 0 {
		if imported, err := cfg.ImportFromGitconfig(); err == nil && imported > 0 {
			fmt.Printf("✨ Auto-imported %d account(s) from .gitconfig\n", imported)
		}
	}

	// Still no accounts? Show helpful message
	if len(cfg.ListAccounts()) == 0 {
		fmt.Println("👋 No accounts configured yet!")
		fmt.Println("\nTo get started:")
		fmt.Println("  1. Configure includeIf in your .gitconfig, or")
		fmt.Println("  2. Use ghmm-cli to add accounts:")
		fmt.Println("\n     go build -o bin/ghmm-cli ./cmd/ghmm-cli")
		fmt.Println("     ./bin/ghmm-cli add work user@work.com ~/code/work")
		return nil
	}

	sshMgr, err := ssh.New()
	if err != nil {
		return fmt.Errorf("failed to initialize SSH manager: %w", err)
	}

	gitMgr, err := git.New()
	if err != nil {
		return fmt.Errorf("failed to initialize Git manager: %w", err)
	}

	shellMgr, err := shell.New()
	if err != nil {
		return fmt.Errorf("failed to initialize shell manager: %w", err)
	}

	// Create table
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Username", Width: 20},
		{Title: "Email", Width: 35},
		{Title: "Directory", Width: 30},
		{Title: "Status", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table:        t,
		config:       cfg,
		sshManager:   sshMgr,
		gitManager:   gitMgr,
		shellManager: shellMgr,
		statusMsg:    fmt.Sprintf("Loaded %d account(s) • Press 'a' to apply configs", len(cfg.ListAccounts())),
	}

	// Load initial data
	m = m.refreshTable()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
