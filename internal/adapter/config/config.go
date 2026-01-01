package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// Loader implements domain.ConfigPort for loading configuration from YAML files.
type Loader struct {
	path   string
	config *domain.Config
}

// NewLoader creates a new config loader for the given path.
func NewLoader(path string) *Loader {
	return &Loader{
		path: path,
	}
}

// Load reads and parses the configuration file.
func (l *Loader) Load() (*domain.Config, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg domain.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := l.validate(&cfg); err != nil {
		return nil, err
	}

	l.config = &cfg
	return &cfg, nil
}

// validate checks that the configuration is valid.
func (l *Loader) validate(cfg *domain.Config) error {
	// Jira validation
	if cfg.Jira.URL == "" {
		return errors.New("jira.url is required")
	}
	if !strings.HasPrefix(cfg.Jira.URL, "https://") {
		return errors.New("jira.url must use HTTPS")
	}
	if cfg.Jira.Username == "" {
		return errors.New("jira.username is required")
	}

	// Okta validation
	if cfg.Okta.Issuer == "" {
		return errors.New("okta.issuer is required")
	}
	if cfg.Okta.ClientID == "" {
		return errors.New("okta.client_id is required")
	}
	if cfg.Okta.CallbackPort <= 0 || cfg.Okta.CallbackPort > 65535 {
		return errors.New("okta.callback_port must be 1-65535")
	}

	// Projects validation (minimum 1)
	if len(cfg.Projects) == 0 {
		return errors.New("at least one project is required")
	}
	for _, p := range cfg.Projects {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("project %s: %w", p.Key, err)
		}
	}

	// Team validation (minimum 1)
	if len(cfg.Team) == 0 {
		return errors.New("at least one team member is required")
	}
	for _, m := range cfg.Team {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("team member %s: %w", m.Name, err)
		}
	}

	return nil
}

// GetProjects returns configured projects.
func (l *Loader) GetProjects() []domain.Project {
	if l.config == nil {
		return nil
	}
	return l.config.Projects
}

// GetTeamMembers returns configured team members.
func (l *Loader) GetTeamMembers() []domain.TeamMember {
	if l.config == nil {
		return nil
	}
	return l.config.Team
}

// ValidateUserAccess checks if email is in team list.
func (l *Loader) ValidateUserAccess(email string) error {
	if l.config == nil {
		return errors.New("config not loaded")
	}

	for _, m := range l.config.Team {
		if strings.EqualFold(m.Email, email) {
			return nil
		}
	}
	return fmt.Errorf("user %s is not a member of the configured team", email)
}

// Ensure Loader implements domain.ConfigPort.
var _ domain.ConfigPort = (*Loader)(nil)
