# Tasks: gojira-tmux TUI Application

**Input**: Design documents from `/specs/001-gojira-tmux/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD is mandatory per constitution (AGENTS.md). Tests are included.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US8)
- Include exact file paths in descriptions

## Path Conventions

Single Go project with Clean Architecture:
- `cmd/gojira/` - Application entrypoint
- `internal/domain/` - Entities and ports
- `internal/usecase/` - Business logic
- `internal/adapter/` - External implementations
- `internal/infrastructure/tui/` - BubbleTea UI

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Initialize Go module with `go mod init github.com/stainedhead/gojira-tmux`
- [X] T002 Install dependencies per research.md (bubbletea, lipgloss, bubbles, go-jira, go-oidc, go-keyring, yaml.v3)
- [X] T003 [P] Create directory structure per plan.md (cmd/, internal/domain/, internal/usecase/, internal/adapter/, internal/infrastructure/tui/)
- [X] T004 [P] Configure golangci-lint with .golangci.yml

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story

**CRITICAL**: No user story work can begin until this phase is complete

### Domain Entities

- [X] T005 [P] Write tests for TeamMember entity in internal/domain/team_member_test.go
- [X] T006 [P] Implement TeamMember entity with validation in internal/domain/team_member.go
- [X] T007 [P] Write tests for Project entity in internal/domain/project_test.go
- [X] T008 [P] Implement Project entity with validation in internal/domain/project.go
- [X] T009 [P] Write tests for Comment entity in internal/domain/comment_test.go
- [X] T010 [P] Implement Comment entity in internal/domain/comment.go
- [X] T011 [P] Write tests for User entity in internal/domain/user_test.go
- [X] T012 [P] Implement User entity with session validation in internal/domain/user.go
- [X] T013 Write tests for Issue entity in internal/domain/issue_test.go
- [X] T014 Implement Issue entity in internal/domain/issue.go (depends on TeamMember, Comment)

### Port Interfaces

- [X] T015 Define port interfaces in internal/domain/ports.go (JiraPort, AuthPort, TokenStorePort, ConfigPort)

### Configuration

- [X] T016 [P] Write tests for Config loading in internal/adapter/config/config_test.go
- [X] T017 [P] Implement Config types in internal/adapter/config/types.go
- [X] T018 Implement Config loading with validation in internal/adapter/config/config.go

### TUI Foundation

- [X] T019 [P] Define TUI message types in internal/infrastructure/tui/messages.go
- [X] T020 [P] Define key bindings in internal/infrastructure/tui/keys.go
- [X] T021 [P] Define lipgloss styles in internal/infrastructure/tui/styles.go
- [X] T022 Define Screen enum and App model skeleton in internal/infrastructure/tui/app.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - First-Time API Token Setup (Priority: P1)

**Goal**: New user can enter and store Jira API token securely

**Independent Test**: Launch app with no stored credentials, verify setup flow completes with token stored in keychain

### Tests for User Story 1

- [X] T023 [P] [US1] Write tests for TokenStore adapter in internal/adapter/auth/token_store_test.go
- [X] T024 [P] [US1] Write tests for SetupToken use case in internal/usecase/setup_token_test.go

### Implementation for User Story 1

- [X] T025 [US1] Implement TokenStore adapter with keyring fallback in internal/adapter/auth/token_store.go
- [X] T026 [US1] Implement SetupToken use case in internal/usecase/setup_token.go
- [X] T027 [US1] Implement SetupScreen TUI component in internal/infrastructure/tui/setup_screen.go
- [X] T028 [US1] Integrate SetupScreen into App model screen routing in internal/infrastructure/tui/app.go

**Checkpoint**: User Story 1 fully functional - can store API token

---

## Phase 4: User Story 2 - Okta SSO Authentication (Priority: P1)

**Goal**: User authenticates via Okta OIDC to access team data

**Independent Test**: Launch app with stored token, complete Okta login, verify user email validated against team list

### Tests for User Story 2

- [X] T029 [P] [US2] Write tests for OAuth callback server in internal/adapter/auth/callback_test.go
- [X] T030 [P] [US2] Write tests for Okta adapter in internal/adapter/auth/okta_test.go
- [X] T031 [P] [US2] Write tests for Authenticate use case in internal/usecase/authenticate_test.go

### Implementation for User Story 2

- [X] T032 [US2] Implement OAuth callback HTTP server in internal/adapter/auth/callback.go
- [X] T033 [US2] Implement Okta OIDC adapter with PKCE in internal/adapter/auth/okta.go
- [X] T034 [US2] Implement Authenticate use case with team validation in internal/usecase/authenticate.go
- [X] T035 [US2] Implement LoginScreen TUI component in internal/infrastructure/tui/login_screen.go
- [X] T036 [US2] Integrate LoginScreen into App model screen routing in internal/infrastructure/tui/app.go

**Checkpoint**: User Stories 1 AND 2 functional - full auth flow works

---

## Phase 5: User Story 3 - View Ticket List (Priority: P1)

**Goal**: Authenticated user sees Jira tickets from configured projects

**Independent Test**: After authentication, verify tickets from configured projects appear with correct columns

### Tests for User Story 3

- [X] T037 [P] [US3] Write tests for JQL builder in internal/adapter/jira/search_test.go
- [X] T038 [P] [US3] Write tests for Jira client in internal/adapter/jira/client_test.go
- [X] T039 [P] [US3] Write tests for ListIssues use case in internal/usecase/list_issues_test.go

### Implementation for User Story 3

- [X] T040 [US3] Implement JQL query builder in internal/adapter/jira/search.go
- [X] T041 [US3] Implement Jira client adapter with Basic Auth in internal/adapter/jira/client.go
- [X] T042 [US3] Implement ListIssues use case in internal/usecase/list_issues.go
- [X] T043 [US3] Implement TicketsTable component in internal/infrastructure/tui/tickets_table.go
- [X] T044 [US3] Implement MainScreen with table and loading state in internal/infrastructure/tui/main_screen.go
- [X] T045 [US3] Integrate MainScreen into App model in internal/infrastructure/tui/app.go
- [X] T046 [US3] Implement refresh ('r' key) functionality in MainScreen

**Checkpoint**: Core MVP complete - user can view tickets

---

## Phase 6: User Story 4 - Filter by Team Member (Priority: P2)

**Goal**: User filters tickets by selecting team member from dropdown

**Independent Test**: Select team member from dropdown, verify only their assigned tickets appear

### Tests for User Story 4

- [X] T047 [P] [US4] Write tests for member filter in internal/usecase/list_issues_test.go (extend existing)

### Implementation for User Story 4

- [X] T048 [US4] Implement FilterBar component with member dropdown in internal/infrastructure/tui/filter_bar.go
- [X] T049 [US4] Update ListIssues use case to filter by assignee email in internal/usecase/list_issues.go
- [X] T050 [US4] Integrate FilterBar into MainScreen in internal/infrastructure/tui/main_screen.go

**Checkpoint**: Member filtering works

---

## Phase 7: User Story 5 - Filter by Project (Priority: P2)

**Goal**: User filters tickets by selecting project from dropdown

**Independent Test**: Select project from dropdown, verify only tickets from that project appear

### Tests for User Story 5

- [X] T051 [P] [US5] Write tests for project filter in internal/usecase/list_issues_test.go (extend existing)

### Implementation for User Story 5

- [X] T052 [US5] Add project dropdown to FilterBar in internal/infrastructure/tui/filter_bar.go
- [X] T053 [US5] Update JQL builder for project filter in internal/adapter/jira/search.go
- [X] T054 [US5] Update ListIssues use case for project filtering in internal/usecase/list_issues.go

**Checkpoint**: Project filtering works

---

## Phase 8: User Story 6 - Filter by Status (Priority: P2)

**Goal**: User filters tickets by selecting status from dropdown

**Independent Test**: Select status from dropdown, verify only tickets in that status appear

### Tests for User Story 6

- [X] T055 [P] [US6] Write tests for status filter and combined filters in internal/usecase/list_issues_test.go (extend existing)

### Implementation for User Story 6

- [X] T056 [US6] Add status dropdown to FilterBar in internal/infrastructure/tui/filter_bar.go
- [X] T057 [US6] Update JQL builder for status filter with mapping in internal/adapter/jira/search.go
- [X] T058 [US6] Update ListIssues use case for combined filtering (AND logic) in internal/usecase/list_issues.go

**Checkpoint**: All filters work with combined AND logic

---

## Phase 9: User Story 7 - View Ticket Details (Priority: P3)

**Goal**: User sees full ticket details when selecting a row

**Independent Test**: Select ticket row, verify properties and comments panels populate correctly

### Tests for User Story 7

- [X] T059 [P] [US7] Write tests for GetIssueDetails use case in internal/usecase/get_issue_details_test.go

### Implementation for User Story 7

- [X] T060 [US7] Implement GetIssueDetails use case in internal/usecase/get_issue_details.go
- [X] T061 [US7] Implement PropertiesPanel component in internal/infrastructure/tui/properties_panel.go
- [X] T062 [US7] Implement CommentsPanel component in internal/infrastructure/tui/comments_panel.go
- [X] T063 [US7] Integrate detail panels into MainScreen with Tab focus cycling in internal/infrastructure/tui/main_screen.go

**Checkpoint**: Ticket details display works

---

## Phase 10: User Story 8 - Attention Indicators (Priority: P3)

**Goal**: Visual indicators highlight tickets needing attention

**Independent Test**: Verify red/yellow dots appear for tickets matching stale/no-due-date criteria

### Tests for User Story 8

- [X] T064 [P] [US8] Write tests for NeedsAttention and IsStale methods in internal/domain/issue_test.go (extend existing)

### Implementation for User Story 8

- [X] T065 [US8] Implement NeedsAttention, IsStale, LastAssigneeComment methods in internal/domain/issue.go
- [X] T066 [US8] Update TicketsTable to render attention indicators in internal/infrastructure/tui/tickets_table.go

**Checkpoint**: All user stories complete

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Final integration, CI/CD, documentation

- [X] T067 Create application entrypoint with dependency wiring in cmd/gojira/main.go
- [X] T068 [P] Create GitHub Actions CI workflow in .github/workflows/ci.yml
- [X] T069 [P] Create GitHub Actions release workflow in .github/workflows/release.yml
- [X] T070 Run quickstart.md validation (manual test of full flow)
- [X] T071 Update README.md with any changes from implementation
- [X] T072 Final lint check and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phase 3-10)**: All depend on Foundational phase completion
  - US1 → US2 → US3 (sequential for auth flow)
  - US4, US5, US6 can run in parallel after US3
  - US7, US8 can run in parallel after US3
- **Polish (Phase 11)**: Depends on all user stories complete

### User Story Dependencies

```
Phase 1: Setup
    ↓
