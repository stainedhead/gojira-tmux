# Research: gojira-tmux TUI Application

**Date**: 2026-01-01
**Feature**: [spec.md](./spec.md)

## Overview

This document captures technology decisions and research findings for implementing the gojira-tmux TUI application.

---

## 1. Okta OIDC with PKCE

### Decision
Use `golang.org/x/oauth2` with `coreos/go-oidc/v3` for Okta OIDC Authorization Code Flow with PKCE.

### Rationale
- Standard library oauth2 provides PKCE support via `oauth2.GenerateVerifier()` and `oauth2.S256ChallengeOption`
- coreos/go-oidc provides ID token verification and claims extraction
- Well-documented combination for CLI applications
- No need for external authentication server

### Alternatives Considered
| Alternative | Reason Rejected |
|-------------|-----------------|
| hashicorp/cap/oidc | More complex, designed for server apps |
| Manual OIDC implementation | Unnecessary complexity, security risks |
| OAuth2 without PKCE | PKCE required for public clients (CLI) |

### Implementation Pattern

```go
// PKCE Flow for CLI
import (
    "golang.org/x/oauth2"
    "github.com/coreos/go-oidc/v3/oidc"
)

// 1. Generate PKCE verifier and challenge
verifier := oauth2.GenerateVerifier()
challenge := oauth2.S256ChallengeOption(verifier)

// 2. Build authorization URL
authURL := oauth2Config.AuthCodeURL(
    state,
    oauth2.AccessTypeOffline,
    challenge,
)

// 3. Open browser, start local callback server
// http://localhost:{port}/callback

// 4. Exchange code for tokens
token, err := oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))

// 5. Extract ID token and verify
rawIDToken := token.Extra("id_token").(string)
idToken, err := verifier.Verify(ctx, rawIDToken)

// 6. Extract email claim
var claims struct {
    Email string `json:"email"`
}
idToken.Claims(&claims)
```

### Callback Server Pattern

```go
// Local HTTP server for OAuth callback
type callbackServer struct {
    server   *http.Server
    codeChan chan string
    errChan  chan error
}

func (s *callbackServer) Start(port int) {
    mux := http.NewServeMux()
    mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        s.codeChan <- code
        w.Write([]byte("Authentication successful. You can close this window."))
    })

    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    go s.server.ListenAndServe()
}

func (s *callbackServer) WaitForCode(ctx context.Context) (string, error) {
    select {
    case code := <-s.codeChan:
        return code, nil
    case <-ctx.Done():
        return "", errors.New("authentication timeout")
    }
}
```

### Session Persistence
- Store refresh token in OS keychain (see go-keyring research)
- Okta session: 8-hour expiry per spec FR-001a
- Check token expiry on app launch, refresh if valid, re-auth if expired

---

## 2. BubbleTea Screen Routing

### Decision
Use enum-based screen state with embedded child models in the root App model.

### Rationale
- Simple, type-safe screen transitions
- Each screen is a separate BubbleTea model
- Root model routes messages and renders active screen
- Clean separation of concerns

### Alternatives Considered
| Alternative | Reason Rejected |
|-------------|-----------------|
| Single monolithic model | Difficult to maintain, violates SRP |
| Stack-based navigation | Overkill for linear flow (setup → login → main) |
| External state machine | Adds unnecessary dependency |

### Implementation Pattern

```go
// Screen state enum
type Screen int

const (
    ScreenSetup Screen = iota
    ScreenLogin
    ScreenMain
)

// Root App model
type App struct {
    screen       Screen
    setupModel   SetupModel
    loginModel   LoginModel
    mainModel    MainModel

    // Shared state
    config       *Config
    user         *domain.User
}

func (m App) Init() tea.Cmd {
    // Check for existing token
    return m.checkToken()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case TokenStoredMsg:
        m.screen = ScreenLogin
        return m, m.loginModel.Init()
    case AuthSuccessMsg:
        m.user = msg.User
        m.screen = ScreenMain
        return m, m.mainModel.Init()
    }

    // Route to active screen
    switch m.screen {
    case ScreenSetup:
        var cmd tea.Cmd
        m.setupModel, cmd = m.setupModel.Update(msg)
        return m, cmd
    case ScreenLogin:
        var cmd tea.Cmd
        m.loginModel, cmd = m.loginModel.Update(msg)
        return m, cmd
    case ScreenMain:
        var cmd tea.Cmd
        m.mainModel, cmd = m.mainModel.Update(msg)
        return m, cmd
    }
    return m, nil
}

func (m App) View() string {
    switch m.screen {
    case ScreenSetup:
        return m.setupModel.View()
    case ScreenLogin:
        return m.loginModel.View()
    case ScreenMain:
        return m.mainModel.View()
    }
    return ""
}
```

