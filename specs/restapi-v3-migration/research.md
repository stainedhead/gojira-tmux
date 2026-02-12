# Jira REST API v3 Migration - Research

**Created:** 2026-02-12
**Source:** `feature-update-restapi-prd.md`
**Status:** Complete

---

## Overview

Research for migrating gojira-tmux from deprecated Jira REST API v2 endpoints to current v3 endpoints. Primary focus: understanding the new `/rest/api/3/search/jql` endpoint and cursor-based pagination model.

**Research Questions:**
1. What are the exact differences between v2 and v3 search endpoints?
2. How does cursor-based pagination work with `nextPageToken`?
3. Are there any response format changes that would break parsing?
4. What error codes should we handle differently in v3?
5. How should we structure the configuration migration?

**Key Findings:**
- Search endpoint path changed from `/rest/api/2/search` to `/rest/api/3/search/jql`
- Pagination changed from offset-based (`startAt`) to cursor-based (`nextPageToken`)
- **CRITICAL**: Description and comment body fields in v3 return ADF (Atlassian Document Format) JSON objects, NOT plain strings
- The `total` field may not be returned by default in v3 search responses
- New `fields` parameter behavior: only issue IDs returned by default; must explicitly request fields
- Unbounded JQL queries (empty string) return `400 Bad Request` in v3
- Maximum 20 comments per response in search results

---

## API Endpoint Specifications

### 1. Search Endpoint: `/rest/api/3/search/jql`

#### Endpoint Summary

| Attribute | Value |
|---|---|
| **v2 Path** | `GET /rest/api/2/search` |
| **v3 Path** | `GET /rest/api/3/search/jql` or `POST /rest/api/3/search/jql` |
| **Auth** | Basic Auth (`base64(email:api_token)`) |
| **Content-Type** | `application/json` |
| **Breaking** | Yes - path changed, pagination model changed, default fields changed |

#### GET Request Parameters

| Parameter | Type | Required | Default | Description |
|---|---|---|---|---|
| `jql` | string | Yes | - | JQL query string (must be URL-encoded). **Must be bounded** (non-empty); empty JQL returns 400 in v3. |
| `maxResults` | integer | No | 50 | Maximum results per page. Max value: 5000. |
| `fields` | string | No | IDs only | Comma-separated field names. Use `*all` for all fields, `*navigable` for navigable fields. **v3 only returns issue IDs by default** — you must specify fields explicitly. |
| `expand` | string | No | - | Expansion options (e.g., `renderedFields`, `names`, `schema`). |
| `nextPageToken` | string | No | - | Cursor token for pagination. Omit on first request. Pass value from previous response for next page. |
| `fieldsByKeys` | boolean | No | false | Return fields by key instead of by ID. |
| `properties` | string | No | - | Comma-separated entity property keys to include. |
| `reconcileIssues` | array | No | - | Issue IDs to ensure appear in results (recently created/updated). |

**Removed v2 Parameters:**
- `startAt` — **REMOVED**. No longer supported. Replaced by `nextPageToken`.

#### POST Request Body Schema

```json
{
  "jql": "project = PROJ ORDER BY updated DESC",
  "maxResults": 100,
  "fields": ["key", "summary", "status", "priority", "assignee", "reporter", "duedate", "created", "updated", "labels"],
  "expand": "renderedFields",
  "nextPageToken": "eyJvcmRlckJ5...",
  "fieldsByKeys": false,
  "properties": [],
  "reconcileIssues": []
}
```

POST is recommended for long JQL queries to avoid URL length limits. All parameters are the same as GET but passed as JSON body.

#### Response Schema

```json
{
  "issues": [
    {
      "id": "10001",
      "key": "PROJ-1",
      "self": "https://your-domain.atlassian.net/rest/api/3/issue/10001",
      "fields": {
        "summary": "Issue title",
        "description": {}, // ADF object or null — SEE CRITICAL NOTE BELOW
        "status": { "name": "In Progress", "id": "3" },
        "priority": { "name": "High", "id": "2" },
        "assignee": {
          "accountId": "5b10a2844c20165700ede21g",
          "displayName": "John Doe",
          "emailAddress": "john@company.com",
          "active": true
        },
        "reporter": {
          "accountId": "5b10a2844c20165700ede21g",
          "displayName": "Jane Smith",
          "emailAddress": "jane@company.com",
          "active": true
        },
        "duedate": "2026-03-15",
        "created": "2026-01-15T10:00:00.000+0000",
        "updated": "2026-02-10T14:30:00.000+0000",
        "labels": ["bug", "critical"]
      }
    }
  ],
  "nextPageToken": "eyJvcmRlckJ5IjoiY3JlYXRlZCBERVNDIiwicGFnZVNpemUiOjEwMCwibGFzdFZhbHVlcyI6WyIyMDI2LTAxLTE1VDEwOjAwOjAwLjAwMCswMDAwIl19",
  "isLast": false,
  "total": 250
}
```

**Response Fields:**

