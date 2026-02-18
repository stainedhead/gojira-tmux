package testutil

import (
	"encoding/json"
	"fmt"
)

// V3 API endpoint paths.
const (
	V3SearchPath = "/rest/api/3/search/jql"
	V3MyselfPath = "/rest/api/3/myself"
	V3IssuePath  = "/rest/api/3/issue/"
	V3StatusPath = "/rest/api/3/status"
)

// MyselfResponse returns a realistic v3 /myself JSON response.
func MyselfResponse(displayName, email, accountID string) json.RawMessage {
	resp := map[string]interface{}{
		"self":         "https://example.atlassian.net/rest/api/3/user?accountId=" + accountID,
		"accountId":    accountID,
		"emailAddress": email,
		"displayName":  displayName,
		"active":       true,
		"timeZone":     "America/New_York",
		"accountType":  "atlassian",
	}
	b, _ := json.Marshal(resp)
	return b
}

// DefaultMyselfResponse returns a /myself response with default test values.
func DefaultMyselfResponse() json.RawMessage {
	return MyselfResponse("Test User", "test@example.com", "abc123def456")
}

// SearchResponse builds a v3 search/jql JSON response with the given issues
// and optional nextPageToken for pagination testing.
func SearchResponse(total int, issues []json.RawMessage, nextPageToken string) json.RawMessage {
	resp := map[string]interface{}{
		"total":  total,
		"issues": issues,
	}
	if nextPageToken != "" {
		resp["nextPageToken"] = nextPageToken
	}
	b, _ := json.Marshal(resp)
	return b
}

// IssueJSON builds a single issue JSON object matching v3 format.
// Use the Option functions to customize fields.
func IssueJSON(key, summary, status string, opts ...IssueOption) json.RawMessage {
	fields := map[string]interface{}{
		"summary": summary,
		"status":  map[string]interface{}{"name": status},
		"created": "2024-01-15T10:00:00.000+0000",
		"updated": "2024-01-20T14:30:00.000+0000",
	}

	issue := map[string]interface{}{
		"key":    key,
		"fields": fields,
	}

	for _, opt := range opts {
		opt(fields)
	}

	b, _ := json.Marshal(issue)
	return b
}

// IssueOption customizes an issue fixture.
type IssueOption func(fields map[string]interface{})

// WithDescription sets the issue description.
func WithDescription(desc string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["description"] = desc
	}
}

// WithPriority sets the issue priority.
func WithPriority(priority string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["priority"] = map[string]interface{}{"name": priority}
	}
}

// WithAssignee sets the issue assignee.
func WithAssignee(displayName, email string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["assignee"] = map[string]interface{}{
			"displayName":  displayName,
			"emailAddress": email,
		}
	}
}

// WithReporter sets the issue reporter.
func WithReporter(displayName, email string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["reporter"] = map[string]interface{}{
			"displayName":  displayName,
			"emailAddress": email,
		}
	}
}

// WithDueDate sets the issue due date (format: "2024-01-31").
func WithDueDate(date string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["duedate"] = date
	}
}

// WithLabels sets the issue labels.
func WithLabels(labels ...string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["labels"] = labels
	}
}

// WithComments adds comments to the issue.
func WithComments(comments ...CommentFixture) IssueOption {
	return func(fields map[string]interface{}) {
		commentObjs := make([]map[string]interface{}, len(comments))
		for i, c := range comments {
			commentObjs[i] = map[string]interface{}{
				"id": c.ID,
				"author": map[string]interface{}{
					"displayName": c.Author,
				},
				"body":    c.Body,
				"created": c.Created,
			}
		}
		fields["comment"] = map[string]interface{}{
			"comments": commentObjs,
		}
	}
}

// WithCreated overrides the default created timestamp.
func WithCreated(timestamp string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["created"] = timestamp
	}
}

// WithUpdated overrides the default updated timestamp.
func WithUpdated(timestamp string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["updated"] = timestamp
	}
}

// CommentFixture holds data for building a comment in an issue fixture.
type CommentFixture struct {
	ID      string
	Author  string
	Body    string
	Created string
}

// --- Pre-built fixtures for common test scenarios ---

// TwoIssueSearchResponse returns a search response with two typical issues.
func TwoIssueSearchResponse() json.RawMessage {
	issues := []json.RawMessage{
		IssueJSON("PROJ-1", "First issue", "Open",
			WithPriority("High"),
			WithDescription("Description of first issue"),
			WithAssignee("John Doe", "john@example.com"),
		),
		IssueJSON("PROJ-2", "Second issue", "Done",
			WithDescription("Description of second issue"),
			WithAssignee("Jane Smith", "jane@example.com"),
		),
	}
	return SearchResponse(2, issues, "")
}

