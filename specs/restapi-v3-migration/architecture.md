# Jira REST API v3 Migration - System Architecture

**Created:** 2026-02-12
**Version:** 1.0
**Status:** Draft
**Last Updated:** 2026-02-12

---

## Architecture Overview

**High-Level Summary:**
This migration updates the adapter layer to use Jira REST API v3 endpoints. The Clean Architecture is preserved - only the outermost adapter layer changes, with no impact to domain or use case layers.

**Architectural Style:** Clean Architecture (maintained)

**Key Principles:**
- Dependency Inversion: Domain remains isolated
- Single Responsibility: Each adapter handles one API concern
- Open/Closed: Adapters updated, domain/use cases closed

**Impact Diagram:**
```
┌─────────────────────────────────────────────────────┐
│          Infrastructure Layer (TUI)                  │
│                  NO CHANGES                          │
└────────────────────┬────────────────────────────────┘
                     │ unchanged
┌────────────────────▼────────────────────────────────┐
│               Adapter Layer                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  Config  │  │ JiraAPI  │  │  Auth    │          │
│  │ MODIFIED │  │ MODIFIED │  │ MODIFIED │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└────────────────────┬────────────────────────────────┘
                     │ unchanged interface
┌────────────────────▼────────────────────────────────┐
│              Use Case Layer                          │
│                  NO CHANGES                          │
└────────────────────┬────────────────────────────────┘
                     │ unchanged
┌────────────────────▼────────────────────────────────┐
│               Domain Layer                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  Config  │  │  Issue   │  │ JiraPort │          │
│  │ MODIFIED │  │NO CHANGE │  │NO CHANGE │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
```

---

## Modified Components

### Component 1: Config Loader Adapter

**Responsibility:**
- Load and validate YAML configuration
- Enforce unified `atlassian` section structure

**Location:** `internal/adapter/config/config.go`

**Changes:**
- Update validation to require `atlassian.url` and `atlassian.email`
- Remove `jira.url` validation
- Add helpful error messages for migration

**Dependencies:**
- Domain Config model (updated)

---

### Component 2: Jira Client Adapter

**Responsibility:**
- Execute HTTP requests to Jira REST API
- Parse responses into domain models

**Location:** `internal/adapter/jira/client.go`

**Changes:**
- Update search endpoint: `/rest/api/2/search` → `/rest/api/3/search/jql`
- Update issue endpoint: `/rest/api/2/issue/{key}` → `/rest/api/3/issue/{key}`
- Add `nextPageToken` handling in response parsing
- Remove `startAt` offset logic

**Dependencies:**
- Domain JiraPort interface (unchanged)
- HTTP client (unchanged)

---

### Component 3: Auth Adapter

**Responsibility:**
- Validate Atlassian API tokens
- Verify user credentials

**Location:** `internal/adapter/auth/atlassian.go`

**Changes:**
- Update myself endpoint: `/rest/api/2/myself` → `/rest/api/3/myself`
- No response parsing changes (format compatible)

**Dependencies:**
- Domain AuthPort interface (unchanged)
- HTTP client (unchanged)

---

## Architectural Decisions

### ADR-001: Single-Page Pagination Initially

**Date:** 2026-02-12
**Status:** Accepted

**Context:**
The v3 API uses cursor-based pagination with `nextPageToken`. We must decide: fetch one page or all pages?

**Decision:**
Implement single-page fetching initially (first 100 results), add TODO for multi-page.

**Rationale:**
- Matches current application behavior (no regression)
- Simpler implementation (lower risk)
- Most queries return <100 results
- Can add multi-page in future sprint

**Consequences:**
**Positive:**
- Faster migration (less code to write/test)
- Lower risk of bugs
- No behavior change for users

**Negative:**
- Queries with >100 results truncated
- TODO technical debt

---

### ADR-002: Consolidate Configuration Sections

**Date:** 2026-02-12
**Status:** Accepted

**Context:**
Current config splits Jira URL and Atlassian email across two sections.

**Decision:**
Merge into single `atlassian` section with both `url` and `email`.

**Rationale:**
- URL and email both identify Atlassian account
- Reduces user confusion
- Clearer that Jira is part of Atlassian platform
- Simpler validation logic

**Consequences:**
**Positive:**
- Clearer configuration structure
- Easier for users to understand
- Simpler code (one section to validate)

**Negative:**
- Breaking change (users must migrate config)
- Requires migration guide

---

## References

- [spec.md](spec.md) - Feature specification
- [data-dictionary.md](data-dictionary.md) - Data structures
- [plan.md](plan.md) - Implementation plan
- [Jira Cloud REST API v3](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/)
