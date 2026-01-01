package tui

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// LoginScreen handles the Okta SSO login flow.
type LoginScreen struct {
	authPort   domain.AuthPort
	configPort domain.ConfigPort

	state     loginState
	authURL   string
	err       error
	keys      LoginKeyMap
	cancelled bool
}

type loginState int

const (
	loginStateReady loginState = iota
	loginStateWaiting
	loginStateSuccess
	loginStateError
)

// NewLoginScreenModel creates a new login screen.
func NewLoginScreenModel(authPort domain.AuthPort, configPort domain.ConfigPort) *LoginScreen {
	return &LoginScreen{
		authPort:   authPort,
		configPort: configPort,
		state:      loginStateReady,
		keys:       DefaultLoginKeyMap(),
	}
}

// Init initializes the login screen.
func (s *LoginScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages for the login screen.
func (s *LoginScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if s.state == loginStateReady {
				return s.startLogin()
			}
		case "esc":
			if s.state == loginStateWaiting {
				return s.cancelLogin()
			}
			return s, tea.Quit
		case "ctrl+c", "q":
			if s.state == loginStateWaiting {
				return s.cancelLogin()
			}
			return s, tea.Quit
		}

	case authStartedInternalMsg:
		s.state = loginStateWaiting
		s.authURL = msg.url
		return s, s.waitForCallback()

	case authCallbackInternalMsg:
		if msg.err != nil {
			s.state = loginStateError
			s.err = msg.err
			return s, nil
		}
		s.state = loginStateSuccess
		return s, func() tea.Msg {
			return AuthSuccessMsg{User: msg.user}
		}

	case authCancelledInternalMsg:
		s.state = loginStateReady
		s.cancelled = true
		return s, nil
	}

	return s, nil
}

// startLogin initiates the OAuth flow.
func (s *LoginScreen) startLogin() (tea.Model, tea.Cmd) {
	s.err = nil
	s.cancelled = false

	return s, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		url, err := s.authPort.StartAuthFlow(ctx)
		if err != nil {
			return authCallbackInternalMsg{err: err}
		}

		// Open browser
		_ = openBrowser(url)

		return authStartedInternalMsg{url: url}
	}
}

// waitForCallback waits for the OAuth callback.
func (s *LoginScreen) waitForCallback() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		user, err := s.authPort.WaitForCallback(ctx)
		if err != nil {
			return authCallbackInternalMsg{err: err}
		}

		// Validate team membership
		if s.configPort != nil {
			if err := s.configPort.ValidateUserAccess(user.Email); err != nil {
				return authCallbackInternalMsg{err: err}
			}
		}

		return authCallbackInternalMsg{user: user}
	}
}

// cancelLogin cancels the OAuth flow.
func (s *LoginScreen) cancelLogin() (tea.Model, tea.Cmd) {
	s.authPort.CancelAuthFlow()
	return s, func() tea.Msg {
		return authCancelledInternalMsg{}
	}
}

// Internal messages
type authStartedInternalMsg struct {
	url string
}

type authCallbackInternalMsg struct {
	user *domain.User
	err  error
}

type authCancelledInternalMsg struct{}

// View renders the login screen.
func (s *LoginScreen) View() string {
	var b strings.Builder

	// Title
	title := Styles.Title.Render("🔑 Okta SSO Login")
	b.WriteString(title)
	b.WriteString("\n\n")

	switch s.state {
	case loginStateReady:
		s.renderReady(&b)
	case loginStateWaiting:
		s.renderWaiting(&b)
	case loginStateSuccess:
		s.renderSuccess(&b)
	case loginStateError:
		s.renderError(&b)
	}

	return Styles.App.Render(b.String())
}

func (s *LoginScreen) renderReady(b *strings.Builder) {
	instructions := Styles.Paragraph.Render(
		"Press Enter to open your browser and log in with Okta.\n\n" +
			"Your browser will open automatically. Complete the login\n" +
			"in the browser, then return here.")
	b.WriteString(instructions)
	b.WriteString("\n\n")

	if s.cancelled {
		msg := Styles.Warning.Render("Login cancelled. Press Enter to try again.")
		b.WriteString(msg)
		b.WriteString("\n\n")
	}

	help := lipgloss.JoinHorizontal(lipgloss.Top,
		Styles.HelpKey.Render("enter"),
		Styles.HelpDesc.Render(" login  "),
		Styles.HelpKey.Render("q/esc"),
		Styles.HelpDesc.Render(" quit"),
	)
	b.WriteString(help)
}

func (s *LoginScreen) renderWaiting(b *strings.Builder) {
	status := Styles.Paragraph.Render("Waiting for authentication...\n\n" +
		"A browser window should have opened. Complete the login\n" +
		"in your browser to continue.")
	b.WriteString(status)
	b.WriteString("\n\n")

	if s.authURL != "" {
		urlLabel := Styles.Muted.Render("If the browser didn't open, visit:")
		b.WriteString(urlLabel)
		b.WriteString("\n")
		url := Styles.Subtitle.Render(truncateURL(s.authURL, 60))
		b.WriteString(url)
		b.WriteString("\n\n")
	}

	help := lipgloss.JoinHorizontal(lipgloss.Top,
		Styles.HelpKey.Render("esc"),
		Styles.HelpDesc.Render(" cancel"),
	)
	b.WriteString(help)
}

func (s *LoginScreen) renderSuccess(b *strings.Builder) {
	msg := Styles.Success.Render("✓ Login successful!")
	b.WriteString(msg)
	b.WriteString("\n\n")
	b.WriteString(Styles.Muted.Render("Redirecting..."))
}

func (s *LoginScreen) renderError(b *strings.Builder) {
	if s.err != nil {
		errMsg := Styles.Error.Render("✗ Login failed: " + s.err.Error())
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	help := lipgloss.JoinHorizontal(lipgloss.Top,
		Styles.HelpKey.Render("enter"),
		Styles.HelpDesc.Render(" try again  "),
		Styles.HelpKey.Render("q/esc"),
		Styles.HelpDesc.Render(" quit"),
	)
	b.WriteString(help)
}

// openBrowser opens the given URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return nil
	}

	return cmd.Start()
}

// truncateURL truncates a URL to the given length.
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
