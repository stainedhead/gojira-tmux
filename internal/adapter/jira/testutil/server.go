// Package testutil provides test infrastructure for Jira API v3 testing.
//
// It includes a configurable mock HTTP server and realistic JSON fixtures
// for all three v3 endpoints (search/jql, myself, issue).
//
// Basic usage:
//
//	srv := testutil.NewMockServer(t)
//	defer srv.Close()
//
//	// Configure responses
//	srv.SetSearchResponse(testutil.TwoIssueSearchResponse())
//	srv.SetMyselfResponse(testutil.DefaultMyselfResponse())
//
//	// Use srv.URL as the base URL for the Jira client
//	client := jira.NewClient(srv.URL, "test@example.com", "test-token", nil, nil)
//
// Error scenarios:
//
//	srv.SetSearchError(http.StatusUnauthorized, testutil.ErrorResponse("Unauthorized"))
//	srv.SetMyselfError(http.StatusForbidden, testutil.ErrorResponse("Forbidden"))
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// MockServer wraps httptest.Server with configurable v3 API handlers.
type MockServer struct {
	*httptest.Server

	mu sync.Mutex
	t  testing.TB

	// Response configuration
	searchResponse   json.RawMessage
	searchStatus     int
	myselfResponse   json.RawMessage
	myselfStatus     int
	issueResponses   map[string]json.RawMessage // keyed by issue key
	issueStatus      int
	paginatedSearch  map[string]json.RawMessage // keyed by nextPageToken value ("" = first page)

	// Request tracking
	searchRequests  []RecordedRequest
	myselfRequests  []RecordedRequest
	issueRequests   []RecordedRequest

	// Auth configuration
	expectedUser string
	expectedPass string
}

// RecordedRequest captures details of an incoming request for assertions.
type RecordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
}

// NewMockServer creates a new mock server with default 200 OK responses.
// The server handles /rest/api/3/search/jql, /rest/api/3/myself,
// and /rest/api/3/issue/{key} routes.
func NewMockServer(t testing.TB) *MockServer {
	ms := &MockServer{
		t:              t,
		searchStatus:   http.StatusOK,
		myselfStatus:   http.StatusOK,
		issueStatus:    http.StatusOK,
		issueResponses: make(map[string]json.RawMessage),
		paginatedSearch: make(map[string]json.RawMessage),
		expectedUser:   "test@example.com",
		expectedPass:   "test-token",
	}

	// Set default responses
	ms.searchResponse = TwoIssueSearchResponse()
	ms.myselfResponse = DefaultMyselfResponse()

	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// handler routes requests to the appropriate v3 endpoint handler.
func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	recorded := RecordedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Header: r.Header.Clone(),
	}

	switch {
	case r.URL.Path == V3SearchPath:
		ms.searchRequests = append(ms.searchRequests, recorded)
		ms.handleSearch(w, r)
	case r.URL.Path == V3MyselfPath:
		ms.myselfRequests = append(ms.myselfRequests, recorded)
		ms.handleMyself(w, r)
	case strings.HasPrefix(r.URL.Path, V3IssuePath):
		ms.issueRequests = append(ms.issueRequests, recorded)
		ms.handleIssue(w, r)
	default:
		ms.t.Errorf("MockServer: unexpected path: %s", r.URL.Path)
		http.NotFound(w, r)
	}
}

func (ms *MockServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if !ms.checkAuth(w, r) {
		return
	}

	if ms.searchStatus != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(ms.searchStatus)
		w.Write(ms.searchResponse)
		return
	}

	// Support paginated responses
	if len(ms.paginatedSearch) > 0 {
		token := r.URL.Query().Get("nextPageToken")
		if resp, ok := ms.paginatedSearch[token]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(ms.searchResponse)
}

func (ms *MockServer) handleMyself(w http.ResponseWriter, r *http.Request) {
	if !ms.checkAuth(w, r) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if ms.myselfStatus != http.StatusOK {
		w.WriteHeader(ms.myselfStatus)
	}
	w.Write(ms.myselfResponse)
}

func (ms *MockServer) handleIssue(w http.ResponseWriter, r *http.Request) {
	if !ms.checkAuth(w, r) {
		return
	}

	// Extract issue key from path: /rest/api/3/issue/PROJ-1
	issueKey := strings.TrimPrefix(r.URL.Path, V3IssuePath)

	if ms.issueStatus != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(ms.issueStatus)
		w.Write(ErrorResponse("Issue not found: " + issueKey))
		return
	}

	if resp, ok := ms.issueResponses[issueKey]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}

	// Return 404 for unknown issues
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write(ErrorResponse("Issue Does Not Exist"))
}

