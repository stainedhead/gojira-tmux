# Jira REST API v3 Migration - Specification

**Created:** 2026-02-12
**Version:** 1.0
**Status:** Draft
**Source PRD:** `feature-update-restapi-prd.md`

---

## Executive Summary

This feature migrates gojira-tmux from deprecated Jira REST API v2 endpoints to current v3 endpoints, addressing critical HTTP 410 errors. Additionally, it consolidates the split `jira` and `atlassian` configuration sections into a unified `atlassian` section for improved clarity.

**Key Deliverables:**
- Migrate three API endpoints to v3 with new pagination model
- Consolidate configuration structure
- Update all adapter layer code
- Provide migration guide for users
- Maintain test coverage >95%

**Timeline:** 10-14 hours (parallel execution) | 18-27 hours (sequential)

---

## Problem Statement

### Current State

The application is currently broken due to deprecated API endpoints:
- **Critical**: `/rest/api/2/search` returns HTTP 410 (Gone) - endpoint removed by Atlassian
- `/rest/api/2/myself` deprecated - still works but will be removed
- `/rest/api/2/issue/{key}` deprecated - still works but will be removed
- Configuration split between `jira` and `atlassian` sections causes user confusion

### Pain Points

- **Application Failure**: Users cannot search for issues, core functionality broken
- **Error Message**: `Jira API error (status 410): The requested API has been removed. Please migrate to /rest/api/3/search/jql`
- **Configuration Confusion**: URL and email separated across two sections
- **Future Risk**: Other deprecated endpoints will also be removed

### Desired State

After implementation:
- All API calls use v3 endpoints (no 410 errors)
- Configuration uses single unified `atlassian` section
- Pagination uses cursor-based `nextPageToken` model
- Users can migrate config in <5 minutes with clear guide
- All tests passing with >95% coverage

---

## Goals and Non-Goals

### Goals

- ✅ Migrate all three API endpoints to v3
- ✅ Implement cursor-based pagination (at least first-page support)
- ✅ Consolidate configuration structure
- ✅ Maintain backward compatibility during transition
- ✅ Provide clear migration guide for users
- ✅ Maintain or improve test coverage

### Non-Goals

- ❌ Multi-page result fetching (future enhancement)
- ❌ Support for API v2 endpoints (fully deprecate)
- ❌ Automatic config file migration (users must manually update)
- ❌ Breaking changes to domain layer (only adapter changes)

---

## User Requirements

### Functional Requirements

#### FR-001: API v3 Search Endpoint Migration
**Priority:** P0 (Critical)

**Description:**
Replace `/rest/api/2/search` with `/rest/api/3/search/jql` in the Jira client adapter.

**Acceptance Criteria:**
- [ ] Search endpoint uses `/rest/api/3/search/jql`
- [ ] JQL parameter format unchanged (backward compatible)
- [ ] Returns first page of results (100 issues max)
- [ ] Response parsing handles `nextPageToken` field
- [ ] No HTTP 410 errors occur
- [ ] All existing search functionality works

**Example:**
```go
// OLD
searchURL := fmt.Sprintf("%s/rest/api/2/search", c.baseURL)

// NEW
searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)
```

---

#### FR-002: API v3 Auth Endpoint Migration
**Priority:** P0 (Critical)

**Description:**
Update token validation to use `/rest/api/3/myself` endpoint.

**Acceptance Criteria:**
- [ ] Auth validation uses `/rest/api/3/myself`
- [ ] Response format compatible (no parsing changes needed)
- [ ] All error scenarios handled (401, 403, CAPTCHA)
- [ ] Basic Auth header format unchanged

**Example:**
```go
// OLD
url := a.jiraURL + "/rest/api/2/myself"

// NEW
url := a.jiraURL + "/rest/api/3/myself"
```

---

#### FR-003: API v3 Issue Endpoint Migration
**Priority:** P0 (Critical)

**Description:**
Update issue retrieval to use `/rest/api/3/issue/{key}` endpoint.

