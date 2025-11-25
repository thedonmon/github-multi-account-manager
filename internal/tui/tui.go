package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	AddAccount  key.Binding
	AutoSync    key.Binding
	GenerateKey key.Binding
	TestConn    key.Binding
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
	AddAccount: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "add account"),
	),
	AutoSync: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "auto-sync"),
	),
	GenerateKey: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "generate SSH key"),
	),
	TestConn: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "test connection"),
	),
}

type viewMode int

const (
	viewTable viewMode = iota
	viewAddAccount
	viewDetails
)

type model struct {
	table        table.Model
	config       *config.Config
	sshManager   *ssh.Manager
	gitManager   *git.Manager
	shellManager *shell.Manager
	statusMsg    string
	errorMsg     string
	mode         viewMode
	detailsText  string
	// Form fields for adding account
	formInputs   []textinput.Model
	formFocused  int
	emptyStartup bool // True if started with no accounts
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If showing details, any key dismisses
		if m.mode == viewDetails {
			m.mode = viewTable
			m.detailsText = ""
			return m, nil
		}

		// If in add account mode, handle form navigation
		if m.mode == viewAddAccount {
			return m.handleFormInput(msg)
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

		case key.Matches(msg, keys.AddAccount):
			m = m.startAddAccount()

		case key.Matches(msg, keys.AutoSync):
			m = m.autoSync()

		case key.Matches(msg, keys.GenerateKey):
			m = m.generateSSHKey()

		case key.Matches(msg, keys.TestConn):
			m = m.testConnection()
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.mode {
	case viewDetails:
		return m.renderDetails()
	case viewAddAccount:
		return m.renderAddAccountForm()
	default:
		return m.renderTable()
	}
}

func (m model) renderTable() string {
	title := titleStyle.Render("📦 GitHub Multi-Account Manager")

	var status string
	if m.errorMsg != "" {
		status = errorStyle.Render(m.errorMsg)
	} else if m.statusMsg != "" {
		status = statusStyle.Render(m.statusMsg)
	}

	help := helpStyle.Render(
		"q:quit • n:add • s:sync • g:gen key • t:test • a:apply • c:copy • enter:details",
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
	m.mode = viewDetails
	return m
}

func (m model) startAddAccount() model {
	// Initialize form inputs
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "work, personal, etc."
	inputs[0].Focus()
	inputs[0].CharLimit = 50
	inputs[0].Width = 40
	inputs[0].Prompt = "Account name: "

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "your-github-username"
	inputs[1].CharLimit = 100
	inputs[1].Width = 40
	inputs[1].Prompt = "GitHub username: "

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "you@example.com"
	inputs[2].CharLimit = 100
	inputs[2].Width = 40
	inputs[2].Prompt = "Email: "

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "~/code/work"
	inputs[3].CharLimit = 200
	inputs[3].Width = 40
	inputs[3].Prompt = "Directory: "

	m.formInputs = inputs
	m.formFocused = 0
	m.mode = viewAddAccount
	m.statusMsg = ""
	m.errorMsg = ""

	return m
}

func (m model) handleFormInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = viewTable
		m.formInputs = nil
		return m, nil

	case "tab", "down":
		m.formInputs[m.formFocused].Blur()
		m.formFocused++
		if m.formFocused >= len(m.formInputs) {
			m.formFocused = 0
		}
		m.formInputs[m.formFocused].Focus()
		return m, nil

	case "shift+tab", "up":
		m.formInputs[m.formFocused].Blur()
		m.formFocused--
		if m.formFocused < 0 {
			m.formFocused = len(m.formInputs) - 1
		}
		m.formInputs[m.formFocused].Focus()
		return m, nil

	case "enter":
		// Validate and save
		name := strings.TrimSpace(m.formInputs[0].Value())
		username := strings.TrimSpace(m.formInputs[1].Value())
		email := strings.TrimSpace(m.formInputs[2].Value())
		directory := strings.TrimSpace(m.formInputs[3].Value())

		if name == "" || username == "" || email == "" || directory == "" {
			m.errorMsg = "❌ All fields are required"
			return m, nil
		}

		// Add account (this also saves the config)
		if err := m.config.AddAccount(name, username, email, directory); err != nil {
			m.errorMsg = fmt.Sprintf("❌ Failed to add account: %v", err)
			return m, nil
		}

		m.mode = viewTable
		m.formInputs = nil
		m = m.refreshTable()
		m.statusMsg = fmt.Sprintf("✓ Added account '%s'! Press 'a' to apply configs", name)
		m.errorMsg = ""

		// If this was the first account during startup, set as default
		if m.emptyStartup && len(m.config.ListAccounts()) == 1 {
			m.config.SetDefaultAccount(name)
		}

		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	m.formInputs[m.formFocused], cmd = m.formInputs[m.formFocused].Update(msg)
	return m, cmd
}

func (m model) renderAddAccountForm() string {
	title := titleStyle.Render("Add New Account")

	var form strings.Builder
	form.WriteString("\n")
	for i, input := range m.formInputs {
		form.WriteString(input.View())
		form.WriteString("\n")
		if i < len(m.formInputs)-1 {
			form.WriteString("\n")
		}
	}

	var status string
	if m.errorMsg != "" {
		status = errorStyle.Render(m.errorMsg)
	}

	help := helpStyle.Render("tab/↓:next • shift+tab/↑:prev • enter:save • esc:cancel")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s\n",
		title,
		baseStyle.Render(form.String()),
		status,
		help,
	)
}