### Screen Transition Messages

```go
// Custom messages for screen transitions
type TokenStoredMsg struct{}
type AuthSuccessMsg struct{ User *domain.User }
type AuthCancelledMsg struct{}
type LogoutMsg struct{}
```

### Keyboard Handling
Each screen defines its own key bindings. Use bubbles/key for key binding management:

```go
type keyMap struct {
    Up    key.Binding
    Down  key.Binding
    Enter key.Binding
    Quit  key.Binding
}

var mainKeys = keyMap{
    Up:    key.NewBinding(key.WithKeys("up", "k")),
    Down:  key.NewBinding(key.WithKeys("down", "j")),
    Enter: key.NewBinding(key.WithKeys("enter")),
    Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c")),
}
```

---

## 3. go-keyring Cross-Platform Storage

### Decision
Use `zalando/go-keyring` with file-based fallback for headless environments.

### Rationale
- Simple API: Set, Get, Delete
- Native OS integration (Keychain, Secret Service, Credential Manager)
- Well-maintained, widely used
- Simple error handling with ErrNotFound

### Alternatives Considered
| Alternative | Reason Rejected |
|-------------|-----------------|
| 99designs/keyring | More complex API, unnecessary features |
| Manual file encryption | Security concerns, wheel reinvention |
| Environment variables only | Insecure for API tokens |

### Implementation Pattern

```go
import "github.com/zalando/go-keyring"

const (
    serviceName = "gojira-tmux"
    tokenUser   = "jira-api-token"
    refreshUser = "okta-refresh-token"
)

// Store token
func StoreJiraToken(token string) error {
    return keyring.Set(serviceName, tokenUser, token)
}

// Retrieve token
func GetJiraToken() (string, error) {
    token, err := keyring.Get(serviceName, tokenUser)
    if err == keyring.ErrNotFound {
        return "", nil // No token stored
    }
    return token, err
}

// Delete token
func DeleteJiraToken() error {
    return keyring.Delete(serviceName, tokenUser)
}
```

### Fallback Strategy
For headless Linux (no D-Bus Secret Service):

```go
// Fallback to encrypted file storage
type TokenStore interface {
    Get(key string) (string, error)
    Set(key, value string) error
    Delete(key string) error
}

type keyringStore struct{}
type fileStore struct {
    path string
    key  []byte
}

func NewTokenStore() TokenStore {
    // Try keyring first
    err := keyring.Set("test", "test", "test")
    if err == nil {
        keyring.Delete("test", "test")
        return &keyringStore{}
    }

    // Fall back to encrypted file
    return &fileStore{
        path: filepath.Join(os.UserHomeDir(), ".config", "gojira", "credentials"),
        key:  deriveKeyFromPassword(),
    }
}
```

### Key Naming Convention
| Key | Purpose |
|-----|---------|
| `gojira-tmux/jira-api-token` | Jira API token for REST calls |
| `gojira-tmux/okta-refresh-token` | Okta refresh token for session persistence |

---

## 4. JQL Query Building

### Decision
Use string builder with parameterized filter conditions and proper escaping.

### Rationale
- JQL is a string-based query language
- Filters combine with AND logic
- Need proper escaping for user-provided values
- Simple builder pattern sufficient for our use case

### Alternatives Considered
| Alternative | Reason Rejected |
|-------------|-----------------|
| ORM-style query builder | Overkill, no Go library for JQL |
| Raw string concatenation | Injection vulnerability |

### Implementation Pattern

```go
type JQLBuilder struct {
    conditions []string
    orderBy    string
    limit      int
}

func NewJQLBuilder() *JQLBuilder {
    return &JQLBuilder{
        orderBy: "updated DESC",
        limit:   100,
    }
}

func (b *JQLBuilder) Project(keys ...string) *JQLBuilder {
    if len(keys) == 0 {
        return b
    }
    quoted := make([]string, len(keys))
    for i, k := range keys {
        quoted[i] = escapeJQL(k)
    }
    b.conditions = append(b.conditions, fmt.Sprintf("project IN (%s)", strings.Join(quoted, ", ")))
    return b
}

func (b *JQLBuilder) Assignee(email string) *JQLBuilder {
    if email == "" {
        return b
    }
    b.conditions = append(b.conditions, fmt.Sprintf("assignee = %q", email))
    return b
}

func (b *JQLBuilder) Status(status string) *JQLBuilder {
    if status == "" || status == "All" {
        return b
    }
    // Map UI status to JQL status
    jqlStatus := mapStatus(status)
    b.conditions = append(b.conditions, fmt.Sprintf("status = %q", jqlStatus))
    return b
}

func (b *JQLBuilder) Build() string {
    if len(b.conditions) == 0 {
        return fmt.Sprintf("ORDER BY %s", b.orderBy)
    }
    return fmt.Sprintf("%s ORDER BY %s", strings.Join(b.conditions, " AND "), b.orderBy)
}

// Status mapping from UI to JQL
func mapStatus(uiStatus string) string {
    switch uiStatus {
    case "Open":
        return "Open"
    case "Ready":
        return "Ready for Development"
    case "In Test":
        return "In Test"
    case "Done":
        return "Done"
    default:
        return uiStatus
    }
}

// Escape special JQL characters
func escapeJQL(s string) string {
    // JQL reserved: + - & | ! ( ) { } [ ] ^ " ~ * ? \ /
    replacer := strings.NewReplacer(
        `\`, `\\`,
        `"`, `\"`,
    )
    return `"` + replacer.Replace(s) + `"`
}
```

### Example Queries

```
# All projects, all members, all statuses
ORDER BY updated DESC