| Field | Type | Always Present | Description |
|---|---|---|---|
| `issues` | array | Yes | Array of issue objects. Empty array if no results. |
| `nextPageToken` | string | No | Present only when more pages exist. Absent on last page. |
| `isLast` | boolean | No | May indicate last page. **Unreliable** — see known issues. Prefer checking absence of `nextPageToken`. |
| `total` | integer | No | Total matching issues. **May not be returned** in v3 — see known issues. Do not depend on this field. |

#### v2 vs v3 Response Comparison

| Field | v2 | v3 |
|---|---|---|
| `total` | Always present | May not be present |
| `startAt` | Present (integer) | **REMOVED** |
| `maxResults` | Echoed back (integer) | **REMOVED** |
| `issues` | Array | Array (same) |
| `nextPageToken` | N/A | Present when more pages exist |
| `isLast` | N/A | May be present (unreliable) |
| `issues[].fields.description` | String (wiki markup) | **ADF JSON object** — see critical note |
| `issues[].fields.comment.comments[].body` | String (wiki markup) | **ADF JSON object** — see critical note |
| Default fields | All navigable fields | **IDs only** — must specify `fields` param |

#### Known Issues & Edge Cases

1. **Pagination bugs**: Community reports of `nextPageToken` not advancing properly, `isLast` never returning `true`, or infinite token chains. Mitigation: use absence of `nextPageToken` as end-of-results indicator, add a max-pages safety limit.
2. **`total` field unreliability**: Some community reports indicate `total` is not always returned. Do not rely on it for pagination termination.
3. **Empty JQL rejection**: v3 returns `400 Bad Request` for empty/unbounded JQL queries. Our JQL builder always produces bounded queries so this should not be an issue.
4. **Sequential-only pagination**: Cannot parallelize page fetches — each request needs the `nextPageToken` from the previous response.

---

### 2. Myself Endpoint: `/rest/api/3/myself`

#### Endpoint Summary

| Attribute | Value |
|---|---|
| **v2 Path** | `GET /rest/api/2/myself` |
| **v3 Path** | `GET /rest/api/3/myself` |
| **Auth** | Basic Auth (`base64(email:api_token)`) |
| **Content-Type** | `application/json` |
| **Breaking** | No — response format is compatible |

#### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `expand` | string | No | Expand options (e.g., `groups`, `applicationRoles`). Not needed for our use case. |

#### Response Schema

```json
{
  "self": "https://your-domain.atlassian.net/rest/api/3/user?accountId=5b10a2844c20165700ede21g",
  "accountId": "5b10a2844c20165700ede21g",
  "accountType": "atlassian",
  "emailAddress": "user@company.com",
  "displayName": "User Name",
  "active": true,
  "timeZone": "America/Los_Angeles",
  "locale": "en_US",
  "avatarUrls": {
    "48x48": "https://avatar-management.services.atlassian.com/.../48",
    "24x24": "https://avatar-management.services.atlassian.com/.../24",
    "16x16": "https://avatar-management.services.atlassian.com/.../16",
    "32x32": "https://avatar-management.services.atlassian.com/.../32"
  }
}
```

**Response Fields (relevant to our use case):**

| Field | Type | Description |
|---|---|---|
| `emailAddress` | string | User's email address. **May be empty** if user has restricted privacy settings. |
| `displayName` | string | User's display name. |
| `accountId` | string | Unique Atlassian account ID. |
| `active` | boolean | Whether account is active. |
| `timeZone` | string/null | User's timezone. Can be null per privacy settings. |

#### v2 vs v3 Differences

**Minimal differences.** The response structure is compatible. Key notes:
- The `self` URL changes from `/rest/api/2/user` to `/rest/api/3/user` in the response
- `name` and `key` fields may be removed (deprecated for privacy — use `accountId` instead)
- `emailAddress` may be empty based on user privacy settings (same in v2)
- Authentication method unchanged (Basic Auth)
- Error codes unchanged (401, 403)

#### Impact on Our Code

**Minimal.** Our `ValidateToken()` in `atlassian.go` only reads `emailAddress` from the response. This field exists in both v2 and v3. Only the URL path needs to change from `/rest/api/2/myself` to `/rest/api/3/myself`.

---

### 3. Issue Endpoint: `/rest/api/3/issue/{issueIdOrKey}`

#### Endpoint Summary

| Attribute | Value |
|---|---|
| **v2 Path** | `GET /rest/api/2/issue/{key}?expand=comments` |
| **v3 Path** | `GET /rest/api/3/issue/{key}?expand=comments` |
| **Auth** | Basic Auth (`base64(email:api_token)`) |
| **Content-Type** | `application/json` |
| **Breaking** | **YES** — description and comment body return ADF, not string |

#### Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `issueIdOrKey` | string (path) | Yes | Issue key (e.g., `PROJ-1`) or numeric ID. |
| `fields` | string (query) | No | Comma-separated field names. Default: all navigable fields. |
| `expand` | string (query) | No | Comma-separated expand options. `comments` expands inline comments. `renderedFields` gives HTML-rendered versions. |
| `properties` | string (query) | No | Comma-separated entity property keys. |
| `fieldsByKeys` | boolean (query) | No | Return fields by key instead of ID. |

