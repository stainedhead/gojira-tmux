package jira_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/jira"
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestClient_SearchIssues(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check auth header
		user, pass, ok := r.BasicAuth()
		if !ok || user != "test@example.com" || pass != "test-token" {
			t.Error("missing or incorrect auth")
		}

		// Return mock response
		response := map[string]interface{}{
			"total":      2,
			"maxResults": 100,
			"issues": []map[string]interface{}{
				{
					"key": "PROJ-1",
					"fields": map[string]interface{}{
						"summary":     "First issue",
						"description": "Description 1",
						"status": map[string]interface{}{
							"name": "Open",
						},
						"priority": map[string]interface{}{
							"name": "High",
						},
						"created": "2024-01-01T10:00:00.000+0000",
						"updated": "2024-01-02T10:00:00.000+0000",
					},
				},
				{
					"key": "PROJ-2",
					"fields": map[string]interface{}{
						"summary":     "Second issue",
						"description": "Description 2",
						"status": map[string]interface{}{
							"name": "Done",
						},
						"created": "2024-01-01T10:00:00.000+0000",
						"updated": "2024-01-02T10:00:00.000+0000",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := jira.NewClient(server.URL, "test@example.com", "test-token", nil, nil)

	// Search issues
	ctx := context.Background()
	filter := domain.IssueFilter{}
	issues, err := client.SearchIssues(ctx, filter)

	if err != nil {
		t.Fatalf("SearchIssues() error = %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("SearchIssues() returned %d issues, want 2", len(issues))
	}

	if issues[0].Key != "PROJ-1" {
		t.Errorf("issues[0].Key = %q, want %q", issues[0].Key, "PROJ-1")
	}

	if issues[0].Summary != "First issue" {
		t.Errorf("issues[0].Summary = %q, want %q", issues[0].Summary, "First issue")
	}
}

func TestClient_GetIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/issue/PROJ-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"key": "PROJ-1",
			"fields": map[string]interface{}{
				"summary":     "Test issue",
				"description": "Test description",
				"status": map[string]interface{}{
					"name": "Open",
				},
				"created": "2024-01-01T10:00:00.000+0000",
				"updated": "2024-01-02T10:00:00.000+0000",
				"comment": map[string]interface{}{
					"comments": []map[string]interface{}{
						{
							"id": "1",
							"author": map[string]interface{}{
								"displayName": "John Doe",
							},
							"body":    "This is a comment",
							"created": "2024-01-01T11:00:00.000+0000",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := jira.NewClient(server.URL, "test@example.com", "test-token", nil, nil)

	ctx := context.Background()
	issue, err := client.GetIssue(ctx, "PROJ-1")

	if err != nil {
		t.Fatalf("GetIssue() error = %v", err)
	}

	if issue == nil {
		t.Fatal("GetIssue() returned nil")
	}

	if issue.Key != "PROJ-1" {
		t.Errorf("issue.Key = %q, want %q", issue.Key, "PROJ-1")
	}

	if len(issue.Comments) != 1 {
		t.Errorf("len(issue.Comments) = %d, want 1", len(issue.Comments))
	}

	if issue.Comments[0].Author != "John Doe" {
		t.Errorf("issue.Comments[0].Author = %q, want %q", issue.Comments[0].Author, "John Doe")
	}
}

func TestClient_SearchIssues_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errorMessages": ["Unauthorized"]}`))
	}))
	defer server.Close()

	client := jira.NewClient(server.URL, "test@example.com", "wrong-token", nil, nil)

	ctx := context.Background()
	_, err := client.SearchIssues(ctx, domain.IssueFilter{})

	if err == nil {
		t.Error("SearchIssues() expected error for 401, got nil")
	}
}
