package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/config"
)

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errContains string
		validate    func(*testing.T, *config.Config)
	}{
		{
			name: "valid config",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.Jira.URL != "https://example.atlassian.net" {
					t.Errorf("Jira.URL = %q, want %q", cfg.Jira.URL, "https://example.atlassian.net")
				}
				if cfg.Atlassian.Email != "user@example.com" {
					t.Errorf("Atlassian.Email = %q, want %q", cfg.Atlassian.Email, "user@example.com")
				}
				if len(cfg.Projects) != 1 {
					t.Errorf("len(Projects) = %d, want %d", len(cfg.Projects), 1)
				}
				if len(cfg.Team) != 1 {
					t.Errorf("len(Team) = %d, want %d", len(cfg.Team), 1)
				}
			},
		},
		{
			name: "valid config with aliases",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Anderson"
    email: "john.anderson@example.com"
    alias: "JohnA"
  - name: "John Flanagan"
    email: "john.flanagan@example.com"
    alias: "JohnF"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.Team[0].Alias != "JohnA" {
					t.Errorf("Team[0].Alias = %q, want %q", cfg.Team[0].Alias, "JohnA")
				}
				if cfg.Team[1].Alias != "JohnF" {
					t.Errorf("Team[1].Alias = %q, want %q", cfg.Team[1].Alias, "JohnF")
				}
			},
		},
		{
			name: "backward compatible - no aliases",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
  - name: "Jane Smith"
    email: "jane@example.com"
`,
			wantErr: false,
			validate: func(t *testing.T, cfg *config.Config) {
				if cfg.Team[0].Alias != "" {
					t.Errorf("Team[0].Alias = %q, want empty", cfg.Team[0].Alias)
				}
			},
		},
		{
			name: "missing jira url",
			yamlContent: `
atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "jira.url is required",
		},
		{
			name: "jira url not https",
			yamlContent: `
jira:
  url: "http://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "jira.url must use HTTPS",
		},
		{
			name: "missing atlassian email",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "atlassian.email is required",
		},
		{
			name: "invalid atlassian email",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "not-an-email"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "atlassian.email must be a valid email address",
		},
		{
			name: "no projects",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects: []

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "at least one project is required",
		},
		{
			name: "no team members",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team: []
`,
			wantErr:     true,
			errContains: "at least one team member is required",
		},
		{
			name: "invalid project key",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "proj123"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "project key must be uppercase letters only",
		},
		{
			name: "invalid team member email",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "invalid-email"
`,
			wantErr:     true,
			errContains: "team member email is invalid",
		},
		{
			name: "duplicate alias",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Anderson"
    email: "john.a@example.com"
    alias: "JohnA"
  - name: "John Adams"
    email: "john.ad@example.com"
    alias: "JohnA"
`,
			wantErr:     true,
			errContains: "duplicate alias",
		},
		{
			name: "invalid alias format",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
    alias: "John D"
`,
			wantErr:     true,
			errContains: "team member alias must be alphanumeric",
		},
		{
			name: "duplicate email",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
  - name: "John D"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "duplicate email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0600)
			if err != nil {
				t.Fatalf("failed to write temp config: %v", err)
			}

			loader := config.NewLoader(configPath)
			cfg, err := loader.Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoader_Load_FileNotFound(t *testing.T) {
	loader := config.NewLoader("/nonexistent/path/config.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
}

func TestLoader_GetProjects(t *testing.T) {
	yamlContent := `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "Project One"
  - key: "TEST"
    name: "Test Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0600)
	if err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	loader := config.NewLoader(configPath)
	_, err = loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	projects := loader.GetProjects()
	if len(projects) != 2 {
		t.Errorf("GetProjects() returned %d projects, want 2", len(projects))
	}
}

func TestLoader_GetTeamMembers(t *testing.T) {
	yamlContent := `
jira:
  url: "https://example.atlassian.net"

atlassian:
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "Project One"

team:
  - name: "John Doe"
    email: "john@example.com"
  - name: "Jane Smith"
    email: "jane@example.com"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0600)
	if err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	loader := config.NewLoader(configPath)
	_, err = loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	team := loader.GetTeamMembers()
	if len(team) != 2 {
		t.Errorf("GetTeamMembers() returned %d members, want 2", len(team))
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