// PaginatedSearchResponse returns two pages of search results.
// Page 1 has nextPageToken set; page 2 does not.
func PaginatedSearchResponse() (page1, page2 json.RawMessage) {
	issues1 := []json.RawMessage{
		IssueJSON("PROJ-1", "First issue", "Open"),
		IssueJSON("PROJ-2", "Second issue", "Open"),
	}
	page1 = SearchResponse(4, issues1, "eyJhbGciOiJIUzI1NiJ9.page2token")

	issues2 := []json.RawMessage{
		IssueJSON("PROJ-3", "Third issue", "Done"),
		IssueJSON("PROJ-4", "Fourth issue", "Done"),
	}
	page2 = SearchResponse(4, issues2, "")

	return page1, page2
}

// EmptySearchResponse returns a search response with zero results.
func EmptySearchResponse() json.RawMessage {
	return SearchResponse(0, []json.RawMessage{}, "")
}

// IssueWithCommentsResponse returns a single issue with comments (for GetIssue).
func IssueWithCommentsResponse() json.RawMessage {
	return IssueJSON("PROJ-1", "Test issue with comments", "Open",
		WithDescription("Detailed description"),
		WithPriority("High"),
		WithAssignee("John Doe", "john@example.com"),
		WithReporter("Jane Smith", "jane@example.com"),
		WithDueDate("2024-02-15"),
		WithLabels("backend", "api"),
		WithComments(
			CommentFixture{
				ID:      "10001",
				Author:  "John Doe",
				Body:    "Working on this now",
				Created: "2024-01-16T09:00:00.000+0000",
			},
			CommentFixture{
				ID:      "10002",
				Author:  "Jane Smith",
				Body:    "Looks good, please add tests",
				Created: "2024-01-17T14:30:00.000+0000",
			},
		),
	)
}

// StatusListResponse returns a /rest/api/3/status response for the given status names.
func StatusListResponse(names ...string) json.RawMessage {
	statuses := make([]map[string]interface{}, len(names))
	for i, name := range names {
		statuses[i] = map[string]interface{}{
			"id":          fmt.Sprintf("%d", i+1),
			"name":        name,
			"description": "",
			"statusCategory": map[string]interface{}{
				"id":   i + 1,
				"key":  "undefined",
				"name": name,
			},
		}
	}
	b, _ := json.Marshal(statuses)
	return b
}

// ErrorResponse returns a Jira error response body.
func ErrorResponse(messages ...string) json.RawMessage {
	resp := map[string]interface{}{
		"errorMessages": messages,
	}
	b, _ := json.Marshal(resp)
	return b
}

// --- ADF (Atlassian Document Format) helpers ---

// ADFDoc creates a simple ADF document JSON object with the given paragraph texts.
func ADFDoc(paragraphs ...string) map[string]interface{} {
	content := make([]interface{}, len(paragraphs))
	for i, text := range paragraphs {
		content[i] = map[string]interface{}{
			"type": "paragraph",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": text,
				},
			},
		}
	}
	return map[string]interface{}{
		"version": 1,
		"type":    "doc",
		"content": content,
	}
}

// WithADFDescription sets the issue description as an ADF document.
func WithADFDescription(paragraphs ...string) IssueOption {
	return func(fields map[string]interface{}) {
		fields["description"] = ADFDoc(paragraphs...)
	}
}

// WithADFComments adds comments with ADF body format to the issue.
func WithADFComments(comments ...CommentFixture) IssueOption {
	return func(fields map[string]interface{}) {
		commentObjs := make([]map[string]interface{}, len(comments))
		for i, c := range comments {
			commentObjs[i] = map[string]interface{}{
				"id": c.ID,
				"author": map[string]interface{}{
					"displayName": c.Author,
				},
				"body":    ADFDoc(c.Body),
				"created": c.Created,
			}
		}
		fields["comment"] = map[string]interface{}{
			"comments": commentObjs,
		}
	}
}

// IssueWithADFResponse returns a v3-style issue with ADF description and comments.
func IssueWithADFResponse() json.RawMessage {
	return IssueJSON("PROJ-1", "Test issue with ADF", "Open",
		WithADFDescription("We need to implement user authentication using OAuth 2.0.", "Requirements:"),
		WithPriority("High"),
		WithAssignee("John Doe", "john@example.com"),
		WithReporter("Jane Smith", "jane@example.com"),
		WithDueDate("2024-02-15"),
		WithLabels("backend", "api"),
		WithADFComments(
			CommentFixture{
				ID:      "10001",
				Author:  "John Doe",
				Body:    "Working on this now",
				Created: "2024-01-16T09:00:00.000+0000",
			},
			CommentFixture{
				ID:      "10002",
				Author:  "Jane Smith",
				Body:    "Looks good, please add tests",
				Created: "2024-01-17T14:30:00.000+0000",
			},
		),
	)
}
