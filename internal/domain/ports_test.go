package domain

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// --- IssueFilter tests ---

func TestIssueFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		filter IssueFilter
		want   bool
	}{
		{
			name:   "zero value filter is empty",
			filter: IssueFilter{},
			want:   true,
		},
		{
			name:   "all sentinel values is empty",
			filter: IssueFilter{Project: "-All-", Assignee: "-All-", Status: "All"},
			want:   true,
		},
		{
			name:   "project set makes non-empty",
			filter: IssueFilter{Project: "PROJ"},
			want:   false,
		},
		{
			name:   "assignee set makes non-empty",
			filter: IssueFilter{Assignee: "john@example.com"},
			want:   false,
		},
		{
			name:   "status set makes non-empty",
			filter: IssueFilter{Status: "Open"},
			want:   false,
		},
		{
			name:   "project -All- with real assignee is non-empty",
			filter: IssueFilter{Project: "-All-", Assignee: "john@example.com"},
			want:   false,
		},
		{
			name:   "all fields set is non-empty",
			filter: IssueFilter{Project: "PROJ", Assignee: "john@example.com", Status: "Done"},
			want:   false,
		},
		{
			name:   "empty project with -All- assignee and All status is empty",
			filter: IssueFilter{Project: "", Assignee: "-All-", Status: "All"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.IsEmpty()
			if got != tt.want {
				t.Errorf("IssueFilter.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Config struct tests ---

func TestConfig_HasAtlassianField(t *testing.T) {
	cfg := Config{
		Atlassian: AtlassianConfig{
			URL:   "https://example.atlassian.net",
			Email: "user@example.com",
		},
	}

	if cfg.Atlassian.URL != "https://example.atlassian.net" {
		t.Errorf("Config.Atlassian.URL = %q, want %q", cfg.Atlassian.URL, "https://example.atlassian.net")
	}
	if cfg.Atlassian.Email != "user@example.com" {
		t.Errorf("Config.Atlassian.Email = %q, want %q", cfg.Atlassian.Email, "user@example.com")
	}
}

func TestConfig_NoJiraField(t *testing.T) {
	// Verify Config YAML unmarshaling ignores old "jira" section
	yamlData := `
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
	var cfg Config
	if err := yaml.Unmarshal([]byte(yamlData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	if cfg.Atlassian.URL != "https://example.atlassian.net" {
		t.Errorf("Config.Atlassian.URL = %q, want %q", cfg.Atlassian.URL, "https://example.atlassian.net")
	}
	if cfg.Atlassian.Email != "user@example.com" {
		t.Errorf("Config.Atlassian.Email = %q, want %q", cfg.Atlassian.Email, "user@example.com")
	}
}

// --- AtlassianConfig tests ---

func TestAtlassianConfig_AllFields(t *testing.T) {
	ac := AtlassianConfig{
		URL:   "https://company.atlassian.net",
		Email: "dev@company.com",
		CustomFields: CustomFieldConfig{
			Sprint:      "customfield_10020",
			Epic:        "customfield_10014",
			StoryPoints: "customfield_10028",
		},
	}

	if ac.URL != "https://company.atlassian.net" {
		t.Errorf("URL = %q, want %q", ac.URL, "https://company.atlassian.net")
	}
	if ac.Email != "dev@company.com" {
		t.Errorf("Email = %q, want %q", ac.Email, "dev@company.com")
	}
	if ac.CustomFields.Sprint != "customfield_10020" {
		t.Errorf("CustomFields.Sprint = %q, want %q", ac.CustomFields.Sprint, "customfield_10020")
	}
	if ac.CustomFields.Epic != "customfield_10014" {
		t.Errorf("CustomFields.Epic = %q, want %q", ac.CustomFields.Epic, "customfield_10014")
	}
	if ac.CustomFields.StoryPoints != "customfield_10028" {
		t.Errorf("CustomFields.StoryPoints = %q, want %q", ac.CustomFields.StoryPoints, "customfield_10028")
	}
}

func TestAtlassianConfig_YAMLUnmarshal(t *testing.T) {
	yamlData := `
url: "https://test.atlassian.net"
email: "test@example.com"
custom_fields:
  sprint: "customfield_10020"
  epic: "customfield_10014"
  story_points: "customfield_10028"
`
	var ac AtlassianConfig
	if err := yaml.Unmarshal([]byte(yamlData), &ac); err != nil {
		t.Fatalf("failed to unmarshal AtlassianConfig: %v", err)
	}

	if ac.URL != "https://test.atlassian.net" {
		t.Errorf("URL = %q, want %q", ac.URL, "https://test.atlassian.net")
	}
	if ac.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", ac.Email, "test@example.com")
	}
	if ac.CustomFields.Sprint != "customfield_10020" {
		t.Errorf("CustomFields.Sprint = %q, want %q", ac.CustomFields.Sprint, "customfield_10020")
	}
	if ac.CustomFields.Epic != "customfield_10014" {
		t.Errorf("CustomFields.Epic = %q, want %q", ac.CustomFields.Epic, "customfield_10014")
	}
	if ac.CustomFields.StoryPoints != "customfield_10028" {
		t.Errorf("CustomFields.StoryPoints = %q, want %q", ac.CustomFields.StoryPoints, "customfield_10028")
	}
}

func TestAtlassianConfig_YAMLUnmarshal_EmailOnly(t *testing.T) {
	yamlData := `email: "minimal@example.com"`
	var ac AtlassianConfig
	if err := yaml.Unmarshal([]byte(yamlData), &ac); err != nil {
		t.Fatalf("failed to unmarshal AtlassianConfig: %v", err)
	}

	if ac.Email != "minimal@example.com" {
		t.Errorf("Email = %q, want %q", ac.Email, "minimal@example.com")
	}
	if ac.URL != "" {
		t.Errorf("URL = %q, want empty", ac.URL)
	}
	if ac.CustomFields.Sprint != "" {
		t.Errorf("CustomFields.Sprint = %q, want empty", ac.CustomFields.Sprint)
	}
}

func TestAtlassianConfig_YAMLMarshal(t *testing.T) {
	ac := AtlassianConfig{
		URL:   "https://test.atlassian.net",
		Email: "test@example.com",
		CustomFields: CustomFieldConfig{
			Sprint: "customfield_10020",
		},
	}

	data, err := yaml.Marshal(&ac)
	if err != nil {
		t.Fatalf("failed to marshal AtlassianConfig: %v", err)
	}

	// Unmarshal back to verify round-trip
	var ac2 AtlassianConfig
	if err := yaml.Unmarshal(data, &ac2); err != nil {
		t.Fatalf("failed to unmarshal marshaled data: %v", err)
	}

	if ac2.URL != ac.URL {
		t.Errorf("round-trip URL = %q, want %q", ac2.URL, ac.URL)
	}
	if ac2.Email != ac.Email {
		t.Errorf("round-trip Email = %q, want %q", ac2.Email, ac.Email)
	}
	if ac2.CustomFields.Sprint != ac.CustomFields.Sprint {
		t.Errorf("round-trip CustomFields.Sprint = %q, want %q", ac2.CustomFields.Sprint, ac.CustomFields.Sprint)
	}
}

// --- Config YAML round-trip tests ---

func TestConfig_YAMLUnmarshal_UnifiedAtlassian(t *testing.T) {
	yamlData := `
atlassian:
  url: "https://company.atlassian.net"
  email: "admin@company.com"
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"
projects:
  - key: "PROJ"
    name: "My Project"
team:
  - name: "John Doe"
    email: "john@company.com"
`
	var cfg Config
	if err := yaml.Unmarshal([]byte(yamlData), &cfg); err != nil {
		t.Fatalf("failed to unmarshal Config: %v", err)
	}

	if cfg.Atlassian.URL != "https://company.atlassian.net" {
		t.Errorf("Atlassian.URL = %q, want %q", cfg.Atlassian.URL, "https://company.atlassian.net")
	}
	if cfg.Atlassian.Email != "admin@company.com" {
		t.Errorf("Atlassian.Email = %q, want %q", cfg.Atlassian.Email, "admin@company.com")
	}
	if cfg.Atlassian.CustomFields.Sprint != "customfield_10020" {
		t.Errorf("Atlassian.CustomFields.Sprint = %q, want %q", cfg.Atlassian.CustomFields.Sprint, "customfield_10020")
	}
	if cfg.Atlassian.CustomFields.Epic != "customfield_10014" {
		t.Errorf("Atlassian.CustomFields.Epic = %q, want %q", cfg.Atlassian.CustomFields.Epic, "customfield_10014")
	}
	if len(cfg.Projects) != 1 {
		t.Errorf("len(Projects) = %d, want 1", len(cfg.Projects))
	}
	if len(cfg.Team) != 1 {
		t.Errorf("len(Team) = %d, want 1", len(cfg.Team))
	}
}

func TestConfig_YAMLMarshalRoundTrip(t *testing.T) {
	cfg := Config{
		Atlassian: AtlassianConfig{
			URL:   "https://test.atlassian.net",
			Email: "user@test.com",
			CustomFields: CustomFieldConfig{
				Sprint:      "customfield_10020",
				Epic:        "customfield_10014",
				StoryPoints: "customfield_10028",
			},
		},
		Projects: []Project{{Key: "TEST", Name: "Test Project"}},
		Team:     []TeamMember{{Name: "Test User", Email: "test@example.com"}},
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal Config: %v", err)
	}

	var cfg2 Config
	if err := yaml.Unmarshal(data, &cfg2); err != nil {
		t.Fatalf("failed to unmarshal Config: %v", err)
	}

	if cfg2.Atlassian.URL != cfg.Atlassian.URL {
		t.Errorf("round-trip Atlassian.URL = %q, want %q", cfg2.Atlassian.URL, cfg.Atlassian.URL)
	}
	if cfg2.Atlassian.Email != cfg.Atlassian.Email {
		t.Errorf("round-trip Atlassian.Email = %q, want %q", cfg2.Atlassian.Email, cfg.Atlassian.Email)
	}
	if cfg2.Atlassian.CustomFields.Sprint != cfg.Atlassian.CustomFields.Sprint {
		t.Errorf("round-trip CustomFields.Sprint = %q, want %q", cfg2.Atlassian.CustomFields.Sprint, cfg.Atlassian.CustomFields.Sprint)
	}
}

// --- CustomFieldConfig tests ---

func TestCustomFieldConfig_YAMLUnmarshal(t *testing.T) {
	yamlData := `
sprint: "customfield_10020"
epic: "customfield_10014"
story_points: "customfield_10028"
`
	var cfc CustomFieldConfig
	if err := yaml.Unmarshal([]byte(yamlData), &cfc); err != nil {
		t.Fatalf("failed to unmarshal CustomFieldConfig: %v", err)
	}

	if cfc.Sprint != "customfield_10020" {
		t.Errorf("Sprint = %q, want %q", cfc.Sprint, "customfield_10020")
	}
	if cfc.Epic != "customfield_10014" {
		t.Errorf("Epic = %q, want %q", cfc.Epic, "customfield_10014")
	}
	if cfc.StoryPoints != "customfield_10028" {
		t.Errorf("StoryPoints = %q, want %q", cfc.StoryPoints, "customfield_10028")
	}
}

func TestCustomFieldConfig_EmptyIsValid(t *testing.T) {
	var cfc CustomFieldConfig
	if cfc.Sprint != "" || cfc.Epic != "" || cfc.StoryPoints != "" {
		t.Error("zero-value CustomFieldConfig should have all empty fields")
	}
}