#### Response Schema

```json
{
  "id": "10001",
  "key": "PROJ-1",
  "self": "https://your-domain.atlassian.net/rest/api/3/issue/10001",
  "fields": {
    "summary": "Issue title",
    "description": {
      "version": 1,
      "type": "doc",
      "content": [
        {
          "type": "paragraph",
          "content": [
            {
              "type": "text",
              "text": "This is the issue description in ADF format."
            }
          ]
        }
      ]
    },
    "status": {
      "name": "In Progress",
      "id": "3",
      "statusCategory": {
        "id": 4,
        "key": "indeterminate",
        "name": "In Progress"
      }
    },
    "priority": {
      "name": "High",
      "id": "2"
    },
    "assignee": {
      "accountId": "5b10a2844c20165700ede21g",
      "displayName": "John Doe",
      "emailAddress": "john@company.com",
      "active": true
    },
    "reporter": {
      "accountId": "5b10ac37c20165700ede21d",
      "displayName": "Jane Smith",
      "emailAddress": "jane@company.com",
      "active": true
    },
    "duedate": "2026-03-15",
    "created": "2026-01-15T10:00:00.000+0000",
    "updated": "2026-02-10T14:30:00.000+0000",
    "labels": ["bug", "critical"],
    "comment": {
      "comments": [
        {
          "id": "10001",
          "author": {
            "accountId": "5b10a2844c20165700ede21g",
            "displayName": "John Doe",
            "emailAddress": "john@company.com"
          },
          "body": {
            "version": 1,
            "type": "doc",
            "content": [
              {
                "type": "paragraph",
                "content": [
                  {
                    "type": "text",
                    "text": "This is a comment in ADF format."
                  }
                ]
              }
            ]
          },
          "created": "2026-01-16T09:30:00.000+0000",
          "updated": "2026-01-16T09:30:00.000+0000"
        }
      ],
      "maxResults": 20,
      "total": 1,
      "startAt": 0
    }
  }
}
```

#### v2 vs v3 Differences

| Field | v2 | v3 |
|---|---|---|
| `description` | `string` (wiki markup/plain text) | **`object` (ADF JSON)** or `null` |
| `comment.comments[].body` | `string` (wiki markup/plain text) | **`object` (ADF JSON)** |
| `assignee.name` | Present | **Removed** (privacy) — use `accountId` |
| `reporter.name` | Present | **Removed** (privacy) — use `accountId` |
| `comment.maxResults` | Present | Present (max 20 in search, full in issue GET) |

---

## CRITICAL: Atlassian Document Format (ADF) Impact

### What Changed

In v3, rich text fields (`description`, `environment`, `comment.body`) are returned as **Atlassian Document Format (ADF)** JSON objects instead of plain text strings.

### ADF Structure

```json
{
  "version": 1,
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        {
          "type": "text",
          "text": "Plain text content here"
        }
      ]
    },
    {
      "type": "heading",
      "attrs": { "level": 2 },
      "content": [
        {
          "type": "text",
          "text": "A heading"
        }
      ]
    },
    {
      "type": "bulletList",
      "content": [
        {
          "type": "listItem",
          "content": [
            {
              "type": "paragraph",
              "content": [
                {
                  "type": "text",
                  "text": "A list item"
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```

### Impact on Our Code

**Current implementation** in `client.go`:
```go
type issueFields struct {
    Description string `json:"description"` // Expects a string — WILL BREAK in v3
    // ...
    Comment     *commentsField `json:"comment"`
}
```

The `Description` field is typed as `string`. In v3, this field is a JSON object (or `null`). JSON unmarshaling will **fail silently** (set to empty string) or **error** depending on the decoder settings.

Similarly, `commentResponse.Body` is `string` but will receive an ADF object in v3.

### Recommended Solutions

#### Option A: Use `expand=renderedFields` (Recommended for MVP)

Request issues with `?expand=renderedFields` to get an additional `renderedFields` object containing HTML-rendered versions of rich text fields.

```go
type issueWithRenderedFields struct {
    Key            string         `json:"key"`
    Fields         issueFields    `json:"fields"`
    RenderedFields renderedFields `json:"renderedFields"`
}

type renderedFields struct {
    Description string `json:"description"` // HTML string
}
```

**Pros:** Simple, get HTML that can be displayed or stripped to plain text.
**Cons:** Additional data per response, HTML may not match UI exactly.

#### Option B: Use `json.RawMessage` for description

```go
type issueFields struct {
    Description json.RawMessage `json:"description"` // Raw JSON — could be string, object, or null
    // ...
}
```

Then write a helper to extract plain text from ADF:

```go
func extractTextFromADF(raw json.RawMessage) string {
    if len(raw) == 0 || string(raw) == "null" {
        return ""
    }
    // Try string first (v2 compat)
    var s string
    if json.Unmarshal(raw, &s) == nil {
        return s
    }
    // Parse as ADF
    var adf ADFDocument
    if json.Unmarshal(raw, &adf) == nil {
        return adf.PlainText()
    }
    return ""
}
```