Phase 2: Foundational (BLOCKS ALL)
    ↓
Phase 3: US1 (Token Setup)
    ↓
Phase 4: US2 (Okta Auth)
    ↓
Phase 5: US3 (View Tickets) ← MVP MILESTONE
    ↓
    ├── Phase 6: US4 (Filter Member) ─┐
    ├── Phase 7: US5 (Filter Project) ├── Can run in parallel
    ├── Phase 8: US6 (Filter Status) ─┘
    ↓
    ├── Phase 9: US7 (Details) ────────┐
    └── Phase 10: US8 (Indicators) ────┴── Can run in parallel
    ↓
Phase 11: Polish
```

### Within Each User Story

1. Tests MUST be written and FAIL before implementation
2. Models before services
3. Services before TUI components
4. TUI components before integration

### Parallel Opportunities

**Phase 2 (Foundational)**:
- T005-T012: All entity tests/implementations can run in parallel
- T016-T017: Config tests and types can run in parallel
- T019-T021: TUI foundation files can run in parallel

**Phase 5+ (After MVP)**:
- US4, US5, US6 can be implemented by different developers
- US7, US8 can be implemented by different developers

---

## Parallel Example: Phase 2 Entities

```bash
# Launch all entity tests together:
Task: "Write tests for TeamMember entity in internal/domain/team_member_test.go"
Task: "Write tests for Project entity in internal/domain/project_test.go"
Task: "Write tests for Comment entity in internal/domain/comment_test.go"
Task: "Write tests for User entity in internal/domain/user_test.go"

