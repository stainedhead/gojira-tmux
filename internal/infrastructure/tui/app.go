package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// Screen represents the current screen in the application.
type Screen int

const (
	// ScreenSetup is the first-time API token setup screen.
	ScreenSetup Screen = iota
	// ScreenMain is the main ticket list view.
	ScreenMain
)

// App is the root model for the application.
type App struct {
	screen Screen
	width  int
	height int

	// Shared state
	config *domain.Config
	user   *domain.User

	// Dependencies (to be injected)
	tokenStore domain.TokenStorePort
	authPort   domain.AuthPort
	jiraPort   domain.JiraPort
	configPort domain.ConfigPort

	// Screen models (initialized lazily)
	setupScreen tea.Model
	mainScreen  tea.Model

	// Error state
	err error
}

// AppOption is a function that configures the App.
type AppOption func(*App)

// WithTokenStore sets the token store port.
func WithTokenStore(ts domain.TokenStorePort) AppOption {
	return func(a *App) {
		a.tokenStore = ts
	}
}

// WithAuthPort sets the auth port.
func WithAuthPort(ap domain.AuthPort) AppOption {
	return func(a *App) {
		a.authPort = ap
	}
}

// WithJiraPort sets the Jira port.
func WithJiraPort(jp domain.JiraPort) AppOption {
	return func(a *App) {
		a.jiraPort = jp
	}
}

// WithConfigPort sets the config port.
func WithConfigPort(cp domain.ConfigPort) AppOption {
	return func(a *App) {
		a.configPort = cp
	}
}

// NewApp creates a new App with the given options.
func NewApp(opts ...AppOption) *App {
	app := &App{
		screen: ScreenSetup,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Init initializes the application.
func (a *App) Init() tea.Cmd {
	return a.checkInitialState()
}

// checkInitialState determines which screen to show based on stored credentials.
func (a *App) checkInitialState() tea.Cmd {
	return func() tea.Msg {
		// Check if we have a stored Jira token
		if a.tokenStore != nil && a.tokenStore.HasJiraToken() {
			// Token exists, go directly to main screen
			return screenChangeMsg{screen: ScreenMain}
		}
		// No token, show setup screen
		return screenChangeMsg{screen: ScreenSetup}
	}
}

// screenChangeMsg is an internal message for screen transitions.
type screenChangeMsg struct {
	screen Screen
}

// Update handles messages and updates the model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		// Global quit handling
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case screenChangeMsg:
		a.screen = msg.screen
		return a, a.initCurrentScreen()

	case TokenStoredMsg:
		// Token validated and saved, go to main screen
		a.screen = ScreenMain
		return a, a.initCurrentScreen()

	case LogoutMsg:
		a.user = nil
		a.screen = ScreenSetup
		return a, a.initCurrentScreen()

	case ErrorMsg:
		a.err = msg.Err
		return a, nil
	}

	// Route to active screen
	return a.updateCurrentScreen(msg)
}

// initCurrentScreen initializes the current screen's model.
func (a *App) initCurrentScreen() tea.Cmd {
	switch a.screen {
	case ScreenSetup:
		if a.setupScreen == nil {
			a.setupScreen = NewSetupScreenModel(a.tokenStore, a.authPort)
		}
		return a.setupScreen.Init()
	case ScreenMain:
		if a.mainScreen == nil {
			a.mainScreen = NewMainScreen(a.jiraPort, a.configPort, a.user)
		}
		return a.mainScreen.Init()
	}
	return nil
}

// updateCurrentScreen routes updates to the current screen.
func (a *App) updateCurrentScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch a.screen {
	case ScreenSetup:
		if a.setupScreen != nil {
			a.setupScreen, cmd = a.setupScreen.Update(msg)
		}
	case ScreenMain:
		if a.mainScreen != nil {
			a.mainScreen, cmd = a.mainScreen.Update(msg)
		}
	}

	return a, cmd
}

// View renders the current screen.
func (a *App) View() string {
	if a.err != nil {
		return Styles.Error.Render("Error: " + a.err.Error())
	}

	switch a.screen {
	case ScreenSetup:
		if a.setupScreen != nil {
			return a.setupScreen.View()
		}
		return "Loading setup..."
	case ScreenMain:
		if a.mainScreen != nil {
			return a.mainScreen.View()
		}
		return "Loading main screen..."
	}

	return "Unknown screen"
}

// NewMainScreen creates a new main screen.
func NewMainScreen(jiraPort domain.JiraPort, configPort domain.ConfigPort, user *domain.User) tea.Model {
	return NewMainScreenModel(jiraPort, configPort, user)
}

// placeholderScreen is a temporary screen implementation.
type placeholderScreen struct {
	name string
}

func (p *placeholderScreen) Init() tea.Cmd {
	return nil
}

func (p *placeholderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return p, nil
}

func (p *placeholderScreen) View() string {
	return Styles.Title.Render(p.name+" Screen") + "\n\n" +
		Styles.Muted.Render("(Not yet implemented)")
}
