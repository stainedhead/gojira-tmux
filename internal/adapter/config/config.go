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
	// Atlassian URL validation
	if cfg.Atlassian.URL == "" {
		return errors.New("atlassian.url is required")
	}
	if !strings.HasPrefix(cfg.Atlassian.URL, "https://") {
		return errors.New("atlassian.url must use HTTPS")
	}
	// Validate URL has a host after the scheme
	host := strings.TrimPrefix(cfg.Atlassian.URL, "https://")
	if host == "" || host == "/" {
		return errors.New("atlassian.url must be a valid URL")
	}

	// Atlassian email validation
	if cfg.Atlassian.Email == "" {
		return errors.New("atlassian.email is required")
	}
	if !isValidEmail(cfg.Atlassian.Email) {
		return errors.New("atlassian.email must be a valid email address")
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

	aliases := make(map[string]bool)
	for i, m := range cfg.Team {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("team member %s: %w", m.Name, err)
		}

		if m.Alias != "" {
			if aliases[m.Alias] {
				return fmt.Errorf("duplicate alias %q found (team member %s)", m.Alias, m.Name)
			}
			aliases[m.Alias] = true
		}

		// Check for duplicate emails
		for j := 0; j < i; j++ {
			if strings.EqualFold(cfg.Team[j].Email, m.Email) {
				return fmt.Errorf("duplicate email %q (team members %s and %s)",
					m.Email, cfg.Team[j].Name, m.Name)
			}
		}
	}

	return nil
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	if local == "" || domain == "" {
		return false
	}
	if !strings.Contains(domain, ".") && len(domain) < 2 {
		return false
	}
	return true
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

// Ensure Loader implements domain.ConfigPort.
var _ domain.ConfigPort = (*Loader)(nil)
