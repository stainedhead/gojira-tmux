# Feature: Security & Team Member Updates

**Document Version:** 1.0
**Date:** 2026-02-09
**Status:** Draft
**Owner:** Development Team

---

## Executive Summary

This feature removes the Okta OAuth authentication flow and replaces it with Atlassian API token-based authentication, aligning with Jira Cloud's recommended authentication method. Additionally, it adds alias support for team members to enable shorter, disambiguated references when multiple team members share the same first name.

**Key Changes:**
1. Replace Okta OIDC flow with Atlassian API token authentication
2. Add alias field to team member configuration
3. Enhance team member matching to support both name and alias
4. Simplify authentication architecture by removing OAuth complexity
5. Update configuration schema and validation

**Impact:** Breaking change requiring configuration updates and token regeneration for all users.

---

## Problem Statement

### Current Issues

1. **Wrong Authentication Method**: The application uses Okta OIDC for user authentication, but Jira API access should use Atlassian API tokens as documented at https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/

2. **Complexity Overhead**: OAuth OIDC flow introduces unnecessary complexity:
   - Callback server running on localhost
   - PKCE implementation
   - Refresh token management
   - Session expiry tracking (8-hour sessions)
   - Browser-based authentication flow

3. **Team Member Ambiguity**: When multiple team members share common first names (e.g., "John Anderson" and "John Flanagan"), there's no short, unambiguous way to reference them in filters or queries.

### Security Implications

- **Current State**: Using OAuth for user auth + separate Jira tokens (mixing two auth mechanisms)
- **Desired State**: Single, straightforward Atlassian API token authentication
- **Benefits**:
  - Simpler credential management
  - Fewer attack surfaces (no callback server, no OAuth state)
  - Aligned with Atlassian's recommendations
  - Token-level revocation control

---

## Solution Overview

### Authentication Simplification

Replace the multi-step OAuth flow with direct API token authentication:

**Before (Okta OIDC):**
```
User → Browser Auth → Okta → Callback Server → Token Exchange → Session Management → Jira API
```

**After (Atlassian Token):**
```
User → API Token Input → Secure Storage → Jira API
```

### Atlassian API Token Authentication

Per Atlassian documentation:
- **Token Generation**: Users create tokens at https://id.atlassian.com/manage/api-tokens
- **Authentication Method**: HTTP Basic Auth with `email:api_token`
- **Header Format**: `Authorization: Basic [base64(email:api_token)]`
- **Security Features**:
  - Individual token revocation
  - MFA compatible
  - No password exposure

### Team Member Aliases

Add optional `alias` field to team members for short, unique identifiers:

```yaml
team:
  - name: "John Anderson"
    email: "john.anderson@company.com"
    alias: "JohnA"
  - name: "John Flanagan"
    email: "john.flanagan@company.com"
    alias: "JohnF"
```

**Matching Logic**: When filtering or searching, try:
1. Exact alias match (case-sensitive)
2. Exact name match (case-sensitive)
3. Case-insensitive alias match
4. Case-insensitive name match (fallback)

---

## Functional Requirements

### FR-1: Atlassian API Token Collection

**Priority:** P0 (Critical)

**Description:** Collect and validate Atlassian API token on first run or when token is missing.

**Acceptance Criteria:**
- [ ] FR-1.1: On first run, display setup screen requesting email and API token
- [ ] FR-1.2: Provide clear instructions with link to https://id.atlassian.com/manage/api-tokens
- [ ] FR-1.3: Validate token format (non-empty, trimmed)
- [ ] FR-1.4: Test token by making authenticated request to Jira API
- [ ] FR-1.5: Store token securely in OS keychain/keyring
- [ ] FR-1.6: Show setup screen when stored token is missing or invalid
- [ ] FR-1.7: Allow token regeneration via config or CLI flag

**User Flow:**
1. User launches app for first time
2. App detects no stored token
3. Setup screen displays:
   ```
   ╔════════════════════════════════════════════════╗
   ║  Atlassian API Token Setup                     ║
   ╠════════════════════════════════════════════════╣
   ║                                                 ║
   ║  Email: [your-email@company.com            ]   ║
   ║  Token: [****************************      ]   ║
   ║                                                 ║
   ║  Get your token at:                            ║
   ║  https://id.atlassian.com/manage/api-tokens    ║
   ║                                                 ║
   ║  [Save]  [Cancel]                              ║
   ╚════════════════════════════════════════════════╝
   ```
4. User enters email and token
5. App validates and stores credentials
6. App proceeds to main screen

### FR-2: Remove Okta Authentication

**Priority:** P0 (Critical)

**Description:** Completely remove Okta OIDC authentication flow and all related code.

**Acceptance Criteria:**
- [ ] FR-2.1: Remove all Okta-specific configuration options
- [ ] FR-2.2: Remove OAuth callback server
- [ ] FR-2.3: Remove session management (no longer needed)
- [ ] FR-2.4: Remove refresh token logic
- [ ] FR-2.5: Remove browser-based auth flow
- [ ] FR-2.6: Remove Okta-specific dependencies (go-oidc, oauth2)

### FR-3: Configuration Schema Updates

**Priority:** P0 (Critical)

**Description:** Update configuration file schema to support new authentication and team aliases.

**Acceptance Criteria:**
- [ ] FR-3.1: Remove `okta` configuration block
- [ ] FR-3.2: Add `atlassian` configuration block with `email` field
- [ ] FR-3.3: Add optional `alias` field to team members
- [ ] FR-3.4: Validate alias uniqueness within team
- [ ] FR-3.5: Validate alias format (alphanumeric, no spaces)
- [ ] FR-3.6: Maintain backward compatibility for team members without aliases
- [ ] FR-3.7: Update config validation with helpful error messages

**New Schema:**
```yaml
jira:
  url: "https://your-company.atlassian.net"
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"
    story_points: "customfield_10016"

atlassian:
  email: "your-email@company.com"
  # Note: API token stored securely in keychain, not in config file

projects:
  - key: "PROJ"
    name: "My Project"
  - key: "INFRA"
    name: "Infrastructure"

team:
  - name: "John Anderson"
    email: "john.anderson@company.com"
    alias: "JohnA"
  - name: "John Flanagan"
    email: "john.flanagan@company.com"
    alias: "JohnF"
  - name: "Sarah Wilson"
    email: "sarah.wilson@company.com"
    alias: "SarahW"
```

