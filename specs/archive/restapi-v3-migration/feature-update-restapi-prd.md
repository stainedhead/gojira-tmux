# Feature: Jira REST API v3 Migration & Configuration Consolidation

**Document Version:** 1.0
**Date:** 2026-02-12
**Status:** Draft
**Owner:** Development Team

---

## Executive Summary

This feature migrates the gojira-tmux application from the deprecated Jira REST API v2 endpoints to the current v3 endpoints, specifically addressing the 410 (Gone) errors encountered with `/rest/api/2/search`. Additionally, it consolidates the split `jira` and `atlassian` configuration sections into a single unified `atlassian` section for improved clarity and maintainability.

**Key Changes:**
1. Migrate from `/rest/api/2/search` to `/rest/api/3/search/jql` with new pagination model
2. Update `/rest/api/2/myself` to `/rest/api/3/myself` for token validation
3. Update `/rest/api/2/issue/{key}` to `/rest/api/3/issue/{key}` for issue retrieval
4. Consolidate `jira` and `atlassian` config sections into single `atlassian` section
5. Update pagination logic from `startAt` offset-based to `nextPageToken` cursor-based
6. Ensure backward compatibility during migration with graceful error handling

**Impact:** Breaking changes to API endpoints requiring code updates in adapter layer. Configuration file format changes requiring user migration of config files.

**Parallel Execution Strategy:** This PRD is structured to enable agent swarms to work in parallel on independent workstreams, with clear dependency markers.

---

## Problem Statement

### Current Issues

1. **API Deprecation Error (HTTP 410)**: The application fails with error:
   ```
   Jira API error (status 410): {"errorMessages":["The requested API has been removed.
   Please migrate to the /rest/api/3/search/jql API. A full migration guideline is
   available at https://developer.atlassian.com/changelog/#CHANGE-2046"],"errors":{}}
   ```
   - Root cause: `/rest/api/2/search` endpoint removed by Atlassian as of August 2025
   - Affected file: `internal/adapter/jira/client.go:40`
   - Impact: Application cannot search for issues, core functionality broken

2. **Outdated API Endpoints**: Multiple deprecated v2 endpoints in use:
   - `/rest/api/2/search` - JQL queries (client.go:40) - **REMOVED, CAUSES 410**
   - `/rest/api/2/myself` - Token validation (auth/atlassian.go:34) - **DEPRECATED**
   - `/rest/api/2/issue/{key}` - Issue details (client.go:75) - **DEPRECATED**

3. **Configuration Confusion**: Split configuration structure:
   ```yaml
   jira:
     url: "https://your-company.atlassian.net"
   atlassian:
     email: "your-email@company.com"
   ```
   - Unclear why Jira and Atlassian are separate (both refer to the same service)
   - URL and email logically belong together as Atlassian credentials
   - Confusing for users setting up the application

4. **Pagination Model Incompatibility**: Current implementation uses v2 pagination:
   - Uses implicit offset-based pagination with `maxResults`
   - New v3 API requires explicit `nextPageToken` cursor-based pagination
   - Missing implementation for handling continuation tokens

### Dependencies Discovered

