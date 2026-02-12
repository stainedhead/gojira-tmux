# Technical Details

## Architecture

### Clean Architecture

```mermaid
graph TB
    subgraph Presentation
        TUI[TUI Components]
    end

    subgraph Application
        UC[Use Cases]
    end

    subgraph Domain
        E[Entities]
        P[Ports/Interfaces]
    end

    subgraph Infrastructure
        JIRA[Jira Adapter]
        AUTH[Auth Adapter]
        CFG[Config Adapter]
    end

    TUI --> UC
    UC --> P
    JIRA --> P
    AUTH --> P
    CFG --> P
    P --> E
```

### Directory Structure

```
cmd/
  gojira/
    main.go                 # Application entrypoint

internal/
  domain/
    issue.go                # Issue entity
    project.go              # Project entity
    team_member.go          # TeamMember entity
    user.go                 # Authenticated user entity
    ports.go                # Interface definitions

  usecase/
    list_issues.go          # List/filter issues
    get_issue_details.go    # Get issue with comments
    authenticate.go         # Authentication orchestration

  adapter/
    auth/
      atlassian.go          # Atlassian API token validation
      token_store.go        # Keyring storage
    config/
      config.go             # YAML loading
      types.go              # Config structs
    jira/
      client.go             # go-jira wrapper
      search.go             # JQL query builder

  infrastructure/
    tui/
      app.go                # Root BubbleTea model
      setup_screen.go       # Email + API token setup UI
      main_screen.go        # Main ticket view
      filter_bar.go         # Filter dropdowns
      tickets_table.go      # Issue list table
      properties_panel.go   # Issue properties
      comments_panel.go     # Issue comments
      styles.go             # lipgloss styles
      keys.go               # Key bindings
      messages.go           # Message types

pkg/                        # Public reusable packages
```

## Authentication Architecture

### Atlassian API Token Authentication

```mermaid
sequenceDiagram
    participant TUI as gojira TUI
    participant Keyring as OS Keyring
    participant Jira as Jira API

    TUI->>Keyring: Check for API token
    alt No token
        TUI->>TUI: Show setup screen (email + token input)
        TUI->>Jira: GET /rest/api/3/myself (validate token)
        Jira->>TUI: 200 OK (email confirmed)
        TUI->>Keyring: Store API token
    end

    TUI->>Keyring: Load Jira API token
    TUI->>Jira: API calls (Basic Auth)
    Jira->>TUI: Issue data
```

### Jira API Authentication

```
Authorization: Basic base64(email:api_token)
```

Token validation is performed against `/rest/api/3/myself` which returns the authenticated user's email, confirming the token is valid.

### Secure Storage

| Platform | Storage Backend |
|----------|-----------------|
| macOS | Keychain |
| Linux | Secret Service (libsecret) |
| Windows | Credential Manager |

**Stored Keys:**
- `gojira-jira-token` - Jira API token

## Jira API Integration

### Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/rest/api/3/myself` | GET | Token validation |
| `/rest/api/3/search/jql` | GET | JQL queries (cursor-based pagination via `nextPageToken`) |
| `/rest/api/3/issue/{key}` | GET | Issue details (with ADF description) |

### JQL Query Patterns

**Filter by assignee:**
```
assignee = "{email}" ORDER BY updated DESC
```

**Filter by project:**
```
project = "{key}" ORDER BY updated DESC
```

**Filter by status:**
```
status = "{status}" ORDER BY updated DESC
```

**Combined filter:**
```
assignee = "{email}" AND project = "{key}" AND status = "{status}" ORDER BY updated DESC
```

### Status Mappings

| UI Filter | JQL Status |
|-----------|------------|
| All | (no filter) |
| Open | `"Open"` |
| Ready | `"Ready for Development"` |
| In Test | `"In Test"` |
| Done | `"Done"` |

## BubbleTea Component Architecture

### Component Hierarchy

```mermaid
graph TD
    App[App Model]
    App --> Setup[SetupScreen]
    App --> Main[MainScreen]

    Setup --> EmailInput[EmailInput]
    Setup --> TokenInput[TokenInput]

    Main --> FilterBar[FilterBar]
    Main --> TicketsTable[TicketsTable]
    Main --> DetailView[DetailView]

    FilterBar --> ProjectSelect
    FilterBar --> MemberSelect
    FilterBar --> StatusSelect

    DetailView --> PropertiesPanel
    DetailView --> CommentsPanel
```

### Message Flow

```mermaid
flowchart LR
    User -->|Key Press| Update
    Update -->|State Change| Model
    Model -->|Render| View
    View -->|Display| User

    Update -->|Cmd| Runtime
    Runtime -->|Msg| Update
```

### State Management

Each screen maintains its own model state:

```go
type MainScreenModel struct {
    filterBar      FilterBarModel
    ticketsTable   TicketsTableModel
    detailView     DetailViewModel
    selectedIssue  *domain.Issue
    issues         []domain.Issue
    loading        bool
    err            error
}
```

## Data Entities

### Issue

```go
type Issue struct {
    Key         string
    Summary     string
    Description string
    Status      string
    Priority    string
    Assignee    *TeamMember
    Reporter    *TeamMember
    DueDate     *time.Time
    Created     time.Time
    Updated     time.Time
    Sprint      string
    Epic        string
    Labels      []string
    StoryPoints int
    Comments    []Comment
}
```

### TeamMember

```go
type TeamMember struct {
    Name  string
    Email string
    Alias string // optional short alias for filtering
}
```

Team members support an optional `Alias` field for disambiguation. The `MatchesIdentifier()` method resolves identifiers with 4-priority matching: exact alias, exact name, case-insensitive alias, case-insensitive name. `DisplayName()` returns "Name (Alias)" when an alias is set.

### Comment

```go
type Comment struct {
    ID        string
    Author    string
    Body      string
    Created   time.Time
}
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `charmbracelet/bubbletea` | TUI framework |
| `charmbracelet/lipgloss` | Styling |
| `charmbracelet/bubbles` | UI components |
| `andygrunwald/go-jira` | Jira REST client |
| `zalando/go-keyring` | Secure storage |
| `gopkg.in/yaml.v3` | Config parsing |

## Related Documentation

- [Product Summary](./product-summary.md) - High-level vision
- [Product Details](./product-details.md) - UI specifications
- [AGENTS.md](../AGENTS.md) - Development practices
