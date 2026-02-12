package testutil_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/jira/testutil"
)

func TestMockServer_SearchEndpoint(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath+"?jql=project+IN+(PROJ)", nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	issues, ok := result["issues"].([]interface{})
	if !ok {
		t.Fatal("expected issues array in response")
	}
	if len(issues) != 2 {
		t.Errorf("got %d issues, want 2", len(issues))
	}

	if srv.SearchRequestCount() != 1 {
		t.Errorf("search request count = %d, want 1", srv.SearchRequestCount())
	}
}

func TestMockServer_SearchError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetSearchError(http.StatusGone, testutil.ErrorResponse("API removed, use v3"))

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath, nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusGone {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusGone)
	}
}

func TestMockServer_PaginatedSearch(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	page1, page2 := testutil.PaginatedSearchResponse()
	srv.SetPaginatedSearch(map[string]json.RawMessage{
		"":                                     page1, // first page (no token)
		"eyJhbGciOiJIUzI1NiJ9.page2token": page2, // second page
	})

	// Request page 1
	req1, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath+"?jql=project+IN+(PROJ)", nil)
	req1.SetBasicAuth("test@example.com", "test-token")
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("page 1 request failed: %v", err)
	}
	defer resp1.Body.Close()

	var result1 map[string]interface{}
	json.NewDecoder(resp1.Body).Decode(&result1)

	token, ok := result1["nextPageToken"].(string)
	if !ok || token == "" {
		t.Fatal("expected nextPageToken in page 1 response")
	}

	// Request page 2 using the token
	req2, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath+"?jql=project+IN+(PROJ)&nextPageToken="+token, nil)
	req2.SetBasicAuth("test@example.com", "test-token")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("page 2 request failed: %v", err)
	}
	defer resp2.Body.Close()

	var result2 map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result2)

	if _, hasToken := result2["nextPageToken"]; hasToken {
		t.Error("page 2 should not have nextPageToken")
	}

	if srv.SearchRequestCount() != 2 {
		t.Errorf("search request count = %d, want 2", srv.SearchRequestCount())
	}
}

func TestMockServer_MyselfEndpoint(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3MyselfPath, nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["emailAddress"] != "test@example.com" {
		t.Errorf("emailAddress = %v, want test@example.com", result["emailAddress"])
	}
	if result["displayName"] != "Test User" {
		t.Errorf("displayName = %v, want Test User", result["displayName"])
	}
}

func TestMockServer_MyselfError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetMyselfError(http.StatusForbidden, testutil.ErrorResponse("Forbidden"))

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3MyselfPath, nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestMockServer_IssueEndpoint(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueResponse("PROJ-1", testutil.IssueWithCommentsResponse())

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3IssuePath+"PROJ-1", nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["key"] != "PROJ-1" {
		t.Errorf("key = %v, want PROJ-1", result["key"])
	}

	if srv.IssueRequestCount() != 1 {
		t.Errorf("issue request count = %d, want 1", srv.IssueRequestCount())
	}
}

func TestMockServer_IssueNotFound(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// Don't set any issue response - should 404
	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3IssuePath+"UNKNOWN-99", nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestMockServer_IssueGlobalError(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetIssueError(http.StatusInternalServerError)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3IssuePath+"PROJ-1", nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestMockServer_AuthFailure(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath, nil)
	req.SetBasicAuth("wrong@example.com", "wrong-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMockServer_NoAuth(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath, nil)
	// No auth header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMockServer_DisabledAuth(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetAuth("", "")

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath, nil)
	// No auth header - should still succeed

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestMockServer_Reset(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetSearchError(http.StatusGone, testutil.ErrorResponse("gone"))

	// Make a request
	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath, nil)
	req.SetBasicAuth("test@example.com", "test-token")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	if srv.SearchRequestCount() != 1 {
		t.Fatalf("expected 1 request before reset")
	}

	srv.Reset()

	if srv.SearchRequestCount() != 0 {
		t.Errorf("expected 0 requests after reset, got %d", srv.SearchRequestCount())
	}

	// Make another request - should get 200 again
	req2, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath, nil)
	req2.SetBasicAuth("test@example.com", "test-token")
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("unexpected error on request after reset: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("after reset, status = %d, want %d", resp2.StatusCode, http.StatusOK)
	}
}

