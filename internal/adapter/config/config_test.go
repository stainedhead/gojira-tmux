package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/config"
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// validMinimalYAML provides a minimal valid config for new unified format.
const validMinimalYAML = `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`

// Helper to write YAML to temp file and create a Loader.
func setupLoader(t *testing.T, yamlContent string) *config.Loader {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return config.NewLoader(configPath)
}

// --- Happy Path Tests (WS2 #1-#8) ---

func TestLoad_ValidMinimalConfig(t *testing.T) {
	loader := setupLoader(t, validMinimalYAML)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Atlassian.URL != "https://example.atlassian.net" {
		t.Errorf("Atlassian.URL = %q, want %q", cfg.Atlassian.URL, "https://example.atlassian.net")
	}
	if cfg.Atlassian.Email != "user@example.com" {
		t.Errorf("Atlassian.Email = %q, want %q", cfg.Atlassian.Email, "user@example.com")
	}
	if len(cfg.Projects) != 1 {
		t.Errorf("len(Projects) = %d, want 1", len(cfg.Projects))
	}
	if len(cfg.Team) != 1 {
		t.Errorf("len(Team) = %d, want 1", len(cfg.Team))
	}
}

func TestLoad_ValidConfigWithCustomFields(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Atlassian.CustomFields.Sprint != "customfield_10020" {
		t.Errorf("CustomFields.Sprint = %q, want %q", cfg.Atlassian.CustomFields.Sprint, "customfield_10020")
	}
	if cfg.Atlassian.CustomFields.Epic != "customfield_10014" {
		t.Errorf("CustomFields.Epic = %q, want %q", cfg.Atlassian.CustomFields.Epic, "customfield_10014")
	}
}

func TestLoad_ValidConfigWithAliases(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
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
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Team[0].Alias != "JohnA" {
		t.Errorf("Team[0].Alias = %q, want %q", cfg.Team[0].Alias, "JohnA")
	}
	if cfg.Team[1].Alias != "JohnF" {
		t.Errorf("Team[1].Alias = %q, want %q", cfg.Team[1].Alias, "JohnF")
	}
}

func TestLoad_ValidConfigWithoutAliases(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
  - name: "Jane Smith"
    email: "jane@example.com"
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Team[0].Alias != "" {
		t.Errorf("Team[0].Alias = %q, want empty", cfg.Team[0].Alias)
	}
}

func TestLoad_MultipleProjects(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "Project One"
  - key: "TEAM"
    name: "Team Project"
  - key: "OPS"
    name: "Operations"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.Projects) != 3 {
		t.Errorf("len(Projects) = %d, want 3", len(cfg.Projects))
	}
}

func TestLoad_MultipleTeamMembers(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "Alice"
    email: "alice@example.com"
  - name: "Bob"
    email: "bob@example.com"
  - name: "Charlie"
    email: "charlie@example.com"
  - name: "Diana"
    email: "diana@example.com"
  - name: "Eve"
    email: "eve@example.com"
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if len(cfg.Team) != 5 {
		t.Errorf("len(Team) = %d, want 5", len(cfg.Team))
	}
}

func TestLoad_CustomFieldsWithStoryPoints(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"
    story_points: "customfield_10028"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Atlassian.CustomFields.StoryPoints != "customfield_10028" {
		t.Errorf("CustomFields.StoryPoints = %q, want %q", cfg.Atlassian.CustomFields.StoryPoints, "customfield_10028")
	}
}

