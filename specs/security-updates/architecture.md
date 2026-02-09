# Security & Team Member Updates - System Architecture

**Created:** 2026-02-09
**Version:** 1.0
**Status:** Draft
**Last Updated:** 2026-02-09

---

## Architecture Overview

**High-Level Summary:**
This feature simplifies the authentication architecture by replacing Okta OAuth OIDC with Atlassian API token authentication. It removes complex OAuth flows (callback servers, PKCE, session management) in favor of direct API token validation. Additionally, it enhances team member disambiguation by adding optional alias support.

**Architectural Style:** Clean Architecture with simplified authentication

**Key Principles:**
- Simplification over complexity (remove OAuth, use direct token auth)
- Backward compatibility (aliases optional)
- Security by design (keychain storage, no token exposure)
- Clean separation of concerns (domain, use case, infrastructure, adapter)

**Architecture Diagram:**
```
┌─────────────────────────────────────────────────────────────┐
│                    User Interface (TUI)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Setup Screen │  │ Filter Bar   │  │ Main Screen  │      │
│  │ (Token Input)│  │ (Show Alias) │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────────────┬────────────────────────────────────┘
                         │ uses
┌────────────────────────▼────────────────────────────────────┐
│                   Use Case Layer                             │
│  ┌──────────────────────────────────────┐                   │
│  │     Authenticate (Simplified)        │                   │
│  │  - ValidateAndSaveToken()            │                   │
│  │  - HasValidToken()                   │                   │
│  │  - ClearToken()                      │                   │
│  └──────────────────────────────────────┘                   │
└────────────────────────┬────────────────────────────────────┘
                         │ uses
┌────────────────────────▼────────────────────────────────────┐
│                   Adapter Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Atlassian    │  │ Config       │  │ JQL Builder  │      │
│  │ Adapter      │  │ Loader       │  │ (Alias       │      │
│  │ (Token Auth) │  │ (Validate)   │  │  Support)    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────────────┬────────────────────────────────────┘
                         │ implements
┌────────────────────────▼────────────────────────────────────┐
│                  Domain Layer                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐              │
│  │ AuthPort │  │TeamMember│  │AtlassianConfig│             │
│  │(Simplified)  │(+Alias)  │  │              │              │
│  └──────────┘  └──────────┘  └──────────────┘              │
└─────────────────────────────────────────────────────────────┘
         │                           │
         ▼                           ▼
┌──────────────┐            ┌──────────────┐
│ Jira REST    │            │OS Keychain   │
│ API          │            │(Token Store) │
└──────────────┘            └──────────────┘
```

---

## System Context

**External Systems:**
```
┌──────────────┐
│     User     │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────────────┐
│      gojira-tmux System               │
│  ┌────────────────────────────────┐  │
│  │  Authentication (Simplified)   │  │
│  │  ┌──────────┐  ┌──────────┐   │  │
│  │  │Atlassian │  │ Token    │   │  │
│  │  │ Adapter  │  │ Store    │   │  │
│  │  └──────────┘  └──────────┘   │  │
│  └────────────────────────────────┘  │
│                                       │
│  ┌────────────────────────────────┐  │
│  │  Team Member Management        │  │
│  │  ┌──────────┐  ┌──────────┐   │  │
│  │  │ Alias    │  │ JQL      │   │  │
│  │  │ Matching │  │ Builder  │   │  │
│  │  └──────────┘  └──────────┘   │  │
│  └────────────────────────────────┘  │
└──────┬───────────────────┬───────────┘
       │                   │
       ▼                   ▼
┌──────────────┐    ┌──────────────┐
│ Jira REST    │    │OS Keychain   │
│ API          │    │              │
└──────────────┘    └──────────────┘
```

**System Boundaries:**
- **Inputs:** User email + API token (one-time setup), filter selections (alias or name)
- **Outputs:** Authenticated API requests, team member matches
- **External Dependencies:** Jira Cloud API, OS keychain service

**Integration Points:**
| System | Type | Protocol | Purpose |
|--------|------|----------|---------|
| Jira Cloud | REST API | HTTPS/JSON | Authentication validation, issue queries |
| OS Keychain | Secure Storage | Native APIs | Token persistence |

