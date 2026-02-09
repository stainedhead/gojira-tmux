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

## Feature Specification Workflow

### Specs Directory Structure

All feature development uses the `specs/` directory for planning and tracking. Each feature gets its own subdirectory named after the feature.

**Directory Structure:**
```
specs/
└── <feature-name>/
    ├── spec.md                  # Feature specification and requirements
    ├── status.md                # **CRITICAL**: Phase progress tracking (update after each task)
    ├── plan.md                  # Implementation plan and architecture decisions
    ├── tasks.md                 # Task breakdown and progress tracking
    ├── research.md              # Research findings, API docs, examples
    ├── data-dictionary.md       # Data structures, types, schemas
    ├── architecture.md          # System architecture and component design
    └── implementation-notes.md  # Implementation details, gotchas, decisions
```

### Progressive Documentation Build

Documents are created progressively as the feature develops:

**Phase 0: Initial Research (PRD/Feature Research)**
- Input: Product Requirement Document, RFC, or feature research
- Purpose: Understand the problem, gather requirements, identify constraints
- **Update status.md**: Mark Phase 0 as "In Progress"

**Phase 1: Specification (spec.md)**
- Define what the feature does
- User requirements and acceptance criteria
- Goals and non-goals
- Success criteria
- **Update status.md**: Mark Phase 0 complete, Phase 1 in progress

**Phase 2: Research & Data Modeling (research.md, data-dictionary.md)**
- Gather API documentation
- Explore existing code and implementations
- Define domain entities and data structures
- Document types, interfaces, and schemas
- **Update status.md**: Mark Phase 1 complete, Phase 2 in progress

**Phase 3: Architecture & Planning (architecture.md, plan.md)**
- Design the implementation approach
- Identify affected layers (Domain, Use Case, Infrastructure, Adapter)
- Document component architecture and data flows
- Create implementation plan with phases and deliverables
- List files to create/modify
- **Update status.md**: Mark Phase 2 complete, Phase 3 in progress

**Phase 4: Task Breakdown (tasks.md)**
- Break down work into concrete, testable tasks
- Define dependencies and critical path
- Estimate durations
- Set up quality gates
- **Update status.md**: Mark Phase 3 complete, Phase 4 in progress

**Phase 5: Implementation (code + implementation-notes.md)**
- Follow TDD (Red-Green-Refactor)
- Record decisions made during implementation
- Document edge cases and solutions
- Note performance optimizations
- Track deviations from plan
- **Update status.md**: After EACH task completion - MANDATORY

**Phase 6: Completion & Archival**
- Update product documentation
- Move spec to specs/archive/
- Capture lessons learned
- **Verify status.md**: Must show 100% completion before archiving

**MANDATORY**: Update `status.md` after completing each task or phase. This file is the single source of truth for progress tracking.

### Specs Workflow Rules

- **Create feature directory** before starting any new feature work
- **Update progressively** as understanding evolves - specs are living documents
- **Update status.md ALWAYS** after completing each task, phase, or milestone - this is MANDATORY
- **Reference from commits** - link to spec directory in commit messages
- **Archive completed** - move to `specs/archive/` when feature is fully implemented and stable
- **Version control** - specs are committed to the repository for team collaboration

**Critical Rule**: Every time you complete a task, update `status.md` immediately to reflect:
- Task completion status
- Phase progress percentage
- Any blockers or issues encountered
- Next steps

### Example Feature Development Flow

```bash
# 1. Initialize specs process (one time)
/prd-to-spec init

# 2. Create feature spec from PRD
/prd-to-spec new-spec docs/prd-gmail-send.md

# 3. Work through phases progressively
# - Fill in spec.md (requirements, goals)
# - UPDATE status.md: Mark Phase 1 complete
# - Research and create research.md, data-dictionary.md
# - UPDATE status.md: Mark Phase 2 complete
# - Design architecture.md, plan.md
# - UPDATE status.md: Mark Phase 3 complete
# - Break down tasks.md
# - UPDATE status.md: Mark Phase 4 complete

# 4. Implement following TDD workflow
# - Update implementation-notes.md as you go
# - **CRITICAL**: Update status.md after EACH task completion

# 5. Archive when complete and stable
# - Verify status.md shows 100% completion
/prd-to-spec archive-spec gmail-send-command
```
