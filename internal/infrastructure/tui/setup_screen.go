package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// SetupScreen handles the first-time API token setup flow.
type SetupScreen struct {
	tokenStore domain.TokenStorePort
	authPort   domain.AuthPort
	emailInput textinput.Model
	tokenInput textinput.Model
	focusIdx   int // 0 = email, 1 = token
	err        error
	submitted  bool
	keys       SetupKeyMap
}

// NewSetupScreenModel creates a new setup screen.
func NewSetupScreenModel(tokenStore domain.TokenStorePort, authPort domain.AuthPort) *SetupScreen {
	ei := textinput.New()
	ei.Placeholder = "your-email@company.com"
	ei.Focus()
	ei.CharLimit = 256
	ei.Width = 50

	ti := textinput.New()
	ti.Placeholder = "Paste your Jira API token here"
	ti.CharLimit = 256
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &SetupScreen{
		tokenStore: tokenStore,
		authPort:   authPort,
		emailInput: ei,
		tokenInput: ti,
		focusIdx:   0,
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
			if s.focusIdx == 0 {
				// Move to token field
				s.focusIdx = 1
				s.emailInput.Blur()
				s.tokenInput.Focus()
				return s, textinput.Blink
			}
			return s.submit()
		case "tab":
			return s.nextField()
		case "shift+tab":
			return s.prevField()
		case "esc":
			return s, tea.Quit
		case "ctrl+c":
			return s, tea.Quit
		}

	case tokenValidatedMsg:
		return s, func() tea.Msg { return TokenStoredMsg{} }

	case tokenValidationErrorMsg:
		s.err = msg.err
		s.submitted = false
		return s, nil
	}

	// Update focused input
	var cmd tea.Cmd
	if s.focusIdx == 0 {
		s.emailInput, cmd = s.emailInput.Update(msg)
	} else {
		s.tokenInput, cmd = s.tokenInput.Update(msg)
	}
	return s, cmd
}

// nextField moves focus to the next input field.
func (s *SetupScreen) nextField() (tea.Model, tea.Cmd) {
	if s.focusIdx == 0 {
		s.focusIdx = 1
		s.emailInput.Blur()
		s.tokenInput.Focus()
	} else {
		s.focusIdx = 0
		s.tokenInput.Blur()
		s.emailInput.Focus()
	}
	return s, textinput.Blink
}

// prevField moves focus to the previous input field.
func (s *SetupScreen) prevField() (tea.Model, tea.Cmd) {
	return s.nextField() // Only 2 fields, so prev = next
}

// submit attempts to validate and save the token.
func (s *SetupScreen) submit() (tea.Model, tea.Cmd) {
	email := strings.TrimSpace(s.emailInput.Value())
	token := strings.TrimSpace(s.tokenInput.Value())

	if email == "" {
		s.err = nil
		return s, nil
	}
	if token == "" {
		s.err = nil
		return s, nil
	}

	s.submitted = true
	s.err = nil

	return s, func() tea.Msg {
		ctx := context.Background()

		// Validate token against Jira API
		if s.authPort != nil {
			_, err := s.authPort.ValidateToken(ctx, email, token)
			if err != nil {
				return tokenValidationErrorMsg{err: err}
			}
		}

		// Save token
		err := s.tokenStore.SetJiraToken(token)
		if err != nil {
			return tokenValidationErrorMsg{err: err}
		}
		return tokenValidatedMsg{}
	}
}

// tokenValidatedMsg indicates the token was validated and saved successfully.
type tokenValidatedMsg struct{}

// tokenValidationErrorMsg indicates an error validating or saving the token.
type tokenValidationErrorMsg struct {
	err error
}

// View renders the setup screen.
func (s *SetupScreen) View() string {
	var b strings.Builder

	// Title
	title := Styles.Title.Render("Atlassian API Token Setup")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Instructions
	instructions := Styles.Paragraph.Render(
		"To use gojira-tmux, you need to provide your Atlassian email\n" +
			"and API token.\n\n" +
			"Get your token at:\n" +
			"https://id.atlassian.com/manage/api-tokens")
	b.WriteString(instructions)
	b.WriteString("\n\n")

	// Email input
	emailLabel := Styles.InputLabel.Render("Email:")
	b.WriteString(emailLabel)
	b.WriteString("\n")
	emailStyle := Styles.Input
	if s.focusIdx == 0 {
		emailStyle = Styles.InputFocused
	}
	b.WriteString(emailStyle.Render(s.emailInput.View()))
	b.WriteString("\n\n")

	// Token input
	tokenLabel := Styles.InputLabel.Render("API Token:")
	b.WriteString(tokenLabel)
	b.WriteString("\n")
	tokenStyle := Styles.Input
	if s.focusIdx == 1 {
		tokenStyle = Styles.InputFocused
	}
	b.WriteString(tokenStyle.Render(s.tokenInput.View()))
	b.WriteString("\n\n")

	// Error message
	if s.err != nil {
		errMsg := Styles.Error.Render("Error: " + s.err.Error())
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	// Status or help
	if s.submitted && s.err == nil {
		status := Styles.Success.Render("Validating token...")
		b.WriteString(status)
	} else {
		help := lipgloss.JoinHorizontal(lipgloss.Top,
			Styles.HelpKey.Render("tab"),
			Styles.HelpDesc.Render(" switch field  "),
			Styles.HelpKey.Render("enter"),
			Styles.HelpDesc.Render(" submit  "),
			Styles.HelpKey.Render("esc"),
			Styles.HelpDesc.Render(" quit"),
		)
		b.WriteString(help)
	}

	return Styles.App.Render(b.String())
}