---

## Component Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────┐
│           Authentication Components                 │
│                                                     │
│  ┌─────────────────┐         ┌─────────────────┐  │
│  │ AtlassianAdapter│────────>│ TokenStorePort  │  │
│  │                 │         │                 │  │
│  │ - ValidateToken()│         │ - GetJiraToken()│  │
│  │ - IsTokenValid()│         │ - SetJiraToken()│  │
│  └─────────────────┘         └─────────────────┘  │
│         │                                           │
│         │ calls                                     │
│         ▼                                           │
│  ┌─────────────────┐                               │
│  │ Jira /myself    │                               │
│  │ Endpoint        │                               │
│  └─────────────────┘                               │
└────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────┐
│           Team Member Components                    │
│                                                     │
│  ┌─────────────────┐         ┌─────────────────┐  │
│  │ TeamMember      │────────>│ JQLBuilder      │  │
│  │                 │         │                 │  │
│  │ - Alias         │         │ - buildAssignee()│  │
│  │ - MatchesID()   │         │                 │  │
│  │ - DisplayName() │         │                 │  │
│  └─────────────────┘         └─────────────────┘  │
└────────────────────────────────────────────────────┘
```

### Component Descriptions

#### Component: AtlassianAdapter

**Responsibility:**
Validates Atlassian API tokens by calling the Jira `/rest/api/2/myself` endpoint.

**Dependencies:**
- TokenStorePort (to retrieve stored token)
- HTTP Client (to call Jira API)
- Jira base URL from config

**Provides:**
- `AuthPort` interface implementation
- Token validation with error handling
- Basic caching strategy (future enhancement)

**Lifecycle:**
- Created during app initialization
- Singleton instance
- Lives for app lifetime

**Concurrency:**
- Thread-safe (HTTP client is thread-safe)
- No internal state mutation after creation

---

#### Component: TeamMember (Enhanced)

**Responsibility:**
Represents team member with optional alias for disambiguation. Provides matching logic for filters.

**Dependencies:**
- None (pure domain entity)

**Provides:**
- Alias storage and validation
- Multi-priority matching algorithm
- Display name formatting

**Lifecycle:**
- Created during config load
- Immutable after validation
- Lives for config lifetime

**Concurrency:**
- Thread-safe (immutable after creation)
- Read-only operations

---

## Layer Responsibilities

### Domain Layer

**Location:** `internal/domain/`

**Responsibility:**
- Define `AuthPort` interface (simplified from Okta version)
- Define `TeamMember` entity with alias field
- Define `AtlassianConfig` struct
- Define `TokenStorePort` interface (remove refresh token methods)

**Contains:**
- `AuthPort` (2 methods vs 7 in Okta version)
- `TeamMember` (+Alias field, +MatchesIdentifier(), +DisplayName())
- `AtlassianConfig` (replaces OktaConfig)
- `TokenStorePort` (4 methods vs 7 in Okta version)

**Dependencies:** None (pure domain logic)

---

### Use Case Layer

**Location:** `internal/usecase/`

**Responsibility:**
- Orchestrate token validation workflow
- Coordinate between domain and adapters

**Contains:**
- `Authenticate` service (simplified from 7 methods to 3)
  - `ValidateAndSaveToken(ctx, email, token)`
  - `HasValidToken(ctx)`
  - `ClearToken()`
- `SetupToken` service (minimal changes)

**Dependencies:**
- Domain layer (AuthPort, TokenStorePort)

---

### Adapter Layer

**Location:** `internal/adapter/`

**Responsibility:**
- Implement `AuthPort` with Atlassian token validation
- Validate config with alias uniqueness checks
- Build JQL queries with alias support

**Contains:**
- `AtlassianAdapter` (new, replaces OktaAdapter)
- `ConfigLoader` (enhanced validation for aliases)
- `JQLBuilder` (enhanced assignee matching)

**Dependencies:**
- Domain layer (interfaces)
- External libraries (HTTP client, YAML parser)

---

### Infrastructure Layer

**Location:** `internal/infrastructure/tui/`

**Responsibility:**
- Display token input form (replaces OAuth flow)
- Show aliases in filter dropdown
- Handle user interactions

**Contains:**
- `SetupScreen` (redesigned for token input)
- `FilterBar` (enhanced to show aliases)
- `LoginScreen` (simplified or removed)
- `App` (updated initialization)

**Dependencies:**
- Use case layer (Authenticate service)
- BubbleTea framework

---

## Data Flow

### Token Validation Flow

**Step-by-Step:**
1. User enters email + token in setup screen
2. Setup screen calls `Authenticate.ValidateAndSaveToken(ctx, email, token)`
3. Authenticate use case calls `AuthPort.ValidateToken(ctx, email, token)`
4. AtlassianAdapter builds HTTP request with Basic Auth
5. AtlassianAdapter calls `GET /rest/api/2/myself`
6. Jira API returns 200 OK with user info (or error)
7. Atlassian Adapter validates response email matches input email
8. Authenticate use case calls `TokenStorePort.SetJiraToken(token)`
9. Token stored in OS keychain
10. Setup screen transitions to main screen

**Sequence Diagram:**
```
User   SetupScreen  Authenticate  AtlassianAdapter  JiraAPI  TokenStore
 |          |           |               |             |          |
 |─input───>|           |               |             |          |
 |          |─validate─>|               |             |          |
 |          |           |─validateToken>|             |          |
 |          |           |               |─GET /myself>|          |
 |          |           |               |<─200 OK────|          |
 |          |           |               |─verify─────|          |
 |          |           |<─email────────|             |          |
 |          |           |─saveToken────────────────────────────>|
 |          |           |<─success──────────────────────────────|
 |          |<─success──|               |             |          |
 |<─proceed─|           |               |             |          |