func (ms *MockServer) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if ms.expectedUser == "" && ms.expectedPass == "" {
		return true
	}

	user, pass, ok := r.BasicAuth()
	if !ok || user != ms.expectedUser || pass != ms.expectedPass {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(ErrorResponse("Unauthorized"))
		return false
	}
	return true
}

// --- Configuration methods ---

// SetSearchResponse configures the search endpoint to return the given JSON.
func (ms *MockServer) SetSearchResponse(response json.RawMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.searchResponse = response
	ms.searchStatus = http.StatusOK
}

// SetSearchError configures the search endpoint to return an error.
func (ms *MockServer) SetSearchError(status int, body json.RawMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.searchStatus = status
	ms.searchResponse = body
}

// SetPaginatedSearch configures paginated search responses.
// Pass "" as the token key for the first page (no token in request).
// Each entry maps a nextPageToken query param value to a response body.
func (ms *MockServer) SetPaginatedSearch(pages map[string]json.RawMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.paginatedSearch = pages
}

// SetMyselfResponse configures the myself endpoint to return the given JSON.
func (ms *MockServer) SetMyselfResponse(response json.RawMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.myselfResponse = response
	ms.myselfStatus = http.StatusOK
}

// SetMyselfError configures the myself endpoint to return an error.
func (ms *MockServer) SetMyselfError(status int, body json.RawMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.myselfStatus = status
	ms.myselfResponse = body
}

// SetIssueResponse configures a response for a specific issue key.
func (ms *MockServer) SetIssueResponse(key string, response json.RawMessage) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.issueResponses[key] = response
	ms.issueStatus = http.StatusOK
}

// SetIssueError configures the issue endpoint to return an error for all keys.
func (ms *MockServer) SetIssueError(status int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.issueStatus = status
}

// SetAuth configures the expected Basic Auth credentials.
// Pass empty strings to disable auth checking.
func (ms *MockServer) SetAuth(user, pass string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.expectedUser = user
	ms.expectedPass = pass
}

// --- Request inspection methods ---

// SearchRequestCount returns the number of search requests received.
func (ms *MockServer) SearchRequestCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.searchRequests)
}

// MyselfRequestCount returns the number of myself requests received.
func (ms *MockServer) MyselfRequestCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.myselfRequests)
}

// IssueRequestCount returns the number of issue requests received.
func (ms *MockServer) IssueRequestCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.issueRequests)
}

// LastSearchRequest returns the most recent search request, or nil if none.
func (ms *MockServer) LastSearchRequest() *RecordedRequest {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.searchRequests) == 0 {
		return nil
	}
	r := ms.searchRequests[len(ms.searchRequests)-1]
	return &r
}

// LastMyselfRequest returns the most recent myself request, or nil if none.
func (ms *MockServer) LastMyselfRequest() *RecordedRequest {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.myselfRequests) == 0 {
		return nil
	}
	r := ms.myselfRequests[len(ms.myselfRequests)-1]
	return &r
}

// LastIssueRequest returns the most recent issue request, or nil if none.
func (ms *MockServer) LastIssueRequest() *RecordedRequest {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(ms.issueRequests) == 0 {
		return nil
	}
	r := ms.issueRequests[len(ms.issueRequests)-1]
	return &r
}

// Reset clears all recorded requests and restores default responses.
func (ms *MockServer) Reset() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.searchRequests = nil
	ms.myselfRequests = nil
	ms.issueRequests = nil
	ms.searchResponse = TwoIssueSearchResponse()
	ms.searchStatus = http.StatusOK
	ms.myselfResponse = DefaultMyselfResponse()
	ms.myselfStatus = http.StatusOK
	ms.issueResponses = make(map[string]json.RawMessage)
	ms.issueStatus = http.StatusOK
	ms.paginatedSearch = make(map[string]json.RawMessage)
}

// --- Helper functions for test config ---

// TestConfig returns a minimal Config-like values suitable for creating a
// Jira client pointing at this mock server.
func (ms *MockServer) BaseURL() string {
	return ms.Server.URL
}

// TestCredentials returns the expected auth credentials.
func (ms *MockServer) TestCredentials() (email, token string) {
	return ms.expectedUser, ms.expectedPass
}
