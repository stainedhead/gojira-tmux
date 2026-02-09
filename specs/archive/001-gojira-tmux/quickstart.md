# Developer Quickstart: gojira-tmux

**Date**: 2026-01-01
**Feature**: [spec.md](./spec.md)

## Overview

This guide helps developers get started with implementing the gojira-tmux TUI application.

---

## Prerequisites

### Required Software

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.21+ | [golang.org/dl](https://golang.org/dl/) |
| Git | 2.x | [git-scm.com](https://git-scm.com/) |
| golangci-lint | 1.55+ | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |

### Jira Account Setup

1. **Generate API Token**:
   - Visit [id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
   - Create new token
   - Copy and save securely (needed for testing)

2. **Okta Application** (for auth testing):
   - Create OIDC application in Okta Admin Console
   - Type: Native Application
   - Grant Type: Authorization Code with PKCE
   - Redirect URI: `http://localhost:8080/callback`
   - Scopes: `openid`, `profile`, `email`

---

## Project Setup

### 1. Clone Repository

```bash
git clone https://github.com/stainedhead/gojira-tmux.git
cd gojira-tmux
```

### 2. Initialize Go Module

```bash
go mod init github.com/stainedhead/gojira-tmux
```

### 3. Install Dependencies

```bash
# TUI Framework
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest

# Jira Client
go get github.com/andygrunwald/go-jira/v2@latest

# Authentication
go get github.com/coreos/go-oidc/v3@latest
go get golang.org/x/oauth2@latest
go get github.com/zalando/go-keyring@latest

# Configuration
go get gopkg.in/yaml.v3@latest

# Tidy
go mod tidy
```

### 4. Create Directory Structure

```bash
mkdir -p cmd/gojira
mkdir -p internal/domain
mkdir -p internal/usecase
mkdir -p internal/adapter/{auth,config,jira}
mkdir -p internal/infrastructure/tui
mkdir -p pkg
```

### 5. Create Configuration File

Create `config.yaml` in project root (for development):

```yaml
jira:
  url: "https://your-company.atlassian.net"
  username: "your-email@company.com"
  custom_fields:
    sprint: "customfield_10020"
    epic: "customfield_10014"
    story_points: "customfield_10016"

okta:
  issuer: "https://your-company.okta.com/oauth2/default"
  client_id: "YOUR_OKTA_CLIENT_ID"
  callback_port: 8080
  scopes:
    - "openid"
    - "profile"
    - "email"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "Your Name"
    email: "your-email@company.com"
```

### 6. Set Environment Variables

```bash
export JIRA_API_TOKEN="your-jira-api-token"
export OKTA_CLIENT_SECRET="your-okta-client-secret"
```

---

## Development Workflow

### TDD Cycle (Mandatory)

All code changes must follow Test-Driven Development:

```bash
# 1. Write test first
vim internal/domain/issue_test.go

# 2. Run test (should fail - RED)
go test ./internal/domain/... -v

# 3. Implement minimum code
vim internal/domain/issue.go

# 4. Run test (should pass - GREEN)
go test ./internal/domain/... -v

# 5. Refactor while keeping green
# 6. Commit
git add . && git commit -m "feat(domain): add Issue entity"
```

### Common Commands

```bash
# Build
go build -o bin/gojira ./cmd/gojira

# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestIssueStaleness ./internal/domain/...

# Lint
golangci-lint run

# Format
go fmt ./...

# Vet
go vet ./...
```

### Running the Application

```bash
# Build and run
go build -o bin/gojira ./cmd/gojira && ./bin/gojira

# Or directly
go run ./cmd/gojira
```

---

## Implementation Order

Based on [spec.md](./spec.md) priorities:

### Phase 1: Core Infrastructure

1. **Domain Entities** (`internal/domain/`)
   - [ ] `issue.go` - Issue entity with NeedsAttention logic
   - [ ] `project.go` - Project entity
   - [ ] `team_member.go` - TeamMember entity
   - [ ] `comment.go` - Comment entity
   - [ ] `user.go` - User entity
   - [ ] `ports.go` - Port interfaces

2. **Configuration** (`internal/adapter/config/`)
   - [ ] `types.go` - Config structs
   - [ ] `config.go` - YAML loading and validation

3. **Token Storage** (`internal/adapter/auth/`)
   - [ ] `token_store.go` - Keyring adapter with fallback

### Phase 2: Authentication

4. **Okta OIDC** (`internal/adapter/auth/`)
   - [ ] `okta.go` - OIDC flow with PKCE
   - [ ] `callback.go` - Local callback server

5. **Auth Use Case** (`internal/usecase/`)
   - [ ] `setup_token.go` - First-time setup
   - [ ] `authenticate.go` - Auth orchestration

### Phase 3: Jira Integration

6. **Jira Client** (`internal/adapter/jira/`)
   - [ ] `client.go` - go-jira wrapper
   - [ ] `search.go` - JQL builder

7. **Issue Use Cases** (`internal/usecase/`)
   - [ ] `list_issues.go` - List with filtering
   - [ ] `get_issue_details.go` - Details with comments

### Phase 4: TUI

8. **TUI Foundation** (`internal/infrastructure/tui/`)
   - [ ] `app.go` - Root model with screen routing
   - [ ] `styles.go` - lipgloss styles
   - [ ] `keys.go` - Key bindings

9. **Screens** (`internal/infrastructure/tui/`)
   - [ ] `setup_screen.go` - Token setup
   - [ ] `login_screen.go` - Okta login
   - [ ] `main_screen.go` - Main view

10. **Components** (`internal/infrastructure/tui/`)
    - [ ] `filter_bar.go` - Filter dropdowns
    - [ ] `tickets_table.go` - Issue list
    - [ ] `properties_panel.go` - Issue properties
    - [ ] `comments_panel.go` - Issue comments

### Phase 5: Entrypoint

11. **Main** (`cmd/gojira/`)
    - [ ] `main.go` - Application bootstrap

---

## Testing Patterns

### Unit Test Example

```go
// internal/domain/issue_test.go
package domain_test

import (
    "testing"
    "time"

    "github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestIssue_NeedsAttention(t *testing.T) {
    tests := []struct {
        name     string
        issue    domain.Issue
        expected domain.AttentionType
    }{
        {
            name: "no attention for non-open issue",
            issue: domain.Issue{
                Status: "Done",
            },
            expected: domain.AttentionNone,
        },
        {
            name: "yellow for open issue without due date",
            issue: domain.Issue{
                Status:  "Open",
                DueDate: nil,
                Created: time.Now(),
            },
            expected: domain.AttentionNoDueDate,
        },
        {
            name: "red for stale open issue",
            issue: domain.Issue{
                Status:   "Open",
                Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
                Created:  time.Now().Add(-15 * 24 * time.Hour),
                Comments: []domain.Comment{},
            },
            expected: domain.AttentionStale,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := tt.issue.NeedsAttention()
            if got != tt.expected {
                t.Errorf("NeedsAttention() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Mock Example

```go
// internal/adapter/jira/mock_client.go
package jira

import (
    "context"

    "github.com/stainedhead/gojira-tmux/internal/domain"
)

type MockJiraClient struct {
    SearchIssuesFunc func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error)
    GetIssueFunc     func(ctx context.Context, key string) (*domain.Issue, error)
}

func (m *MockJiraClient) SearchIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
    return m.SearchIssuesFunc(ctx, filter)
}

func (m *MockJiraClient) GetIssue(ctx context.Context, key string) (*domain.Issue, error) {
    return m.GetIssueFunc(ctx, key)
}
```

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `keyring: secret not found` | No token stored - run setup flow |
| `401 Unauthorized` from Jira | Invalid API token - regenerate |
| `OIDC: token verification failed` | Check Okta issuer URL and client ID |
| `D-Bus error` on Linux | Install `gnome-keyring` or use file fallback |
| `TUI not rendering` | Check terminal supports ANSI colors |

### Debug Mode

Set environment variable for verbose logging:

```bash
GOJIRA_DEBUG=1 ./bin/gojira
```

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [spec.md](./spec.md) | Feature specification |
| [research.md](./research.md) | Technology decisions |
| [data-model.md](./data-model.md) | Entity definitions |
| [contracts/](./contracts/) | Interface contracts |
| [AGENTS.md](../../AGENTS.md) | Development practices |