**Research Sources:**
- [Atlassian REST API Search Endpoints Deprecation](https://docs.adaptavist.com/sr4jc/latest/release-notes/breaking-changes/atlassian-rest-api-search-endpoints-deprecation)
- [Jira API Migration Discussion](https://community.atlassian.com/forums/Jira-questions/Jira-API-Migration-to-rest-api-3-search-jql/qaq-p/3111339)
- [Run JQL Search Query Using Jira Cloud REST API](https://confluence.atlassian.com/jirakb/run-jql-search-query-using-jira-cloud-rest-api-1289424308.html)
- [GitHub Issue on API Deprecation](https://github.com/pycontribs/jira/issues/2369)

---

## Solution Overview

### API Migration Strategy

**Endpoint Migrations:**

| Current (v2)                              | New (v3)                          | Location                           |
|-------------------------------------------|-----------------------------------|------------------------------------|
| `GET /rest/api/2/search`                  | `GET /rest/api/3/search/jql`      | `internal/adapter/jira/client.go:40` |
| `GET /rest/api/2/myself`                  | `GET /rest/api/3/myself`          | `internal/adapter/auth/atlassian.go:34` |
| `GET /rest/api/2/issue/{key}?expand=...`  | `GET /rest/api/3/issue/{key}?expand=...` | `internal/adapter/jira/client.go:75` |

**Key API Changes:**

1. **Search Endpoint (`/rest/api/3/search/jql`)**:
   - **Request Changes**:
     - JQL passed as `jql` parameter (same as v2)
     - `maxResults` parameter retained (limits page size)
     - `fields` parameter retained (specifies returned fields)
     - **NEW**: `nextPageToken` parameter for pagination continuation
     - Supports both GET and POST methods (POST recommended for long JQL)

   - **Response Changes**:
     - Same basic structure with `total`, `issues` array
     - **NEW**: `nextPageToken` field in response (present when more results available)
     - **REMOVED**: `startAt` field (no longer used)
     - **REMOVED**: `maxResults` echo (no longer returned)

   - **Pagination Model**:
     ```
     OLD (v2): startAt + maxResults offset pagination
     Request 1: ?jql=...&maxResults=100&startAt=0
     Request 2: ?jql=...&maxResults=100&startAt=100

     NEW (v3): cursor-based with nextPageToken
     Request 1: ?jql=...&maxResults=100
     Response 1: {..., "nextPageToken": "abc123"}
     Request 2: ?jql=...&maxResults=100&nextPageToken=abc123
     Response 2: {..., "nextPageToken": "def456"} or no token if done
     ```

2. **Myself Endpoint (`/rest/api/3/myself`)**:
   - Minimal changes, primarily version number update
   - Response format remains compatible
   - Authentication method unchanged (Basic Auth)

3. **Issue Endpoint (`/rest/api/3/issue/{key}`)**:
   - `expand` parameter works the same
   - Response format compatible with v2
   - Field structure unchanged

### Configuration Consolidation

**Current Structure:**
```yaml
jira:
  url: "https://your-company.atlassian.net"

atlassian:
  email: "your-email@company.com"
```

**New Structure:**
```yaml
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"

# Projects and team remain unchanged
projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
```

**Migration Strategy:**
1. Update domain models to use single `AtlassianConfig` with both fields
2. Update config loader validation to require `atlassian.url` and `atlassian.email`
3. Remove `JiraConfig` struct (consolidate into `AtlassianConfig`)
4. Update all references from `cfg.Jira.URL` to `cfg.Atlassian.URL`
5. Provide clear migration guide for users updating config files

---

## Workstream Architecture (Parallel Execution)

This migration is structured into **5 independent workstreams** that can be executed in parallel by agent swarms, plus 2 sequential integration phases.

```
PHASE 1 (PARALLEL - No Dependencies):
┌─────────────────────────────────────────────────────────────────┐
│ WS1: Domain Model Updates          [INDEPENDENT]                │
│ WS2: Configuration Research         [INDEPENDENT]                │
│ WS3: API Endpoint Research          [INDEPENDENT]                │
│ WS4: Test Infrastructure Prep      [INDEPENDENT]                │
└─────────────────────────────────────────────────────────────────┘

PHASE 2 (PARALLEL - Depends on Phase 1):
┌─────────────────────────────────────────────────────────────────┐
│ WS5: Config Adapter Implementation  [DEPENDS: WS1, WS2]         │
│ WS6: Auth Adapter Implementation    [DEPENDS: WS1, WS3, WS4]    │
│ WS7: Jira Client Implementation     [DEPENDS: WS1, WS3, WS4]    │
└─────────────────────────────────────────────────────────────────┘

PHASE 3 (SEQUENTIAL - Integration):
┌─────────────────────────────────────────────────────────────────┐
│ WS8: Integration Testing            [DEPENDS: WS5, WS6, WS7]    │
│ WS9: Documentation & Migration      [DEPENDS: WS8]              │
└─────────────────────────────────────────────────────────────────┘
```

---

## Detailed Workstream Specifications

### **Workstream 1: Domain Model Updates** ⚡ PARALLEL-SAFE

**Objective**: Update domain configuration models to support unified Atlassian config.

**Dependencies**: NONE (Foundation work)

**Deliverables**:
1. Update `internal/domain/ports.go:74-98`:
   - Remove `JiraConfig` struct or make it alias for backward compat
   - Expand `AtlassianConfig` to include `URL` field
   - Add `CustomFields` to `AtlassianConfig` (move from `JiraConfig`)

2. New domain model:
   ```go
   type Config struct {
       Atlassian AtlassianConfig `yaml:"atlassian"`
       Projects  []Project       `yaml:"projects"`
       Team      []TeamMember    `yaml:"team"`
   }

   type AtlassianConfig struct {
       URL          string            `yaml:"url"`
       Email        string            `yaml:"email"`
       CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"`
   }
   ```

3. Write comprehensive unit tests for new model:
   - Test YAML unmarshaling
   - Test validation rules
   - Test backward compatibility (if applicable)

**Files Modified**:
- `internal/domain/ports.go`
- `internal/domain/ports_test.go` (new file if needed)

**Acceptance Criteria**:
- [ ] `Config` struct uses single `Atlassian` field
- [ ] `AtlassianConfig` contains `URL`, `Email`, `CustomFields`
- [ ] Unit tests achieve >95% coverage
- [ ] No compilation errors in domain package
- [ ] Changes are backward compatible with existing code (no breaking changes yet)

**Estimated Effort**: 1-2 hours

---

### **Workstream 2: Configuration Research & Validation** ⚡ PARALLEL-SAFE

**Objective**: Research configuration migration patterns and create validation strategy.

**Dependencies**: NONE (Research work)

**Deliverables**:
1. Document validation requirements:
   - URL must be HTTPS
   - URL must be valid Atlassian domain pattern
   - Email must be valid format
   - Create validation rule specification

2. Research error messages:
   - What errors should be shown for missing `atlassian.url`?
   - What errors for invalid URL format?
   - How to guide users migrating from old config?

3. Create validation test cases:
   - Valid config examples
   - Invalid config examples (missing url, invalid email, etc.)
   - Edge cases (trailing slashes, http vs https, etc.)

4. Document migration steps for users:
   - Before/after config examples
   - Step-by-step migration guide
   - Troubleshooting common issues

**Files Created**:
- `docs/config-migration-guide.md` (or similar)
- `specs/restapi-v3/validation-rules.md` (or similar)

**Acceptance Criteria**:
- [ ] Validation rules documented with examples
- [ ] Migration guide complete with before/after examples
- [ ] Test cases cover all validation scenarios
- [ ] User-facing documentation clear and actionable

**Estimated Effort**: 2-3 hours

---

### **Workstream 3: API Endpoint Research & Specification** ⚡ PARALLEL-SAFE

**Objective**: Create detailed specification for v3 API implementation.

**Dependencies**: NONE (Research work)

**Deliverables**:
1. **Search Endpoint Specification** (`/rest/api/3/search/jql`):
   - Request format (GET vs POST)
   - Query parameters (`jql`, `maxResults`, `fields`, `nextPageToken`)
   - Response format with examples
   - Pagination algorithm (nextPageToken handling)
   - Error handling (what errors can occur?)

2. **Myself Endpoint Specification** (`/rest/api/3/myself`):
   - Request format
   - Response format
   - Comparison to v2 (what changed?)
   - Error cases

3. **Issue Endpoint Specification** (`/rest/api/3/issue/{key}`):
   - Request format
   - Query parameters (`expand`)
   - Response format
   - Comparison to v2

4. **Response Type Definitions**:
   - Create Go struct definitions for all response types
   - Document JSON field mappings
   - Identify any breaking changes in response format

5. **Test Request/Response Examples**:
   - Real JQL query examples
   - Expected response JSON
   - Pagination sequences
   - Error response examples

**Files Created**:
- `specs/restapi-v3/api-specification.md`
- `specs/restapi-v3/response-types.md`
- `specs/restapi-v3/test-data-examples.json`

**Acceptance Criteria**:
- [ ] All three endpoints fully specified
- [ ] Response type Go structs defined
- [ ] Pagination algorithm documented with examples
- [ ] Test data includes success and error cases
- [ ] Differences from v2 clearly documented

**Estimated Effort**: 3-4 hours

---

### **Workstream 4: Test Infrastructure Preparation** ⚡ PARALLEL-SAFE

**Objective**: Create test infrastructure for v3 API testing.

**Dependencies**: NONE (Foundation work)

**Deliverables**:
1. **HTTP Mock Server Helpers**:
   - Create reusable HTTP test server for Jira API v3
   - Implement mock handlers for all three endpoints
   - Support pagination scenarios (multiple pages)
   - Support error scenarios (401, 403, 410, 500)

2. **Test Data Fixtures**:
   - Create realistic Jira issue JSON responses (v3 format)
   - Create search response fixtures with multiple pages
   - Create error response fixtures
   - Create myself response fixture

3. **Test Utilities**:
   - Helper to assert HTTP requests (headers, params, body)
   - Helper to build test configs
   - Helper to compare domain models

4. **Update Existing Tests** (if needed):
   - Identify tests that need updating for config changes
   - Mark tests that will break with v3 changes

**Files Created/Modified**:
- `internal/adapter/jira/testutil/server.go` (new)
- `internal/adapter/jira/testutil/fixtures.go` (new)
- `internal/adapter/auth/testutil/` (new if needed)

**Acceptance Criteria**:
- [ ] Mock server can simulate v3 search endpoint with pagination
- [ ] Mock server can simulate error responses
- [ ] Test fixtures cover all response types
- [ ] Helper functions well-documented and reusable
- [ ] All tests pass before implementation changes

**Estimated Effort**: 2-3 hours

---

### **Workstream 5: Config Adapter Implementation** 🔗 DEPENDS: WS1, WS2

**Objective**: Implement consolidated configuration loading and validation.

**Dependencies**:
- ✅ WS1 (needs new domain models)
- ✅ WS2 (needs validation rules)

**Deliverables**:
1. Update `internal/adapter/config/config.go`:
   - Update validation function to expect `atlassian.url` and `atlassian.email`
   - Remove `jira.url` validation or add deprecation warning
   - Implement URL format validation (HTTPS, valid domain)
   - Add helpful error messages for migration

2. Update `config.example.yaml`:
   - Show new unified structure
   - Add comments explaining migration from old format
   - Include example custom_fields under atlassian

3. Comprehensive unit tests `internal/adapter/config/config_test.go`:
   - Test valid unified config loads successfully
   - Test missing `atlassian.url` produces clear error
   - Test missing `atlassian.email` produces clear error
   - Test invalid URL format (http instead of https) fails
   - Test invalid email format fails
   - Test backward compatibility (if supporting old format during transition)

4. Integration with existing code:
   - Update all references to `cfg.Jira.URL` → `cfg.Atlassian.URL`
   - Ensure `cfg.Atlassian.Email` still works (already in use)

**Files Modified**:
- `internal/adapter/config/config.go`
- `internal/adapter/config/config_test.go`
- `config.example.yaml`

**Files Needing Updates** (for `cfg.Jira.URL` → `cfg.Atlassian.URL`):
- `cmd/gojira/main.go` (likely passes URL to constructors)
- Any other adapter constructors

**Acceptance Criteria**:
- [ ] Config loader validates unified `atlassian` section
- [ ] Clear error messages for missing or invalid fields
- [ ] Unit tests achieve >95% coverage
- [ ] Example config file demonstrates new format
- [ ] All existing tests updated and passing
- [ ] No references to `cfg.Jira.URL` remain

**Estimated Effort**: 2-3 hours

---

### **Workstream 6: Auth Adapter Implementation** 🔗 DEPENDS: WS1, WS3, WS4

**Objective**: Update authentication adapter to use v3 `/rest/api/3/myself` endpoint.

**Dependencies**:
- ✅ WS1 (needs config model changes)
- ✅ WS3 (needs API specification)
- ✅ WS4 (needs test infrastructure)

**Deliverables**:
1. Update `internal/adapter/auth/atlassian.go`:
   - Change URL construction from `/rest/api/2/myself` to `/rest/api/3/myself`
   - Verify response format compatibility (should be same)
   - Update error handling if response changed
   - Ensure Basic Auth header format unchanged

2. Update constructor signature (if needed):
   - If constructor takes `jiraURL`, rename parameter to `atlassianURL`
   - Update any documentation/comments

3. Comprehensive unit tests `internal/adapter/auth/atlassian_test.go`:
   - Use mock server from WS4
   - Test successful token validation with v3 endpoint
   - Test 401 Unauthorized response
   - Test 403 Forbidden response
   - Test CAPTCHA lockout (X-Seraph-LoginReason header)
   - Test network errors
   - Verify correct endpoint called (`/rest/api/3/myself`)
   - Verify correct Basic Auth header sent

**Files Modified**:
- `internal/adapter/auth/atlassian.go` (line 34)
- `internal/adapter/auth/atlassian_test.go`

**Acceptance Criteria**:
- [ ] `/rest/api/3/myself` endpoint used
- [ ] All error scenarios tested
- [ ] Unit tests achieve >95% coverage
- [ ] No regression in token validation functionality
- [ ] All tests pass

**Estimated Effort**: 1-2 hours

---

### **Workstream 7: Jira Client Implementation** 🔗 DEPENDS: WS1, WS3, WS4

**Objective**: Migrate Jira client to v3 search and issue endpoints with new pagination.

**Dependencies**:
- ✅ WS1 (needs config model changes)
- ✅ WS3 (needs API specification, especially pagination)
- ✅ WS4 (needs test infrastructure and fixtures)

**Deliverables**:
1. **Update Search Implementation** (`internal/adapter/jira/client.go`):
   - Change URL from `/rest/api/2/search` to `/rest/api/3/search/jql` (line 40)
   - Add pagination support with `nextPageToken`
   - Decide: Do we fetch all pages or just first page?
     - **Recommendation**: Fetch first page only (100 results) to match current behavior
     - Add TODO comment for future enhancement (fetch all pages)
   - Update response parsing (remove `startAt`, add `nextPageToken` handling)
   - Consider using POST method for long JQL queries (currently using GET)

2. **Update GetIssue Implementation**:
   - Change URL from `/rest/api/2/issue/{key}` to `/rest/api/3/issue/{key}` (line 75)
   - Verify `expand=comments` still works
   - Update response parsing if format changed

3. **New Response Types** (if response format changed):
   - Update `searchResponse` struct if needed
   - Update `issueResponse` struct if needed
   - Add `nextPageToken` field to `searchResponse`

4. **Comprehensive Unit Tests** (`internal/adapter/jira/client_test.go`):
   - Update existing tests to expect v3 endpoints
   - Test search with single page (no nextPageToken)
   - Test search with multiple pages (nextPageToken present, but we ignore it)
   - Test GetIssue with v3 endpoint
   - Test error responses (410 should not occur anymore)
   - Verify correct URL path in mock assertions (line 17, 96)

**Files Modified**:
- `internal/adapter/jira/client.go` (lines 40, 75)
- `internal/adapter/jira/client_test.go` (lines 17, 96, others)

**Pagination Decision**:
```go
// Option A: Single page only (simpler, matches current behavior)
func (c *Client) SearchIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
    // Build URL with /rest/api/3/search/jql
    // Set maxResults=100
    // Parse response
    // Ignore nextPageToken (TODO: implement multi-page fetching)
    // Return first 100 results
}

// Option B: Fetch all pages (more complete, but more complex)
func (c *Client) SearchIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
    // Initial request
    // Loop while nextPageToken exists:
    //   Make request with nextPageToken
    //   Append results
    // Return all results
}

RECOMMENDATION: Start with Option A for minimal change, add TODO for Option B
```

**Acceptance Criteria**:
- [ ] Search uses `/rest/api/3/search/jql` endpoint
- [ ] GetIssue uses `/rest/api/3/issue/{key}` endpoint
- [ ] Pagination handled (even if just first page)
- [ ] Response parsing works correctly
- [ ] All error scenarios tested
- [ ] Unit tests achieve >95% coverage
- [ ] No 410 errors occur
- [ ] All tests pass

**Estimated Effort**: 3-4 hours

---

### **Workstream 8: Integration Testing & Validation** 🔗 DEPENDS: WS5, WS6, WS7

**Objective**: End-to-end testing of all changes together.

**Dependencies**:
- ✅ WS5 (config changes)
- ✅ WS6 (auth changes)
- ✅ WS7 (Jira client changes)

**Deliverables**:
1. **Manual Integration Testing**:
   - Create new config file with unified `atlassian` section
   - Run application with real Jira instance
   - Verify token validation works (`/rest/api/3/myself`)
   - Verify issue search works (`/rest/api/3/search/jql`)
   - Verify issue details work (`/rest/api/3/issue/{key}`)
   - Verify no 410 errors occur
   - Test with various JQL queries
   - Test with different project configurations

2. **Automated Integration Tests** (optional but recommended):
   - Create integration test suite (if not exists)
   - Test full flow: config load → auth → search → get issue
   - Use mock server or test Jira instance

3. **Performance Validation**:
   - Compare response times with v2 (should be similar or better)
   - Verify 100-result limit still works
   - Check memory usage with large result sets

4. **Error Scenario Testing**:
   - Test with invalid token (expect 401)
   - Test with invalid URL (expect connection error)
   - Test with invalid JQL (expect Jira error response)
   - Test with network timeout

**Test Matrix**:
```
✓ Config: Unified atlassian section loads correctly
✓ Auth: Token validation succeeds
✓ Auth: Invalid token fails gracefully
✓ Search: Returns results with v3 endpoint
✓ Search: Handles empty results
✓ Search: Handles pagination (first page)
✓ Issue: Fetches issue details
✓ Issue: Fetches comments
✓ Error: No 410 errors
✓ Error: Proper error messages for API failures
```

**Acceptance Criteria**:
- [ ] All integration tests pass
- [ ] No 410 errors in any scenario
- [ ] Application works end-to-end with real Jira
- [ ] Performance acceptable (no regressions)
- [ ] Error handling works correctly

**Estimated Effort**: 2-3 hours

---

### **Workstream 9: Documentation & User Migration Guide** 🔗 DEPENDS: WS8

**Objective**: Complete user-facing documentation and migration instructions.

**Dependencies**:
- ✅ WS8 (integration testing confirms everything works)

**Deliverables**:
1. **Update README.md**:
   - Update configuration section with new format
   - Remove references to old `jira` section
   - Add migration note for existing users

2. **Create Migration Guide** (`MIGRATION-v3.md` or similar):
   - Explain why migration is needed (API deprecation)
   - Step-by-step config update instructions
   - Before/after config examples
   - Troubleshooting section
   - FAQ

3. **Update Technical Documentation**:
   - Update `documentation/technical-details.md`:
     - Change API endpoint references from v2 to v3
     - Update pagination documentation
     - Update API table (lines 131-135)
   - Update `documentation/product-research.md`:
     - Update API endpoint references (lines 40-42)

4. **Update Changelog/Release Notes**:
   - Document breaking changes
   - Document config migration requirement
   - Document API version upgrade

5. **Update Example Files**:
   - `config.example.yaml` (already done in WS5)
   - Any other example configurations

**Files Modified**:
- `README.md`
- `MIGRATION-v3.md` (new)
- `documentation/technical-details.md`
- `documentation/product-research.md`
- `CHANGELOG.md` or release notes

**Acceptance Criteria**:
- [ ] Migration guide is clear and actionable
- [ ] All technical docs reflect v3 API
- [ ] Example configs use new format
- [ ] User can migrate config in <5 minutes with guide
- [ ] Troubleshooting covers common issues

**Estimated Effort**: 2-3 hours

---

## Testing Strategy

### Unit Testing Requirements

**Coverage Targets**:
- Domain models: >95%
- Adapters: >95%
- Use cases: >95%
- TUI: Best effort (BubbleTea is hard to test)

**Critical Test Cases**:
1. **Config Validation**:
   - Valid unified config loads
   - Missing URL/email rejected
   - Invalid URL format rejected
   - All validation rules enforced

2. **API Endpoint Correctness**:
   - Verify exact URLs called (v3 paths)
   - Verify query parameters correct
   - Verify headers correct (Basic Auth)

3. **Pagination Handling**:
   - Single page response parsed
   - Multi-page token captured (even if not used)
   - No pagination token handled

4. **Response Parsing**:
   - v3 response JSON parsed correctly
   - All fields mapped to domain models
   - Null/missing fields handled

5. **Error Handling**:
   - 401/403 handled gracefully
   - Network errors handled
   - Invalid JSON handled
   - No 410 errors occur

### Integration Testing Requirements

**Test Environment**:
- Use real Jira Cloud test instance OR
- Use comprehensive mock server

**Test Scenarios**:
1. End-to-end auth flow with v3 API
2. Search with various filters
3. Pagination (first page)
4. Issue detail retrieval
5. Comment retrieval
6. Error scenarios

**Regression Testing**:
- All existing functionality still works
- No performance degradation
- Error messages still clear

---

## Migration Path for Users

### User Impact

**Breaking Changes**:
- Configuration file format changes
- Users must update `config.yaml`

**Non-Breaking Changes**:
- API version change is transparent (same functionality)
- Token validation still works the same way

### Migration Steps

1. **Backup Current Config**:
   ```bash
   cp config.yaml config.yaml.backup
   ```

2. **Update Config Structure**:
   ```yaml
   # OLD FORMAT (remove this):
   # jira:
   #   url: "https://your-company.atlassian.net"
   # atlassian:
   #   email: "your-email@company.com"

   # NEW FORMAT (use this):
   atlassian:
     url: "https://your-company.atlassian.net"
     email: "your-email@company.com"

   # projects and team sections unchanged
   ```

3. **Validate Config**:
   ```bash
   # Run application, it will validate config on startup
   ./gojira
   ```

4. **Verify Functionality**:
   - Token validation works
   - Issue search works
   - No 410 errors

### Rollback Plan

If issues occur:
1. Restore backup config: `cp config.yaml.backup config.yaml`
2. Revert to previous version of application
3. Report issue to development team

---

## Risk Assessment

### High Risks

1. **API Compatibility Changes**:
   - **Risk**: v3 response format different from v2, breaking parsing
   - **Mitigation**: Thorough testing with real API responses
   - **Likelihood**: Medium
   - **Impact**: High

2. **Pagination Breaking Change**:
   - **Risk**: New pagination model breaks existing assumptions
   - **Mitigation**: Start with single-page implementation (matches current behavior)
   - **Likelihood**: Low (if we only fetch first page)
   - **Impact**: Medium

3. **User Migration Confusion**:
   - **Risk**: Users don't understand how to update config
   - **Mitigation**: Clear migration guide with examples
   - **Likelihood**: Medium
   - **Impact**: Medium

### Medium Risks

1. **Performance Regression**:
   - **Risk**: v3 API slower than v2
   - **Mitigation**: Performance testing during integration
   - **Likelihood**: Low
   - **Impact**: Medium

2. **Missing Test Coverage**:
   - **Risk**: Edge cases not covered by tests
   - **Mitigation**: Comprehensive test plan with checklist
   - **Likelihood**: Medium
   - **Impact**: Medium

### Low Risks

1. **Endpoint URL Typos**:
   - **Risk**: Typo in endpoint path (/rest/api/3/serch instead of /search)
   - **Mitigation**: Unit tests verify exact URLs
   - **Likelihood**: Low
   - **Impact**: High (but easy to fix)

---

## Success Metrics

### Technical Metrics

- [ ] Zero HTTP 410 errors
- [ ] All unit tests pass (>95% coverage)
- [ ] All integration tests pass
- [ ] No performance regression (response times within 10% of v2)
- [ ] Zero production errors related to API calls

### User Metrics

- [ ] Users can migrate config in <5 minutes
- [ ] Zero support tickets related to config migration
- [ ] Application startup succeeds on first try after migration

### Quality Metrics

- [ ] Code review approved by 2+ developers
- [ ] All acceptance criteria met across all workstreams
- [ ] Documentation complete and clear
- [ ] No regression in existing functionality

---

## Timeline Estimate

### Sequential Execution (Single Developer)
- **Phase 1 (Parallel Work)**: 8-12 hours if done sequentially
  - WS1: 1-2 hours
  - WS2: 2-3 hours
  - WS3: 3-4 hours
  - WS4: 2-3 hours

- **Phase 2 (Implementation)**: 6-9 hours
  - WS5: 2-3 hours (after WS1, WS2)
  - WS6: 1-2 hours (after WS1, WS3, WS4)
  - WS7: 3-4 hours (after WS1, WS3, WS4)

- **Phase 3 (Integration)**: 4-6 hours
  - WS8: 2-3 hours
  - WS9: 2-3 hours

**Total Sequential**: 18-27 hours (2.5-3.5 days)

### Parallel Execution (Agent Swarm - 4 Agents)
- **Phase 1**: 3-4 hours (all 4 workstreams in parallel)
- **Phase 2**: 3-4 hours (3 workstreams in parallel after dependencies met)
- **Phase 3**: 4-6 hours (sequential integration)

**Total Parallel**: 10-14 hours (1.5-2 days)

**Efficiency Gain**: ~50% time reduction with parallel execution

---

## Appendix

### A. API Endpoint Quick Reference

```
Auth:
  OLD: GET /rest/api/2/myself
  NEW: GET /rest/api/3/myself

Search:
  OLD: GET /rest/api/2/search?jql=...&maxResults=100&startAt=0
  NEW: GET /rest/api/3/search/jql?jql=...&maxResults=100
       GET /rest/api/3/search/jql?jql=...&maxResults=100&nextPageToken=abc123

Issue:
  OLD: GET /rest/api/2/issue/PROJ-1?expand=comments
  NEW: GET /rest/api/3/issue/PROJ-1?expand=comments
```

### B. Response Format Comparison

**Search Response v2:**
```json
{
  "total": 250,
  "startAt": 0,
  "maxResults": 100,
  "issues": [...]
}
```

**Search Response v3:**
```json
{
  "total": 250,
  "issues": [...],
  "nextPageToken": "abc123"
}
```

### C. Configuration Migration Checklist

For Users:
- [ ] Backup existing config.yaml
- [ ] Create new `atlassian` section
- [ ] Move `url` from `jira.url` to `atlassian.url`
- [ ] Verify `email` already under `atlassian.email`
- [ ] Remove old `jira` section
- [ ] Test application startup
- [ ] Verify issue search works

For Developers:
- [ ] Update domain models
- [ ] Update config loader
- [ ] Update all references to `cfg.Jira.URL`
- [ ] Update example config
- [ ] Update tests
- [ ] Update documentation

### D. Troubleshooting Guide

**Problem**: Application shows "jira.url is required"
- **Cause**: Old config format
- **Solution**: Rename `jira` to `atlassian`, ensure `url` field present

**Problem**: HTTP 410 errors persist
- **Cause**: Not using updated application version
- **Solution**: Rebuild application (`go build ./cmd/gojira`)

**Problem**: Config validation fails
- **Cause**: URL not HTTPS or invalid format
- **Solution**: Verify URL starts with `https://` and ends with `.atlassian.net`

**Problem**: No search results returned
- **Cause**: Pagination or JQL issues
- **Solution**: Check logs for API errors, verify JQL syntax

---

## Approval & Sign-off

- [ ] Technical Lead Review
- [ ] Architecture Review
- [ ] Security Review (if applicable)
- [ ] Product Owner Approval
- [ ] Ready for Implementation

**Next Steps**: Break down into implementation tasks and assign to agent swarm workstreams.