**Pros:** Handles both v2 and v3 responses, no extra API expansion.
**Cons:** More complex, need ADF parser.

#### Option C: Request `renderedFields` in search `expand` param

For the search endpoint, add `expand=renderedFields` to the query:
```
GET /rest/api/3/search/jql?jql=...&fields=summary,status,...&expand=renderedFields
```

This returns `renderedFields.description` as HTML alongside the normal `fields.description` as ADF.

### Decision: Option A + C (renderedFields)

**Rationale:**
- Simplest approach for MVP
- Our TUI doesn't need rich formatting — plain text from HTML stripping is sufficient
- `renderedFields` provides HTML that's easy to convert to plain text
- Avoids building an ADF parser
- Can enhance with full ADF support later if needed
- For search results, we don't display description — so this only matters for `GetIssue`

**Implementation:**
1. For `SearchIssues`: Keep requesting specific fields via `fields` param. Description is not displayed in list view, so ADF vs string doesn't matter. Change `Description` to `json.RawMessage` for safe deserialization.
2. For `GetIssue`: Add `expand=renderedFields` to request. Use `renderedFields.description` for display. For comments, use `renderedFields` or extract text from ADF body.

---

## Pagination Algorithm

### Current Implementation (v2)

```go
// Single request, first 100 results
searchURL := fmt.Sprintf("%s/rest/api/2/search", c.baseURL)
params := url.Values{}
params.Set("jql", jql)
params.Set("maxResults", "100")
// Implicit: startAt=0 (default)
```

### New Implementation (v3) - Option A: Single Page (MVP)

```go
// Single request, first 100 results (matches current behavior)
searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)
params := url.Values{}
params.Set("jql", jql)
params.Set("maxResults", "100")
params.Set("fields", "key,summary,status,priority,assignee,reporter,duedate,created,updated,labels")
// No nextPageToken on first request

// Response will include nextPageToken if more results exist
// We ignore it for MVP — first 100 results only
```

**Response parsing:**
```go
type searchResponse struct {
    Issues        []issueResponse `json:"issues"`
    NextPageToken string          `json:"nextPageToken,omitempty"`
    // total may not be present — do not depend on it
}
```

### New Implementation (v3) - Option B: Multi-Page (Future Enhancement)

```go
func (c *Client) SearchIssuesAllPages(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
    jql := c.jqlBuilder.Build(filter)
    allIssues := []domain.Issue{}
    nextPageToken := ""
    maxPages := 10 // Safety limit to prevent infinite loops

    for page := 0; page < maxPages; page++ {
        searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)
        params := url.Values{}
        params.Set("jql", jql)
        params.Set("maxResults", "100")
        params.Set("fields", "key,summary,status,priority,assignee,reporter,duedate,created,updated,labels")

        if nextPageToken != "" {
            params.Set("nextPageToken", nextPageToken)
        }

        // Make request...
        var result searchResponse
        // ...decode response...

        allIssues = append(allIssues, c.convertIssues(result.Issues)...)

        // End-of-results: nextPageToken absent
        if result.NextPageToken == "" {
            break
        }
        nextPageToken = result.NextPageToken
    }

    return allIssues, nil
}
```

**Key design decisions for multi-page:**
- Use `nextPageToken == ""` as termination condition (most reliable)
- Do NOT rely on `isLast` field (unreliable per community reports)
- Do NOT rely on `total` field (may not be present)
- Add `maxPages` safety limit to prevent infinite loops from pagination bugs
- Pages must be fetched sequentially — each needs the token from the previous response

### Pagination Flow Diagram

```
┌─────────────────┐
│ First Request    │
│ (no token)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     nextPageToken present?
│ Parse Response   │──── YES ──┐
└────────┬────────┘            │
         │ NO                  │
         ▼                     ▼
┌─────────────────┐  ┌─────────────────┐
│ Return Results   │  │ Request with     │
│ (done)           │  │ nextPageToken    │
└─────────────────┘  └────────┬────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │ Parse Response   │──── (loop)
                     └─────────────────┘
```

**Decision:** Start with Option A (single page) for MVP, add TODO for Option B.

**Rationale:**
- Matches current application behavior (no regression)
- Simpler implementation (lower risk)
- Can add multi-page in future sprint
- Most users won't notice (queries typically return <100 results)
- Avoids known pagination bugs in the v3 API

---

## Error Handling

### HTTP Status Codes

**Common Errors (v2 and v3):**

| Code | Meaning | Our Handling |
|---|---|---|
| `400 Bad Request` | Invalid JQL, empty JQL (v3), malformed request | Return error with Jira's error message |
| `401 Unauthorized` | Invalid or expired token | Return "invalid token or email" |
| `403 Forbidden` | Insufficient permissions | Return "insufficient permissions" |
| `404 Not Found` | Issue/resource doesn't exist | Return "not found" error |
| `429 Too Many Requests` | Rate limited | Return error (consider retry in future) |
| `500 Internal Server Error` | Jira server error | Return generic server error |