**Acceptance Criteria:**
- [ ] Issue retrieval uses `/rest/api/3/issue/{key}`
- [ ] `expand=comments` parameter still works
- [ ] Response parsing unchanged (format compatible)
- [ ] All issue details retrieved correctly

---

#### FR-004: Configuration Consolidation
**Priority:** P0 (Critical)

**Description:**
Merge `jira` and `atlassian` config sections into single `atlassian` section with both `url` and `email` fields.

**Acceptance Criteria:**
- [ ] Domain model updated (`AtlassianConfig` includes URL)
- [ ] Config loader validates `atlassian.url` and `atlassian.email`
- [ ] Example config file shows new format
- [ ] All references to `cfg.Jira.URL` updated to `cfg.Atlassian.URL`
- [ ] Clear error messages for missing fields

**Example:**
```yaml
# NEW format
atlassian:
  url: "https://company.atlassian.net"
  email: "user@company.com"

# OLD format (deprecated)
# jira:
#   url: "https://company.atlassian.net"
# atlassian:
#   email: "user@company.com"
```

---

#### FR-005: Cursor-Based Pagination Support
**Priority:** P1 (High)

**Description:**
Update response parsing to handle `nextPageToken` field (first-page implementation).

**Acceptance Criteria:**
- [ ] `searchResponse` struct includes `NextPageToken` field
- [ ] Response parsing extracts token correctly
- [ ] First 100 results returned (matches current behavior)
- [ ] TODO comment for future multi-page implementation
- [ ] No errors if token is missing (single page result)

---

### Non-Functional Requirements

#### NFR-001: Test Coverage
**Category:** Quality

**Description:**
Maintain high test coverage across all layers.

**Metrics:**
- Domain layer: >95% coverage
- Adapter layer: >95% coverage
- Use case layer: >95% coverage

**Acceptance Criteria:**
- [ ] All existing tests updated
- [ ] New tests for v3 endpoints
- [ ] Mock server for v3 API testing
- [ ] Integration tests with real/mock Jira

---

#### NFR-002: Performance
**Category:** Performance

**Description:**
API migration must not degrade performance.

**Metrics:**
- Search response time: <1s (same as v2)
- Auth validation: <500ms (same as v2)
- Memory usage: No increase

**Acceptance Criteria:**
- [ ] Response times within 10% of v2
- [ ] No memory leaks
- [ ] Pagination doesn't add latency

---

#### NFR-003: User Experience
**Category:** Usability

**Description:**
Config migration must be simple and clear.

**Metrics:**
- Migration time: <5 minutes
- Support tickets: 0 related to migration
- Success rate: 100% on first try

**Acceptance Criteria:**
- [ ] Migration guide with before/after examples
- [ ] Troubleshooting section
- [ ] Clear error messages
- [ ] Example config updated

---

## System Architecture

### Affected Layers

- [x] Domain Layer - Config model updates
- [ ] Use Case Layer - No changes (transparent)
- [x] Infrastructure Layer - No direct changes
- [x] Adapter Layer - All API client changes

### New Components

No new components - only modifications to existing adapters.

### Modified Components

- **`internal/domain/ports.go`**: Update `AtlassianConfig` struct
- **`internal/adapter/config/config.go`**: Update validation logic
- **`internal/adapter/jira/client.go`**: Update API endpoints and pagination
- **`internal/adapter/auth/atlassian.go`**: Update auth endpoint
- **`cmd/gojira/main.go`**: Update config references

---

## Scope of Changes

### Files to Create

- `docs/MIGRATION-v3.md` - User migration guide
- `internal/adapter/jira/testutil/server.go` - Mock server for testing
- `internal/adapter/jira/testutil/fixtures.go` - Test data fixtures

### Files to Modify

- `internal/domain/ports.go` - Config struct updates
- `internal/adapter/config/config.go` - Validation updates
- `internal/adapter/config/config_test.go` - Updated tests
- `internal/adapter/jira/client.go` - API endpoint migrations
- `internal/adapter/jira/client_test.go` - Updated tests
- `internal/adapter/auth/atlassian.go` - Auth endpoint update
- `internal/adapter/auth/atlassian_test.go` - Updated tests
- `cmd/gojira/main.go` - Config reference updates
- `config.example.yaml` - New format
- `README.md` - Config section updates
- `documentation/technical-details.md` - API table updates
- `documentation/product-research.md` - Endpoint updates

