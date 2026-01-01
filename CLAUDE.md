# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> **IMPORTANT**: [AGENTS.md](./AGENTS.md) is the single source of truth for all project rules and conventions. **Do not modify this file** to add or change guidelines. All updates must be made in AGENTS.md.

## Quick Reference

See [AGENTS.md](./AGENTS.md) for the complete, authoritative guide including:
- Technology stack and architecture
- Development commands
- Testing patterns
- Code conventions

## Essential Commands

```bash
go test ./...                    # Run all tests
go test -run TestName ./...      # Run single test
golangci-lint run                # Lint
go build -o bin/gojira ./cmd/gojira  # Build
```

## Architecture Summary

Clean Architecture with BubbleTea TUI:
- `cmd/` - entrypoints
- `internal/domain/` - business entities (no dependencies)
- `internal/usecase/` - application logic
- `internal/adapter/` - Jira client, TUI components, config
- `internal/infrastructure/` - external concerns

## TDD is Mandatory

Always follow red-green-refactor:
1. Write failing test first
2. Implement minimum to pass
3. Refactor with tests green

## Key Packages

- BubbleTea for TUI (Elm architecture: Model, Init, Update, View)
- go-jira for Jira REST API
- lipgloss for styling

## Before Committing

```bash
go test ./... && golangci-lint run && go build ./...
```

## Recent Changes
- 001-gojira-tmux: Added [if applicable, e.g., PostgreSQL, CoreData, files or N/A]