# Launch all entity implementations together:
Task: "Implement TeamMember entity in internal/domain/team_member.go"
Task: "Implement Project entity in internal/domain/project.go"
Task: "Implement Comment entity in internal/domain/comment.go"
Task: "Implement User entity in internal/domain/user.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1-3)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all)
3. Complete Phase 3: US1 - Token Setup
4. Complete Phase 4: US2 - Okta Auth
5. Complete Phase 5: US3 - View Tickets
6. **STOP and VALIDATE**: MVP complete, all core flows work
7. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Token storage works
3. Add US2 → Auth flow works
4. Add US3 → **MVP! Users can view tickets**
5. Add US4-6 → Filtering works
6. Add US7-8 → Details and indicators work
7. Polish → Production ready

### Parallel Team Strategy

With 3 developers after Phase 5:

- Developer A: US4 (Member Filter)
- Developer B: US5 (Project Filter)
- Developer C: US6 (Status Filter)

Then:

- Developer A: US7 (Details)
- Developer B: US8 (Indicators)
- Developer C: Polish/CI

---

## Summary

| Phase | Tasks | Story |
|-------|-------|-------|
| Setup | 4 | - |
| Foundational | 18 | - |
| US1 | 6 | Token Setup |
| US2 | 8 | Okta Auth |
| US3 | 10 | View Tickets |
| US4 | 4 | Filter Member |
| US5 | 4 | Filter Project |
| US6 | 4 | Filter Status |
| US7 | 5 | View Details |
| US8 | 3 | Indicators |
| Polish | 6 | - |
| **Total** | **72** | |

**MVP Scope**: Phases 1-5 (46 tasks) delivers Token Setup, Okta Auth, View Tickets

**Parallel Opportunities**: 32 tasks marked [P]

**Independent Tests per Story**: Each story has defined acceptance criteria

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- TDD mandatory: tests must fail before implementation
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
