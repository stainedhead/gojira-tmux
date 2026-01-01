package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/config"
)

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name       string
		yamlContent string
		wantErr    bool
		errContains string
		validate   func(*testing.T, *config.Config)
	}{
		{
			name: "valid config",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"
  username: "user@example.com"
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"
    story_points: "customfield_10016"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080
  scopes:
    - "openid"
    - "profile"
    - "email"

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
				if cfg.Jira.Username != "user@example.com" {
					t.Errorf("Jira.Username = %q, want %q", cfg.Jira.Username, "user@example.com")
				}
				if cfg.Okta.Issuer != "https://example.okta.com/oauth2/default" {
					t.Errorf("Okta.Issuer = %q, want correct value", cfg.Okta.Issuer)
				}
				if cfg.Okta.ClientID != "0oaexample" {
					t.Errorf("Okta.ClientID = %q, want %q", cfg.Okta.ClientID, "0oaexample")
				}
				if cfg.Okta.CallbackPort != 8080 {
					t.Errorf("Okta.CallbackPort = %d, want %d", cfg.Okta.CallbackPort, 8080)
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
			name: "missing jira url",
			yamlContent: `
jira:
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
			name: "missing jira username",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "jira.username is required",
		},
		{
			name: "jira url not https",
			yamlContent: `
jira:
  url: "http://example.atlassian.net"
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
			name: "missing okta issuer",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"
  username: "user@example.com"

okta:
  client_id: "0oaexample"
  callback_port: 8080

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "okta.issuer is required",
		},
		{
			name: "missing okta client_id",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  callback_port: 8080

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "okta.client_id is required",
		},
		{
			name: "invalid callback port",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 0

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`,
			wantErr:     true,
			errContains: "okta.callback_port must be 1-65535",
		},
		{
			name: "no projects",
			yamlContent: `
jira:
  url: "https://example.atlassian.net"
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with YAML content
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0600)
			if err != nil {
				t.Fatalf("failed to write temp config: %v", err)
			}

			// Create loader and load config
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
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

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

func TestLoader_ValidateUserAccess(t *testing.T) {
	yamlContent := `
jira:
  url: "https://example.atlassian.net"
  username: "user@example.com"

okta:
  issuer: "https://example.okta.com/oauth2/default"
  client_id: "0oaexample"
  callback_port: 8080

projects:
  - key: "PROJ"
    name: "Project One"

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

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid user",
			email:   "john@example.com",
			wantErr: false,
		},
		{
			name:    "valid user case insensitive",
			email:   "JOHN@EXAMPLE.COM",
			wantErr: false,
		},
		{
			name:    "invalid user",
			email:   "stranger@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidateUserAccess(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
