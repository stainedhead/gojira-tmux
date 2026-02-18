# Security & Team Member Updates - Research

**Created:** 2026-02-09
**Source:** `Feature-Security-Updates.md`
**Status:** In Progress

---

## Overview

Research findings for replacing Okta OAuth with Atlassian API token authentication and adding team member alias support.

**Research Questions:**
1. How does Atlassian API token authentication work?
2. What is the recommended approach for token validation?
3. What are best practices for storing API tokens securely?
4. How should team member alias matching be prioritized?
5. What are common pitfalls when migrating from OAuth to tokens?

---

## Atlassian API Token Authentication

### Token Generation

**URL:** https://id.atlassian.com/manage/api-tokens

**Process:**
1. User logs into Atlassian account
2. Navigates to API token management page
3. Clicks "Create API token"
4. Provides label (e.g., "gojira-tmux")
5. Token is generated and displayed (only once)
6. User must copy and store securely

**Token Properties:**
- Format: Alphanumeric string (exact format undocumented by Atlassian)
- Length: Variable, typically 50-100 characters
- Expiry: No automatic expiry (user-revocable only)
- Scope: Full API access (same permissions as password)
- Revocation: Can be revoked individually from management page

### Authentication Method

**HTTP Basic Auth:**
```
Authorization: Basic base64(email:api_token)
```

**Example in Go:**
```go
req.SetBasicAuth(email, token)
// This sets: Authorization: Basic [base64-encoded email:token]
```

**Headers Required:**
```
Authorization: Basic [base64(email:token)]
Accept: application/json
Content-Type: application/json
```

### Token Validation Endpoint

**Endpoint:** `GET /rest/api/2/myself`

**Purpose:** Lightweight endpoint to validate credentials and get user info

**Request:**
```http
GET https://your-domain.atlassian.net/rest/api/2/myself
Authorization: Basic base64(email:token)
Accept: application/json
```

**Response (Success - 200 OK):**
```json
{
  "self": "https://your-domain.atlassian.net/rest/api/2/user?accountId=...",
  "accountId": "5b10ac8d82e05b22cc7d4ef5",
  "emailAddress": "user@company.com",
  "displayName": "John Doe",
  "active": true,
  "timeZone": "America/New_York",
  "locale": "en_US"
}
```

**Error Responses:**
- **401 Unauthorized**: Invalid email or token
  ```json
  {
    "errorMessages": ["Basic authentication with passwords is deprecated."],
    "errors": {}
  }
  ```

- **403 Forbidden**: Valid credentials but insufficient permissions
  ```json
  {
    "errorMessages": ["You do not have permission to access this resource."],
    "errors": {}
  }
  ```

- **429 Too Many Requests**: Rate limited
  ```
  X-RateLimit-Limit: 300
  X-RateLimit-Remaining: 0
  X-RateLimit-Reset: 1234567890
  ```

- **CAPTCHA Triggered**:
  ```
  X-Seraph-LoginReason: AUTHENTICATION_DENIED
  ```
  Message: "Too many failed login attempts. Account temporarily locked."

### Security Benefits

**Compared to OAuth:**
- Simpler implementation (no callback server, no PKCE)
- Fewer attack surfaces (no state management, no redirect flow)
- Direct API access without session management
- Individual token revocation (per application)

**Best Practices:**
- Never log tokens
- Store in OS keychain/keyring (not in config files)
- Never commit tokens to version control
- Use separate tokens for different applications
- Rotate tokens periodically
- Revoke immediately if compromised

**MFA Compatibility:**
- API tokens work with MFA-enabled accounts
- Token bypasses MFA requirement (by design)
- Primary account password remains protected

---

## Token Storage Strategy

### OS-Specific Secure Storage

**macOS:**
- **Storage**: Keychain Services
- **Access**: `security` command-line tool or APIs
- **Example**:
  ```bash
  # Store
  security add-generic-password -a "gojira-tmux" -s "jira-token" -w "token-value"

  # Retrieve
  security find-generic-password -a "gojira-tmux" -s "jira-token" -w
  ```