### FR-4: Team Member Alias Matching

**Priority:** P1 (High)

**Description:** Support both name and alias when filtering or matching team members.

**Acceptance Criteria:**
- [ ] FR-4.1: Match by exact alias (case-sensitive) first
- [ ] FR-4.2: Fall back to exact name match if alias not found
- [ ] FR-4.3: Support case-insensitive matching as last resort
- [ ] FR-4.4: Display both name and alias in UI where space permits
- [ ] FR-4.5: Update JQL query builder to use email regardless of match method
- [ ] FR-4.6: Maintain performance (no regex, simple string comparison)

**Matching Algorithm:**
```go
func FindTeamMember(identifier string, team []TeamMember) *TeamMember {
    // Priority 1: Exact alias match (case-sensitive)
    for _, m := range team {
        if m.Alias == identifier {
            return &m
        }
    }

    // Priority 2: Exact name match (case-sensitive)
    for _, m := range team {
        if m.Name == identifier {
            return &m
        }
    }

    // Priority 3: Case-insensitive alias match
    lower := strings.ToLower(identifier)
    for _, m := range team {
        if strings.ToLower(m.Alias) == lower {
            return &m
        }
    }

    // Priority 4: Case-insensitive name match (fallback)
    for _, m := range team {
        if strings.ToLower(m.Name) == lower {
            return &m
        }
    }

    return nil
}
```

### FR-5: Token Validation

**Priority:** P0 (Critical)

**Description:** Validate API token on app startup and handle invalid tokens gracefully.

**Acceptance Criteria:**
- [ ] FR-5.1: Test token with lightweight Jira API call on startup
- [ ] FR-5.2: Handle 401 Unauthorized by requesting new token
- [ ] FR-5.3: Handle 403 Forbidden with clear permission error
- [ ] FR-5.4: Display helpful error messages for common failures
- [ ] FR-5.5: Support token refresh without app restart

**Validation API Call:**
```
GET /rest/api/2/myself
```
This lightweight endpoint confirms authentication without fetching large datasets.

---

## Technical Requirements

### TR-1: Domain Model Changes

**File:** `internal/domain/ports.go`

**Changes Required:**
1. Simplify `AuthPort` interface to remove OAuth methods
2. Add `AtlassianConfig` struct
3. Remove `OktaConfig` struct
4. Update `Config` struct

**New AuthPort Interface:**
```go
// AuthPort defines the interface for authentication operations.
type AuthPort interface {
    // ValidateToken validates the Atlassian API token by making a test request.
    // Returns user email if valid, error otherwise.
    ValidateToken(ctx context.Context, email, token string) (string, error)

    // IsTokenValid checks if the stored token is valid.
    // Returns true if token exists and passes validation.
    IsTokenValid(ctx context.Context) bool
}
```

**New Config Structs:**
```go
// Config represents the application configuration.
type Config struct {
    Jira       JiraConfig       `yaml:"jira"`
    Atlassian  AtlassianConfig  `yaml:"atlassian"`
    Projects   []Project        `yaml:"projects"`
    Team       []TeamMember     `yaml:"team"`
}

// AtlassianConfig holds Atlassian-specific configuration.
type AtlassianConfig struct {
    Email string `yaml:"email"`
    // Note: API token stored in keychain, not here
}
```

**Removed:**
- `OktaConfig` struct
- `AuthPort.StartAuthFlow()`
- `AuthPort.WaitForCallback()`
- `AuthPort.CancelAuthFlow()`
- `AuthPort.RefreshSession()`
- `AuthPort.IsSessionValid()`
- `AuthPort.Logout()`

### TR-2: Team Member Model Updates

**File:** `internal/domain/team_member.go`

**Changes Required:**
1. Add `Alias` field
2. Update validation to check alias format and uniqueness
3. Add helper methods for matching

**Updated TeamMember:**
```go
// TeamMember represents a team member for filtering and display.
type TeamMember struct {
    Name  string `yaml:"name" json:"name"`
    Email string `yaml:"email" json:"email"`
    Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`
}

// Validate checks that the TeamMember has valid data.
func (t *TeamMember) Validate() error {
    if t.Name == "" {
        return errors.New("team member name is required")
    }
    if t.Email == "" {
        return errors.New("team member email is required")
    }
    if !isValidEmail(t.Email) {
        return errors.New("team member email is invalid")
    }
    if t.Alias != "" && !isValidAlias(t.Alias) {
        return errors.New("team member alias must be alphanumeric (no spaces)")
    }
    return nil
}

// isValidAlias checks if alias contains only alphanumeric characters.
func isValidAlias(alias string) bool {
    if alias == "" {
        return true // empty is valid (optional field)
    }
    for _, r := range alias {
        if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
            return false
        }
    }
    return true
}

// MatchesIdentifier returns true if the identifier matches name or alias.
func (t *TeamMember) MatchesIdentifier(identifier string) bool {
    // Exact matches (case-sensitive)
    if t.Alias == identifier || t.Name == identifier {
        return true
    }
    // Case-insensitive fallback
    lower := strings.ToLower(identifier)
    return strings.ToLower(t.Alias) == lower || strings.ToLower(t.Name) == lower
}

// DisplayName returns the best display name (with alias if available).
func (t *TeamMember) DisplayName() string {
    if t.Alias != "" {
        return fmt.Sprintf("%s (%s)", t.Name, t.Alias)
    }
    return t.Name
}
```

### TR-3: Authentication Adapter Changes

**Files to DELETE:**
- `internal/adapter/auth/okta.go`
- `internal/adapter/auth/okta_test.go`
- `internal/adapter/auth/callback.go`
- `internal/adapter/auth/callback_test.go`

**Files to CREATE:**
- `internal/adapter/auth/atlassian.go`
- `internal/adapter/auth/atlassian_test.go`

**New Atlassian Adapter:**
```go
package auth

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/stainedhead/gojira-tmux/internal/domain"
)