func TestLoad_URLWithTrailingPath(t *testing.T) {
	yaml := `
atlassian:
  url: "https://jira.company.com/jira"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	loader := setupLoader(t, yaml)
	_, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
}

// --- URL Validation Tests (WS2 #9-#15) ---

func TestLoad_URLValidation(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing URL",
			url:         "",
			wantErr:     true,
			errContains: "atlassian.url is required",
		},
		{
			name:        "HTTP URL",
			url:         "http://example.atlassian.net",
			wantErr:     true,
			errContains: "atlassian.url must use HTTPS",
		},
		{
			name:        "no scheme",
			url:         "example.atlassian.net",
			wantErr:     true,
			errContains: "atlassian.url must use HTTPS",
		},
		{
			name:        "FTP scheme",
			url:         "ftp://example.atlassian.net",
			wantErr:     true,
			errContains: "atlassian.url must use HTTPS",
		},
		{
			name:    "valid HTTPS",
			url:     "https://example.atlassian.net",
			wantErr: false,
		},
		{
			name:    "HTTPS custom domain",
			url:     "https://jira.company.com",
			wantErr: false,
		},
		{
			name:    "HTTPS with path",
			url:     "https://company.com/jira",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
atlassian:
  url: "` + tt.url + `"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
			loader := setupLoader(t, yaml)
			_, err := loader.Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %q, want containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error: %v", err)
				}
			}
		})
	}
}

// --- Email Validation Tests (WS2 #16-#22) ---

func TestLoad_EmailValidation(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing email",
			email:       "",
			wantErr:     true,
			errContains: "atlassian.email is required",
		},
		{
			name:        "no @ symbol",
			email:       "not-an-email",
			wantErr:     true,
			errContains: "atlassian.email must be a valid email address",
		},
		{
			name:        "multiple @",
			email:       "a@b@c.com",
			wantErr:     true,
			errContains: "atlassian.email must be a valid email address",
		},
		{
			name:        "empty local part",
			email:       "@company.com",
			wantErr:     true,
			errContains: "atlassian.email must be a valid email address",
		},
		{
			name:        "empty domain",
			email:       "user@",
			wantErr:     true,
			errContains: "atlassian.email must be a valid email address",
		},
		{
			name:    "valid email",
			email:   "user@company.com",
			wantErr: false,
		},
		{
			name:    "plus addressing",
			email:   "user+tag@company.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "` + tt.email + `"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
			loader := setupLoader(t, yaml)
			_, err := loader.Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %q, want containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error: %v", err)
				}
			}
		})
	}
}

// --- Project Validation Tests (WS2 #23-#29) ---

func TestLoad_ProjectValidation(t *testing.T) {
	tests := []struct {
		name        string
		projects    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no projects",
			projects:    "projects: []",
			wantErr:     true,
			errContains: "at least one project is required",
		},
		{
			name: "missing key",
			projects: `projects:
  - key: ""
    name: "My Project"`,
			wantErr:     true,
			errContains: "project key is required",
		},
		{
			name: "lowercase key",
			projects: `projects:
  - key: "proj"
    name: "My Project"`,
			wantErr:     true,
			errContains: "project key must be uppercase letters only",
		},
		{
			name: "mixed case key",
			projects: `projects:
  - key: "Proj"
    name: "My Project"`,
			wantErr:     true,
			errContains: "project key must be uppercase letters only",
		},
		{
			name: "key with numbers",
			projects: `projects:
  - key: "PROJ1"
    name: "My Project"`,
			wantErr:     true,
			errContains: "project key must be uppercase letters only",
		},
		{
			name: "missing name",
			projects: `projects:
  - key: "PROJ"
    name: ""`,
			wantErr:     true,
			errContains: "project name is required",
		},
		{
			name: "valid project",
			projects: `projects:
  - key: "PROJ"
    name: "My Project"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

` + tt.projects + `

team:
  - name: "John Doe"
    email: "john@example.com"
`
			loader := setupLoader(t, yaml)
			_, err := loader.Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %q, want containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error: %v", err)
				}
			}
		})
	}
}

// --- Team Validation Tests (WS2 #30-#39) ---

func TestLoad_TeamValidation(t *testing.T) {
	tests := []struct {
		name        string
		team        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no team members",
			team:        "team: []",
			wantErr:     true,
			errContains: "at least one team member is required",
		},
		{
			name: "missing name",
			team: `team:
  - name: ""
    email: "john@example.com"`,
			wantErr:     true,
			errContains: "team member name is required",
		},
		{
			name: "missing email",
			team: `team:
  - name: "John Doe"
    email: ""`,
			wantErr:     true,
			errContains: "team member email is required",
		},
		{
			name: "invalid email",
			team: `team:
  - name: "John Doe"
    email: "invalid"`,
			wantErr:     true,
			errContains: "team member email is invalid",
		},
		{
			name: "alias with spaces",
			team: `team:
  - name: "John Doe"
    email: "john@example.com"
    alias: "John D"`,
			wantErr:     true,
			errContains: "team member alias must be alphanumeric",
		},
		{
			name: "alias with special chars",
			team: `team:
  - name: "John Doe"
    email: "john@example.com"
    alias: "John@D"`,
			wantErr:     true,
			errContains: "team member alias must be alphanumeric",
		},
		{
			name: "duplicate alias",
			team: `team:
  - name: "John Anderson"
    email: "john.a@example.com"
    alias: "JohnA"
  - name: "John Adams"
    email: "john.ad@example.com"
    alias: "JohnA"`,
			wantErr:     true,
			errContains: "duplicate alias",
		},
		{
			name: "duplicate email",
			team: `team:
  - name: "John Doe"
    email: "john@example.com"
  - name: "John D"
    email: "john@example.com"`,
			wantErr:     true,
			errContains: "duplicate email",
		},
		{
			name: "case-insensitive duplicate email",
			team: `team:
  - name: "John Doe"
    email: "John@CO.com"
  - name: "Jane Doe"
    email: "john@co.com"`,
			wantErr:     true,
			errContains: "duplicate email",
		},
		{
			name: "valid alias",
			team: `team:
  - name: "John Doe"
    email: "john@example.com"
    alias: "JohnA"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

` + tt.team + `
`
			loader := setupLoader(t, yaml)
			_, err := loader.Load()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %q, want containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error: %v", err)
				}
			}
		})
	}
}

// --- Edge Case Tests (WS2 #40-#44) ---

func TestLoad_FileNotFound(t *testing.T) {
	loader := config.NewLoader("/nonexistent/path/config.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Load() error = %q, want containing %q", err.Error(), "failed to read config file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
  email: [this is not valid yaml
`
	loader := setupLoader(t, yaml)
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse config file") {
		t.Errorf("Load() error = %q, want containing %q", err.Error(), "failed to parse config file")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	loader := setupLoader(t, "")
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() expected error for empty file, got nil")
	}
}

