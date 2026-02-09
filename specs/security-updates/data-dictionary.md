# Security & Team Member Updates - Data Dictionary

**Created:** 2026-02-09
**Version:** 1.0
**Status:** Draft

---

## Overview

This document defines all data structures, types, interfaces, and constants for the Security & Team Member Updates implementation. Organized by Clean Architecture layers.

**Purpose:**
- Single source of truth for all data types
- Ensure consistency across layers
- Document validation rules and constraints

---

## Domain Layer

Location: `internal/domain/`

### 1. AtlassianConfig (Value Object)

**File:** `ports.go`

```go
// AtlassianConfig holds Atlassian-specific configuration.
type AtlassianConfig struct {
    Email string `yaml:"email"`
    // Note: API token stored in keychain, not here
}
```

**Validation Rules:**
- `Email` must be valid email format
- `Email` must match Jira account email

**Example:**
```go
config := AtlassianConfig{
    Email: "user@company.com",
}
```

---

### 2. TeamMember (Entity)

**File:** `team_member.go`

```go
// TeamMember represents a team member for filtering and display.
type TeamMember struct {
    Name  string `yaml:"name" json:"name"`
    Email string `yaml:"email" json:"email"`
    Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`
}
```

**Methods:**
```go
// Validate checks that the TeamMember has valid data.
func (t *TeamMember) Validate() error

// MatchesIdentifier returns true if the identifier matches name or alias.
func (t *TeamMember) MatchesIdentifier(identifier string) bool

// DisplayName returns the best display name (with alias if available).
func (t *TeamMember) DisplayName() string
```

**Validation Rules:**
- `Name` must be non-empty
- `Email` must be valid email format
- `Email` must be unique within team
- `Alias` optional, if provided must be alphanumeric (no spaces)
- `Alias` must be unique within team

**Business Rules:**
- Alias matching priority: exact alias → exact name → case-insensitive alias → case-insensitive name
- Display format: "Name (Alias)" if alias exists, otherwise just "Name"

**Example:**
```go
member := &TeamMember{
    Name:  "John Anderson",
    Email: "john.anderson@company.com",
    Alias: "JohnA",
}

if err := member.Validate(); err != nil {
    // Handle validation error
}

// Matching
member.MatchesIdentifier("JohnA")  // true
member.MatchesIdentifier("johna")  // true (case-insensitive fallback)
member.MatchesIdentifier("John Anderson")  // true

// Display
member.DisplayName()  // "John Anderson (JohnA)"
```

---

### 3. AuthPort (Interface)

**File:** `ports.go`

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

**Expected Behavior:**
- `ValidateToken`: Calls `/rest/api/2/myself`, returns validated email or error
- `IsTokenValid`: Checks token existence and basic validity (may use cache)

**Error Conditions:**
- Returns error with message "invalid token or email" if 401
- Returns error with message "insufficient permissions" if 403
- Returns error with network details if connection fails
- Returns error with CAPTCHA message if `X-Seraph-LoginReason: AUTHENTICATION_DENIED`

---

### 4. TokenStorePort (Interface)

**File:** `ports.go`

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

**Removed Methods:**
- `GetRefreshToken()` - No longer needed (no OAuth)
- `SetRefreshToken(token string)` - No longer needed
- `DeleteRefreshToken()` - No longer needed

---

### 5. Config (Aggregate)

**File:** `ports.go`

```go
// Config represents the application configuration.
type Config struct {
    Jira       JiraConfig       `yaml:"jira"`
    Atlassian  AtlassianConfig  `yaml:"atlassian"`  // Changed from Okta
    Projects   []Project        `yaml:"projects"`
    Team       []TeamMember     `yaml:"team"`
}
```

**Removed:**
- `OktaConfig` struct - Replaced by `AtlassianConfig`

---

## Adapter Layer

Location: `internal/adapter/auth/`

### AtlassianAdapter (Implementation)

**File:** `atlassian.go`

```go
// AtlassianAdapter implements the AuthPort interface for Atlassian API token authentication.
type AtlassianAdapter struct {
    tokenStore domain.TokenStorePort
    httpClient *http.Client
    jiraURL    string
}

// NewAtlassianAdapter creates a new Atlassian adapter.
func NewAtlassianAdapter(tokenStore domain.TokenStorePort, jiraURL string) *AtlassianAdapter

// ValidateToken validates the Atlassian API token.
func (a *AtlassianAdapter) ValidateToken(ctx context.Context, email, token string) (string, error)

// IsTokenValid checks if the stored token is valid.
func (a *AtlassianAdapter) IsTokenValid(ctx context.Context) bool
```

**Dependencies:**
- `domain.TokenStorePort` for token storage
- `http.Client` for API calls
- Jira base URL from config

**API Integration:**
```
GET /rest/api/2/myself
Authorization: Basic base64(email:token)
Accept: application/json
```

---

## Configuration

Location: `internal/config/`

### Atlassian Configuration

**YAML Format:**
```yaml
atlassian:
  email: "user@company.com"
```

**Go Struct:**
```go
type AtlassianConfig struct {
    Email string `yaml:"email"`
}
```

**Environment Variable Overrides:**
- `JIRA_API_TOKEN` → Stored in keychain (not in config)
- Email must be in config file (not environment variable for security)

---

## Type Aliases & Enums

### Alias Validation

```go
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
```

**Valid Aliases:**
- "JohnA" ✅
- "John123" ✅
- "SA" ✅

**Invalid Aliases:**
- "John A" ❌ (contains space)
- "John-A" ❌ (contains hyphen)
- "John.A" ❌ (contains dot)
- "John@A" ❌ (contains special char)

---

## Constants

### Error Messages

```go
const (
    ErrMsgInvalidToken = "Invalid email or API token. Please verify your credentials."
    ErrMsgInsufficientPermissions = "Valid credentials but insufficient permissions. Contact your Jira administrator."
    ErrMsgNetworkError = "Could not reach Jira API. Check your internet connection."
    ErrMsgCAPTCHA = "Too many failed login attempts. Your account is temporarily locked."
    ErrMsgDuplicateAlias = "Duplicate alias %q found in team configuration"
    ErrMsgInvalidAlias = "Team member alias must be alphanumeric (no spaces)"
)
```

### Timeouts

```go
const (
    TokenValidationTimeout = 10 * time.Second
    TokenCacheValidation   = 1 * time.Hour
)
```

---

## API Types

### Validation Response

**Endpoint:** `GET /rest/api/2/myself`

**Response Type:**
```go
type JiraUserResponse struct {
    EmailAddress string `json:"emailAddress"`
    DisplayName  string `json:"displayName"`
    Active       bool   `json:"active"`
}
```

**Example Response:**
```json
{
  "self": "https://your-domain.atlassian.net/rest/api/2/user?accountId=...",
  "accountId": "5b10ac8d82e05b22cc7d4ef5",
  "emailAddress": "user@company.com",
  "displayName": "John Doe",
  "active": true
}
```

---

## Type Mapping Reference

**Domain → Configuration:**
| Domain Type | YAML Type | Conversion |
|-------------|-----------|------------|
| `string` | `string` | Direct |
| `[]TeamMember` | `array` | YAML unmarshal |
| `AtlassianConfig` | `object` | YAML unmarshal |

**Domain → API:**
| Domain Type | JSON Type | HTTP Method |
|-------------|-----------|-------------|
| `email, token` | Basic Auth header | GET /rest/api/2/myself |
| `emailAddress` | `string` | Response field |

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-09 | Initial version with Atlassian types and TeamMember alias |
