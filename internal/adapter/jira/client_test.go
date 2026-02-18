package jira_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/jira"
	"github.com/stainedhead/gojira-tmux/internal/adapter/jira/testutil"
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// --- Search Endpoint Tests ---

func TestClient_SearchIssues_V3Endpoint(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	last := srv.LastSearchRequest()
	if last == nil {
		t.Fatal("no search request recorded")
	}
	if last.Path != testutil.V3SearchPath {
		t.Errorf("path = %q, want %q", last.Path, testutil.V3SearchPath)
	}
}

func TestClient_SearchIssues_SendsFieldsParam(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, _ = client.SearchIssues(context.Background(), domain.IssueFilter{})

	last := srv.LastSearchRequest()
	if last == nil {
		t.Fatal("no search request recorded")
	}
	if !strings.Contains(last.Query, "fields=") {
		t.Errorf("query missing 'fields' parameter: %s", last.Query)
	}
}

func TestClient_SearchIssues_SendsMaxResults(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, _ = client.SearchIssues(context.Background(), domain.IssueFilter{})

	last := srv.LastSearchRequest()
	if last == nil {
		t.Fatal("no search request recorded")
	}
	if !strings.Contains(last.Query, "maxResults=100") {
		t.Errorf("query missing maxResults=100: %s", last.Query)
	}
}

func TestClient_SearchIssues_ParsesV3Response(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issues, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("got %d issues, want 2", len(issues))
	}
	if issues[0].Key != "PROJ-1" {
		t.Errorf("issues[0].Key = %q, want PROJ-1", issues[0].Key)
	}
	if issues[0].Summary != "First issue" {
		t.Errorf("issues[0].Summary = %q, want %q", issues[0].Summary, "First issue")
	}
	if issues[0].Priority != "High" {
		t.Errorf("issues[0].Priority = %q, want High", issues[0].Priority)
	}
	if issues[1].Key != "PROJ-2" {
		t.Errorf("issues[1].Key = %q, want PROJ-2", issues[1].Key)
	}
}

func TestClient_SearchIssues_EmptyResults(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetSearchResponse(testutil.EmptySearchResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issues, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("got %d issues, want 0", len(issues))
	}
}

func TestClient_SearchIssues_IgnoresNextPageToken(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// Set response with nextPageToken present
	page1, _ := testutil.PaginatedSearchResponse()
	srv.SetSearchResponse(page1)

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issues, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return first page only, ignoring nextPageToken
	if len(issues) != 2 {
		t.Errorf("got %d issues, want 2 (first page only)", len(issues))
	}
	// Should have made only one request
	if srv.SearchRequestCount() != 1 {
		t.Errorf("search requests = %d, want 1 (single page only)", srv.SearchRequestCount())
	}
}

func TestClient_SearchIssues_WithADFDescription(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	issues := []json.RawMessage{
		testutil.IssueJSON("PROJ-1", "Issue with ADF", "Open",
			testutil.WithADFDescription("This is ADF text"),
		),
	}
	srv.SetSearchResponse(testutil.SearchResponse(1, issues, ""))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	result, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d issues, want 1", len(result))
	}
	if result[0].Description != "This is ADF text" {
		t.Errorf("description = %q, want %q", result[0].Description, "This is ADF text")
	}
}

func TestClient_SearchIssues_WithStringDescription(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// v2 compat: description as plain string
	issues := []json.RawMessage{
		testutil.IssueJSON("PROJ-1", "Issue with string desc", "Open",
			testutil.WithDescription("Plain text description"),
		),
	}
	srv.SetSearchResponse(testutil.SearchResponse(1, issues, ""))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	result, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Description != "Plain text description" {
		t.Errorf("description = %q, want %q", result[0].Description, "Plain text description")
	}
}

func TestClient_SearchIssues_Unauthorized(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "wrong@example.com", "wrong-token", nil, nil)
	_, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want to contain '401'", err.Error())
	}
}

func TestClient_SearchIssues_BadRequest(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetSearchError(http.StatusBadRequest, testutil.ErrorResponse("Invalid JQL"))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err == nil {
		t.Fatal("expected error for bad request")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error = %q, want to contain '400'", err.Error())
	}
}

func TestClient_SearchIssues_ServerError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetSearchError(http.StatusInternalServerError, testutil.ErrorResponse("Internal error"))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain '500'", err.Error())
	}
}

// --- GetIssue Endpoint Tests ---

