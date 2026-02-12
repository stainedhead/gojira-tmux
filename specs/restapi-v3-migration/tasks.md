# Jira REST API v3 Migration - Task Breakdown

**Feature:** Jira REST API v3 Migration
**Version:** 1.0
**Created:** 2026-02-12
**Status:** Planning
**Estimated Total Duration:** 10-14 hours (parallel) | 18-27 hours (sequential)

---

## Progress Summary

**Overall Progress:** 0/9 workstreams complete (0%)

**By Phase:**
- Phase 1 (Foundation): 0/4 workstreams complete
- Phase 2 (Implementation): 0/3 workstreams complete
- Phase 3 (Integration): 0/2 workstreams complete

**Critical Path Status:** Not started

---

## Workstream 1: Domain Model Updates (1-2 hours)

**ID:** WS1
**Dependencies:** NONE
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Can Run in Parallel With:** WS2, WS3, WS4

### Tasks

#### WS1.1: Update Config Struct
**Duration:** 30 min

**Files to Modify:**
- `internal/domain/ports.go`

**Changes:**
```go
// Remove or comment out JiraConfig field
// Add/expand AtlassianConfig field
type Config struct {
    Atlassian AtlassianConfig `yaml:"atlassian"`  // MODIFIED
    Projects  []Project       `yaml:"projects"`
    Team      []TeamMember    `yaml:"team"`
}
```

**Acceptance Criteria:**
- [ ] `Config` uses single `Atlassian` field
- [ ] Compiles without errors
- [ ] Existing tests updated

**Verification:**
```bash
go build ./internal/domain
```

---

#### WS1.2: Expand AtlassianConfig
**Duration:** 45 min

**Files to Modify:**
- `internal/domain/ports.go`

**Changes:**
```go
type AtlassianConfig struct {
    URL          string            `yaml:"url"`              // NEW
    Email        string            `yaml:"email"`            // EXISTING
    CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"` // NEW
}
```

**Acceptance Criteria:**
- [ ] `AtlassianConfig` includes URL field
- [ ] `CustomFields` moved from JiraConfig
- [ ] YAML tags correct

**Verification:**
```bash
go test ./internal/domain -v
```

---

#### WS1.3: Write Unit Tests
**Duration:** 15 min

**Files to Create/Modify:**
- `internal/domain/ports_test.go` (if needed)

**Test Cases:**
- [ ] Config unmarshal from YAML
- [ ] AtlassianConfig validation
- [ ] Custom fields optional

**Acceptance Criteria:**
- [ ] >95% coverage
- [ ] All tests pass

**Verification:**
```bash
go test ./internal/domain -cover
```

---

## Workstream 2: Configuration Research (2-3 hours)

**ID:** WS2
**Dependencies:** NONE
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Can Run in Parallel With:** WS1, WS3, WS4

### Tasks

#### WS2.1: Document Validation Requirements
**Duration:** 1 hour

**Deliverables:**
- Validation rule specification
- Error message examples
- Test case scenarios

**Acceptance Criteria:**
- [ ] URL validation rules documented
- [ ] Email validation rules documented
- [ ] Error messages helpful and clear

---

#### WS2.2-WS2.4: Complete Research
**Duration:** 1-2 hours

**See research.md for detailed research findings**

---

## Workstream 3: API Endpoint Research (3-4 hours)

**ID:** WS3
**Dependencies:** NONE
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Can Run in Parallel With:** WS1, WS2, WS4

### Tasks

**See research.md for complete API specification**

**Key Deliverables:**
- [ ] `/rest/api/3/search/jql` specification
- [ ] `/rest/api/3/myself` specification
- [ ] `/rest/api/3/issue/{key}` specification
- [ ] Response type Go structs
- [ ] Test request/response examples

---

## Workstream 4: Test Infrastructure Prep (2-3 hours)

**ID:** WS4
**Dependencies:** NONE
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Can Run in Parallel With:** WS1, WS2, WS3

### Tasks

#### WS4.1: Create Mock Server
**Duration:** 1 hour

**Files to Create:**
- `internal/adapter/jira/testutil/server.go`

**Implementation:**
```go
package testutil

// NewMockServerV3 creates HTTP test server for v3 API
func NewMockServerV3() *httptest.Server {
    // Implement handlers for:
    // - /rest/api/3/search/jql
    // - /rest/api/3/myself
    // - /rest/api/3/issue/{key}
}
```

**Acceptance Criteria:**
- [ ] Supports all three endpoints
- [ ] Handles pagination scenarios
- [ ] Returns realistic responses
- [ ] Simulates error conditions

---

#### WS4.2: Create Test Fixtures
**Duration:** 1 hour

**Files to Create:**
- `internal/adapter/jira/testutil/fixtures.go`

**Fixtures Needed:**
- [ ] Search response (single page)
- [ ] Search response (with nextPageToken)
- [ ] Myself response
- [ ] Issue response with comments
- [ ] Error responses (401, 403, 500)

---

## Workstream 5: Config Adapter Implementation (2-3 hours)

**ID:** WS5
**Dependencies:** WS1 (domain models), WS2 (validation rules)
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Blocks:** WS8

### Tasks

#### WS5.1: Update Validation Logic
**Duration:** 1 hour

**Files to Modify:**
- `internal/adapter/config/config.go`

**Changes:**
- Update `validate()` function
- Check for `atlassian.url` and `atlassian.email`
- Remove `jira.url` check
- Add helpful error messages

**Acceptance Criteria:**
- [ ] Validates `atlassian.url` (HTTPS, .atlassian.net)
- [ ] Validates `atlassian.email` (valid format)
- [ ] Clear error messages
- [ ] Tests updated and passing

---

#### WS5.2: Update Example Config
**Duration:** 15 min

**Files to Modify:**
- `config.example.yaml`

**New Format:**
```yaml
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"
```

---

#### WS5.3: Update Code References
**Duration:** 45 min

**Files to Modify:**
- `cmd/gojira/main.go`
- Any files using `cfg.Jira.URL`

**Changes:**
```go
// OLD: cfg.Jira.URL
// NEW: cfg.Atlassian.URL
```

**Verification:**
```bash
grep -r "cfg.Jira.URL" . --exclude-dir=specs
# Should return no results
```

---

## Workstream 6: Auth Adapter Implementation (1-2 hours)

**ID:** WS6
**Dependencies:** WS1, WS3, WS4
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Blocks:** WS8

### Task WS6.1: Update Auth Endpoint
**Duration:** 1-2 hours

**Files to Modify:**
- `internal/adapter/auth/atlassian.go` (line 34)
- `internal/adapter/auth/atlassian_test.go`

**Changes:**
```go
// OLD
url := a.jiraURL + "/rest/api/2/myself"