# Single project
project = "PROJ" ORDER BY updated DESC

# Combined filters
project IN ("PROJ1", "PROJ2") AND assignee = "john@example.com" AND status = "Open" ORDER BY updated DESC
```

---

## 5. go-jira Client Setup

### Decision
Use `andygrunwald/go-jira` with Basic Auth transport for API token authentication.

### Rationale
- Most popular Go Jira client
- Simple API for search, issue details, comments
- Supports Basic Auth with API tokens
- Handles pagination automatically for search

### Alternatives Considered
| Alternative | Reason Rejected |
|-------------|-----------------|
| Raw HTTP client | Unnecessary boilerplate |
| atlassian/go-jira | Deprecated, points to andygrunwald |

### Implementation Pattern

```go
import jira "github.com/andygrunwald/go-jira"

// Client initialization with Basic Auth
func NewJiraClient(url, username, apiToken string) (*jira.Client, error) {
    tp := jira.BasicAuthTransport{
        Username: username,
        Password: apiToken, // API token, not password
    }

    return jira.NewClient(tp.Client(), url)
}

// Search issues with JQL
func SearchIssues(client *jira.Client, jql string, maxResults int) ([]jira.Issue, error) {
    opts := &jira.SearchOptions{
        MaxResults: maxResults,
        StartAt:    0,
        Fields:     []string{
            "key", "summary", "description", "status",
            "priority", "assignee", "reporter", "duedate",
            "created", "updated", "labels", "customfield_10020", // sprint
            "customfield_10014", // epic
            "customfield_10016", // story points
        },
    }

    issues, resp, err := client.Issue.Search(jql, opts)
    if err != nil {
        return nil, fmt.Errorf("jira search failed: %w", err)
    }
    _ = resp // Can check resp.Total for pagination info

    return issues, nil
}

// Get issue with comments
func GetIssueWithComments(client *jira.Client, key string) (*jira.Issue, error) {
    issue, _, err := client.Issue.Get(key, &jira.GetQueryOptions{
        Expand: "comments",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get issue %s: %w", key, err)
    }
    return issue, nil
}
```

### Custom Fields
Custom field IDs vary by Jira instance. Common patterns:
- Sprint: `customfield_10020` or similar
- Epic Link: `customfield_10014`
- Story Points: `customfield_10016`

Consider making these configurable in config.yaml:

```yaml
jira:
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"
    story_points: "customfield_10016"
```

### Error Handling

```go
func handleJiraError(resp *jira.Response, err error) error {
    if err != nil {
        if resp != nil {
            switch resp.StatusCode {
            case 401:
                return errors.New("invalid API token or expired session")
            case 403:
                return errors.New("insufficient permissions")
            case 404:
                return errors.New("resource not found")
            case 429:
                return errors.New("rate limited, try again later")
            }
        }
        return err
    }
    return nil
}
```

---

## Dependencies Summary

| Package | Version | Purpose |
|---------|---------|---------|
| `charmbracelet/bubbletea` | v0.27+ | TUI framework |
| `charmbracelet/lipgloss` | v0.13+ | Styling |
| `charmbracelet/bubbles` | v0.20+ | UI components |
| `andygrunwald/go-jira` | v2.0+ | Jira REST client |
| `coreos/go-oidc/v3` | v3.11+ | OIDC verification |
| `golang.org/x/oauth2` | latest | OAuth2 with PKCE |
| `zalando/go-keyring` | v0.2+ | Secure storage |
| `gopkg.in/yaml.v3` | v3.0+ | Config parsing |

---

## Next Steps

1. Generate `data-model.md` with entity definitions
2. Generate `contracts/` with port interfaces
3. Generate `quickstart.md` for developer setup
