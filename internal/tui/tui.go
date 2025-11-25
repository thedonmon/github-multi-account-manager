package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/donbowman/github-multi-account-manager/internal/config"
	"github.com/donbowman/github-multi-account-manager/internal/git"
	"github.com/donbowman/github-multi-account-manager/internal/ssh"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table      table.Model
	config     *config.Config
	sshManager *ssh.Manager
	gitManager *git.Manager
	statusMsg  string
	err        error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m = m.refreshTable()
			m.statusMsg = "Refreshed"
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(m.statusMsg)

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("\nq: quit • r: refresh")

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("📦 GitHub Multi-Account Manager\n\n")

	return title + baseStyle.Render(m.table.View()) + "\n" + statusBar + help + "\n"
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

		status := "✅ Ready"
		// Could check if SSH key exists here

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

// Run starts the TUI application
func Run() error {
	// Initialize managers
	cfg, err := config.New("")
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	sshMgr, err := ssh.New()
	if err != nil {
		return fmt.Errorf("failed to initialize SSH manager: %w", err)
	}

	gitMgr, err := git.New()
	if err != nil {
		return fmt.Errorf("failed to initialize Git manager: %w", err)
	}

	// Create table
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Username", Width: 20},
		{Title: "Email", Width: 30},
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
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table:      t,
		config:     cfg,
		sshManager: sshMgr,
		gitManager: gitMgr,
		statusMsg:  "Ready",
	}

	// Load initial data
	m = m.refreshTable()

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
