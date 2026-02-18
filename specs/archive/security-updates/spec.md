# Security & Team Member Updates - Specification

**Created:** 2026-02-09
**Version:** 1.0
**Status:** Draft
**Source PRD:** `Feature-Security-Updates.md`

---

## Executive Summary

Replace Okta OAuth authentication with Atlassian API token-based authentication and add team member alias support. This simplifies the authentication architecture while improving team member disambiguation.

**Key Deliverables:**
- Atlassian API token authentication (replacing Okta OIDC)
- Team member alias system for disambiguation
- Simplified configuration schema
- Updated TUI for token setup
- Migration guide for users

**Timeline:** Estimated 12 hours with parallelization (6 phases)

---

## Problem Statement

### Current State
The application uses Okta OIDC for user authentication with a complex OAuth flow (callback server, PKCE, session management), while Jira API access should use Atlassian API tokens directly. Team members with similar names (e.g., "John Anderson" and "John Flanagan") cannot be easily distinguished in filters.

### Pain Points
- **Wrong authentication method**: Using OAuth instead of recommended Atlassian API tokens
- **Unnecessary complexity**: Callback server, PKCE, refresh tokens, 8-hour session tracking
- **Mixed auth mechanisms**: OAuth for users + separate Jira tokens
- **Team member ambiguity**: No way to distinguish "John A" from "John F"
- **Security overhead**: Multiple attack surfaces (callback server, OAuth state management)

### Desired State
- Single, straightforward authentication using Atlassian API tokens
- Simple token input on first run (email + token)
- Team member aliases for easy disambiguation
- Simplified codebase with fewer dependencies
- Better security posture with fewer components

---

## Goals and Non-Goals

### Goals
- Replace Okta OAuth with Atlassian API token authentication
- Remove all OAuth-related code and dependencies
- Add optional alias field to team member configuration
- Support alias-based team member matching
- Maintain backward compatibility for team members without aliases
- Provide clear migration path for existing users
- Simplify TUI authentication flow

### Non-Goals
- Multiple Atlassian account support (future enhancement)
- Auto-generated aliases (manual configuration required)
- Token rotation/expiry detection (future enhancement)
- Biometric authentication (future enhancement)
- Validation of aliases against Jira usernames

---

## User Requirements

### Functional Requirements

#### FR-001: Atlassian API Token Collection
**Priority:** P0 (Critical)

**Description:**
Collect and validate Atlassian API token on first run or when token is missing.

**Acceptance Criteria:**
- [ ] FR-1.1: Display setup screen requesting email and API token on first run
- [ ] FR-1.2: Provide instructions with link to https://id.atlassian.com/manage/api-tokens
- [ ] FR-1.3: Validate token format (non-empty, trimmed)
- [ ] FR-1.4: Test token with authenticated request to `/rest/api/2/myself`
- [ ] FR-1.5: Store token securely in OS keychain/keyring
- [ ] FR-1.6: Show setup screen when stored token is missing or invalid
- [ ] FR-1.7: Allow token regeneration via config or CLI flag

**Example:**
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

#### FR-002: Remove Okta Authentication
**Priority:** P0 (Critical)

**Description:**
Completely remove Okta OIDC authentication flow and all related code.

**Acceptance Criteria:**
- [ ] FR-2.1: Remove all Okta-specific configuration options
- [ ] FR-2.2: Remove OAuth callback server implementation
- [ ] FR-2.3: Remove session management (no longer needed)
- [ ] FR-2.4: Remove refresh token logic
- [ ] FR-2.5: Remove browser-based auth flow
- [ ] FR-2.6: Remove Okta dependencies (go-oidc, oauth2)

**Example:**
Files to delete:
- `internal/adapter/auth/okta.go`
- `internal/adapter/auth/okta_test.go`
- `internal/adapter/auth/callback.go`
- `internal/adapter/auth/callback_test.go`

#### FR-003: Configuration Schema Updates
**Priority:** P0 (Critical)

**Description:**
Update configuration file schema to support new authentication and team aliases.

**Acceptance Criteria:**
- [ ] FR-3.1: Remove `okta` configuration block
- [ ] FR-3.2: Add `atlassian` configuration block with `email` field
- [ ] FR-3.3: Add optional `alias` field to team members
- [ ] FR-3.4: Validate alias uniqueness within team
- [ ] FR-3.5: Validate alias format (alphanumeric, no spaces)
- [ ] FR-3.6: Maintain backward compatibility for team members without aliases
- [ ] FR-3.7: Update config validation with helpful error messages