**v2-Only (now gone):**

| Code | Meaning | Notes |
|---|---|---|
| `410 Gone` | Endpoint removed | **This is the error we're fixing** — will not occur with v3 endpoints |

**v3-Specific:**

| Code | Meaning | Notes |
|---|---|---|
| `400 Bad Request` | Empty/unbounded JQL | New in v3 — our JQL builder always produces bounded queries |

### Error Response Format

```json
{
  "errorMessages": [
    "An error occurred"
  ],
  "errors": {
    "fieldName": "specific field error"
  }
}
```

Same format in both v2 and v3. Our existing error handling (`fmt.Errorf("Jira API error (status %d): %s", ...)`) works without changes.

### CAPTCHA Lockout

Detected via `X-Seraph-LoginReason: AUTHENTICATION_DENIED` header. Same in v2 and v3. Our existing handling in `atlassian.go:52` is correct.

---

## Go Response Type Definitions

### Search Response Types (Updated for v3)

```go
// searchResponse represents the v3 /rest/api/3/search/jql response.
type searchResponse struct {
	Issues        []issueResponse `json:"issues"`
	NextPageToken string          `json:"nextPageToken,omitempty"` // Cursor for next page; absent on last page
	// Note: 'total' and 'isLast' may not be reliably present in v3.
	// Do NOT depend on them for pagination logic.
}
```

**Changes from current:**
- Removed implicit reliance on `total` (was parsed but not used)
- Added `NextPageToken` field
- Removed `startAt`/`maxResults` fields (not in v3 response)

### Issue Response Types (Updated for v3)

```go
// issueResponse represents a single issue in API responses.
type issueResponse struct {
	Key    string      `json:"key"`
	Fields issueFields `json:"fields"`
}

// issueFields represents the fields of an issue.
type issueFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description"` // CHANGED: ADF object in v3, string in v2, or null
	Status      statusField     `json:"status"`
	Priority    *priorityField  `json:"priority"`
	Assignee    *userField      `json:"assignee"`
	Reporter    *userField      `json:"reporter"`
	DueDate     string          `json:"duedate"`
	Created     string          `json:"created"`
	Updated     string          `json:"updated"`
	Labels      []string        `json:"labels"`
	Comment     *commentsField  `json:"comment"`
}

// statusField represents a Jira status.
type statusField struct {
	Name string `json:"name"`
}

// priorityField represents a Jira priority.
type priorityField struct {
	Name string `json:"name"`
}

// userField represents a Jira user.
type userField struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	AccountId    string `json:"accountId"` // NEW: v3 uses accountId instead of name/key
}

// commentsField represents the comments section of an issue.
type commentsField struct {
	Comments []commentResponse `json:"comments"`
}

// commentResponse represents a single comment.
type commentResponse struct {
	ID      string          `json:"id"`
	Author  userField       `json:"author"`
	Body    json.RawMessage `json:"body"` // CHANGED: ADF object in v3, string in v2
	Created string          `json:"created"`
}
```

### ADF Helper Types (for future use)

```go
// ADFDocument represents a top-level Atlassian Document Format document.
type ADFDocument struct {
	Version int       `json:"version"`
	Type    string    `json:"type"` // Always "doc"
	Content []ADFNode `json:"content"`
}

// ADFNode represents a node in an ADF document tree.
type ADFNode struct {
	Type    string            `json:"type"`    // "paragraph", "heading", "text", "bulletList", etc.
	Text    string            `json:"text,omitempty"` // Present only on "text" type nodes
	Attrs   map[string]any    `json:"attrs,omitempty"`
	Content []ADFNode         `json:"content,omitempty"`
	Marks   []ADFMark         `json:"marks,omitempty"`
}

// ADFMark represents formatting applied to a text node.
type ADFMark struct {
	Type  string         `json:"type"` // "strong", "em", "code", "link", etc.
	Attrs map[string]any `json:"attrs,omitempty"`
}
```

### Plain Text Extraction from ADF

```go
// extractPlainText extracts plain text from an ADF JSON field.
// Handles v3 ADF objects, v2 plain strings, and null values.
func extractPlainText(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	// Try as plain string first (v2 compatibility)
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Parse as ADF document
	var doc ADFDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "" // Unparseable — return empty
	}

	// Walk the tree and collect text nodes
	var buf strings.Builder
	extractTextNodes(&buf, doc.Content)
	return strings.TrimSpace(buf.String())
}

