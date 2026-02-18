package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// Client implements the JiraPort interface for the Jira REST API.
type Client struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
	jqlBuilder *JQLBuilder
}

// NewClient creates a new Jira client.
func NewClient(baseURL, username, token string, projects []domain.Project, team []domain.TeamMember) *Client {
	return &Client{
		baseURL:    baseURL,
		username:   username,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		jqlBuilder: NewJQLBuilder(projects, team),
	}
}

// SearchIssues searches for issues matching the filter criteria.
func (c *Client) SearchIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
	jql := c.jqlBuilder.Build(filter)

	// Build search URL (v3 endpoint with cursor-based pagination)
	searchURL := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)
	params := url.Values{}
	params.Set("jql", jql)
	params.Set("maxResults", "100")
	params.Set("fields", "key,summary,description,status,priority,assignee,reporter,duedate,created,updated,labels")
	// TODO: implement multi-page pagination using nextPageToken

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertIssues(result.Issues), nil
}

// GetIssue retrieves a single issue with all details including comments.
func (c *Client) GetIssue(ctx context.Context, key string) (*domain.Issue, error) {
	issueURL := fmt.Sprintf("%s/rest/api/3/issue/%s?expand=renderedFields,comments", c.baseURL, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issueURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result issueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.convertIssue(result), nil
}

// ListStatuses returns all available issue status names from the Jira instance.
func (c *Client) ListStatuses(ctx context.Context) ([]string, error) {
	statusURL := fmt.Sprintf("%s/rest/api/3/status", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Jira API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	names := make([]string, 0, len(result))
	for _, s := range result {
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}
	return names, nil
}

// GetIssueComments retrieves comments for an issue.
func (c *Client) GetIssueComments(ctx context.Context, key string) ([]domain.Comment, error) {
	issue, err := c.GetIssue(ctx, key)
	if err != nil {
		return nil, err
	}
	return issue.Comments, nil
}

// Response types for JSON parsing

type searchResponse struct {
	Issues        []issueResponse `json:"issues"`
	NextPageToken string          `json:"nextPageToken,omitempty"`
}

type issueResponse struct {
	Key    string      `json:"key"`
	Fields issueFields `json:"fields"`
}

type issueFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description"`
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

type statusField struct {
	Name string `json:"name"`
}

type priorityField struct {
	Name string `json:"name"`
}

type userField struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type commentsField struct {
	Comments []commentResponse `json:"comments"`
}

type commentResponse struct {
	ID      string          `json:"id"`
	Author  userField       `json:"author"`
	Body    json.RawMessage `json:"body"`
	Created string          `json:"created"`
}

// Conversion methods

func (c *Client) convertIssues(issues []issueResponse) []domain.Issue {
	result := make([]domain.Issue, len(issues))
	for i, issue := range issues {
		converted := c.convertIssue(issue)
		if converted != nil {
			result[i] = *converted
		}
	}
	return result
}

func (c *Client) convertIssue(issue issueResponse) *domain.Issue {
	domainIssue := &domain.Issue{
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Description: extractPlainText(issue.Fields.Description),
		Status:      issue.Fields.Status.Name,
		Labels:      issue.Fields.Labels,
	}

	if issue.Fields.Priority != nil {
		domainIssue.Priority = issue.Fields.Priority.Name
	}

	if issue.Fields.Assignee != nil {
		domainIssue.Assignee = &domain.TeamMember{
			Name:  issue.Fields.Assignee.DisplayName,
			Email: issue.Fields.Assignee.EmailAddress,
		}
	}

	if issue.Fields.Reporter != nil {
		domainIssue.Reporter = &domain.TeamMember{
			Name:  issue.Fields.Reporter.DisplayName,
			Email: issue.Fields.Reporter.EmailAddress,
		}
	}

	if issue.Fields.DueDate != "" {
		if t, err := time.Parse("2006-01-02", issue.Fields.DueDate); err == nil {
			domainIssue.DueDate = &t
		}
	}

	if t, err := parseJiraTime(issue.Fields.Created); err == nil {
		domainIssue.Created = t
	}

	if t, err := parseJiraTime(issue.Fields.Updated); err == nil {
		domainIssue.Updated = t
	}

	// Convert comments (ADF body → plain text)
	if issue.Fields.Comment != nil {
		domainIssue.Comments = make([]domain.Comment, len(issue.Fields.Comment.Comments))
		for i, comment := range issue.Fields.Comment.Comments {
			created, _ := parseJiraTime(comment.Created)
			domainIssue.Comments[i] = domain.Comment{
				ID:      comment.ID,
				Author:  comment.Author.DisplayName,
				Body:    extractPlainText(comment.Body),
				Created: created,
			}
		}
	}

	return domainIssue
}

// parseJiraTime parses Jira's datetime format.
func parseJiraTime(s string) (time.Time, error) {
	// Jira uses format: "2024-01-01T10:00:00.000+0000"
	layouts := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05Z",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// Ensure Client implements domain.JiraPort.
var _ domain.JiraPort = (*Client)(nil)
