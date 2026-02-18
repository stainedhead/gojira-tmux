# Security & Team Member Updates - Task Breakdown

**Feature:** Security & Team Member Updates
**Version:** 1.0
**Created:** 2026-02-09
**Status:** Planning
**Estimated Total Duration:** 12-18 hours

---

## Task Organization

Tasks are organized by implementation phase following Clean Architecture layers (bottom-up). Each task includes:
- **ID:** Unique task identifier (e.g., P1.1, P1.2)
- **Phase:** Implementation phase (1-6)
- **Dependencies:** Prerequisites that must complete first
- **Estimated Duration:** Time estimate in hours
- **Priority:** P0 (Critical), P1 (High), P2 (Medium), P3 (Low)
- **Description:** What needs to be done
- **Files to Create/Modify:** Specific files affected
- **Acceptance Criteria:** Definition of done (checkboxes)
- **Verification Commands:** Commands to verify completion

**Development Approach:** TDD + Bottom-Up Clean Architecture
- Red → Green → Refactor for each component
- Domain → Adapter → Use Case → Infrastructure
- Each phase produces working, tested code

---

## Progress Summary

**Overall Progress:** [0/25] tasks complete ([0%])

**By Phase:**
- Phase 1: [0/3] tasks complete
- Phase 2: [0/4] tasks complete
- Phase 3: [0/2] tasks complete
- Phase 4: [0/4] tasks complete
- Phase 5: [0/3] tasks complete
- Phase 6: [0/2] tasks complete

**By Priority:**
- P0 (Critical): [0/18] complete
- P1 (High): [0/2] complete
- P2 (Medium): [0/0] complete

---

## Phase 1: Domain Model Updates (1.5 hours)

### Task P1.1: Update Domain Ports

**ID:** P1.1
**Dependencies:** None
**Duration:** 30 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Simplify `AuthPort` interface to remove OAuth methods and add Atlassian API token validation. Add `AtlassianConfig` struct and remove `OktaConfig`. Update `TokenStorePort` to remove refresh token methods.

**Files to Modify:**
- `internal/domain/ports.go` - Simplify AuthPort, add AtlassianConfig, remove OktaConfig, update TokenStorePort

**Acceptance Criteria:**
- [ ] `AuthPort` has only 2 methods: `ValidateToken(ctx, email, token)` and `IsTokenValid(ctx)`
- [ ] Removed methods: StartAuthFlow, WaitForCallback, CancelAuthFlow, RefreshSession, Logout
- [ ] `AtlassianConfig` struct created with `Email string` field
- [ ] `OktaConfig` struct removed
- [ ] `Config` struct uses `Atlassian AtlassianConfig` instead of `Okta OktaConfig`
- [ ] `TokenStorePort` has 4 methods (removed GetRefreshToken, SetRefreshToken, DeleteRefreshToken)
- [ ] Comments updated to reflect new auth model
- [ ] Code compiles without errors

**Implementation Details:**

```go
// New AuthPort interface
type AuthPort interface {
    ValidateToken(ctx context.Context, email, token string) (string, error)
    IsTokenValid(ctx context.Context) bool
}

// New AtlassianConfig
type AtlassianConfig struct {
    Email string `yaml:"email"`
}

// Updated Config
type Config struct {
    Jira       JiraConfig       `yaml:"jira"`
    Atlassian  AtlassianConfig  `yaml:"atlassian"`  // Changed
    Projects   []Project        `yaml:"projects"`
    Team       []TeamMember     `yaml:"team"`
}

// Updated TokenStorePort (refresh methods removed)
type TokenStorePort interface {
    GetJiraToken() (string, error)
    SetJiraToken(token string) error
    DeleteJiraToken() error
    HasJiraToken() bool
}
```

**Verification Commands:**
```bash
go build ./internal/domain
```

---

### Task P1.2: Update Team Member Model

**ID:** P1.2
**Dependencies:** None (can run parallel to P1.1)
**Duration:** 45 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Add `Alias` field to `TeamMember` struct. Add `isValidAlias()` helper function. Update `Validate()` to check alias format. Add `MatchesIdentifier()` method for multi-priority matching. Add `DisplayName()` method for UI display.