```

---

### Team Member Alias Matching Flow

**Step-by-Step:**
1. User selects "JohnA" in filter dropdown
2. Filter bar calls `JQLBuilder.buildAssigneeCondition("JohnA")`
3. JQLBuilder iterates through team members
4. For each member, calls `TeamMember.MatchesIdentifier("JohnA")`
5. TeamMember returns true for exact alias match
6. JQLBuilder uses matched member's email in JQL
7. JQL query: `assignee = "john.anderson@company.com"`

**Matching Algorithm:**
```
Input: "JohnA"

TeamMember 1: Name="John Anderson", Alias="JohnA", Email="john.anderson@company.com"
  - Exact alias match? "JohnA" == "JohnA" → YES ✅
  - Return true, use email

Result: JQL = 'assignee = "john.anderson@company.com"'
```

---

## Integration Points

### Integration 1: Jira REST API

**Type:** REST API
**Purpose:** Token validation and issue queries
**Protocol:** HTTPS/JSON

**Endpoint:**
```
GET /rest/api/2/myself
Authorization: Basic base64(email:token)
Accept: application/json
```

**Response:**
```json
{
  "emailAddress": "user@company.com",
  "displayName": "John Doe",
  "active": true
}
```

**Error Handling:**
- 401: Invalid credentials → Clear error message, prompt retry
- 403: Insufficient permissions → Suggest contacting admin
- 429: Rate limited → Backoff and retry
- Network errors: Check connection, verify URL

---

### Integration 2: OS Keychain

**Type:** Secure Storage
**Purpose:** Token persistence
**Protocol:** Native OS APIs

**Implementation:**
- macOS: Keychain Services
- Linux: Secret Service API
- Windows: Credential Manager

**Error Handling:**
- Keychain locked: Prompt user to unlock
- Permission denied: Fallback to encrypted file
- Service unavailable: Warn user, use in-memory (insecure)

---

## Architectural Decisions

### ADR-001: Replace Okta OAuth with Atlassian API Tokens

**Date:** 2026-02-09
**Status:** Accepted

**Context:**
The application was using Okta OIDC for user authentication, but Jira API access should use Atlassian API tokens directly. This created unnecessary complexity and mixed authentication mechanisms.

**Decision:**
Remove Okta OAuth entirely and replace with direct Atlassian API token authentication.

**Rationale:**
- Aligns with Atlassian's recommended approach
- Simplifies architecture (no callback server, no PKCE, no session management)
- Reduces attack surface (fewer components, fewer dependencies)
- Easier for users (one-time token setup vs browser auth flow)
- Better security model (individual token revocation, MFA compatible)

**Consequences:**
**Positive:**
- ~700 lines of code removed
- 2 dependencies removed (go-oidc, oauth2)
- Simpler testing (no OAuth mocks needed)
- Faster startup (no OAuth provider initialization)
- Better user experience (no browser redirects)

**Negative:**
- Breaking change (users must migrate config)
- Requires migration guide
- Users must generate API tokens manually

**Alternatives Considered:**
1. **Keep Okta + Add Atlassian tokens**
   - Pros: No breaking change
   - Cons: Maintains complexity, mixed auth mechanisms
   - Why rejected: Doesn't solve core problem

2. **OAuth 2.0 with Atlassian**
   - Pros: Standards-based auth
   - Cons: Still complex, requires app registration
   - Why rejected: Overkill for CLI tool, Atlassian recommends tokens

---

### ADR-002: Add Optional Alias Field to Team Members

**Date:** 2026-02-09
**Status:** Accepted

**Context:**
When multiple team members share common first names (e.g., "John Anderson" and "John Flanagan"), there's no short way to distinguish them in filters.

**Decision:**
Add optional `alias` field to team member configuration with multi-priority matching.

**Rationale:**
- Solves disambiguation problem
- Maintains backward compatibility (optional field)
- Simple implementation (no regex, just string comparison)
- User-friendly (short identifiers like "JohnA", "JohnF")

**Consequences:**
**Positive:**
- Better UX for teams with similar names
- Faster filtering (type "JohnA" vs "John Anderson")
- More flexible (users choose their own aliases)

**Negative:**
- Additional configuration required
- Need to validate alias uniqueness
- Slightly more complex matching logic

**Alternatives Considered:**
1. **Auto-generate aliases from names**
   - Pros: No user configuration
   - Cons: Might generate ambiguous aliases (JA, JA)
   - Why rejected: User choice is better

2. **Use email prefixes as aliases**
   - Pros: Automatic, always unique
   - Cons: Less readable ("john.anderson" vs "JohnA")
   - Why rejected: User-defined is more flexible

---

## Trade-offs

### Trade-off 1: Simplicity vs Feature Completeness

**Choice:** Remove OAuth entirely in favor of API tokens

**Benefits:**
- Significantly simpler codebase
- Easier to understand and maintain
- Fewer failure modes

**Costs:**
- Breaking change for existing users
- No SSO integration (users manage tokens manually)
- Users must visit Atlassian to generate tokens

**Mitigation:**
- Detailed migration guide with screenshots
- Clear setup instructions in TUI
- Environment variable support for testing/CI

---

### Trade-off 2: Performance vs Simplicity

**Choice:** O(n) linear search for team member matching

**Benefits:**
- Simple, easy to understand
- No premature optimization
- Acceptable for typical team sizes (<100)

**Costs:**
- Not optimal for very large teams
- Multiple passes (4 max) per lookup

**Mitigation:**
- Early exit on first match
- Future: Build index if performance issues observed
- Document O(1) optimization strategy for future

---

## Security Architecture

**Security Layers:**
1. **Token Input** (TUI layer): Masked input, no logging
2. **Token Storage** (Infrastructure layer): OS keychain, never in config
3. **Token Transmission** (Adapter layer): HTTPS only, Basic Auth header
4. **Token Validation** (Use case layer): Verify with Jira API

**Threat Model:**
- **Token exposure in logs**: Mitigation: Never log tokens
- **Token in config file**: Mitigation: Store only in keychain
- **Token in memory**: Mitigation: Minimize lifetime, no dumps
- **MITM attacks**: Mitigation: HTTPS required for Jira URL

**Security Controls:**
- Input validation (email format, token non-empty)
- HTTPS enforcement (reject HTTP URLs)
- Secure storage (keychain with OS-level encryption)
- Error message sanitization (don't expose internals)

---

## References

- [spec.md](spec.md) - Feature specification
- [data-dictionary.md](data-dictionary.md) - Data structures
- [plan.md](plan.md) - Implementation plan
- [Feature-Security-Updates.md](../../Feature-Security-Updates.md) - Original PRD
