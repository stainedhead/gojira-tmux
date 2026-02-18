# Post-Deploy Bug Fixes PRD

**Date:** 2026-02-18
**Status:** Draft
**Scope:** Filter persistence, status filter correctness, filter selection UX

---

## Overview

Three issues were identified after initial deployment of gojira-tmux. This PRD covers the investigation and resolution of each, providing enough context for implementation without requiring codebase archaeology.

---

## Issue 1: Filter Settings Lost on Restart

### Problem

Filter selections (person, project, status) are held in-memory only within the `MainScreen` struct. On restart, all filters reset to `-All-`. There is no mechanism to persist or restore them.

### Current State

- `FilterBar` holds state in `memberIdx`, `projectIdx`, `statusIdx` (indices, not values)
- `domain.IssueFilter` carries string values: `Project`, `Assignee`, `Status`
- `MainScreen.filter` holds the current `IssueFilter` but is never persisted
- The `domain.Config` struct has no filter fields
- The config adapter (`internal/adapter/config/config.go`) only reads config — there is no `Save()` method

### Config File Locations

Resolved at startup in `cmd/gojira/main.go` (lines 75-99):
1. `./config.yaml` or `./config.yml` in working directory
2. `~/.config/gojira/config.yaml` (XDG standard)
3. Default: `./config.yaml`

### Requirements

1. **Add filter state to domain config model** (`internal/domain/ports.go`)
   ```go
   // New section in Config struct
   type FilterState struct {
       Project  string `yaml:"project,omitempty"`
       Assignee string `yaml:"assignee,omitempty"`
       Status   string `yaml:"status,omitempty"`
   }

   // Add to Config struct:
   LastFilter FilterState `yaml:"last_filter,omitempty"`
   ```

2. **Implement `Save()` on the config adapter** (`internal/adapter/config/config.go`)
   - Must write YAML back to the same file path used to load
   - Must preserve all existing config fields (Atlassian, Projects, Team)
   - Must be safe to call concurrently (use a mutex or channel-based write queue)

3. **Update `ConfigLoader` port** (`internal/domain/ports.go`) to include a Save method:
   ```go
   type ConfigLoader interface {
       Load() (*Config, error)
       Save(config *Config) error
   }
   ```

4. **Wire save into the TUI** (`internal/infrastructure/tui/main_screen.go`)
   - On `FilterChangedMsg`, after updating `MainScreen.filter`, call `configLoader.Save()` with the updated config (including the new filter state)
   - Save must be **non-blocking** — fire a `tea.Cmd` that writes to disk and returns a `FilterSavedMsg` (or silent error msg)
   - On error, display a brief status message but do not crash

5. **Restore filter state on startup** (`cmd/gojira/main.go` and/or `main_screen.go`)
   - After loading config, pass `cfg.LastFilter` to `MainScreen` initialisation
   - Call `filterBar.SetFilter()` with the restored filter before first render

6. **Validation constraint**: `LastFilter` values must be validated against the current config's project list and team list on load. If a saved project key no longer exists in the config, silently reset that filter to `-All-`.

---

## Issue 2: Status Filters Return No Results

### Problem

Selecting any status other than `-All-` (e.g. "Open", "Ready", "In Test", "Done") returns zero rows. The JQL is being generated and sent to Jira, but the status names in the query likely do not match the actual workflow status names configured in the Jira project.

### Current State

Status names are **hardcoded** in `filter_bar.go` (line 56):
```go
statuses := []string{"All", "Open", "Ready", "In Test", "Done"}
```

The JQL builder maps these to Jira status names in `internal/adapter/jira/search.go` (`MapStatus`, lines 101-117):

| UI Label | JQL status value |
|----------|-----------------|
| `Open`   | `"Open"` |
| `Ready`  | `"Ready for Development"` |
| `In Test`| `"In Test"` |
| `Done`   | `"Done"` |

The resulting JQL: `status = "Open"` (quoted, via `escapeJQL`).

### Root Cause Analysis

There are two likely causes that should both be investigated:

**Cause A: Status names don't match the Jira instance's workflow**
Jira status names are configurable per-project. "Open" in one instance may be "To Do" or "Backlog" in another. The hardcoded mappings were guesses and may simply not match.

**Cause B: JQL syntax issue with quoting**
The `escapeJQL` function wraps all values in double quotes: `status = "Open"`. If the status name contains no spaces or special characters, this is valid JQL. However, certain Jira Cloud configurations may behave differently. The JQL `status in ("Open")` form is also valid and worth testing.

### Investigation Steps (for implementer)

1. Add a **debug/diagnostic mode**: log the full JQL string being sent (to stderr or a log file) so it can be manually tested in Jira's issue navigator
2. Fetch available statuses from Jira API: `GET /rest/api/3/status` — this returns all statuses with their exact names as Jira knows them
3. Compare returned status names against the hardcoded mappings

### Requirements

1. **Fetch statuses dynamically from Jira at startup**
   - Add a `ListStatuses() ([]string, error)` method to the Jira client port (`internal/domain/ports.go`, `JiraClient` interface)
   - Implement it in `internal/adapter/jira/client.go` calling `GET /rest/api/3/status`
   - Pass the live status list to `FilterBar` instead of the hardcoded slice