**Files to Modify:**
- `internal/domain/team_member.go` - Add Alias field and methods
- `internal/domain/team_member_test.go` - Add comprehensive alias tests

**Acceptance Criteria:**
- [ ] `Alias` field added: `Alias string` with yaml and json tags and omitempty
- [ ] `isValidAlias(alias string) bool` function validates alphanumeric only
- [ ] `Validate()` checks alias format if alias is provided
- [ ] `MatchesIdentifier(identifier string) bool` method implements 4-priority matching
- [ ] `DisplayName() string` method returns "Name (Alias)" if alias exists, else "Name"
- [ ] Tests cover: valid aliases, invalid aliases, matching logic, display name
- [ ] All tests pass
- [ ] Code coverage >90%

**Implementation Details:**

```go
type TeamMember struct {
    Name  string `yaml:"name" json:"name"`
    Email string `yaml:"email" json:"email"`
    Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`  // NEW
}

func isValidAlias(alias string) bool {
    if alias == "" {
        return true
    }
    for _, r := range alias {
        if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
            return false
        }
    }
    return true
}

func (t *TeamMember) Validate() error {
    // ... existing validation ...
    if t.Alias != "" && !isValidAlias(t.Alias) {
        return errors.New("team member alias must be alphanumeric (no spaces)")
    }
    return nil
}

func (t *TeamMember) MatchesIdentifier(identifier string) bool {
    // Priority 1: Exact alias match
    if t.Alias != "" && t.Alias == identifier {
        return true
    }
    // Priority 2: Exact name match
    if t.Name == identifier {
        return true
    }
    // Priority 3: Case-insensitive alias
    if t.Alias != "" && strings.EqualFold(t.Alias, identifier) {
        return true
    }
    // Priority 4: Case-insensitive name
    return strings.EqualFold(t.Name, identifier)
}

func (t *TeamMember) DisplayName() string {
    if t.Alias != "" {
        return fmt.Sprintf("%s (%s)", t.Name, t.Alias)
    }
    return t.Name
}
```

**Test Cases:**
- [ ] Test valid aliases: "JohnA", "SA", "John123"
- [ ] Test invalid aliases: "John A", "John-A", "John.A"
- [ ] Test exact alias match (case-sensitive)
- [ ] Test exact name match
- [ ] Test case-insensitive alias match
- [ ] Test case-insensitive name match
- [ ] Test no match returns false
- [ ] Test display name with alias
- [ ] Test display name without alias

**Verification Commands:**
```bash
go test ./internal/domain -run TestTeamMember -v
go test ./internal/domain -cover
```

---

### Task P1.3: Update User Model

**ID:** P1.3
**Dependencies:** None (can run parallel)
**Duration:** 15 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Remove session management from User model (SessionExpiry field, IsSessionValid method, ValidateTeamMembership method).

**Files to Modify:**
- `internal/domain/user.go` - Remove session fields and methods
- `internal/domain/user_test.go` - Update tests

**Acceptance Criteria:**
- [ ] `SessionExpiry` field removed
- [ ] `IsSessionValid()` method removed
- [ ] `ValidateTeamMembership()` method removed
- [ ] `Validate()` only checks email
- [ ] All tests updated and passing

**Implementation Details:**

```go
// Simplified User struct
type User struct {
    Email string `json:"email"`
}

// Simplified validation
func (u *User) Validate() error {
    if u.Email == "" {
        return errors.New("user email is required")
    }
    if !isValidEmail(u.Email) {
        return errors.New("user email is invalid")
    }
    return nil
}
```

**Verification Commands:**
```bash
go test ./internal/domain -run TestUser -v
```

---

## Phase 2: Adapter Layer (4 hours)

### Task P2.1: Create Atlassian Auth Adapter

**ID:** P2.1
**Dependencies:** P1.1 (need AuthPort interface)
**Duration:** 2 hours
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Create new `AtlassianAdapter` that implements `AuthPort` interface. Implement token validation by calling `/rest/api/2/myself` endpoint with Basic Auth. Handle all error cases (401, 403, network, CAPTCHA).

**Files to Create:**
- `internal/adapter/auth/atlassian.go` - Atlassian adapter implementation
- `internal/adapter/auth/atlassian_test.go` - Comprehensive tests

**Acceptance Criteria:**
- [ ] `AtlassianAdapter` struct created with tokenStore and httpClient
- [ ] `ValidateToken(ctx, email, token)` method implemented
- [ ] Builds Basic Auth header: `base64(email:token)`
- [ ] Calls `GET /rest/api/2/myself` with Authorization header
- [ ] Parses response and returns validated email
- [ ] Handles 401: "Invalid token or email"
- [ ] Handles 403: "Insufficient permissions"
- [ ] Handles network errors with clear messages
- [ ] Detects CAPTCHA via `X-Seraph-LoginReason` header
- [ ] `IsTokenValid(ctx)` method implemented (checks token existence)
- [ ] Tests cover all success and error scenarios
- [ ] Tests use mock HTTP server
- [ ] All tests pass
- [ ] Code coverage >90%

**Implementation Details:**

```go
package auth

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"

    "github.com/stainedhead/gojira-tmux/internal/domain"
)