func (m model) autoSync() model {
	imported, err := m.config.ImportFromExistingSetup()
	if err != nil {
		m.errorMsg = fmt.Sprintf("❌ Auto-sync failed: %v", err)
		m.statusMsg = ""
		return m
	}

	if imported == 0 {
		m.statusMsg = "No accounts found in .gitconfig or .ssh/config"
		m.errorMsg = ""
	} else {
		m = m.refreshTable()
		m.statusMsg = fmt.Sprintf("✓ Imported %d account(s) from existing setup!", imported)
		m.errorMsg = ""
	}

	return m
}

func (m model) generateSSHKey() model {
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

	// Check if key already exists
	if _, err := os.Stat(account.SSHKeyPath); err == nil {
		m.errorMsg = fmt.Sprintf("⚠️  SSH key already exists for %s", account.Name)
		m.statusMsg = ""
		return m
	}

	// Generate the key
	if err := m.sshManager.GenerateKey(account.SSHKeyPath, account.Email, "ed25519"); err != nil {
		m.errorMsg = fmt.Sprintf("❌ Failed to generate key: %v", err)
		m.statusMsg = ""
		return m
	}

	// Get and copy the public key
	pubKey, err := m.sshManager.GetPublicKey(account.SSHKeyPath)
	if err != nil {
		m.errorMsg = fmt.Sprintf("❌ Failed to read public key: %v", err)
		m.statusMsg = ""
		return m
	}

	if err := clipboard.WriteAll(pubKey); err != nil {
		m.statusMsg = fmt.Sprintf("✓ Key generated for %s! Press 'enter' to see details", account.Name)
	} else {
		m.statusMsg = fmt.Sprintf("✓ SSH key generated and copied! Add it to GitHub, then press 't' to test", account.Name)
	}
	m.errorMsg = ""
	m = m.refreshTable()

	return m
}

func (m model) testConnection() model {
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

	// Check if SSH key exists
	if _, err := os.Stat(account.SSHKeyPath); err != nil {
		m.errorMsg = fmt.Sprintf("❌ No SSH key found for %s. Press 'g' to generate one", account.Name)
		m.statusMsg = ""
		return m
	}

	m.statusMsg = fmt.Sprintf("Testing connection for %s...", account.Name)

	// Test the connection
	success, message := m.sshManager.TestConnection(account.HostAlias)

	if success {
		m.statusMsg = fmt.Sprintf("✓ %s connected successfully!", account.Name)
		m.errorMsg = ""
	} else {
		m.errorMsg = fmt.Sprintf("❌ Connection failed for %s: %s", account.Name, message)
		m.statusMsg = ""
	}

	m = m.refreshTable()
	return m
}

// Run starts the TUI application
func Run() error {
	// Initialize managers
	cfg, err := config.New("")
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	emptyStartup := false

	// Auto-import from existing setup if no accounts configured
	if len(cfg.ListAccounts()) == 0 {
		emptyStartup = true
		if imported, err := cfg.ImportFromExistingSetup(); err == nil && imported > 0 {
			fmt.Printf("✨ Auto-detected %d account(s) from .gitconfig and .ssh/config\n", imported)
			emptyStartup = false
		}
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
		emptyStartup: emptyStartup,
	}

	// Load initial data or show welcome
	if emptyStartup {
		m.mode = viewTable
		m.statusMsg = "👋 Welcome! Press 'n' to add an account or 's' to auto-sync from existing setup"
	} else {
		m.statusMsg = fmt.Sprintf("Loaded %d account(s) • Press 'a' to apply configs", len(cfg.ListAccounts()))
	}

	m = m.refreshTable()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