**Example:**
```yaml
jira:
  url: "https://your-company.atlassian.net"

atlassian:
  email: "your-email@company.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Anderson"
    email: "john.anderson@company.com"
    alias: "JohnA"
  - name: "John Flanagan"
    email: "john.flanagan@company.com"
    alias: "JohnF"
```

#### FR-004: Team Member Alias Matching
**Priority:** P1 (High)

**Description:**
Support both name and alias when filtering or matching team members.

**Acceptance Criteria:**
- [ ] FR-4.1: Match by exact alias (case-sensitive) first
- [ ] FR-4.2: Fall back to exact name match if alias not found
- [ ] FR-4.3: Support case-insensitive matching as last resort
- [ ] FR-4.4: Display both name and alias in UI where space permits
- [ ] FR-4.5: Update JQL query builder to use email regardless of match method
- [ ] FR-4.6: Maintain performance (O(n), no regex)

**Example:**
```go
// Priority 1: Exact alias match (case-sensitive)
// Priority 2: Exact name match (case-sensitive)
// Priority 3: Case-insensitive alias match
// Priority 4: Case-insensitive name match (fallback)
```

#### FR-005: Token Validation
**Priority:** P0 (Critical)

**Description:**
Validate API token on app startup and handle invalid tokens gracefully.

**Acceptance Criteria:**
- [ ] FR-5.1: Test token with lightweight API call on startup (`/rest/api/2/myself`)
- [ ] FR-5.2: Handle 401 Unauthorized by requesting new token
- [ ] FR-5.3: Handle 403 Forbidden with clear permission error
- [ ] FR-5.4: Display helpful error messages for common failures
- [ ] FR-5.5: Support token refresh without app restart

**Example:**
```
GET /rest/api/2/myself
Authorization: Basic base64(email:token)
```

---

### Non-Functional Requirements

#### NFR-001: Performance
**Category:** Performance

**Description:**
Authentication and team member operations must remain fast and responsive.

**Metrics:**
- App startup < 2 seconds (with valid token)
- Token validation < 500ms
- Team member matching < 1ms

#### NFR-002: Security
**Category:** Security

**Description:**
API tokens must be stored securely and never exposed in logs or UI.

**Metrics:**
- Tokens stored in OS keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager)
- Tokens never logged or displayed unmasked
- Token validation errors don't expose internals

#### NFR-003: Backward Compatibility
**Category:** Usability

**Description:**
Team members without aliases must continue to work without changes.

**Metrics:**
- Existing configs load successfully (after migration)
- Team member filtering works with or without aliases
- No breaking changes to domain models (except config schema)

---

## System Architecture

### Affected Layers
- [x] Domain Layer (ports, team member, user models)
- [x] Use Case Layer (authenticate, setup token)
- [x] Infrastructure Layer (TUI screens)
- [x] Adapter Layer (auth, config, jira)

### New Components
- `AtlassianAdapter`: Handles API token validation
- `TeamMember.Alias`: Optional field for short identifiers
- `TeamMember.MatchesIdentifier()`: Alias matching logic
- `TeamMember.DisplayName()`: Format name with alias

### Modified Components
- `AuthPort`: Simplified to token validation only
- `Config`: Replace `Okta` with `Atlassian`
- `TokenStorePort`: Remove refresh token methods
- `Authenticate` use case: Simplified token flow
- `SetupScreen`: Replace OAuth with token input
- `FilterBar`: Display aliases
- `JQLBuilder`: Support alias matching
- `main.go`: Use AtlassianAdapter

---

## Scope of Changes

### Files to Create
- `internal/adapter/auth/atlassian.go` - Atlassian API token adapter
- `internal/adapter/auth/atlassian_test.go` - Tests for token validation

### Files to Modify
- `internal/domain/ports.go` - Simplify AuthPort, update Config structs
- `internal/domain/team_member.go` - Add Alias field and matching methods
- `internal/domain/user.go` - Remove session management
- `internal/adapter/config/config.go` - Update validation for Atlassian + aliases
- `internal/adapter/jira/search.go` - Update assignee matching for aliases
- `internal/usecase/authenticate.go` - Simplify to token validation
- `internal/usecase/setup_token.go` - Update for email validation
- `internal/infrastructure/tui/setup_screen.go` - Replace OAuth with token input
- `internal/infrastructure/tui/login_screen.go` - Simplify or remove
- `internal/infrastructure/tui/filter_bar.go` - Display aliases
- `internal/infrastructure/tui/app.go` - Update initialization
- `cmd/gojira/main.go` - Use AtlassianAdapter
- `go.mod` - Remove OAuth dependencies

