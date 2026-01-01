# AGENTS.md

This file is the centralized source of truth for AI coding assistants (Claude Code, GitHub Copilot, Cursor, etc.) working in this repository.

## Context Engineering

This project uses context engineering with agentic tooling. Documentation is structured to support both human developers and AI agents working on this product.

### Documentation Requirements

**All code changes must consider updates to these files:**

| File | Purpose | Audience |
|------|---------|----------|
| `README.md` | Project overview, quick start, installation | Humans, GitHub visitors |
| `documentation/product-summary.md` | High-level product vision and features | Humans, agents needing context |
| `documentation/product-details.md` | Detailed specifications, UI/UX, workflows | Humans, agents implementing features |
| `documentation/technical-details.md` | Architecture, APIs, data flows, integrations | Agents, technical implementers |

### Documentation Standards

- **Language**: Concise, professional, unambiguous
- **Diagrams**: Mermaid preferred for rendering; ASCII for terminal/agent compatibility
- **Updates**: Every code change that affects behavior, architecture, or user experience must update relevant documentation
- **Cross-references**: Link between documents to avoid duplication

### Agent Responsibilities

When making code changes, agents must:
1. Review affected documentation files
2. Update product-summary.md if features or vision change
3. Update product-details.md if UI, workflows, or specifications change
4. Update technical-details.md if architecture, APIs, or data flows change
5. Update README.md if installation, usage, or quick start changes

## Project Overview

**gojira-tmux** - A team-based Jira viewer built as a TUI (Terminal User Interface) application in Go using the BubbleTea framework. Provides real-time access to Jira data via the Jira REST API.

## Technology Stack

- **Language**: Go (latest stable version)
- **TUI Framework**: [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- **Jira Integration**: [andygrunwald/go-jira](https://github.com/andygrunwald/go-jira)
- **Styling**: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- **CI/CD**: GitHub Actions (multi-platform builds: macOS, Windows, Linux)
- **Distribution**: GitHub Packages

## Architecture

### Clean Architecture Layers

```
cmd/                    # Application entrypoints
internal/
  domain/              # Business entities and interfaces (no external dependencies)
  usecase/             # Application business logic
  adapter/
    jira/              # Jira REST API client implementation
    tui/               # BubbleTea UI components
    config/            # Configuration handling
  infrastructure/      # External concerns (logging, persistence)
pkg/                   # Reusable public packages
```

### BubbleTea Model Pattern

All TUI components follow the Elm architecture:
- `Model` - state container
- `Init()` - initial command
- `Update(msg) (Model, Cmd)` - state transitions
- `View() string` - render to string

## Development Commands

```bash
# Build
go build -o bin/gojira ./cmd/gojira

# Run tests
go test ./...

# Run single test
go test -run TestName ./path/to/package

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint
golangci-lint run

# Format
go fmt ./...

# Vet
go vet ./...
```

## Development Practices

### TDD Workflow (Mandatory)

1. Write test first - define expected behavior
2. Run test - verify it fails (red)
3. Implement minimum code to pass
4. Run test - verify it passes (green)
5. Refactor while keeping tests green
6. Commit

### Code Quality Gates

All PRs must pass:
- `go test ./...` - all tests pass
- `golangci-lint run` - no lint errors
- `go build` - successful compilation for all target platforms

### Testing Patterns

```go
// Table-driven tests
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}

// Interface mocking for Jira client
type MockJiraClient struct {
    // mock implementation
}
```

### Jira API Integration

- Use interfaces for Jira client to enable testing without API calls
- Rate limiting and retry logic in adapter layer
- Configuration via environment variables or config file

### Configuration

Environment variables:
- `JIRA_API_TOKEN` - Jira API token for REST API authentication
- `OKTA_CLIENT_SECRET` - Okta OAuth client secret for OIDC flow

See `documentation/technical-details.md` for full authentication architecture.

## GitHub Actions Workflow

The CI pipeline:
1. Runs on push to main and all PRs
2. Tests on Go latest
3. Lints with golangci-lint
4. Builds binaries for darwin/amd64, darwin/arm64, linux/amd64, windows/amd64
5. Publishes releases to GitHub Packages

## Commit Guidelines

- Atomic commits with clear messages
- Format: `type(scope): description`
- Types: feat, fix, refactor, test, docs, ci

## File Naming Conventions

- Go files: `snake_case.go`
- Test files: `snake_case_test.go`
- Interfaces in separate files: `interfaces.go` or `ports.go`
