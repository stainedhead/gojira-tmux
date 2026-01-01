# GitHub Copilot Instructions

This file provides guidance to GitHub Copilot when working in this repository.

> **IMPORTANT**: [AGENTS.md](../AGENTS.md) is the single source of truth for all project rules and conventions. **Do not modify this file** to add or change guidelines. All updates must be made in AGENTS.md.

## Reference

See [AGENTS.md](../AGENTS.md) for the complete, authoritative guide including technology stack, architecture, commands, and conventions.

## Project Context

gojira-tmux is a Go TUI application using BubbleTea for terminal UI and go-jira for Jira REST API integration. It follows Clean Architecture principles.

## Key Patterns

### BubbleTea Components

```go
type Model struct {
    // state fields
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    // handle messages
    }
    return m, nil
}

func (m Model) View() string {
    // return rendered string
}
```

### Table-Driven Tests

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "expected1"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // assertions
        })
    }
}
```

### Interface-Based Design

Define interfaces in domain layer, implement in adapter layer for testability.

## TDD Required

Write tests before implementation. Red-green-refactor cycle is mandatory.

## Commands

```bash
go test ./...           # Tests
golangci-lint run       # Lint
go build ./cmd/gojira   # Build
```
