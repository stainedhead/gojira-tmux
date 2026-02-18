package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/adapter/config"
	"github.com/stainedhead/gojira-tmux/internal/adapter/jira"
	"github.com/stainedhead/gojira-tmux/internal/domain"
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
	atlassianAdapter := auth.NewAtlassianAdapter(tokenStore, cfg.Atlassian.URL)

	// Initialize Jira client using Atlassian email for authentication
	storedToken, _ := tokenStore.GetJiraToken()
	jiraClient := jira.NewClient(
		cfg.Atlassian.URL,
		cfg.Atlassian.Email,
		storedToken,
		cfg.Projects,
		cfg.Team,
	)

	// Fetch available Jira statuses (best-effort; falls back to config or defaults)
	statuses := fetchStatuses(jiraClient, cfg)

	// Create application
	app := tui.NewApp(
		tui.WithTokenStore(tokenStore),
		tui.WithAuthPort(atlassianAdapter),
		tui.WithJiraPort(jiraClient),
		tui.WithConfigPort(configLoader),
		tui.WithConfig(cfg),
		tui.WithStatuses(statuses),
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

// fetchStatuses retrieves the list of Jira status names from the API.
// Falls back to cfg.Statuses, then a built-in default list if the API call fails.
func fetchStatuses(client *jira.Client, cfg *domain.Config) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if statuses, err := client.ListStatuses(ctx); err == nil && len(statuses) > 0 {
		return statuses
	}

	// Config-defined fallback
	if len(cfg.Statuses) > 0 {
		return cfg.Statuses
	}

	// Built-in defaults
	return []string{"To Do", "In Progress", "In Review", "Done"}
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