type AtlassianAdapter struct {
    tokenStore domain.TokenStorePort
    httpClient *http.Client
    jiraURL    string
}

func NewAtlassianAdapter(tokenStore domain.TokenStorePort, jiraURL string) *AtlassianAdapter {
    return &AtlassianAdapter{
        tokenStore: tokenStore,
        httpClient: &http.Client{Timeout: 10 * time.Second},
        jiraURL:    jiraURL,
    }
}

func (a *AtlassianAdapter) ValidateToken(ctx context.Context, email, token string) (string, error) {
    url := a.jiraURL + "/rest/api/2/myself"
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    // Set Basic Auth
    auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
    req.Header.Set("Authorization", "Basic "+auth)
    req.Header.Set("Accept", "application/json")

    resp, err := a.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to validate token: %w", err)
    }
    defer resp.Body.Close()

    // Check for CAPTCHA
    if resp.Header.Get("X-Seraph-LoginReason") == "AUTHENTICATION_DENIED" {
        return "", errors.New("too many failed login attempts. account temporarily locked")
    }

    // Handle errors
    if resp.StatusCode == http.StatusUnauthorized {
        return "", errors.New("invalid token or email")
    }
    if resp.StatusCode == http.StatusForbidden {
        return "", errors.New("insufficient permissions")
    }
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    // Parse response
    var result struct {
        EmailAddress string `json:"emailAddress"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("failed to decode response: %w", err)
    }

    return result.EmailAddress, nil
}

func (a *AtlassianAdapter) IsTokenValid(ctx context.Context) bool {
    token, err := a.tokenStore.GetJiraToken()
    return err == nil && token != ""
}

var _ domain.AuthPort = (*AtlassianAdapter)(nil)
```

**Verification Commands:**
```bash
go test ./internal/adapter/auth -run TestAtlassian -v
go test ./internal/adapter/auth -cover
```

---

### Task P2.2: Delete Okta Components

**ID:** P2.2
**Dependencies:** P2.1 (ensure replacement exists)
**Duration:** 10 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Delete all Okta-related files. Run `go mod tidy` to remove OAuth dependencies.

**Files to Delete:**
- `internal/adapter/auth/okta.go`
- `internal/adapter/auth/okta_test.go`
- `internal/adapter/auth/callback.go`
- `internal/adapter/auth/callback_test.go`

**Acceptance Criteria:**
- [ ] All 4 Okta files deleted
- [ ] `go mod tidy` executed
- [ ] `go-oidc` dependency removed from go.mod
- [ ] `oauth2` dependency removed from go.mod
- [ ] `go build ./...` succeeds (no import errors)

**Verification Commands:**
```bash
rm internal/adapter/auth/okta.go
rm internal/adapter/auth/okta_test.go
rm internal/adapter/auth/callback.go
rm internal/adapter/auth/callback_test.go
go mod tidy
go build ./...
```

---

### Task P2.3: Update Config Loader

**ID:** P2.3
**Dependencies:** P1.1 (AtlassianConfig), P1.2 (TeamMember.Alias)
**Duration:** 1 hour
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Update config validation to check Atlassian email instead of Okta. Add alias uniqueness validation. Add duplicate email check.

**Files to Modify:**
- `internal/adapter/config/config.go` - Update validate() method
- `internal/adapter/config/config_test.go` - Add alias validation tests

**Acceptance Criteria:**
- [ ] Okta validation removed (lines 60-69)
- [ ] Atlassian email validation added
- [ ] Alias uniqueness check added
- [ ] Duplicate email check added
- [ ] `ValidateUserAccess()` method removed (no longer needed)
- [ ] Tests cover all validation scenarios
- [ ] All tests pass

**Implementation Details:**

```go
func (l *Loader) validate(cfg *domain.Config) error {
    // Jira validation (unchanged)
    if cfg.Jira.URL == "" {
        return errors.New("jira.url is required")
    }
    if !strings.HasPrefix(cfg.Jira.URL, "https://") {
        return errors.New("jira.url must use HTTPS")
    }

    // NEW: Atlassian validation
    if cfg.Atlassian.Email == "" {
        return errors.New("atlassian.email is required")
    }
    if !isValidEmail(cfg.Atlassian.Email) {
        return errors.New("atlassian.email must be a valid email address")
    }

    // Projects validation (unchanged)
    if len(cfg.Projects) == 0 {
        return errors.New("at least one project is required")
    }
    for _, p := range cfg.Projects {
        if err := p.Validate(); err != nil {
            return fmt.Errorf("project %s: %w", p.Key, err)
        }
    }

    // Team validation (enhanced)
    if len(cfg.Team) == 0 {
        return errors.New("at least one team member is required")
    }

    // NEW: Check alias uniqueness
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

        // NEW: Check for duplicate emails
        for j := 0; j < i; j++ {
            if strings.EqualFold(cfg.Team[j].Email, m.Email) {
                return fmt.Errorf("duplicate email %q (team members %s and %s)",
                    m.Email, cfg.Team[j].Name, m.Name)
            }
        }
    }

    return nil
}
```

**Test Cases:**
- [ ] Test valid Atlassian config
- [ ] Test missing Atlassian email error
- [ ] Test invalid email format error
- [ ] Test duplicate alias error
- [ ] Test invalid alias format error
- [ ] Test duplicate email error
- [ ] Test backward compatibility (no aliases)

**Verification Commands:**
```bash
go test ./internal/adapter/config -v
go test ./internal/adapter/config -cover
```

---

### Task P2.4: Update JQL Builder

**ID:** P2.4
**Dependencies:** P1.2 (TeamMember.MatchesIdentifier)
**Duration:** 45 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Update `buildAssigneeCondition()` to use `TeamMember.MatchesIdentifier()` for alias support.

**Files to Modify:**
- `internal/adapter/jira/search.go` - Update buildAssigneeCondition()
- `internal/adapter/jira/search_test.go` - Add alias matching tests

**Acceptance Criteria:**
- [ ] `buildAssigneeCondition()` uses `MatchesIdentifier()`
- [ ] Iterates through team members to find match
- [ ] Uses email from matched member for JQL
- [ ] Tests cover matching by alias
- [ ] Tests cover matching by name
- [ ] Tests cover case-insensitive matching
- [ ] Tests cover no match (returns empty)
- [ ] All tests pass

**Implementation Details:**

```go
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

**Test Cases:**
- [ ] Test exact alias match: "JohnA" matches member with alias="JohnA"
- [ ] Test exact name match: "John Anderson" matches member
- [ ] Test case-insensitive alias: "johna" matches "JohnA"
- [ ] Test case-insensitive name: "john anderson" matches "John Anderson"
- [ ] Test no match: "Unknown" returns empty string
- [ ] Test backward compat: member without alias matches by name

**Verification Commands:**
```bash
go test ./internal/adapter/jira -run TestJQLBuilder -v
go test ./internal/adapter/jira -cover
```

---

## Phase 3: Use Case Layer (1.5 hours)

### Task P3.1: Simplify Authenticate Use Case

**ID:** P3.1
**Dependencies:** P2.1 (Atlassian Adapter), P2.3 (Config validation)
**Duration:** 1 hour
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Remove all OAuth methods from Authenticate use case. Keep only token validation methods.

**Files to Modify:**
- `internal/usecase/authenticate.go` - Remove OAuth methods, simplify
- `internal/usecase/authenticate_test.go` - Update tests

**Acceptance Criteria:**
- [ ] Removed methods: StartLogin, CompleteLogin, CancelLogin, CheckSession, Logout
- [ ] Kept methods: ValidateAndSaveToken, HasValidToken, ClearToken
- [ ] Constructor updated (remove configPort dependency)
- [ ] All tests updated
- [ ] All tests pass
- [ ] Code coverage >90%

**Implementation Details:**

```go
package usecase

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/stainedhead/gojira-tmux/internal/domain"
)

