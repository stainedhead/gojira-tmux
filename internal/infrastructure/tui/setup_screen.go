package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// SetupScreen handles the first-time API token setup flow.
type SetupScreen struct {
	tokenStore domain.TokenStorePort
	input      textinput.Model
	err        error
	submitted  bool
	keys       SetupKeyMap
}

// NewSetupScreenModel creates a new setup screen.
func NewSetupScreenModel(tokenStore domain.TokenStorePort) *SetupScreen {
	ti := textinput.New()
	ti.Placeholder = "Paste your Jira API token here"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &SetupScreen{
		tokenStore: tokenStore,
		input:      ti,
		keys:       DefaultSetupKeyMap(),
	}
}

// Init initializes the setup screen.
func (s *SetupScreen) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the setup screen.
func (s *SetupScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return s.submit()
		case "esc":
			return s, tea.Quit
		case "ctrl+c":
			return s, tea.Quit
		}

	case tokenSavedMsg:
		return s, func() tea.Msg { return TokenStoredMsg{} }

	case tokenSaveErrorMsg:
		s.err = msg.err
		s.submitted = false
		return s, nil
	}

	// Update text input
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

// submit attempts to save the token.
func (s *SetupScreen) submit() (tea.Model, tea.Cmd) {
	token := strings.TrimSpace(s.input.Value())
	if token == "" {
		s.err = nil // Clear error, just don't submit
		return s, nil
	}

	s.submitted = true
	s.err = nil

	return s, func() tea.Msg {
		err := s.tokenStore.SetJiraToken(token)
		if err != nil {
			return tokenSaveErrorMsg{err: err}
		}
		return tokenSavedMsg{}
	}
}

// tokenSavedMsg indicates the token was saved successfully.
type tokenSavedMsg struct{}

// tokenSaveErrorMsg indicates an error saving the token.
type tokenSaveErrorMsg struct {
	err error
}

// View renders the setup screen.
func (s *SetupScreen) View() string {
	var b strings.Builder

	// Title
	title := Styles.Title.Render("🔐 Jira API Token Setup")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instructions
	instructions := Styles.Paragraph.Render(
		"To use gojira-tmux, you need to provide your Jira API token.\n\n" +
			"You can generate a token at:\n" +
			"https://id.atlassian.com/manage-profile/security/api-tokens")
	b.WriteString(instructions)
	b.WriteString("\n\n")

	// Input field
	label := Styles.InputLabel.Render("API Token:")
	b.WriteString(label)
	b.WriteString("\n")

	inputStyle := Styles.Input
	if s.input.Focused() {
		inputStyle = Styles.InputFocused
	}
	b.WriteString(inputStyle.Render(s.input.View()))
	b.WriteString("\n\n")

	// Error message
	if s.err != nil {
		errMsg := Styles.Error.Render("Error: " + s.err.Error())
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	// Status or help
	if s.submitted && s.err == nil {
		status := Styles.Success.Render("Saving token...")
		b.WriteString(status)
	} else {
		help := lipgloss.JoinHorizontal(lipgloss.Top,
			Styles.HelpKey.Render("enter"),
			Styles.HelpDesc.Render(" submit  "),
			Styles.HelpKey.Render("esc"),
			Styles.HelpDesc.Render(" quit"),
		)
		b.WriteString(help)
	}

	return Styles.App.Render(b.String())
}