func TestClient_GetIssue_V3Endpoint(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithCommentsResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.GetIssue(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	last := srv.LastIssueRequest()
	if last == nil {
		t.Fatal("no issue request recorded")
	}
	if !strings.HasPrefix(last.Path, testutil.V3IssuePath) {
		t.Errorf("path = %q, want prefix %q", last.Path, testutil.V3IssuePath)
	}
}

func TestClient_GetIssue_ExpandsRenderedFieldsAndComments(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithCommentsResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, _ = client.GetIssue(context.Background(), "PROJ-1")

	last := srv.LastIssueRequest()
	if last == nil {
		t.Fatal("no issue request recorded")
	}
	if !strings.Contains(last.Query, "renderedFields") {
		t.Errorf("query missing renderedFields expand: %s", last.Query)
	}
	if !strings.Contains(last.Query, "comments") {
		t.Errorf("query missing comments expand: %s", last.Query)
	}
}

func TestClient_GetIssue_ParsesAllFields(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithCommentsResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issue, err := client.GetIssue(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if issue.Key != "PROJ-1" {
		t.Errorf("Key = %q, want PROJ-1", issue.Key)
	}
	if issue.Summary != "Test issue with comments" {
		t.Errorf("Summary = %q, want %q", issue.Summary, "Test issue with comments")
	}
	if issue.Status != "Open" {
		t.Errorf("Status = %q, want Open", issue.Status)
	}
	if issue.Priority != "High" {
		t.Errorf("Priority = %q, want High", issue.Priority)
	}
	if issue.Description != "Detailed description" {
		t.Errorf("Description = %q, want %q", issue.Description, "Detailed description")
	}
	if issue.Assignee == nil || issue.Assignee.Name != "John Doe" {
		t.Errorf("Assignee = %v, want John Doe", issue.Assignee)
	}
	if issue.Reporter == nil || issue.Reporter.Name != "Jane Smith" {
		t.Errorf("Reporter = %v, want Jane Smith", issue.Reporter)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Labels count = %d, want 2", len(issue.Labels))
	}
	if issue.DueDate == nil {
		t.Error("DueDate is nil, want non-nil")
	}
}

func TestClient_GetIssue_WithADFDescription(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithADFResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issue, err := client.GetIssue(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(issue.Description, "We need to implement user authentication using OAuth 2.0.") {
		t.Errorf("Description = %q, want to contain ADF text", issue.Description)
	}
	if !strings.Contains(issue.Description, "Requirements:") {
		t.Errorf("Description = %q, want to contain 'Requirements:'", issue.Description)
	}
}

func TestClient_GetIssue_WithADFComments(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithADFResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issue, err := client.GetIssue(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issue.Comments) != 2 {
		t.Fatalf("Comments count = %d, want 2", len(issue.Comments))
	}
	if issue.Comments[0].Body != "Working on this now" {
		t.Errorf("Comments[0].Body = %q, want %q", issue.Comments[0].Body, "Working on this now")
	}
	if issue.Comments[0].Author != "John Doe" {
		t.Errorf("Comments[0].Author = %q, want John Doe", issue.Comments[0].Author)
	}
	if issue.Comments[1].Body != "Looks good, please add tests" {
		t.Errorf("Comments[1].Body = %q, want %q", issue.Comments[1].Body, "Looks good, please add tests")
	}
	if issue.Comments[1].Author != "Jane Smith" {
		t.Errorf("Comments[1].Author = %q, want Jane Smith", issue.Comments[1].Author)
	}
}

func TestClient_GetIssue_WithStringComments(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// v2 compat: comments with plain string body
	srv.SetIssueResponse("PROJ-1", testutil.IssueWithCommentsResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issue, err := client.GetIssue(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issue.Comments) != 2 {
		t.Fatalf("Comments count = %d, want 2", len(issue.Comments))
	}
	if issue.Comments[0].Body != "Working on this now" {
		t.Errorf("Comments[0].Body = %q, want %q", issue.Comments[0].Body, "Working on this now")
	}
}

func TestClient_GetIssue_NullDescription(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// Issue with no description set (will be null/absent in JSON)
	srv.SetIssueResponse("PROJ-1", testutil.IssueJSON("PROJ-1", "No description", "Open"))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	issue, err := client.GetIssue(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Description != "" {
		t.Errorf("Description = %q, want empty string for null/missing description", issue.Description)
	}
}

func TestClient_GetIssue_NotFound(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// Don't set any issue response — unknown key returns 404
	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.GetIssue(context.Background(), "UNKNOWN-99")
	if err == nil {
		t.Fatal("expected error for nonexistent issue")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want to contain '404'", err.Error())
	}
}

func TestClient_GetIssue_ServerError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueError(http.StatusInternalServerError)

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.GetIssue(context.Background(), "PROJ-1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain '500'", err.Error())
	}
}

func TestClient_GetIssue_Unauthorized(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueJSON("PROJ-1", "Test", "Open"))

	client := jira.NewClient(srv.BaseURL(), "wrong@example.com", "wrong-token", nil, nil)
	_, err := client.GetIssue(context.Background(), "PROJ-1")
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want to contain '401'", err.Error())
	}
}

// --- GetIssueComments Tests ---

func TestClient_GetIssueComments(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithADFResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	comments, err := client.GetIssueComments(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("Comments count = %d, want 2", len(comments))
	}
	if comments[0].Body != "Working on this now" {
		t.Errorf("Comments[0].Body = %q, want %q", comments[0].Body, "Working on this now")
	}
	if comments[1].Body != "Looks good, please add tests" {
		t.Errorf("Comments[1].Body = %q, want %q", comments[1].Body, "Looks good, please add tests")
	}
}

func TestClient_GetIssueComments_NoComments(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueJSON("PROJ-1", "No comments", "Open"))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	comments, err := client.GetIssueComments(context.Background(), "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("Comments count = %d, want 0", len(comments))
	}
}