### Files to Delete
- `internal/adapter/auth/okta.go`
- `internal/adapter/auth/okta_test.go`
- `internal/adapter/auth/callback.go`
- `internal/adapter/auth/callback_test.go`

### Dependencies
- **Remove**: `github.com/coreos/go-oidc/v3`, `golang.org/x/oauth2`
- **Keep**: All existing dependencies (no new external deps)

---

## Breaking Changes

### API Changes
No public API changes (internal refactoring only).

### Configuration Changes
**Migration Required:**
```yaml
# OLD
okta:
  issuer: "..."
  client_id: "..."
  callback_port: 8080
  scopes: [...]

# NEW
atlassian:
  email: "your-email@company.com"
```

**Migration Path:**
1. Generate Atlassian API token at https://id.atlassian.com/manage/api-tokens
2. Update config.yaml (remove `okta`, add `atlassian`)
3. Clear old credentials: `rm -rf ~/.config/gojira/credentials`
4. Run app and enter token on setup screen

### Database Schema Changes
None (no database).

---

## Success Criteria

### Acceptance Criteria
- [ ] All Okta code removed
- [ ] Atlassian token authentication working
- [ ] Team member aliases supported
- [ ] Config validation includes alias uniqueness
- [ ] TUI displays aliases correctly
- [ ] Migration guide complete

### Quality Gates
- [ ] All tests pass
- [ ] Code coverage >85%
- [ ] Documentation complete (README, migration guide)
- [ ] No OAuth dependencies in go.mod
- [ ] Manual testing checklist complete

### User Validation
- [ ] First-run setup < 2 minutes
- [ ] Token instructions clear
- [ ] Aliases display in filter dropdown
- [ ] Filtering by alias works
- [ ] Error messages are actionable

---

## Risks and Mitigation

### Risk 1: Token Validation Endpoint Failures
**Likelihood:** Medium
**Impact:** High

**Mitigation:**
- Implement robust error handling (401, 403, network errors)
- Cache validation results (1 hour TTL)
- Provide clear error messages with troubleshooting steps
- Support environment variable override for testing

### Risk 2: User Migration Confusion
**Likelihood:** High
**Impact:** Medium

**Mitigation:**
- Create detailed migration guide with screenshots
- Provide example configs (old vs new)
- Include troubleshooting section
- Test migration flow with sample users

### Risk 3: Alias Conflicts
**Likelihood:** Low
**Impact:** Low

**Mitigation:**
- Validate alias uniqueness at config load
- Provide clear error messages for duplicates
- Document alias naming conventions
- Make aliases optional (backward compatible)

---

## Timeline and Milestones

### Phase 1: Domain Model Updates (1.5 hours)
**Deliverables:**
- Updated `ports.go` (AuthPort, Config structs)
- Updated `team_member.go` (Alias field, matching methods)
- Updated `user.go` (remove session management)

### Phase 2: Adapter Layer (4 hours)
**Deliverables:**
- New `AtlassianAdapter` with tests
- Deleted Okta components
- Updated config loader with alias validation
- Updated JQL builder for alias matching

### Phase 3: Use Case Layer (1.5 hours)
**Deliverables:**
- Simplified `Authenticate` use case
- Updated `SetupToken` use case
- Updated tests

### Phase 4: Infrastructure (4.5 hours)
**Deliverables:**
- Updated setup screen (token input form)
- Updated/removed login screen
- Updated filter bar (alias display)
- Updated app initialization

### Phase 5: Entry Point & Config (1.5 hours)
**Deliverables:**
- Updated `main.go` (use AtlassianAdapter)
- Updated README and docs
- Cleaned up dependencies (go mod tidy)

### Phase 6: Testing & Documentation (5 hours)
**Deliverables:**
- Integration tests
- Manual testing
- Migration guide
- Updated architecture docs

**Total Estimated Duration:** 12-18 hours

---

## References

- **Source PRD:** `Feature-Security-Updates.md`
- **Atlassian API Docs:** https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
- **Token Generation:** https://id.atlassian.com/manage/api-tokens