// NEW
url := a.jiraURL + "/rest/api/3/myself"
```

**Acceptance Criteria:**
- [ ] Endpoint updated to v3
- [ ] All tests use mock server from WS4
- [ ] All error scenarios tested
- [ ] >95% coverage
- [ ] All tests pass

---

## Workstream 7: Jira Client Implementation (3-4 hours)

**ID:** WS7
**Dependencies:** WS1, WS3, WS4
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Blocks:** WS8

### Task WS7.1: Update Search Endpoint
**Duration:** 2 hours

**Files to Modify:**
- `internal/adapter/jira/client.go` (line 40)
- `internal/adapter/jira/client_test.go`

**Changes:**
```go
// OLD
searchURL := fmt.Sprintf("%s/rest/api/2/search", c.baseURL)

// NEW
searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)
```

**Add Pagination Support:**
```go
type searchResponse struct {
    Total         int             `json:"total"`
    Issues        []issueResponse `json:"issues"`
    NextPageToken string          `json:"nextPageToken,omitempty"` // NEW
}
```

**Acceptance Criteria:**
- [ ] Uses `/rest/api/3/search/jql`
- [ ] Parses `nextPageToken`
- [ ] Returns first 100 results
- [ ] TODO added for multi-page

---

### Task WS7.2: Update Issue Endpoint
**Duration:** 1 hour

**Files to Modify:**
- `internal/adapter/jira/client.go` (line 75)

**Changes:**
```go
// OLD
issueURL := fmt.Sprintf("%s/rest/api/2/issue/%s?expand=comments", c.baseURL, key)

// NEW
issueURL := fmt.Sprintf("%s/rest/api/3/issue/%s?expand=comments", c.baseURL, key)
```

**Acceptance Criteria:**
- [ ] Uses `/rest/api/3/issue/{key}`
- [ ] `expand=comments` works
- [ ] Tests updated

---

## Workstream 8: Integration Testing (2-3 hours)

**ID:** WS8
**Dependencies:** WS5, WS6, WS7
**Priority:** P0 (Critical)
**Status:** ⬜ Not Started
**Blocks:** WS9

### Tasks

#### WS8.1: Manual Testing
**Duration:** 1 hour

**Test Checklist:**
- [ ] Create test config with new format
- [ ] Run application against test Jira
- [ ] Verify token validation (WS6)
- [ ] Verify issue search (WS7)
- [ ] Verify issue details (WS7)
- [ ] Verify no 410 errors
- [ ] Test various JQL queries

---

#### WS8.2: Automated Tests
**Duration:** 1 hour

**Create integration test suite:**
- [ ] Config load test
- [ ] Auth flow test
- [ ] Search flow test
- [ ] Issue detail flow test
- [ ] Error scenario tests

---

## Workstream 9: Documentation (2-3 hours)

**ID:** WS9
**Dependencies:** WS8
**Priority:** P1 (High)
**Status:** ⬜ Not Started

### Tasks

#### WS9.1: Migration Guide
**Duration:** 1 hour

**File to Create:**
- `MIGRATION-v3.md`

**Content:**
- Why migration needed
- Step-by-step instructions
- Before/after config examples
- Troubleshooting section

---

#### WS9.2: Update Documentation
**Duration:** 1 hour

**Files to Modify:**
- `README.md` - Config section
- `documentation/technical-details.md` - API table
- `documentation/product-research.md` - Endpoints

---

## Quality Gate Checklist

**Before marking ANY workstream complete:**

- [ ] All tests passing (`go test ./...`)
- [ ] Build succeeds (`go build ./cmd/gojira`)
- [ ] Code formatted (`go fmt ./...`)
- [ ] Dependencies tidy (`go mod tidy`)
- [ ] Coverage >95% for modified packages
- [ ] status.md updated

---

## Dependencies Graph

```
Phase 1 (Parallel):
WS1 ─┐
WS2 ─┤
WS3 ─┼─→ Phase 2
WS4 ─┘

Phase 2 (Parallel after WS1):
WS5 (needs WS1, WS2) ─┐
WS6 (needs WS1, WS3, WS4) ─┼─→ Phase 3
WS7 (needs WS1, WS3, WS4) ─┘

Phase 3 (Sequential):
WS8 (needs WS5, WS6, WS7) → WS9 (needs WS8)
```

---

## Notes

**CRITICAL:** Update status.md after completing each workstream!

**Parallel Execution Tips:**
- Assign WS1-WS4 to different agents
- WS5-WS7 can start once WS1 complete
- WS8-WS9 must be sequential