// extractTextNodes recursively walks ADF nodes and collects text content.
func extractTextNodes(buf *strings.Builder, nodes []ADFNode) {
	for _, node := range nodes {
		if node.Type == "text" && node.Text != "" {
			buf.WriteString(node.Text)
		}
		if node.Type == "hardBreak" {
			buf.WriteString("\n")
		}
		if node.Type == "paragraph" && buf.Len() > 0 {
			buf.WriteString("\n")
		}
		if len(node.Content) > 0 {
			extractTextNodes(buf, node.Content)
		}
	}
}
```

### Myself Response Type (for reference)

```go
// myselfResponse represents the /rest/api/3/myself response.
// Only the emailAddress field is used in our ValidateToken flow.
type myselfResponse struct {
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
	AccountId    string `json:"accountId"`
	Active       bool   `json:"active"`
}
```

No changes needed — our existing inline struct is sufficient.

---

## Test Data Examples

### Success Cases

#### Search — First Page with Results

**Request:**
```http
GET /rest/api/3/search/jql?jql=project%3DPROJ+ORDER+BY+updated+DESC&maxResults=100&fields=key,summary,status,priority,assignee,reporter,duedate,created,updated,labels HTTP/1.1
Host: your-company.atlassian.net
Authorization: Basic dXNlckBjb21wYW55LmNvbTpteS1hcGktdG9rZW4=
Accept: application/json
```

**Response (200 OK):**
```json
{
  "issues": [
    {
      "key": "PROJ-42",
      "fields": {
        "summary": "Implement user authentication",
        "status": { "name": "In Progress" },
        "priority": { "name": "High" },
        "assignee": {
          "displayName": "John Doe",
          "emailAddress": "john@company.com",
          "accountId": "5b10a2844c20165700ede21g"
        },
        "reporter": {
          "displayName": "Jane Smith",
          "emailAddress": "jane@company.com",
          "accountId": "5b10ac37c20165700ede21d"
        },
        "duedate": "2026-03-15",
        "created": "2026-01-15T10:00:00.000+0000",
        "updated": "2026-02-10T14:30:00.000+0000",
        "labels": ["feature", "auth"]
      }
    },
    {
      "key": "PROJ-41",
      "fields": {
        "summary": "Fix login page CSS",
        "status": { "name": "Done" },
        "priority": { "name": "Low" },
        "assignee": null,
        "reporter": {
          "displayName": "Jane Smith",
          "emailAddress": "jane@company.com",
          "accountId": "5b10ac37c20165700ede21d"
        },
        "duedate": null,
        "created": "2026-01-10T08:00:00.000+0000",
        "updated": "2026-02-08T16:00:00.000+0000",
        "labels": []
      }
    }
  ],
  "nextPageToken": "eyJvcmRlckJ5IjoidXBkYXRlZCBERVNDIiwibGFzdFZhbHVlcyI6WyIyMDI2LTAyLTA4VDE2OjAwOjAwLjAwMCswMDAwIl19"
}
```

#### Search — Last Page (No More Results)

**Request:**
```http
GET /rest/api/3/search/jql?jql=project%3DPROJ&maxResults=100&fields=key,summary,status,priority,assignee,reporter,duedate,created,updated,labels&nextPageToken=eyJvcmRlckJ5... HTTP/1.1
```

**Response (200 OK):**
```json
{
  "issues": [
    {
      "key": "PROJ-1",
      "fields": {
        "summary": "Initial project setup",
        "status": { "name": "Done" },
        "priority": { "name": "Medium" },
        "assignee": {
          "displayName": "John Doe",
          "emailAddress": "john@company.com",
          "accountId": "5b10a2844c20165700ede21g"
        },
        "reporter": null,
        "duedate": null,
        "created": "2025-12-01T09:00:00.000+0000",
        "updated": "2025-12-15T12:00:00.000+0000",
        "labels": ["setup"]
      }
    }
  ]
}
```

Note: No `nextPageToken` field — this indicates it's the last page.

#### Search — Empty Results

**Response (200 OK):**
```json
{
  "issues": []
}
```

#### GetIssue — Full Issue with Comments

**Request:**
```http
GET /rest/api/3/issue/PROJ-42?expand=renderedFields,comments HTTP/1.1
Host: your-company.atlassian.net
Authorization: Basic dXNlckBjb21wYW55LmNvbTpteS1hcGktdG9rZW4=
Accept: application/json
```

**Response (200 OK):**
```json
{
  "key": "PROJ-42",
  "fields": {
    "summary": "Implement user authentication",
    "description": {
      "version": 1,
      "type": "doc",
      "content": [
        {
          "type": "paragraph",
          "content": [
            {
              "type": "text",
              "text": "We need to implement user authentication using OAuth 2.0."
            }
          ]
        },
        {
          "type": "paragraph",
          "content": [
            {
              "type": "text",
              "text": "Requirements:"
            }
          ]
        },
        {
          "type": "bulletList",
          "content": [
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    { "type": "text", "text": "Login page" }
                  ]
                }
              ]
            },
            {
              "type": "listItem",
              "content": [
                {
                  "type": "paragraph",
                  "content": [
                    { "type": "text", "text": "Token refresh" }
                  ]
                }
              ]
            }
          ]
        }
      ]
    },
    "status": { "name": "In Progress" },
    "priority": { "name": "High" },
    "assignee": {
      "displayName": "John Doe",
      "emailAddress": "john@company.com",
      "accountId": "5b10a2844c20165700ede21g"
    },
    "reporter": {
      "displayName": "Jane Smith",
      "emailAddress": "jane@company.com",
      "accountId": "5b10ac37c20165700ede21d"
    },
    "duedate": "2026-03-15",
    "created": "2026-01-15T10:00:00.000+0000",
    "updated": "2026-02-10T14:30:00.000+0000",
    "labels": ["feature", "auth"],
    "comment": {
      "comments": [
        {
          "id": "10001",
          "author": {
            "displayName": "Jane Smith",
            "emailAddress": "jane@company.com",
            "accountId": "5b10ac37c20165700ede21d"
          },
          "body": {
            "version": 1,
            "type": "doc",
            "content": [
              {
                "type": "paragraph",
                "content": [
                  {
                    "type": "text",
                    "text": "Can we use the existing OAuth library?"
                  }
                ]
              }
            ]
          },
          "created": "2026-01-16T09:30:00.000+0000"
        },
        {
          "id": "10002",
          "author": {
            "displayName": "John Doe",
            "emailAddress": "john@company.com",
            "accountId": "5b10a2844c20165700ede21g"
          },
          "body": {
            "version": 1,
            "type": "doc",
            "content": [
              {
                "type": "paragraph",
                "content": [
                  { "type": "text", "text": "Yes, I'll use " },
                  { "type": "text", "text": "golang.org/x/oauth2", "marks": [{ "type": "code" }] },
                  { "type": "text", "text": " for the implementation." }
                ]
              }
            ]
          },
          "created": "2026-01-16T10:15:00.000+0000"
        }
      ],
      "maxResults": 20,
      "total": 2,
      "startAt": 0
    }
  },
  "renderedFields": {
    "description": "<p>We need to implement user authentication using OAuth 2.0.</p>\n<p>Requirements:</p>\n<ul>\n<li>Login page</li>\n<li>Token refresh</li>\n</ul>"
  }
}
```

#### Myself — Successful Validation

**Request:**
```http
GET /rest/api/3/myself HTTP/1.1
Host: your-company.atlassian.net
Authorization: Basic dXNlckBjb21wYW55LmNvbTpteS1hcGktdG9rZW4=
Accept: application/json
```

**Response (200 OK):**
```json
{
  "self": "https://your-company.atlassian.net/rest/api/3/user?accountId=5b10a2844c20165700ede21g",
  "accountId": "5b10a2844c20165700ede21g",
  "accountType": "atlassian",
  "emailAddress": "user@company.com",
  "displayName": "User Name",
  "active": true,
  "timeZone": "America/Los_Angeles",
  "locale": "en_US",
  "avatarUrls": {
    "48x48": "https://avatar-management.services.atlassian.com/default/48",
    "24x24": "https://avatar-management.services.atlassian.com/default/24",
    "16x16": "https://avatar-management.services.atlassian.com/default/16",
    "32x32": "https://avatar-management.services.atlassian.com/default/32"
  }
}
```

### Error Cases

#### 401 Unauthorized — Invalid Token

**Response:**
```json
{
  "errorMessages": ["Client must be authenticated to access this resource."],
  "errors": {}
}
```

#### 403 Forbidden — Insufficient Permissions

**Response:**
```json
{
  "errorMessages": ["You do not have the permission to see the specified issue."],
  "errors": {}
}
```

#### 400 Bad Request — Invalid JQL

**Response:**
```json
{
  "errorMessages": [],
  "errors": {
    "jql": "Error in the JQL Query: 'projectt' is not a valid JQL field name."
  }
}
```

#### 400 Bad Request — Empty/Unbounded JQL (v3 only)

**Response:**
```json
{
  "errorMessages": ["Invalid request payload. Refer to the REST API documentation and try again."],
  "errors": {}
}
```

#### 404 Not Found — Issue Does Not Exist

**Response:**
```json
{
  "errorMessages": ["Issue does not exist or you do not have permission to see it."],
  "errors": {}
}
```

#### CAPTCHA Lockout

**Response (403):**
```
HTTP/1.1 403 Forbidden
X-Seraph-LoginReason: AUTHENTICATION_DENIED
Content-Type: application/json