### Dependencies

**External:**
- None (using standard library HTTP client)

**Internal:**
- Domain models must be updated first (WS1)
- Test infrastructure needed before implementation (WS4)

---

## Breaking Changes

### API Changes

**No public API changes** - this is an internal implementation detail.

**Migration path:** Transparent to users (same functionality, different endpoints).

---

### Configuration Changes

**Breaking Change:** Config file format

**Old Format:**
```yaml
jira:
  url: "https://company.atlassian.net"
atlassian:
  email: "user@company.com"
```

**New Format:**
```yaml
atlassian:
  url: "https://company.atlassian.net"
  email: "user@company.com"
```

**Migration Path:**
1. Backup `config.yaml`
2. Move `url` from `jira` section to `atlassian` section
3. Remove empty `jira` section
4. Restart application

**Timeline:** Must be done before running updated version

---

### Database Schema Changes

None (no database used).

---

## Success Criteria

### Acceptance Criteria

- [ ] Zero HTTP 410 errors
- [ ] All three endpoints use v3 API
- [ ] Configuration consolidated
- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] Documentation updated
- [ ] Migration guide complete

### Quality Gates

- [ ] All tests pass (`go test ./...`)
- [ ] Code coverage >95%
- [ ] No linter warnings (`golangci-lint run`)
- [ ] Builds successfully (`go build ./cmd/gojira`)
- [ ] Manual testing with real Jira instance

### User Validation

- [ ] Config migration takes <5 minutes
- [ ] Application starts without errors
- [ ] Issue search returns results
- [ ] No 410 errors in logs
- [ ] Performance acceptable

---

## Risks and Mitigation

### Risk 1: API Response Format Changes
**Likelihood:** Medium
**Impact:** High

**Mitigation:**
- Thorough testing with real Jira API responses
- Mock server with realistic v3 responses
- Unit tests verify response parsing
- Integration tests catch format mismatches

---

### Risk 2: User Config Migration Errors
**Likelihood:** Medium
**Impact:** Medium

**Mitigation:**
- Clear migration guide with examples
- Helpful error messages in validation
- Troubleshooting section in docs
- Example config file updated

---

### Risk 3: Pagination Breaking Changes
**Likelihood:** Low
**Impact:** Medium

**Mitigation:**
- Start with single-page implementation (matches current behavior)
- Add TODO for multi-page support
- Test with queries returning >100 results
- Document pagination limitations

---

## Timeline and Milestones

### Phase 1: Foundation & Research (3-4 hours)
**Deliverables:**
- Domain model updates (WS1)
- Configuration research (WS2)
- API specification (WS3)
- Test infrastructure (WS4)

### Phase 2: Implementation (3-4 hours)
**Deliverables:**
- Config adapter implementation (WS5)
- Auth adapter implementation (WS6)
- Jira client implementation (WS7)

### Phase 3: Integration & Documentation (4-6 hours)
**Deliverables:**
- Integration testing (WS8)
- Documentation & migration guide (WS9)

**Total Estimated Duration:** 10-14 hours (parallel) | 18-27 hours (sequential)

---

## References

- **Source PRD:** `feature-update-restapi-prd.md`
- **Related Specs:** None (standalone migration)
- **External Documentation:**
  - [Atlassian REST API Search Endpoints Deprecation](https://docs.adaptavist.com/sr4jc/latest/release-notes/breaking-changes/atlassian-rest-api-search-endpoints-deprecation)
  - [Jira API Migration Discussion](https://community.atlassian.com/forums/Jira-questions/Jira-API-Migration-to-rest-api-3-search-jql/qaq-p/3111339)
  - [Run JQL Search Query Using Jira Cloud REST API](https://confluence.atlassian.com/jirakb/run-jql-search-query-using-jira-cloud-rest-api-1289424308.html)
  - [GitHub Issue on API Deprecation](https://github.com/pycontribs/jira/issues/2369)