func TestLoad_OldConfigFormat(t *testing.T) {
	// Old format with jira.url but no atlassian.url should fail
	yaml := `
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
`
	loader := setupLoader(t, yaml)
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() expected error for old config format, got nil")
	}
	if !strings.Contains(err.Error(), "atlassian.url is required") {
		t.Errorf("Load() error = %q, want containing %q", err.Error(), "atlassian.url is required")
	}
}

func TestLoad_BothOldAndNewConfig(t *testing.T) {
	// Both old jira.url and new atlassian.url present - should succeed (old section ignored)
	yaml := `
jira:
  url: "https://old.atlassian.net"

atlassian:
  url: "https://example.atlassian.net"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	loader := setupLoader(t, yaml)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Atlassian.URL != "https://example.atlassian.net" {
		t.Errorf("Atlassian.URL = %q, want %q", cfg.Atlassian.URL, "https://example.atlassian.net")
	}
}

// --- Accessor Tests ---

func TestLoader_GetProjects(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
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
	loader := setupLoader(t, yaml)
	_, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	projects := loader.GetProjects()
	if len(projects) != 2 {
		t.Errorf("GetProjects() returned %d projects, want 2", len(projects))
	}
}

func TestLoader_GetProjects_BeforeLoad(t *testing.T) {
	loader := config.NewLoader("unused")
	projects := loader.GetProjects()
	if projects != nil {
		t.Errorf("GetProjects() before Load() = %v, want nil", projects)
	}
}

func TestLoader_GetTeamMembers(t *testing.T) {
	yaml := `
atlassian:
  url: "https://example.atlassian.net"
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
	loader := setupLoader(t, yaml)
	_, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	team := loader.GetTeamMembers()
	if len(team) != 2 {
		t.Errorf("GetTeamMembers() returned %d members, want 2", len(team))
	}
}