**Linux:**
- **Storage**: Secret Service API (GNOME Keyring, KWallet)
- **Access**: DBus interface
- **Fallback**: Encrypted file with user permissions

**Windows:**
- **Storage**: Credential Manager
- **Access**: Windows Credential APIs
- **Example**: `cmdkey` command

**Current Implementation:**
- Using `internal/adapter/auth/token_store.go`
- Already implements keychain storage
- Need to remove refresh token methods
- Keep Jira token methods unchanged

---

## Team Member Alias Matching

### Matching Algorithm

**Priority Order (from PRD):**
1. Exact alias match (case-sensitive)
2. Exact name match (case-sensitive)
3. Case-insensitive alias match
4. Case-insensitive name match (fallback)

**Rationale:**
- Aliases are short and explicit (user's intent clear)
- Exact matches preferred for performance
- Case-insensitive fallback for user convenience
- No regex (simple string comparison for speed)

**Example:**
```go
func (t *TeamMember) MatchesIdentifier(identifier string) bool {
    // Priority 1: Exact alias match
    if t.Alias != "" && t.Alias == identifier {
        return true
    }

    // Priority 2: Exact name match
    if t.Name == identifier {
        return true
    }

    // Priority 3: Case-insensitive alias match
    if t.Alias != "" && strings.EqualFold(t.Alias, identifier) {
        return true
    }

    // Priority 4: Case-insensitive name match
    return strings.EqualFold(t.Name, identifier)
}
```

**Performance:**
- O(n) where n = team size (typically < 20)
- 4 comparisons maximum per team member
- Early exit on first match
- No regex, no complex string operations

**Optimization (Future):**
- Build alias → member map at startup
- O(1) lookup instead of O(n) iteration
- Rebuild only on config reload

### Alias Format Validation

**Rules:**
- Optional field (backward compatible)
- Alphanumeric characters only (letters + numbers)
- No spaces, no special characters
- Case-sensitive storage
- Must be unique within team

**Validation Function:**
```go
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

**Error Messages:**
- "Team member alias must be alphanumeric (no spaces)"
- "Duplicate alias 'JohnA' found (used by John Anderson and John Andrews)"

---

## Migration Considerations

### Breaking Changes

**Configuration File:**
```yaml
# OLD (Okta)
okta:
  issuer: "https://company.okta.com/oauth2/default"
  client_id: "client-id"
  callback_port: 8080
  scopes: ["openid", "profile", "email"]

# NEW (Atlassian)
atlassian:
  email: "user@company.com"
  # Note: Token stored in keychain, not here
```

**Migration Steps for Users:**
1. Generate Atlassian API token
2. Update config.yaml (remove okta, add atlassian)
3. Clear old credentials: `rm -rf ~/.config/gojira/credentials`
4. Run app and enter token on setup screen
5. Optional: Add aliases to team members

### Code Removal

**Files to Delete:**
- `internal/adapter/auth/okta.go` (303 lines)
- `internal/adapter/auth/okta_test.go` (106 lines)
- `internal/adapter/auth/callback.go` (estimated 200 lines)
- `internal/adapter/auth/callback_test.go` (estimated 100 lines)

**Dependencies to Remove:**
- `github.com/coreos/go-oidc/v3`
- `golang.org/x/oauth2`

**Total Code Reduction:** ~700 lines of OAuth-related code

### Testing Migration

**Test Scenarios:**
1. **First-time setup**: App starts without token, user enters email + token
2. **Invalid token**: User enters wrong token, sees error, tries again
3. **Valid token**: Token validates, app proceeds to main screen
4. **Token persistence**: Token stored, survives app restart
5. **Token invalidation**: Token revoked externally, app detects on next API call
6. **Alias filtering**: Filter by alias "JohnA", correct team member selected
7. **Backward compatibility**: Team members without aliases still work

---

## Error Handling Strategy

### Token Validation Errors

**401 Unauthorized:**
```go
if resp.StatusCode == http.StatusUnauthorized {
    return errors.New("Invalid email or API token. Please verify your credentials at https://id.atlassian.com/manage/api-tokens")
}
```

**403 Forbidden:**
```go
if resp.StatusCode == http.StatusForbidden {
    return errors.New("Valid credentials but insufficient permissions. Contact your Jira administrator.")
}
```

**Network Errors:**
```go
if err != nil {
    return fmt.Errorf("Could not reach Jira API. Check your internet connection and verify Jira URL in config.")
}
```

**CAPTCHA Detection:**
```go
if resp.Header.Get("X-Seraph-LoginReason") == "AUTHENTICATION_DENIED" {
    return errors.New("Too many failed login attempts. Your account is temporarily locked. Wait 10 minutes or reset your password.")
}
```

### User-Friendly Messages

**Principles:**
- Don't expose internal details
- Don't specify which credential (email or token) is wrong
- Provide actionable next steps
- Include links to help resources

**Examples:**
- ❌ "Database error: connection timeout"
- ✅ "Could not connect to Jira. Check your internet connection."

- ❌ "HTTP 401: Invalid token"
- ✅ "Authentication failed. Verify your email and token."

---

## Performance Optimization

### Token Validation Caching

**Strategy:**
- Validate on startup (~100-200ms)
- Cache validation result for 1 hour
- Skip validation if token used successfully in last hour
- Revalidate on 401 errors from API

**Implementation:**
```go
type TokenValidator struct {
    lastValidation time.Time
    isValid        bool
    mu             sync.RWMutex
}

func (v *TokenValidator) IsValid(ctx context.Context) bool {
    v.mu.RLock()
    defer v.mu.RUnlock()

    // Cache hit (validated within last hour)
    if v.isValid && time.Since(v.lastValidation) < 1*time.Hour {
        return true
    }

    // Cache miss - need to revalidate
    return false
}
```

### Team Member Lookup Optimization

**Current:** O(n) linear search
**Future:** O(1) map lookup

```go
type TeamMemberIndex struct {
    byAlias map[string]*domain.TeamMember
    byName  map[string]*domain.TeamMember
}

func BuildIndex(team []domain.TeamMember) *TeamMemberIndex {
    idx := &TeamMemberIndex{
        byAlias: make(map[string]*domain.TeamMember),
        byName:  make(map[string]*domain.TeamMember),
    }

    for i := range team {
        member := &team[i]
        if member.Alias != "" {
            idx.byAlias[member.Alias] = member
            idx.byAlias[strings.ToLower(member.Alias)] = member
        }
        idx.byName[member.Name] = member
        idx.byName[strings.ToLower(member.Name)] = member
    }

    return idx
}
```

**When to Optimize:**
- Current O(n) is acceptable for teams < 100
- Only optimize if performance issues observed
- Keep simple implementation initially

---

## Open Questions

### Question 1: Environment Variable for Email

**Context:** Should we support `JIRA_EMAIL` environment variable in addition to config file?

**Research Needed:**
- Check if other CLI tools support this pattern
- Consider security implications (emails in env vars)
- Evaluate convenience vs. security trade-off

**Decision:** TBD - propose in plan.md

### Question 2: Alias Maximum Length

**Context:** Should we enforce a maximum length for aliases?

**Research Needed:**
- Check typical alias patterns (initials, short names)
- Consider UI display constraints
- Balance between flexibility and usability

**Decision:** Recommend 20 characters maximum

### Question 3: Token Expiry Handling

**Context:** Atlassian tokens don't expire automatically. Should we warn users to rotate?

**Research Needed:**
- Security best practices for long-lived tokens
- Other tools' approaches to token rotation
- User experience trade-offs

**Decision:** Out of scope for initial implementation. Mark as future enhancement.

---

## References

**Official Documentation:**
- Atlassian API Tokens: https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
- Token Management: https://id.atlassian.com/manage/api-tokens
- Jira REST API: https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/

**Security Resources:**
- OWASP Credential Storage: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- Keychain Security (macOS): https://developer.apple.com/documentation/security/keychain_services

**Source PRD:** `Feature-Security-Updates.md`