func TestClient_GetIssueComments_Error(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// Don't set any issue response — will return 404
	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.GetIssueComments(context.Background(), "UNKNOWN-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- ListStatuses Tests ---

func TestClient_ListStatuses_ReturnsNames(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetStatusResponse(testutil.StatusListResponse("To Do", "In Progress", "In Review", "Done"))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	statuses, err := client.ListStatuses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(statuses) != 4 {
		t.Fatalf("got %d statuses, want 4", len(statuses))
	}
	if statuses[0] != "To Do" {
		t.Errorf("statuses[0] = %q, want %q", statuses[0], "To Do")
	}
	if statuses[3] != "Done" {
		t.Errorf("statuses[3] = %q, want %q", statuses[3], "Done")
	}
}

func TestClient_ListStatuses_Unauthorized(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "wrong@example.com", "wrong-token", nil, nil)
	_, err := client.ListStatuses(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestClient_ListStatuses_ServerError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetStatusError(http.StatusInternalServerError)

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.ListStatuses(context.Background())
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain '500'", err.Error())
	}
}

// --- ListProjectStatuses Tests ---

func TestClient_ListProjectStatuses_ReturnsNames(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetProjectStatusesResponse("MYPROJ", testutil.ProjectStatusesResponse("To Do", "In Progress", "Done"))

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	statuses, err := client.ListProjectStatuses(context.Background(), "MYPROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(statuses) != 3 {
		t.Fatalf("got %d statuses, want 3", len(statuses))
	}
	// Results are sorted
	if statuses[0] != "Done" {
		t.Errorf("statuses[0] = %q, want Done (sorted first)", statuses[0])
	}
}

func TestClient_ListProjectStatuses_DeduplicatesAcrossIssueTypes(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetProjectStatusesResponse("MYPROJ", testutil.ProjectStatusesMultiTypeResponse())

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	statuses, err := client.ListProjectStatuses(context.Background(), "MYPROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Multi-type fixture has: To Do, In Progress (x2), Done → 3 unique names
	if len(statuses) != 3 {
		t.Fatalf("got %d statuses, want 3 (deduplicated)", len(statuses))
	}
}

func TestClient_ListProjectStatuses_EmptyProject(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// No response configured → mock returns empty array
	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	statuses, err := client.ListProjectStatuses(context.Background(), "EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("got %d statuses, want 0", len(statuses))
	}
}

func TestClient_ListProjectStatuses_Unauthorized(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetProjectStatusesResponse("MYPROJ", testutil.ProjectStatusesResponse("To Do"))

	client := jira.NewClient(srv.BaseURL(), "wrong@example.com", "wrong-token", nil, nil)
	_, err := client.ListProjectStatuses(context.Background(), "MYPROJ")
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want to contain '401'", err.Error())
	}
}

func TestClient_ListProjectStatuses_ServerError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetProjectStatusesError(http.StatusInternalServerError)

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.ListProjectStatuses(context.Background(), "MYPROJ")
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain '500'", err.Error())
	}
}

// --- Basic Auth Tests ---

func TestClient_SearchIssues_SendsBasicAuth(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	client := jira.NewClient(srv.BaseURL(), "test@example.com", "test-token", nil, nil)
	_, err := client.SearchIssues(context.Background(), domain.IssueFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Auth is validated by mock server — 200 means auth was correct
	if srv.SearchRequestCount() != 1 {
		t.Errorf("request count = %d, want 1", srv.SearchRequestCount())
	}
}