2. **Fall back to a configurable status list**
   - If the API call fails, fall back to statuses defined in `config.yaml` under a new optional `statuses` key
   - If no config key and API fails, fall back to current hardcoded list with a warning log

3. **Remove the `MapStatus` translation layer**
   - The UI labels and Jira status names should be the same thing once populated from the API
   - `buildStatusCondition` should use the status string directly (still quoted in JQL)
   - Delete `MapStatus` function and its tests after migration

4. **Config option for status override** (fallback):
   ```yaml
   # config.yaml
   statuses:         # Optional. If set, used instead of fetching from Jira API.
     - "To Do"
     - "In Progress"
     - "In Review"
     - "Done"
   ```

5. **Update `domain.Config`** to include optional `Statuses []string`

6. **Update FilterBar initialisation** to accept `[]string` of statuses (already accepts members and projects as slices — same pattern)

7. **JQL quoting**: switch from `status = "X"` to `status in ("X")` form — more robust and consistent with multi-value future extension

---

## Issue 3: Filter Selection Has No Visual List

### Problem

Selecting a filter value requires blindly pressing `up`/`down` to cycle through options one at a time, with no visibility of the full option set. Users with many team members or projects have no way to see the list or jump to a specific entry.

### Current State

- `FilterBar` renders as a single horizontal row: `Member: -All-   Project: -All-   Status: All`
- Focused dropdown shows brackets and underline: `[Open]`
- Help text shows: `(←/→ switch, ↑/↓ change)`
- Interaction: `up`/`down` or `enter` cycles options one at a time — no overlay, no list view

### Requirements

1. **Dropdown overlay component**
   - When a filter is focused and the user presses `enter` or `space`, open a vertical list overlay anchored below the filter label
   - The overlay lists all options for that filter, one per line
   - The currently selected option is highlighted
   - Keyboard navigation: `up`/`down` move cursor within the list, `enter` confirms selection and closes overlay, `esc` cancels and closes overlay without changing selection
   - The overlay must not break the BubbleTea rendering model — use an overlay rendered on top of the table via `lipgloss` z-layering or by composing the view strings with the overlay inserted at the right position

2. **Visual design**
   - Overlay is a bordered box using `lipgloss` styles consistent with the app's existing palette
   - Selected item uses a highlight background (match table row highlight style)
   - If the list is longer than ~10 items, add a scroll indicator (`▲`/`▼` at top/bottom)
   - Overlay width: fit longest option + 2 padding chars

3. **State machine**
   - `FilterBar` gains a `dropdownOpen bool` field
   - When dropdown is open, all key events are captured by the dropdown handler, not the outer filter bar
   - Closing the dropdown returns focus to the filter bar (not to the main table)

4. **Retain cycling shortcut**
   - Keep existing `up`/`down` cycling behaviour when dropdown is **not** open — it's fast for single-step changes
   - Only opening the dropdown (`enter`/`space`) activates the list view

5. **Accessibility**: filter bar help text updates to show `(enter: open list, ↑/↓ cycle)` when no dropdown is open, and `(↑/↓ navigate, enter: select, esc: cancel)` when dropdown is open

---

## Affected Files Summary

| File | Issue | Change Type |
|------|-------|-------------|
| `internal/domain/ports.go` | 1, 2 | Add `FilterState`, `Statuses`, update `ConfigLoader`, `JiraClient` interfaces |
| `internal/adapter/config/config.go` | 1 | Add `Save()` method |
| `internal/adapter/jira/client.go` | 2 | Add `ListStatuses()` method |
| `internal/adapter/jira/search.go` | 2 | Remove `MapStatus`, update `buildStatusCondition`, switch to `status in (...)` |
| `internal/adapter/jira/search_test.go` | 2 | Update JQL tests |
| `internal/infrastructure/tui/filter_bar.go` | 2, 3 | Accept dynamic statuses, add dropdown overlay |
| `internal/infrastructure/tui/main_screen.go` | 1, 2 | Restore filter on startup, save on change, pass statuses to filter bar |
| `cmd/gojira/main.go` | 1, 2 | Pass config loader to TUI, fetch statuses, restore filter |
| `config.example.yaml` | 2 | Document optional `statuses` key |

---

## Implementation Order

Implement in this order to avoid circular refactoring:

1. **Issue 2 first** — fix status fetch/dynamic statuses; this changes the `JiraClient` port and `FilterBar` constructor signature which both Issues 1 and 3 depend on
2. **Issue 1** — add config save; requires no UI changes, pure adapter + wiring work
3. **Issue 3** — dropdown overlay; pure UI work, no domain or adapter changes

---

## Testing Requirements (TDD)

Each issue must follow red-green-refactor:

- **Issue 1**: Unit test `config.Save()`, test that `FilterChangedMsg` triggers save cmd, test filter restore on model init
- **Issue 2**: Test `ListStatuses()` with mocked HTTP response, test that `FilterBar` uses provided status list, update JQL tests to use `status in (...)` form, integration test with fixture Jira responses
- **Issue 3**: Test `FilterBar` state machine transitions (closed → open → select → closed), test dropdown renders correct options, test `esc` reverts selection

---

## Out of Scope

- Adding new filter dimensions (sprint, epic, label) — future work
- Saved filter presets / named filters — future work
- Multi-select status filtering (e.g. "Open OR In Progress") — future work