func TestLoader_GetTeamMembers_BeforeLoad(t *testing.T) {
	loader := config.NewLoader("unused")
	team := loader.GetTeamMembers()
	if team != nil {
		t.Errorf("GetTeamMembers() before Load() = %v, want nil", team)
	}
}

// --- Save Tests ---

func TestLoader_Save_PersistsFilterState(t *testing.T) {
	loader := setupLoader(t, validMinimalYAML)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cfg.LastFilter.Assignee = "John Doe"
	cfg.LastFilter.Project = "PROJ"
	cfg.LastFilter.Status = "In Progress"

	if err := loader.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Reload and verify
	cfg2, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if cfg2.LastFilter.Assignee != "John Doe" {
		t.Errorf("LastFilter.Assignee = %q, want %q", cfg2.LastFilter.Assignee, "John Doe")
	}
	if cfg2.LastFilter.Project != "PROJ" {
		t.Errorf("LastFilter.Project = %q, want %q", cfg2.LastFilter.Project, "PROJ")
	}
	if cfg2.LastFilter.Status != "In Progress" {
		t.Errorf("LastFilter.Status = %q, want %q", cfg2.LastFilter.Status, "In Progress")
	}
}

func TestLoader_Save_PreservesExistingFields(t *testing.T) {
	loader := setupLoader(t, validMinimalYAML)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cfg.LastFilter.Status = "Done"
	if err := loader.Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	cfg2, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	// Original fields must still be present
	if cfg2.Atlassian.URL != "https://example.atlassian.net" {
		t.Errorf("Atlassian.URL lost after Save: %q", cfg2.Atlassian.URL)
	}
	if len(cfg2.Projects) != 1 {
		t.Errorf("Projects lost after Save: got %d", len(cfg2.Projects))
	}
	if len(cfg2.Team) != 1 {
		t.Errorf("Team lost after Save: got %d", len(cfg2.Team))
	}
}

func TestLoader_Save_ClearsFilter(t *testing.T) {
	loader := setupLoader(t, validMinimalYAML)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cfg.LastFilter.Status = "Open"
	_ = loader.Save(cfg)

	// Clear and save again
	cfg.LastFilter = domain.FilterState{}
	if err := loader.Save(cfg); err != nil {
		t.Fatalf("Save() after clear error: %v", err)
	}

	cfg2, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg2.LastFilter.Status != "" {
		t.Errorf("LastFilter.Status = %q after clear, want empty", cfg2.LastFilter.Status)
	}
}

func TestLoader_Save_InvalidPath(t *testing.T) {
	loader := config.NewLoader("/nonexistent/directory/config.yaml")
	cfg := &domain.Config{
		Atlassian: domain.AtlassianConfig{URL: "https://example.atlassian.net", Email: "u@e.com"},
		Projects:  []domain.Project{{Key: "PROJ", Name: "P"}},
		Team:      []domain.TeamMember{{Name: "A", Email: "a@b.com"}},
	}
	err := loader.Save(cfg)
	if err == nil {
		t.Error("Save() to invalid path should return error")
	}
}

// --- URL Edge Cases ---

func TestLoad_URLSchemeOnly(t *testing.T) {
	yaml := `
atlassian:
  url: "https://"
  email: "user@example.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john@example.com"
`
	loader := setupLoader(t, yaml)
	_, err := loader.Load()
	if err == nil {
		t.Error("Load() expected error for URL with scheme only, got nil")
	}
	if !strings.Contains(err.Error(), "atlassian.url must be a valid URL") {
		t.Errorf("Load() error = %q, want containing %q", err.Error(), "atlassian.url must be a valid URL")
	}
}