type Authenticate struct {
    authPort   domain.AuthPort
    tokenStore domain.TokenStorePort
}

func NewAuthenticate(authPort domain.AuthPort, tokenStore domain.TokenStorePort) *Authenticate {
    return &Authenticate{
        authPort:   authPort,
        tokenStore: tokenStore,
    }
}

func (a *Authenticate) ValidateAndSaveToken(ctx context.Context, email, token string) error {
    email = strings.TrimSpace(email)
    token = strings.TrimSpace(token)

    if email == "" || token == "" {
        return errors.New("email and token are required")
    }

    validatedEmail, err := a.authPort.ValidateToken(ctx, email, token)
    if err != nil {
        return fmt.Errorf("token validation failed: %w", err)
    }

    if !strings.EqualFold(validatedEmail, email) {
        return fmt.Errorf("token email mismatch: expected %s, got %s", email, validatedEmail)
    }

    if err := a.tokenStore.SetJiraToken(token); err != nil {
        return fmt.Errorf("failed to save token: %w", err)
    }

    return nil
}

func (a *Authenticate) HasValidToken(ctx context.Context) bool {
    return a.authPort.IsTokenValid(ctx)
}

func (a *Authenticate) ClearToken() error {
    return a.tokenStore.DeleteJiraToken()
}
```

**Test Cases:**
- [ ] Test ValidateAndSaveToken success
- [ ] Test ValidateAndSaveToken with invalid credentials
- [ ] Test ValidateAndSaveToken with email mismatch
- [ ] Test ValidateAndSaveToken with save failure
- [ ] Test HasValidToken returns true when token exists
- [ ] Test HasValidToken returns false when token missing
- [ ] Test ClearToken success

**Verification Commands:**
```bash
go test ./internal/usecase -run TestAuthenticate -v
go test ./internal/usecase -cover
```

---

### Task P3.2: Update Setup Token Use Case

**ID:** P3.2
**Dependencies:** None (minimal changes)
**Duration:** 30 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Update SetupToken use case to validate email format if needed. Ensure it works with new token store interface.

**Files to Modify:**
- `internal/usecase/setup_token.go` - Minor updates
- `internal/usecase/setup_token_test.go` - Update tests

**Acceptance Criteria:**
- [ ] Email validation added if not already present
- [ ] Works with updated TokenStorePort
- [ ] All tests updated and passing

**Verification Commands:**
```bash
go test ./internal/usecase -run TestSetupToken -v
```

---

## Phase 4: Infrastructure Layer (4.5 hours)

### Task P4.1: Update Setup Screen

**ID:** P4.1
**Dependencies:** P3.1 (Authenticate use case), P3.2 (SetupToken)
**Duration:** 2 hours
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Replace OAuth flow with email + token input form. Add instructions with link to token generation. Handle validation and errors.

**Files to Modify:**
- `internal/infrastructure/tui/setup_screen.go` - Complete redesign

**Acceptance Criteria:**
- [ ] Email input field added
- [ ] Token input field added (masked)
- [ ] Instructions displayed with https://id.atlassian.com/manage/api-tokens link
- [ ] Tab key switches between fields
- [ ] Enter key validates and submits
- [ ] Validation errors displayed clearly
- [ ] Success transitions to main screen
- [ ] Manual testing passes

**Implementation Details:**
- Use BubbleTea's textinput component
- Mask token input with EchoMode = EchoPassword
- Display helpful error messages
- Show loading indicator during validation

**Verification Commands:**
```bash
go build -o bin/gojira ./cmd/gojira
# Manual testing: Run app without token, verify setup screen
```

---

### Task P4.2: Update/Remove Login Screen

**ID:** P4.2
**Dependencies:** P4.1 (Setup screen working)
**Duration:** 1 hour
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Evaluate if login screen is still needed. If not, remove it. If needed, simplify to token validation only.

**Files to Modify:**
- `internal/infrastructure/tui/login_screen.go` - Simplify or remove

**Acceptance Criteria:**
- [ ] Evaluated need for login screen
- [ ] If removed: all references deleted
- [ ] If kept: simplified to token validation
- [ ] Manual testing passes

**Verification Commands:**
```bash
go build -o bin/gojira ./cmd/gojira
```

---

### Task P4.3: Update Filter Bar

**ID:** P4.3
**Dependencies:** P1.2 (TeamMember.DisplayName)
**Duration:** 1 hour
**Priority:** P1 (High)
**Status:** ⬜ Not Started

**Description:**
Update filter bar to display aliases using `DisplayName()` method. Test alias filtering.

**Files to Modify:**
- `internal/infrastructure/tui/filter_bar.go` - Update team member display

**Acceptance Criteria:**
- [ ] Team member dropdown shows "Name (Alias)" format
- [ ] Filtering by alias works correctly
- [ ] Filtering by name still works
- [ ] Backward compatible (members without aliases)
- [ ] Manual testing passes

**Verification Commands:**
```bash
go build -o bin/gojira ./cmd/gojira
# Manual testing: Filter by alias, verify correct member selected
```

---

### Task P4.4: Update App Initialization

**ID:** P4.4
**Dependencies:** P4.1, P4.2 (TUI screens updated)
**Duration:** 30 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Update app initialization to use AtlassianAdapter instead of OktaAdapter. Add Atlassian email option.

**Files to Modify:**
- `internal/infrastructure/tui/app.go` - Update initialization

**Acceptance Criteria:**
- [ ] OAuth-related initialization removed
- [ ] AtlassianAdapter used instead of OktaAdapter
- [ ] Atlassian email option added (if needed)
- [ ] All tests updated and passing
- [ ] Manual testing passes

**Verification Commands:**
```bash
go test ./internal/infrastructure/tui -v
go build -o bin/gojira ./cmd/gojira
```

---

## Phase 5: Entry Point & Configuration (1.5 hours)

### Task P5.1: Update Main Entry Point

**ID:** P5.1
**Dependencies:** P4.4 (App initialization complete)
**Duration:** 45 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Update `main.go` to use AtlassianAdapter and Atlassian email from config.

**Files to Modify:**
- `cmd/gojira/main.go` - Update initialization

**Acceptance Criteria:**
- [ ] `NewOktaAdapter` replaced with `NewAtlassianAdapter`
- [ ] Jira client uses `cfg.Atlassian.Email` instead of `cfg.Jira.Username`
- [ ] OAuth provider initialization removed
- [ ] Environment variable support works (JIRA_API_TOKEN)
- [ ] Full app builds and runs

**Implementation Details:**

```go
// Initialize Atlassian adapter
atlassianAdapter := auth.NewAtlassianAdapter(tokenStore, cfg.Jira.URL)

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
)
```

**Verification Commands:**
```bash
go build -o bin/gojira ./cmd/gojira
./bin/gojira --help
# Test with valid token
# Test without token (should show setup screen)
```

---

### Task P5.2: Update Configuration Examples

**ID:** P5.2
**Dependencies:** None (can run anytime)
**Duration:** 30 minutes
**Priority:** P1 (High)
**Status:** ⬜ Not Started

**Description:**
Update README with new auth. Create example config file. Create migration guide.

**Files to Create/Modify:**
- `README.md` - Update authentication section
- `config.example.yaml` - New format
- `docs/MIGRATION.md` - Migration guide

**Acceptance Criteria:**
- [ ] README updated (remove Okta, add Atlassian)
- [ ] Quick start guide updated
- [ ] Environment variables section updated
- [ ] `config.example.yaml` created with Atlassian config
- [ ] Migration guide created with step-by-step instructions
- [ ] Migration guide includes before/after config examples
- [ ] Migration guide includes troubleshooting section

**Verification Commands:**
```bash
# Review README changes
# Test example config
```

---

### Task P5.3: Update Dependencies

**ID:** P5.3
**Dependencies:** P2.2 (Okta deleted)
**Duration:** 15 minutes
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Run `go mod tidy` and verify OAuth dependencies are removed.

**Acceptance Criteria:**
- [ ] `go mod tidy` executed
- [ ] `go-oidc` removed from go.mod
- [ ] `oauth2` removed from go.mod
- [ ] No unused dependencies remain
- [ ] `go mod verify` passes
- [ ] All builds pass

**Verification Commands:**
```bash
go mod tidy
go mod verify
go build ./...
go test ./...
```

---

## Phase 6: Testing & Documentation (5 hours)

### Task P6.1: Integration Testing

**ID:** P6.1
**Dependencies:** P5.1 (Full app working)
**Duration:** 3 hours
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started

**Description:**
Create integration test suite covering all major workflows.

**Acceptance Criteria:**
- [ ] Test: First-time setup (no token → setup → success)
- [ ] Test: Invalid token (error → retry → success)
- [ ] Test: Token validation (valid/invalid credentials)
- [ ] Test: Alias filtering (filter by alias → correct member)
- [ ] Test: Backward compatibility (no aliases → works normally)
- [ ] Test: Network errors handled gracefully
- [ ] Test: Malformed config detected
- [ ] Test: Duplicate alias detected
- [ ] All integration tests pass
- [ ] Manual testing checklist complete

**Manual Testing Checklist:**
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
- [ ] Token stored securely in keychain
- [ ] Token persists across app restarts
- [ ] Environment variable override works

**Verification Commands:**
```bash
go test ./... -v
go test ./... -cover
```

---

### Task P6.2: Update Documentation

**ID:** P6.2
**Dependencies:** P5.2 (Migration guide exists)
**Duration:** 2 hours
**Priority:** P1 (High)
**Status:** ⬜ Not Started

**Description:**
Update all documentation with new authentication approach.

**Files to Modify:**
- `documentation/product-summary.md` - Update features
- `documentation/product-details.md` - Update auth workflow
- `documentation/technical-details.md` - Update architecture
- `specs/security-updates/implementation-notes.md` - Add lessons learned

**Acceptance Criteria:**
- [ ] Product summary updated (remove Okta, add Atlassian)
- [ ] Product details updated (new setup flow)
- [ ] Technical details updated (simplified architecture)
- [ ] Architecture diagrams updated
- [ ] API documentation updated
- [ ] Troubleshooting guide updated
- [ ] Implementation notes complete
- [ ] All documentation reviewed

**Verification Commands:**
```bash
# Manual review of all docs
```

---

## Blocked Tasks

**Tasks currently blocked:**
- None at start. Update as blockers arise.

---

## Quality Gate Checklist

Before marking ANY task as complete:

- [ ] **Tests Pass** - `go test ./...`
- [ ] **Code Formatted** - `go fmt ./...`
- [ ] **Dependencies Tidy** - `go mod tidy`
- [ ] **Vet Passes** - `go vet ./...`
- [ ] **Linter Passes** - `golangci-lint run`
- [ ] **Build Succeeds** - `go build -o bin/gojira ./cmd/gojira`
- [ ] **Documentation** - Updated if needed
- [ ] **Status.md** - Updated with progress

**Never mark a task complete until ALL gates pass.**

---

## Task Dependencies Graph

**Visual representation of task dependencies:**

```
P1.1 ──┐
P1.2 ──┤─→ P2.1 ─→ P2.2 ──→ P5.3
P1.3 ──┘     │         │
             │         └──→ P3.1 ─→ P4.1 ──┐
             │                │            │
             └──→ P2.3 ───────┤            ├─→ P4.4 ─→ P5.1 ─→ P6.1
             └──→ P2.4 ───────┘            │
                                 P4.2 ──┘
                                 P4.3
                                           P5.2 ─→ P6.2

Legend:
─→  Dependency (must complete before)
┐   Merge point (all parents must complete)
```

**Critical Path:**
```
P1.1 → P2.1 → P3.1 → P4.1 → P4.4 → P5.1 → P6.1
Total: ~9 hours
```

---

## Notes

**Assumptions:**
- User has Atlassian account for token generation
- Jira instance is Cloud (not Server/Data Center)
- Team size <100 (for O(n) matching performance)

**Risks:**
- Token validation endpoint availability
- User confusion during migration
- Keychain access issues on some systems

**TDD Reminder:**
- Write test first (Red)
- Implement minimum to pass (Green)
- Refactor while keeping tests green
- Commit after each task

**Status.md Updates:**
- Update after EACH task completion
- Mark phase progress
- Note any blockers or deviations