// AtlassianAdapter implements the AuthPort interface for Atlassian API token authentication.
type AtlassianAdapter struct {
    tokenStore domain.TokenStorePort
    httpClient *http.Client
}

// NewAtlassianAdapter creates a new Atlassian adapter.
func NewAtlassianAdapter(tokenStore domain.TokenStorePort) *AtlassianAdapter {
    return &AtlassianAdapter{
        tokenStore: tokenStore,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// ValidateToken validates the Atlassian API token by calling /rest/api/2/myself.
func (a *AtlassianAdapter) ValidateToken(ctx context.Context, email, token string) (string, error) {
    // Create test request to Jira API
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/rest/api/2/myself", nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    // Set Basic Auth header
    auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
    req.Header.Set("Authorization", "Basic "+auth)
    req.Header.Set("Accept", "application/json")

    // Execute request
    resp, err := a.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to validate token: %w", err)
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode == http.StatusUnauthorized {
        return "", errors.New("invalid token or email")
    }
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    // Parse response to get email
    var result struct {
        EmailAddress string `json:"emailAddress"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    return result.EmailAddress, nil
}

// IsTokenValid checks if the stored token is valid.
func (a *AtlassianAdapter) IsTokenValid(ctx context.Context) bool {
    token, err := a.tokenStore.GetJiraToken()
    if err != nil || token == "" {
        return false
    }
    // Could validate here, but for performance we trust the stored token
    // until it fails in actual use
    return true
}

// Ensure AtlassianAdapter implements domain.AuthPort.
var _ domain.AuthPort = (*AtlassianAdapter)(nil)
```

### TR-4: Configuration Loader Updates

**File:** `internal/adapter/config/config.go`

**Changes Required:**
1. Update validation to check Atlassian config instead of Okta
2. Add team alias uniqueness validation
3. Update error messages

**Updated Validation:**
```go
// validate checks that the configuration is valid.
func (l *Loader) validate(cfg *domain.Config) error {
    // Jira validation
    if cfg.Jira.URL == "" {
        return errors.New("jira.url is required")
    }
    if !strings.HasPrefix(cfg.Jira.URL, "https://") {
        return errors.New("jira.url must use HTTPS")
    }

    // Atlassian validation
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

    // Check alias uniqueness
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

// Remove ValidateUserAccess - no longer needed without Okta
```

### TR-5: Use Case Updates

**File:** `internal/usecase/authenticate.go`

**Changes Required:**
1. Simplify to just token setup and validation
2. Remove OAuth flow methods
3. Remove team membership validation (handled at config level)

**Updated Authenticate Use Case:**
```go
package usecase

import (
    "context"
    "errors"
    "strings"

    "github.com/stainedhead/gojira-tmux/internal/domain"
)

// Authenticate handles API token validation.
type Authenticate struct {
    authPort   domain.AuthPort
    tokenStore domain.TokenStorePort
}

// NewAuthenticate creates a new Authenticate use case.
func NewAuthenticate(authPort domain.AuthPort, tokenStore domain.TokenStorePort) *Authenticate {
    return &Authenticate{
        authPort:   authPort,
        tokenStore: tokenStore,
    }
}

// ValidateAndSaveToken validates the token and saves it if valid.
func (a *Authenticate) ValidateAndSaveToken(ctx context.Context, email, token string) error {
    // Trim whitespace
    email = strings.TrimSpace(email)
    token = strings.TrimSpace(token)

    if email == "" || token == "" {
        return errors.New("email and token are required")
    }

    // Validate token
    validatedEmail, err := a.authPort.ValidateToken(ctx, email, token)
    if err != nil {
        return fmt.Errorf("token validation failed: %w", err)
    }

    // Verify email matches
    if !strings.EqualFold(validatedEmail, email) {
        return fmt.Errorf("token email mismatch: expected %s, got %s", email, validatedEmail)
    }

    // Save token
    if err := a.tokenStore.SetJiraToken(token); err != nil {
        return fmt.Errorf("failed to save token: %w", err)
    }

    return nil
}

// HasValidToken checks if a valid token exists.
func (a *Authenticate) HasValidToken(ctx context.Context) bool {
    return a.authPort.IsTokenValid(ctx)
}

// ClearToken removes the stored token.
func (a *Authenticate) ClearToken() error {
    return a.tokenStore.DeleteJiraToken()
}
```

**File:** `internal/usecase/setup_token.go`

**Changes Required:**
- Integrate email validation
- Update to work with new auth flow

### TR-6: JQL Builder Enhancement

**File:** `internal/adapter/jira/search.go`

**Changes Required:**
1. Update `buildAssigneeCondition` to use new team member matching
2. Support both name and alias in filters

**Updated buildAssigneeCondition:**
```go
// buildAssigneeCondition builds the assignee filter condition.
func (b *JQLBuilder) buildAssigneeCondition(identifier string) string {
    if identifier == "" || identifier == "-All-" {
        return ""
    }

    // Find team member by name or alias
    var member *domain.TeamMember
    for i := range b.team {
        if b.team[i].MatchesIdentifier(identifier) {
            member = &b.team[i]
            break
        }
    }

    if member == nil {
        return ""
    }

    return fmt.Sprintf(`assignee = %s`, escapeJQL(member.Email))
}
```

### TR-7: Jira Client Updates

**File:** `internal/adapter/jira/client.go`

**Changes Required:**
- No changes needed! Already uses Basic Auth with username:token
- Continue using `req.SetBasicAuth(c.username, c.token)` at line 51

**Verification:**
The Jira client already implements Atlassian's recommended authentication:
```go
req.SetBasicAuth(c.username, c.token)  // Line 51 in client.go
```
This is exactly what Atlassian requires: `email:api_token` in Basic Auth header.

### TR-8: Main Entry Point Updates

**File:** `cmd/gojira/main.go`

**Changes Required:**
1. Replace OktaAdapter with AtlassianAdapter
2. Update to use email from config
3. Simplify initialization (no OAuth provider)

**Updated run() function:**
```go
func run() error {
    // Determine config path
    configPath := getConfigPath()

    // Load configuration
    configLoader := config.NewLoader(configPath)
    cfg, err := configLoader.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Initialize token store
    credentialsPath := getCredentialsPath()
    tokenStore := auth.NewKeyringTokenStore(credentialsPath)

    // Check for environment token (for testing/CI)
    if jiraToken := os.Getenv("JIRA_API_TOKEN"); jiraToken != "" {
        if err := tokenStore.SetJiraToken(jiraToken); err != nil {
            return fmt.Errorf("failed to store token: %w", err)
        }
    }

    // Initialize Atlassian adapter (simplified - no OAuth)
    atlassianAdapter := auth.NewAtlassianAdapter(tokenStore)

    // Get stored token
    storedToken, _ := tokenStore.GetJiraToken()

    // Initialize Jira client with email from config
    jiraClient := jira.NewClient(
        cfg.Jira.URL,
        cfg.Atlassian.Email,  // Changed from cfg.Jira.Username
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
        tui.WithAtlassianEmail(cfg.Atlassian.Email),  // New option
    )

    // Run TUI
    p := tea.NewProgram(app, tea.WithAltScreen())
    _, err = p.Run()
    return err
}
```

### TR-9: TUI Updates

**Files to Update:**
- `internal/infrastructure/tui/app.go` - Update initialization
- `internal/infrastructure/tui/setup_screen.go` - Replace OAuth flow with token input
- `internal/infrastructure/tui/login_screen.go` - Simplify or remove
- `internal/infrastructure/tui/filter_bar.go` - Support alias display

**Key TUI Changes:**
1. **Setup Screen**: Replace browser-based OAuth with email + token input form
2. **Login Screen**: May be removable - no "login" flow, just token validation
3. **Filter Bar**: Show aliases in team member dropdown when available

### TR-10: Dependency Updates

**File:** `go.mod`

**Dependencies to REMOVE:**
```
github.com/coreos/go-oidc/v3 v3.x.x
golang.org/x/oauth2 v0.x.x
```

**Command:**
```bash
go mod tidy
```

### TR-11: Token Store Updates

**File:** `internal/adapter/auth/token_store.go`

**Changes Required:**
- Remove `GetRefreshToken()`, `SetRefreshToken()`, `DeleteRefreshToken()` methods
- Keep Jira token methods unchanged

**Updated TokenStorePort (in ports.go):**
```go
// TokenStorePort defines the interface for secure credential storage.
type TokenStorePort interface {
    // GetJiraToken retrieves the stored Jira API token.
    GetJiraToken() (string, error)

    // SetJiraToken stores the Jira API token securely.
    SetJiraToken(token string) error

    // DeleteJiraToken removes the stored Jira API token.
    DeleteJiraToken() error

    // HasJiraToken returns true if a Jira token exists.
    HasJiraToken() bool
}
```

---

## Implementation Plan

### Phase 1: Domain Model Updates (No Dependencies)

**Workers:** Can run in parallel

#### Task 1.1: Update Domain Ports
**File:** `internal/domain/ports.go`
**Estimated Time:** 30 minutes
**Dependencies:** None
**Assigned To:** Worker A

**Steps:**
1. Remove `OktaConfig` struct (lines 121-127)
2. Add `AtlassianConfig` struct:
   ```go
   type AtlassianConfig struct {
       Email string `yaml:"email"`
   }
   ```
3. Update `Config` struct - replace `Okta OktaConfig` with `Atlassian AtlassianConfig`
4. Simplify `AuthPort` interface - remove all methods except token validation
5. Remove refresh token methods from `TokenStorePort`
6. Update comments to reflect new auth model

**Test:** Compile check `go build ./internal/domain`

#### Task 1.2: Update Team Member Model
**File:** `internal/domain/team_member.go`
**Estimated Time:** 45 minutes
**Dependencies:** None
**Assigned To:** Worker B

**Steps:**
1. Add `Alias` field to `TeamMember` struct
2. Add `isValidAlias()` helper function
3. Update `Validate()` to check alias format
4. Add `MatchesIdentifier(identifier string) bool` method
5. Add `DisplayName() string` method
6. Update tests in `team_member_test.go`:
   - Test valid aliases
   - Test invalid aliases (with spaces, special chars)
   - Test matching logic (exact, case-insensitive)
   - Test display name formatting

**Test:** `go test ./internal/domain -run TestTeamMember`

#### Task 1.3: Update User Model
**File:** `internal/domain/user.go`
**Estimated Time:** 15 minutes
**Dependencies:** None
**Assigned To:** Worker B (after 1.2)

**Steps:**
1. Remove `SessionExpiry` field (no longer needed)
2. Remove `IsSessionValid()` method
3. Remove `ValidateTeamMembership()` method (handled in config)
4. Simplify to just email validation
5. Update tests

**Test:** `go test ./internal/domain -run TestUser`

### Phase 2: Adapter Layer (Depends on Phase 1)

**Workers:** Can run in parallel after Phase 1 completes

#### Task 2.1: Create Atlassian Auth Adapter
**Files:** `internal/adapter/auth/atlassian.go`, `atlassian_test.go`
**Estimated Time:** 2 hours
**Dependencies:** Task 1.1
**Assigned To:** Worker C

**Steps:**
1. Create `AtlassianAdapter` struct
2. Implement `ValidateToken(ctx, email, token)` method:
   - Build Basic Auth header
   - Call `/rest/api/2/myself` endpoint
   - Parse response
   - Handle errors (401, 403, network)
3. Implement `IsTokenValid(ctx)` method
4. Write comprehensive tests:
   - Test valid token
   - Test invalid token (401)
   - Test forbidden (403)
   - Test network errors
   - Test malformed responses
   - Mock HTTP client for testing

**Test:** `go test ./internal/adapter/auth -run TestAtlassian`

#### Task 2.2: Delete Okta Components
**Files:** `okta.go`, `okta_test.go`, `callback.go`, `callback_test.go`
**Estimated Time:** 10 minutes
**Dependencies:** Task 2.1 (ensure replacement exists)
**Assigned To:** Worker C (after 2.1)

**Steps:**
1. Delete `internal/adapter/auth/okta.go`
2. Delete `internal/adapter/auth/okta_test.go`
3. Delete `internal/adapter/auth/callback.go`
4. Delete `internal/adapter/auth/callback_test.go`
5. Run `go mod tidy` to remove unused dependencies

**Test:** `go build ./...` (verify no import errors)

#### Task 2.3: Update Config Loader
**File:** `internal/adapter/config/config.go`
**Estimated Time:** 1 hour
**Dependencies:** Task 1.1, 1.2
**Assigned To:** Worker D

**Steps:**
1. Update `validate()` method:
   - Remove Okta validation (lines 60-69)
   - Add Atlassian email validation
   - Add team alias uniqueness check
   - Add duplicate email check
2. Remove `ValidateUserAccess()` method
3. Update tests in `config_test.go`:
   - Test Atlassian config validation
   - Test alias uniqueness validation
   - Test duplicate alias error
   - Test missing email error
   - Test invalid email format
   - Remove Okta-related tests

**Test:** `go test ./internal/adapter/config`

#### Task 2.4: Update JQL Builder
**File:** `internal/adapter/jira/search.go`
**Estimated Time:** 45 minutes
**Dependencies:** Task 1.2
**Assigned To:** Worker E

**Steps:**
1. Update `buildAssigneeCondition()` method:
   - Replace direct name matching with `MatchesIdentifier()`
   - Iterate through team members
   - Use email from matched member
2. Update tests in `search_test.go`:
   - Test matching by alias
   - Test matching by name
   - Test case-insensitive matching
   - Test no match returns empty condition

**Test:** `go test ./internal/adapter/jira -run TestJQLBuilder`

### Phase 3: Use Case Layer (Depends on Phase 2)

**Workers:** Can run in parallel after Phase 2 completes

#### Task 3.1: Simplify Authenticate Use Case
**File:** `internal/usecase/authenticate.go`
**Estimated Time:** 1 hour
**Dependencies:** Task 2.1
**Assigned To:** Worker F

**Steps:**
1. Remove all OAuth-related methods:
   - `StartLogin()`
   - `CompleteLogin()`
   - `CancelLogin()`
   - `CheckSession()`
   - `Logout()`
2. Keep only:
   - `ValidateAndSaveToken(ctx, email, token) error`
   - `HasValidToken(ctx) bool`
   - `ClearToken() error`
3. Update constructor to remove `configPort` dependency
4. Update tests in `authenticate_test.go`:
   - Test token validation success
   - Test token validation failure
   - Test token save
   - Test clear token
   - Remove all OAuth flow tests

**Test:** `go test ./internal/usecase -run TestAuthenticate`

#### Task 3.2: Update Setup Token Use Case
**File:** `internal/usecase/setup_token.go`
**Estimated Time:** 30 minutes
**Dependencies:** None (minimal changes)
**Assigned To:** Worker F (after 3.1)

**Steps:**
1. Update `SaveToken()` to validate email format if needed
2. Update tests
3. Ensure it works with new token store interface

**Test:** `go test ./internal/usecase -run TestSetupToken`

### Phase 4: Infrastructure Layer (Depends on Phase 3)

**Workers:** Sequential (TUI components depend on each other)

#### Task 4.1: Update Setup Screen
**File:** `internal/infrastructure/tui/setup_screen.go`
**Estimated Time:** 2 hours
**Dependencies:** Task 3.1, 3.2
**Assigned To:** Worker G

**Steps:**
1. Replace OAuth flow with simple form:
   - Email input field
   - Token input field (masked)
   - Instructions with link to token generation
2. Update `Init()` - no callback server
3. Update `Update()`:
   - Handle email input
   - Handle token input
   - Validate on submit
   - Show validation errors
4. Update `View()` - render form with instructions
5. Handle success/failure messages
6. Write tests for new flow

**Test:** Manual TUI testing + unit tests

#### Task 4.2: Update/Remove Login Screen
**File:** `internal/infrastructure/tui/login_screen.go`
**Estimated Time:** 1 hour
**Dependencies:** Task 4.1
**Assigned To:** Worker G (after 4.1)

**Steps:**
1. Evaluate if login screen is still needed
2. If not needed, remove file and references
3. If needed, simplify to just token validation screen
4. Update tests

**Test:** Manual TUI testing

#### Task 4.3: Update Filter Bar
**File:** `internal/infrastructure/tui/filter_bar.go`
**Estimated Time:** 1 hour
**Dependencies:** Task 1.2
**Assigned To:** Worker H

**Steps:**
1. Update team member display to show aliases:
   - Use `DisplayName()` method
   - Format: "John Anderson (JohnA)"
2. Support filtering by alias
3. Update dropdown rendering
4. Test alias display and filtering

**Test:** Manual TUI testing

#### Task 4.4: Update App Initialization
**File:** `internal/infrastructure/tui/app.go`
**Estimated Time:** 45 minutes
**Dependencies:** Task 4.1, 4.2
**Assigned To:** Worker G (after 4.2)

**Steps:**
1. Remove OAuth-related initialization
2. Update options to include Atlassian email
3. Simplify auth flow startup
4. Update state management
5. Update tests

**Test:** `go test ./internal/infrastructure/tui`

### Phase 5: Entry Point & Configuration (Depends on Phase 4)

**Workers:** Sequential (affects entire app)

#### Task 5.1: Update Main Entry Point
**File:** `cmd/gojira/main.go`
**Estimated Time:** 45 minutes
**Dependencies:** All previous tasks
**Assigned To:** Worker I

**Steps:**
1. Replace `NewOktaAdapter` with `NewAtlassianAdapter`
2. Update Jira client initialization to use `cfg.Atlassian.Email`
3. Add `WithAtlassianEmail()` option to app creation
4. Remove OAuth provider initialization
5. Test with environment variables
6. Test without environment variables

**Test:** Full integration test

#### Task 5.2: Update Configuration Examples
**Files:** Documentation, README
**Estimated Time:** 30 minutes
**Dependencies:** None (can run early)
**Assigned To:** Worker J

**Steps:**
1. Update README.md:
   - Remove Okta authentication section
   - Add Atlassian token section
   - Update config example
   - Update quick start guide
   - Update environment variables section
2. Create example config file:
   - `config.example.yaml`
   - Show Atlassian config
   - Show team with aliases
3. Create migration guide (see Migration section below)

**Test:** Documentation review

#### Task 5.3: Update Dependencies
**File:** `go.mod`, `go.sum`
**Estimated Time:** 15 minutes
**Dependencies:** Task 2.2
**Assigned To:** Worker I (after 5.1)

**Steps:**
1. Run `go mod tidy` to remove unused dependencies
2. Verify `go-oidc` and `oauth2` are removed
3. Check for any indirect dependencies that can be cleaned up
4. Run `go mod verify`

**Test:** `go build ./... && go test ./...`

### Phase 6: Testing & Documentation

#### Task 6.1: Integration Testing
**Estimated Time:** 3 hours
**Dependencies:** All previous tasks
**Assigned To:** Worker K

**Steps:**
1. Test complete flow:
   - First run (no token)
   - Token setup
   - Token validation
   - Main app with valid token
   - Alias filtering
   - Invalid token handling
2. Test error scenarios:
   - Invalid email
   - Invalid token
   - Network errors
   - Malformed config
   - Duplicate aliases
3. Test backward compatibility:
   - Team members without aliases
   - Old config migration

**Test:** Full E2E test suite

#### Task 6.2: Update Documentation
**Estimated Time:** 2 hours
**Dependencies:** Task 5.2
**Assigned To:** Worker J

**Steps:**
1. Update architecture documentation
2. Update API documentation
3. Create migration guide
4. Update troubleshooting guide
5. Update development setup docs

---

## Testing Strategy

### Unit Tests

**Coverage Target:** 85%

**Critical Test Areas:**
1. **Team Member Matching** (`team_member_test.go`):
   - Exact alias match
   - Exact name match
   - Case-insensitive matches
   - No match scenarios
   - Invalid alias formats
   - Display name formatting

2. **Config Validation** (`config_test.go`):
   - Atlassian email validation
   - Alias uniqueness
   - Duplicate email detection
   - Missing required fields
   - Invalid formats

3. **Token Validation** (`atlassian_test.go`):
   - Valid token response
   - Invalid credentials (401)
   - Forbidden access (403)
   - Network errors
   - Timeout handling
   - Malformed responses

4. **JQL Building** (`search_test.go`):
   - Alias-based filtering
   - Name-based filtering
   - Fallback matching
   - No match handling

### Integration Tests

**Test Scenarios:**
1. **First Run Flow**:
   - App starts without token
   - Setup screen displays
   - User enters email + token
   - Token validates successfully
   - App proceeds to main screen

2. **Invalid Token Flow**:
   - App starts with invalid token
   - Validation fails
   - Setup screen redisplays
   - User enters new token
   - App proceeds

3. **Alias Filtering**:
   - Load config with aliases
   - Filter by alias "JohnA"
   - Verify correct team member selected
   - Verify JQL uses correct email

4. **Backward Compatibility**:
   - Load config without aliases
   - Team members still work
   - Filtering by name still works

### Manual Testing Checklist

- [ ] First-time setup flow
- [ ] Token generation instructions clear
- [ ] Token input masked properly
- [ ] Validation errors display correctly
- [ ] Successful login transitions smoothly
- [ ] Filter bar shows aliases
- [ ] Filtering by alias works
- [ ] Filtering by name still works
- [ ] Invalid token prompts re-setup
- [ ] Network errors display helpful messages
- [ ] Config validation errors are clear
- [ ] Duplicate alias error is clear
- [ ] Token stored securely in keychain
- [ ] Token persists across app restarts
- [ ] Environment variable override works

---

## Migration Guide

### For Users

**Breaking Changes:**
1. Configuration file format changed
2. Okta authentication removed
3. API token required

**Migration Steps:**

#### Step 1: Generate Atlassian API Token
1. Visit https://id.atlassian.com/manage/api-tokens
2. Click "Create API token"
3. Label it "gojira-tmux"
4. Copy the generated token (save it securely)

#### Step 2: Update Configuration File

**Old config.yaml:**
```yaml
jira:
  url: "https://your-company.atlassian.net"
  username: "your-email@company.com"

okta:
  issuer: "https://your-company.okta.com/oauth2/default"
  client_id: "your-okta-client-id"
  callback_port: 8080
  scopes: ["openid", "profile", "email"]

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
```

**New config.yaml:**
```yaml
jira:
  url: "https://your-company.atlassian.net"

atlassian:
  email: "your-email@company.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
    alias: "JohnD"  # Optional but recommended
```

#### Step 3: Clear Old Credentials
```bash
# Remove old Okta tokens
rm -rf ~/.config/gojira/credentials
```

#### Step 4: Run Application
```bash
gojira
```

The app will prompt for your email and API token on first run.

#### Step 5: Add Aliases (Optional)
Edit `config.yaml` to add aliases for team members:
```yaml
team:
  - name: "John Anderson"
    email: "john.anderson@company.com"
    alias: "JohnA"
  - name: "John Flanagan"
    email: "john.flanagan@company.com"
    alias: "JohnF"
```

### For Developers

**Build Changes:**
1. Run `go mod tidy` after pulling updates
2. Verify dependencies removed: `go-oidc`, `oauth2`
3. Update any custom integrations using `AuthPort`

**Testing Changes:**
1. No more OAuth mock servers needed
2. Simpler HTTP mocking for token validation
3. Update config fixtures to use Atlassian format

---

## Security Considerations

### Token Storage

**Current:** OS keychain/keyring (unchanged)
- macOS: Keychain
- Linux: Secret Service API
- Windows: Credential Manager

**Best Practices:**
- Never log tokens
- Never commit tokens to version control
- Rotate tokens periodically
- Use separate tokens for different apps

### Token Validation

**On Startup:**
1. Load token from secure storage
2. Validate with lightweight API call (`/rest/api/2/myself`)
3. On failure, prompt for new token

**During Operation:**
1. Trust stored token until API returns 401
2. On 401, prompt for token re-entry
3. Clear invalid token from storage

### Error Messages

**Safe Error Messages:**
- "Invalid token or email" (don't specify which)
- "Authentication failed" (don't expose internals)
- "Network error - please try again"

**Avoid:**
- Exposing full error messages from Jira API
- Logging token values
- Displaying token in UI (except masked during input)

### CAPTCHA Handling

Per Atlassian docs, detect CAPTCHA lockout:
```
X-Seraph-LoginReason: AUTHENTICATION_DENIED
```

Display clear message:
```
"Too many failed login attempts. Please wait and try again, or reset your password on Atlassian."
```

---

## Performance Considerations

### Token Validation

**Current Approach:**
- Validate on startup: ~100-200ms API call
- Trust token during operation
- Revalidate on 401 errors

**Optimization:**
- Cache validation result for 1 hour
- Skip validation if token used successfully in last hour
- Background validation (non-blocking UI)

### Team Member Matching

**Algorithm Complexity:**
- O(n) where n = team size (typically < 20)
- 4 passes maximum (exact alias, exact name, case-insensitive)
- Early exit on first match
- No regex, no complex string operations

**Optimization:**
- Build alias → member map at startup
- O(1) lookup instead of O(n) iteration
- Rebuild map only on config reload

### JQL Query Building

**No Change:**
- Still O(n) for assignee lookup
- Acceptable for small teams
- Consider caching if team > 100

---

## Dependencies and Parallelization

### Dependency Graph

```
Phase 1: Domain Model Updates
├─ Task 1.1: Update Ports (Worker A) ────────┐
├─ Task 1.2: Update TeamMember (Worker B) ───┤
└─ Task 1.3: Update User (Worker B) ─────────┤
                                              │
                                              ▼
Phase 2: Adapter Layer                       │
├─ Task 2.1: Atlassian Adapter (Worker C) ◄──┤
├─ Task 2.2: Delete Okta (Worker C) ◄────────┤
├─ Task 2.3: Config Loader (Worker D) ◄──────┤
└─ Task 2.4: JQL Builder (Worker E) ◄────────┘
                   │
                   ▼
Phase 3: Use Case Layer
├─ Task 3.1: Authenticate (Worker F)
└─ Task 3.2: Setup Token (Worker F)
                   │
                   ▼
Phase 4: Infrastructure
├─ Task 4.1: Setup Screen (Worker G)
├─ Task 4.2: Login Screen (Worker G)
├─ Task 4.3: Filter Bar (Worker H)
└─ Task 4.4: App Init (Worker G)
                   │
                   ▼
Phase 5: Entry Point
├─ Task 5.1: Main (Worker I)
├─ Task 5.2: Docs (Worker J) ◄── Can start early
└─ Task 5.3: Dependencies (Worker I)
                   │
                   ▼
Phase 6: Testing
├─ Task 6.1: Integration Tests (Worker K)
└─ Task 6.2: Documentation (Worker J)
```

### Maximum Parallelization

**Phase 1:** 2 workers (A, B)
**Phase 2:** 4 workers (C, D, E, + J for docs)
**Phase 3:** 1 worker (F) - sequential use cases
**Phase 4:** 2 workers (G, H) - some TUI parallelism
**Phase 5:** 2 workers (I, J)
**Phase 6:** 2 workers (K, J)

### Critical Path

```
1.1 → 2.1 → 2.2 → 3.1 → 4.1 → 4.2 → 4.4 → 5.1 → 6.1
Total: ~12 hours (with parallelization)
```

### File Contention Matrix

| File | Phase 1 | Phase 2 | Phase 3 | Phase 4 | Phase 5 |
|------|---------|---------|---------|---------|---------|
| ports.go | W-A | R | R | R | R |
| team_member.go | W-B | R | R | R | R |
| user.go | W-B | R | R | R | R |
| atlassian.go | - | W-C | R | R | R |
| config.go | - | W-D | R | R | R |
| search.go | - | W-E | R | R | R |
| authenticate.go | - | - | W-F | R | R |
| setup_screen.go | - | - | - | W-G | R |
| main.go | - | - | - | - | W-I |

**Legend:** W-X = Write (Worker X), R = Read

**No Contention:** Each file has only one writer per phase.

---

## Rollback Plan

### If Critical Issues Found

**Immediate Rollback:**
1. Revert to previous git commit
2. Re-release previous version
3. Notify users via GitHub release notes

**Data Safety:**
- Old tokens in keychain remain (different keys)
- Config files can coexist (users rename back)
- No data loss risk

### Partial Rollback

**If only TUI issues:**
- Can rollback TUI components
- Keep domain/adapter changes
- Use CLI flags for token input temporarily

**If only token validation issues:**
- Can skip validation temporarily
- Trust stored tokens
- Fix validation logic separately

---

## Success Criteria

### Technical Success

- [ ] All unit tests passing (coverage > 85%)
- [ ] All integration tests passing
- [ ] No OAuth dependencies remaining
- [ ] App builds without errors
- [ ] App runs without errors
- [ ] Token validation working
- [ ] Alias matching working
- [ ] Config validation working
- [ ] Secure token storage working

### User Experience Success

- [ ] First-run setup < 2 minutes
- [ ] Token instructions clear and helpful
- [ ] Errors display actionable messages
- [ ] Aliases display in UI
- [ ] Filtering by alias works intuitively
- [ ] Migration guide complete and clear
- [ ] Documentation updated

### Performance Success

- [ ] App startup < 2 seconds (with valid token)
- [ ] Token validation < 500ms
- [ ] Team member matching < 1ms
- [ ] No noticeable performance degradation

---

## Future Enhancements

### Out of Scope (For Later)

1. **Multiple Atlassian Accounts:**
   - Support switching between accounts
   - Profile management

2. **Token Rotation:**
   - Auto-detect expiring tokens
   - Guided token refresh flow

3. **Advanced Alias Features:**
   - Auto-generate aliases from names
   - Validate alias against Jira usernames

4. **Enhanced Security:**
   - Token encryption at rest
   - Biometric unlock support
   - Session timeout configuration

---

## Open Questions

### To Be Resolved Before Implementation

1. **Q:** Should we support environment variable for email in addition to config?
   **A:** TBD - discuss with team

2. **Q:** Should aliases be case-sensitive in config but case-insensitive for matching?
   **A:** Recommend case-sensitive storage, case-insensitive matching

3. **Q:** Maximum alias length?
   **A:** Recommend 20 characters

4. **Q:** Should we validate aliases against Jira usernames?
   **A:** No - keep aliases as UI-only feature

5. **Q:** Should we support alias auto-generation?
   **A:** No - require manual configuration

6. **Q:** Deprecation timeline for old config format?
   **A:** Immediate (breaking change in next major version)

---

## Appendix A: API Reference

### Atlassian API Endpoints

**Token Validation:**
```
GET /rest/api/2/myself
Authorization: Basic base64(email:token)
```

**Response:**
```json
{
  "self": "https://your-domain.atlassian.net/rest/api/2/user?accountId=...",
  "accountId": "...",
  "emailAddress": "user@company.com",
  "displayName": "John Doe",
  "active": true
}
```

**Error Responses:**
- `401 Unauthorized`: Invalid email or token
- `403 Forbidden`: Valid credentials but insufficient permissions
- `429 Too Many Requests`: Rate limited
- `X-Seraph-LoginReason: AUTHENTICATION_DENIED`: CAPTCHA triggered

### Token Generation

**URL:** https://id.atlassian.com/manage/api-tokens

**Steps:**
1. Log in to Atlassian account
2. Click "Create API token"
3. Enter label (e.g., "gojira-tmux")
4. Copy token (shown only once)
5. Store securely

**Token Properties:**
- Format: Alphanumeric string (exact format undocumented)
- Length: Variable (typically 50-100 characters)
- Expiry: No automatic expiry (user-revocable)
- Scope: Full API access (same as password)

---

## Appendix B: Configuration Schema

### Full Schema Definition

```yaml
# Required: Jira instance configuration
jira:
  # Required: Jira Cloud URL (must be HTTPS)
  url: string

  # Optional: Custom field mappings
  custom_fields:
    sprint: string       # e.g., "customfield_10020"
    epic: string         # e.g., "customfield_10014"
    story_points: string # e.g., "customfield_10016"

# Required: Atlassian account configuration
atlassian:
  # Required: Email address for API authentication
  # Must match the email of the Atlassian account
  email: string

# Required: List of Jira projects to query (minimum 1)
projects:
  - key: string   # Required: Project key (uppercase letters)
    name: string  # Required: Display name

# Required: Team members for filtering (minimum 1)
team:
  - name: string   # Required: Full display name
    email: string  # Required: Jira email address
    alias: string  # Optional: Short unique identifier
```

### Validation Rules

**jira.url:**
- Must start with "https://"
- Must be valid URL format
- Typically: `https://*.atlassian.net`

**atlassian.email:**
- Must be valid email format
- Must match Jira user account email

**projects[]:**
- Minimum 1 project required
- `key`: Must be uppercase letters only (regex: `^[A-Z]+$`)
- `name`: Non-empty string

**team[]:**
- Minimum 1 member required
- `name`: Non-empty string
- `email`: Valid email format, unique within team
- `alias`: Optional, alphanumeric only, unique within team

---

## Appendix C: Error Messages

### User-Facing Error Messages

**Configuration Errors:**
```
❌ Configuration Error

atlassian.email is required
→ Add your Atlassian email to config.yaml

Example:
  atlassian:
    email: "you@company.com"
```

```
❌ Configuration Error

Duplicate alias "JohnA" found
→ Each team member must have a unique alias
→ Found in: John Anderson, John Andrews
```

**Authentication Errors:**
```
❌ Authentication Failed

Invalid email or API token
→ Verify your credentials at:
   https://id.atlassian.com/manage/api-tokens

→ Check your email matches your Atlassian account
```

```
❌ Too Many Failed Attempts

Your account has been temporarily locked
→ Wait 10 minutes and try again
→ Or reset your password at Atlassian
```

**Network Errors:**
```
❌ Connection Failed

Could not reach Jira API
→ Check your internet connection
→ Verify Jira URL in config: https://your-domain.atlassian.net
```

---

## Appendix D: Code Examples

### Example 1: Token Validation Flow

```go
// In TUI setup screen Update() method
case tea.KeyEnter:
    if m.emailInput.Focused() {
        m.emailInput.Blur()
        m.tokenInput.Focus()
        return m, nil
    }

    if m.tokenInput.Focused() {
        email := strings.TrimSpace(m.emailInput.Value())
        token := strings.TrimSpace(m.tokenInput.Value())

        if email == "" || token == "" {
            m.error = "Email and token are required"
            return m, nil
        }

        // Validate and save
        return m, func() tea.Msg {
            err := m.authUseCase.ValidateAndSaveToken(context.Background(), email, token)
            if err != nil {
                return SetupErrorMsg{err}
            }
            return SetupCompleteMsg{}
        }
    }
```

### Example 2: Team Member Lookup with Alias

```go
// In filter bar Update() method
func (m *FilterBar) selectTeamMember(identifier string) {
    // Find team member by alias or name
    for _, member := range m.teamMembers {
        if member.MatchesIdentifier(identifier) {
            m.selectedMember = member
            m.filterChanged = true
            return
        }
    }

    // Not found - clear selection
    m.selectedMember = nil
}
```

### Example 3: Config Validation

```go
// In config loader
func (l *Loader) validateAliases(team []domain.TeamMember) error {
    seen := make(map[string]string) // alias -> name

    for _, member := range team {
        if member.Alias == "" {
            continue // optional field
        }

        if existingName, exists := seen[member.Alias]; exists {
            return fmt.Errorf(
                "duplicate alias %q found: used by both %q and %q",
                member.Alias, existingName, member.Name,
            )
        }

        seen[member.Alias] = member.Name
    }

    return nil
}
```

---

## Sign-off

**Prepared By:** Claude (AI Assistant)
**Review Required:** Development Team Lead, Security Team
**Approval Required:** Product Owner, Engineering Manager

**Next Steps:**
1. Team review of PRD
2. Estimate refinement
3. Sprint planning
4. Implementation kickoff

---

**Document Status:** Ready for Review
**Last Updated:** 2026-02-09
**Version:** 1.0