func TestMockServer_LastSearchRequest(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	if srv.LastSearchRequest() != nil {
		t.Error("expected nil before any requests")
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3SearchPath+"?jql=project+%3D+PROJ&maxResults=50", nil)
	req.SetBasicAuth("test@example.com", "test-token")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	last := srv.LastSearchRequest()
	if last == nil {
		t.Fatal("expected non-nil LastSearchRequest")
	}
	if last.Method != http.MethodGet {
		t.Errorf("Method = %q, want GET", last.Method)
	}
	if last.Path != testutil.V3SearchPath {
		t.Errorf("Path = %q, want %q", last.Path, testutil.V3SearchPath)
	}
}

func TestFixtures_EmptySearchResponse(t *testing.T) {
	data := testutil.EmptySearchResponse()

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	total, _ := result["total"].(float64)
	if total != 0 {
		t.Errorf("total = %v, want 0", total)
	}

	issues, _ := result["issues"].([]interface{})
	if len(issues) != 0 {
		t.Errorf("issues length = %d, want 0", len(issues))
	}
}

func TestFixtures_IssueJSON_WithOptions(t *testing.T) {
	data := testutil.IssueJSON("TEST-42", "My summary", "In Progress",
		testutil.WithDescription("Some desc"),
		testutil.WithPriority("Critical"),
		testutil.WithAssignee("Alice", "alice@example.com"),
		testutil.WithLabels("bug", "urgent"),
	)

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result["key"] != "TEST-42" {
		t.Errorf("key = %v, want TEST-42", result["key"])
	}

	fields, _ := result["fields"].(map[string]interface{})
	if fields["summary"] != "My summary" {
		t.Errorf("summary = %v, want My summary", fields["summary"])
	}
	if fields["description"] != "Some desc" {
		t.Errorf("description = %v, want Some desc", fields["description"])
	}

	priority, _ := fields["priority"].(map[string]interface{})
	if priority["name"] != "Critical" {
		t.Errorf("priority.name = %v, want Critical", priority["name"])
	}

	assignee, _ := fields["assignee"].(map[string]interface{})
	if assignee["displayName"] != "Alice" {
		t.Errorf("assignee.displayName = %v, want Alice", assignee["displayName"])
	}

	labels, _ := fields["labels"].([]interface{})
	if len(labels) != 2 {
		t.Errorf("labels length = %d, want 2", len(labels))
	}
}

func TestFixtures_ErrorResponse(t *testing.T) {
	data := testutil.ErrorResponse("Not found", "Issue does not exist")

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	msgs, _ := result["errorMessages"].([]interface{})
	if len(msgs) != 2 {
		t.Errorf("errorMessages length = %d, want 2", len(msgs))
	}
}

func TestMockServer_TestCredentials(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	email, token := srv.TestCredentials()
	if email != "test@example.com" {
		t.Errorf("email = %q, want test@example.com", email)
	}
	if token != "test-token" {
		t.Errorf("token = %q, want test-token", token)
	}
}

func TestMockServer_BaseURL(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	url := srv.BaseURL()
	if url == "" {
		t.Error("BaseURL() returned empty string")
	}

	// Should be a valid httptest URL
	resp, err := http.Get(url + testutil.V3SearchPath)
	if err != nil {
		t.Fatalf("failed to reach server: %v", err)
	}
	defer resp.Body.Close()
	// Without auth, should get 401
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestMockServer_CustomMyselfResponse(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetMyselfResponse(testutil.MyselfResponse("Custom User", "custom@example.com", "custom-id"))

	req, _ := http.NewRequest(http.MethodGet, srv.URL+testutil.V3MyselfPath, nil)
	req.SetBasicAuth("test@example.com", "test-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["displayName"] != "Custom User" {
		t.Errorf("displayName = %v, want Custom User", result["displayName"])
	}
	if result["emailAddress"] != "custom@example.com" {
		t.Errorf("emailAddress = %v, want custom@example.com", result["emailAddress"])
	}
}
