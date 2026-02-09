package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/adapter/config"
	"github.com/stainedhead/gojira-tmux/internal/adapter/jira"
	"github.com/stainedhead/gojira-tmux/internal/infrastructure/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Determine config path
	configPath := getConfigPath()

	// Load configuration
	configLoader := config.NewLoader(configPath)
	cfg, err := configLoader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get Jira API token from environment or keychain
	jiraToken := os.Getenv("JIRA_API_TOKEN")

	// Initialize token store
	credentialsPath := getCredentialsPath()
	tokenStore := auth.NewKeyringTokenStore(credentialsPath)

	// If environment token is set, use it
	if jiraToken != "" {
		if err := tokenStore.SetJiraToken(jiraToken); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}
	}

	// Initialize Atlassian adapter
	atlassianAdapter := auth.NewAtlassianAdapter(tokenStore, cfg.Jira.URL)

	// Initialize Jira client using Atlassian email for authentication
	storedToken, _ := tokenStore.GetJiraToken()
	jiraClient := jira.NewClient(
		cfg.Jira.URL,
		cfg.Atlassian.Email,
		storedToken,
		cfg.Projects,
		cfg.Team,
	)

	// Create application
	app := tui.NewApp(
		tui.WithTokenStore(tokenStore),
		tui.WithAuthPort(atlassianAdapter),
		tui.WithJiraPort(jiraClient),
		tui.WithConfigPort(configLoader),
	)

	// Run TUI
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// getConfigPath returns the path to the config file.
func getConfigPath() string {
	// Check for config in current directory first
	if _, err := os.Stat("config.yaml"); err == nil {
		return "config.yaml"
	}
	if _, err := os.Stat("config.yml"); err == nil {
		return "config.yml"
	}

	// Check in user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "config.yaml"
	}

	gojiraDir := filepath.Join(configDir, "gojira")
	configPath := filepath.Join(gojiraDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Default to current directory
	return "config.yaml"
}

// getCredentialsPath returns the path to store credentials.
func getCredentialsPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".gojira", "credentials")
	}
	return filepath.Join(configDir, "gojira", "credentials")
}
