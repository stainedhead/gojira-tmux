# Jira REST API v3 Migration - Data Dictionary

**Created:** 2026-02-12
**Version:** 1.0
**Status:** Draft

---

## Overview

This document defines all data structures, types, interfaces, and constants for the Jira REST API v3 Migration implementation.

**Purpose:**
- Document configuration model changes
- Define API response types for v3
- Specify validation rules
- Track type modifications

---

## Domain Layer

Location: `internal/domain/`

### 1. Config (Entity)

**File:** `ports.go`

```go
// Config represents the application configuration
type Config struct {
    Atlassian AtlassianConfig `yaml:"atlassian"`  // MODIFIED: was separate Jira + Atlassian
    Projects  []Project       `yaml:"projects"`   // UNCHANGED
    Team      []TeamMember    `yaml:"team"`       // UNCHANGED
}
```

**Changes from Current:**
- Removed `Jira JiraConfig` field
- Consolidated into single `Atlassian AtlassianConfig` field

---

### 2. AtlassianConfig (Value Object)

**File:** `ports.go`

```go
// AtlassianConfig holds Atlassian-specific configuration
type AtlassianConfig struct {
    URL          string            `yaml:"url"`              // NEW - moved from JiraConfig
    Email        string            `yaml:"email"`            // EXISTING
    CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"` // NEW - moved from JiraConfig
}
```

**New Fields:**
- `URL`: Atlassian instance URL (moved from `JiraConfig.URL`)
- `CustomFields`: Optional custom field mappings (moved from `JiraConfig.CustomFields`)

**Validation Rules:**
- `URL`: Required, must start with `https://`, should end with `.atlassian.net`
- `Email`: Required, must be valid email format
- `CustomFields`: Optional, no validation

---

### 3. JiraConfig (Deprecated)

**File:** `ports.go`

```go
// JiraConfig - DEPRECATED: Consolidated into AtlassianConfig
// This struct will be removed in this migration
type JiraConfig struct {
    URL          string            `yaml:"url"`
    Username     string            `yaml:"username"`          // Already unused
    CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"`
}
```

**Status:** TO BE REMOVED
**Migration:** Fields moved to `AtlassianConfig`

---

## Adapter Layer

Location: `internal/adapter/jira/`

### 1. searchResponse (API Response Type)

**File:** `client.go`

```go
// searchResponse represents v3 search API response
type searchResponse struct {
    Total         int             `json:"total"`                   // Total results
    Issues        []issueResponse `json:"issues"`                  // Issue array
    NextPageToken string          `json:"nextPageToken,omitempty"` // NEW in v3 - cursor for pagination
}
```

**Changes from v2:**
- Added `NextPageToken string` field
- Removed `StartAt int` field (no longer in v3 response)
- Removed `MaxResults int` field (no longer echoed in v3 response)

---

### 2. issueResponse (Unchanged)

**File:** `client.go`

```go
// issueResponse represents a single issue from API
type issueResponse struct {
    Key    string      `json:"key"`
    Fields issueFields `json:"fields"`
}
```

**Status:** NO CHANGES NEEDED (v3 compatible)

---

## Configuration

Location: `config.yaml`

### Atlassian Configuration (YAML)

```yaml
# NEW format (after migration)
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"
  custom_fields:                    # optional
    sprint: "customfield_10020"
    epic: "customfield_10014"
```

**Go Struct Mapping:**
```go
type AtlassianConfig struct {
    URL          string            `yaml:"url"`
    Email        string            `yaml:"email"`
    CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"`
}

type CustomFieldConfig struct {
    Sprint      string `yaml:"sprint,omitempty"`
    Epic        string `yaml:"epic,omitempty"`
    StoryPoints string `yaml:"story_points,omitempty"`
}
```

---

## API Endpoints

### Endpoint Changes

| Endpoint Type | v2 Path | v3 Path | Status |
|---------------|---------|---------|--------|
| Search | `/rest/api/2/search` | `/rest/api/3/search/jql` | MIGRATING |
| Auth | `/rest/api/2/myself` | `/rest/api/3/myself` | MIGRATING |
| Issue | `/rest/api/2/issue/{key}` | `/rest/api/3/issue/{key}` | MIGRATING |

---

## Constants

### API Endpoints

```go
const (
    // v3 API paths
    SearchEndpoint = "/rest/api/3/search/jql"
    MyselfEndpoint = "/rest/api/3/myself"
    IssueEndpoint  = "/rest/api/3/issue/%s"  // Format with issue key
)
```

### Pagination

```go
const (
    DefaultMaxResults = 100  // Default page size
)
```

---

## Type Mapping Reference

**Domain → API:**
| Domain Type | API JSON Field | v2 vs v3 Change |
|-------------|----------------|-----------------|
| `[]Issue` | `issues` array | No change |
| `int` (total) | `total` | No change |
| N/A | `nextPageToken` | NEW in v3 |
| N/A | `startAt` | REMOVED in v3 |

**Configuration → YAML:**
| Go Field | YAML Key | Required | Validation |
|----------|----------|----------|------------|
| `AtlassianConfig.URL` | `atlassian.url` | Yes | HTTPS, .atlassian.net |
| `AtlassianConfig.Email` | `atlassian.email` | Yes | Valid email |
| `CustomFieldConfig.*` | `atlassian.custom_fields.*` | No | None |

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-12 | Initial version - v3 migration planning |
