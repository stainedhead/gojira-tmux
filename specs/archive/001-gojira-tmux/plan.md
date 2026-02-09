# Implementation Plan: gojira-tmux TUI Application

**Branch**: `001-gojira-tmux` | **Date**: 2026-01-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-gojira-tmux/spec.md`

## Summary

Team-based Jira TUI viewer with dual authentication (Okta OIDC for user identity + Jira API tokens for API access). Clean Architecture with BubbleTea TUI framework. Features: secure token storage, team filtering, attention indicators, keyboard navigation.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: charmbracelet/bubbletea, charmbracelet/lipgloss, charmbracelet/bubbles, andygrunwald/go-jira, coreos/go-oidc/v3, zalando/go-keyring
**Storage**: OS Keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager)
**Testing**: go test with table-driven tests, interface mocking
**Target Platform**: macOS, Linux, Windows (cross-platform CLI)
**Project Type**: single
**Performance Goals**: Ticket list load <3 seconds, details update <1 second
**Constraints**: 100 ticket limit per query, 8-hour Okta session persistence
**Scale/Scope**: Team-based filtering, read-only Jira access, ~5 screens

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| TDD Mandatory | PASS | All components designed with testable interfaces |
| Clean Architecture | PASS | Domain/usecase/adapter/infrastructure separation |
| Code Quality Gates | PASS | CI with tests, lint, multi-platform build |
| CI/CD Pipeline | PASS | GitHub Actions configured |

## Project Structure

### Documentation (this feature)

```text
specs/001-gojira-tmux/
├── plan.md              # This file
├── research.md          # Technology decisions (COMPLETE)
├── data-model.md        # Entity definitions (COMPLETE)
├── quickstart.md        # Developer onboarding (COMPLETE)
├── contracts/           # Interface definitions (COMPLETE)
│   ├── README.md
│   └── ports.go
└── tasks.md             # Implementation tasks (via /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
  gojira/
    main.go                 # Application entrypoint

internal/
  domain/
    issue.go                # Issue entity
    project.go              # Project entity
    team_member.go          # TeamMember entity
    user.go                 # Authenticated user entity
    comment.go              # Comment entity
    ports.go                # Interface definitions

  usecase/
    list_issues.go          # List/filter issues use case
    get_issue_details.go    # Get issue with comments
    authenticate.go         # Authentication orchestration
    setup_token.go          # First-time API token setup

  adapter/
    auth/
      okta.go               # Okta OIDC implementation
      token_store.go        # Keyring storage adapter
      callback.go           # OAuth callback HTTP server
    config/
      config.go             # YAML configuration loading
      types.go              # Config structs
    jira/
      client.go             # go-jira wrapper
      search.go             # JQL query builder

  infrastructure/
    tui/
      app.go                # Root BubbleTea model (screen router)
      setup_screen.go       # Token setup UI
      login_screen.go       # Okta login UI
      main_screen.go        # Main ticket view
      filter_bar.go         # Filter dropdowns
      tickets_table.go      # Issue list table
      properties_panel.go   # Issue properties
      comments_panel.go     # Issue comments
      styles.go             # lipgloss styles
      keys.go               # Key bindings
```

**Structure Decision**: Single Go project with Clean Architecture layers. Domain entities have no external dependencies. Adapters implement domain ports. TUI is in infrastructure layer.

## Implementation Priorities

Based on spec user story priorities:

| Priority | User Stories | Key Components |
|----------|-------------|----------------|
| P1 | Token Setup, Okta Auth, Ticket List | domain/*, adapter/auth/*, adapter/jira/*, tui/setup_screen, tui/login_screen |
| P2 | Filter by Member/Project/Status | tui/filter_bar, usecase/list_issues |
| P3 | Ticket Details, Attention Indicators | tui/properties_panel, tui/comments_panel |

## Phase Outputs

### Phase 0: Research (COMPLETE)

- [x] [research.md](./research.md) - Technology decisions documented
  - Okta OIDC with PKCE
  - BubbleTea screen routing patterns
  - go-keyring cross-platform storage
  - JQL query building
  - go-jira authentication

### Phase 1: Design (COMPLETE)

- [x] [data-model.md](./data-model.md) - Entity definitions with validation
- [x] [contracts/](./contracts/) - Port interface definitions
- [x] [quickstart.md](./quickstart.md) - Developer onboarding guide

### Phase 2: Tasks (NEXT)

Run `/speckit.tasks` to generate implementation tasks.

## Complexity Tracking

No constitution violations. All gates pass.

## Next Steps

1. Run `/speckit.tasks` to generate implementation tasks
2. Begin implementation following TDD workflow
3. Start with P1 priorities: Domain entities, Config, Token storage