{
  "errorMessages": ["Login denied"],
  "errors": {}
}
```

---

## Configuration Migration Research

### Current Structure

```yaml
jira:
  url: "https://your-company.atlassian.net"
  username: "deprecated-field"  # Not used anymore
  custom_fields:  # Optional
    sprint: "customfield_10020"
    epic: "customfield_10014"

atlassian:
  email: "your-email@company.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
```

### New Structure

```yaml
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"
  custom_fields:  # Optional (moved from jira section)
    sprint: "customfield_10020"
    epic: "customfield_10014"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
```

**Rationale:**
- URL and email both relate to Atlassian account — should be together
- Removes dead `jira.username` field
- Reduces confusion (both sections referred to same service)
- Custom fields naturally belong with Atlassian config

**Validation Rules:**
- `atlassian.url`: Required, must start with `https://`
- `atlassian.email`: Required, must be valid email format
- `custom_fields`: Optional, no validation needed

---

## Summary: Changes Required per File

### `internal/adapter/jira/client.go`

| Line | Current | New | Notes |
|---|---|---|---|
| 40 | `/rest/api/2/search` | `/rest/api/3/search/jql` | Path change |
| 44 | (none) | Add explicit `fields` param | v3 only returns IDs by default |
| 75 | `/rest/api/2/issue/%s?expand=comments` | `/rest/api/3/issue/%s?expand=renderedFields,comments` | Path + expand change |
| 115-118 | `searchResponse` struct | Add `NextPageToken` field | New pagination |
| 125 | `Description string` | `Description json.RawMessage` | ADF support |
| 159 | `Body string` | `Body json.RawMessage` | ADF support (comments) |
| 176-232 | `convertIssue()` | Add ADF text extraction | Handle new description format |

### `internal/adapter/auth/atlassian.go`

| Line | Current | New | Notes |
|---|---|---|---|
| 34 | `/rest/api/2/myself` | `/rest/api/3/myself` | Path change only |

### `internal/adapter/jira/client_test.go`

| Line | Current | New | Notes |
|---|---|---|---|
| 17 | `/rest/api/2/search` | `/rest/api/3/search/jql` | Path assertion |
| 96 | `/rest/api/2/issue/PROJ-1` | `/rest/api/3/issue/PROJ-1` | Path assertion |
| All | v2 response format | v3 response format | Test data update |

### `internal/domain/ports.go`

| Line | Current | New | Notes |
|---|---|---|---|
| 74-79 | `Config` with `Jira` + `Atlassian` | `Config` with unified `Atlassian` | Remove `Jira` field |
| 82-86 | `JiraConfig` struct | Remove entirely | Consolidate into `AtlassianConfig` |
| 96-98 | `AtlassianConfig` (email only) | Add `URL` + `CustomFields` | Expand struct |

---

## External Resources

### Official Documentation
- [Jira Cloud REST API v3 Intro](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/)
- [Issue Search API Group](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/)
- [Myself API Group](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/)
- [Atlassian Document Format](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
- [Atlassian Changelog #CHANGE-2046](https://developer.atlassian.com/changelog/#CHANGE-2046)

### Migration Guides
- [Atlassian REST API Search Endpoints Deprecation](https://docs.adaptavist.com/sr4jc/latest/release-notes/breaking-changes/atlassian-rest-api-search-endpoints-deprecation)
- [Run JQL Search Query Using Jira Cloud REST API](https://confluence.atlassian.com/jirakb/run-jql-search-query-using-jira-cloud-rest-api-1289424308.html)
- [How to use the new Jira cloud issue search API](https://community.atlassian.com/forums/Jira-articles/How-to-use-the-new-Jira-cloud-issue-search-API/ba-p/3006109)

### Community Discussions & Known Issues
- [v3 search/jql endpoint issues](https://community.atlassian.com/forums/Jira-questions/REST-The-new-rest-api-3-search-jql-endpoint-is-a-complete/qaq-p/3101716)
- [nextPageToken pagination issues](https://community.atlassian.com/forums/Jira-questions/JIRA-Search-pagination-with-nextPageToken-doesn-t-updated-on-2nd/qaq-p/3101873)
- [Slower fetching with nextPageToken](https://community.developer.atlassian.com/t/jira-cloud-rest-api-v3-search-jql-slower-fetching-with-nextpagetoken-no-totalissues-any-workarounds/90176)
- [isLast field request](https://jira.atlassian.com/browse/JRACLOUD-94648)
- [nextPageToken bug](https://jira.atlassian.com/browse/JRACLOUD-94632)
- [Jira API Migration Discussion](https://community.atlassian.com/forums/Jira-questions/Jira-API-Migration-to-rest-api-3-search-jql/qaq-p/3111339)

---

## Open Questions & Decisions

### Question 1: Multi-Page Fetching Strategy

**Decision:** Start with single-page (Option A), add TODO for multi-page
**Rationale:**
- Matches current application behavior (no regression)
- Simpler implementation (lower risk)
- Avoids known pagination bugs in v3 API
- Most queries return <100 results

### Question 2: POST vs GET for Search

**Decision:** Keep using GET
**Rationale:**
- Current JQL queries are short (<256 chars)
- GET is simpler and more cacheable
- Can switch to POST if URL length becomes an issue

### Question 3: ADF Description Handling

**Decision:** Use `json.RawMessage` for `description` and `comment.body` fields, with a plain text extraction helper
**Rationale:**
- `json.RawMessage` safely handles both v2 strings and v3 ADF objects
- Plain text extraction is sufficient for TUI display
- `expand=renderedFields` provides HTML fallback for GetIssue
- Avoids full ADF parser dependency for MVP

### Question 4: Backward Compatibility During Transition

**Decision:** No backward compatibility — require immediate config migration
**Rationale:**
- API v2 is completely removed (410 errors)
- Cannot support v2 endpoints anymore
- Forcing migration is safer than silent behavior changes
- Migration guide makes it easy (<5 minutes)
