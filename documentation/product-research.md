# Product Research: gojira-tmux

## Product Vision

A team-based Jira ticket viewer and manager built as a TUI (Terminal User Interface) application. Provides developers with real-time access to Jira data directly from the terminal, enabling quick ticket browsing and filtering by team member without leaving the command line workflow.

## Core Features

### Team-Based Ticket Filtering
- Select team members by name in the TUI interface
- Filter Jira tickets using the member's email address via JQL queries
- View all tickets assigned to a specific team member

### Multi-Project Support
- Configure multiple Jira projects in a single config file
- Switch between projects within the TUI
- Minimum of one project required in configuration

### Terminal-Native Interface
- TUI built with BubbleTea framework
- Keyboard-driven navigation
- Works in any terminal environment (TMUX compatible)

## Technology Stack

### Jira REST API Integration

**Package**: [andygrunwald/go-jira](https://github.com/andygrunwald/go-jira)

The go-jira library provides a Go client for the Atlassian Jira REST API with the following capabilities:

- **Authentication**: API token-based authentication using Basic Auth scheme
- **IssueService**: Create, read, update issues
- **SearchService**: JQL-based queries for filtered ticket retrieval
- **ProjectService**: List and validate configured projects

**Key API Endpoints**:
| Endpoint | Purpose |
|----------|---------|
| `GET /rest/api/3/search/jql` | JQL queries filtered by assignee email (cursor-based pagination) |
| `GET /rest/api/3/issue/{issueKey}` | Retrieve issue details (ADF format descriptions) |
| `GET /rest/api/3/myself` | Token validation |

### Configuration Management

**Format**: YAML
**Package**: gopkg.in/yaml.v3

Configuration file stores all non-sensitive settings. The Jira API token is stored via environment variable or OS-native secure storage (keyring).

## Configuration Structure

```yaml
# config.yaml

atlassian:
  url: "https://company.atlassian.net"
  email: "your-email@company.com"
  custom_fields:  # optional
    sprint: "customfield_10020"
    epic: "customfield_10014"

projects:
  - key: "PROJ1"
    name: "Project One"
  - key: "PROJ2"
    name: "Project Two"
  # Minimum 1 project required

team:
  - name: "John Doe"
    email: "john.doe@company.com"
    alias: "JohnD"  # optional short alias
  - name: "Jane Smith"
    email: "jane.smith@company.com"
  - name: "Bob Wilson"
    email: "bob.wilson@company.com"
```

### Configuration Fields

| Section | Field | Description |
|---------|-------|-------------|
| `atlassian.url` | string | Jira instance base URL (must be HTTPS) |
| `atlassian.email` | string | Atlassian account email for API authentication |
| `atlassian.custom_fields` | object | Optional custom field ID mappings |
| `projects[].key` | string | Jira project key (e.g., "PROJ") |
| `projects[].name` | string | Display name for TUI |
| `team[].name` | string | Team member display name (used in TUI selection) |
| `team[].email` | string | Email for Jira assignee filtering |
| `team[].alias` | string | Optional short alias for quick filtering |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `JIRA_API_TOKEN` | Jira API token for REST API authentication |

## Architecture Mapping

The configuration and Jira integration map to the Clean Architecture layers defined in AGENTS.md:

```
internal/
  domain/
    issue.go          # Issue entity
    project.go        # Project entity
    team_member.go    # TeamMember entity
    user.go           # Authenticated user entity
    ports.go          # JiraClient, TokenStore, UserAuthenticator interfaces

  adapter/
    auth/
      atlassian.go    # Atlassian API token validation
      token_store.go  # Secure credential storage (keyring)
    config/
      config.go       # YAML config loading
      types.go        # Config struct definitions
    jira/
      client.go       # go-jira wrapper implementing JiraClient interface
      search.go       # JQL query building
```

## Authentication

### Atlassian API Token Authentication

gojira-tmux uses Atlassian API token authentication with HTTP Basic Auth:

- **Email**: From `atlassian.email` in config
- **Password**: API token (generated in Atlassian account settings, stored in OS keyring)

### Authentication Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   gojira    │     │  OS Keyring │     │    Jira     │
│    TUI      │     │             │     │   API v3    │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       │ 1. Check keyring  │                   │
       ├──────────────────>│                   │
       │                   │                   │
       │ No token?         │                   │
       │ Show setup screen │                   │
       │                   │                   │
       │ 2. Validate token │                   │
       ├───────────────────────────────────────>│
       │   GET /rest/api/3/myself              │
       │                   │                   │
       │ 3. 200 OK         │                   │
       │<───────────────────────────────────────┤
       │                   │                   │
       │ 4. Store token    │                   │
       ├──────────────────>│                   │
       │                   │                   │
       │ 5. API calls (Basic Auth)             │
       ├───────────────────────────────────────>│
       │                   │                   │
       │ 6. Jira data      │                   │
       │<───────────────────────────────────────┤
```

**Request Header**:
```
Authorization: Basic base64(email:api_token)
```

**Generating an API Token**:
1. Log in to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Name the token (e.g., "gojira-tmux")
4. Copy and store securely

### Authentication States

```
┌──────────────┐   No token   ┌──────────────┐   Token OK   ┌──────────────┐
│   Startup    ├─────────────>│    Setup     ├─────────────>│  Main Screen │
│  (check      │              │   Screen     │              │              │
│   keyring)   │              └──────────────┘              └──────────────┘
└──────┬───────┘
       │ Token exists
       └──────────────────────────────────────────────────>│  Main Screen │
                                                           └──────────────┘
```

### Secure Credential Storage

Credentials stored using OS-native secure storage:

| Platform | Storage |
|----------|---------|
| macOS | Keychain |
| Linux | Secret Service API (libsecret) |
| Windows | Windows Credential Manager |

**Stored Items**:
| Key | Value |
|-----|-------|
| `gojira-jira-token` | Jira API token |

**Go Package**: [zalando/go-keyring](https://github.com/zalando/go-keyring)

### First-Time Setup Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│                           ╔═══════════════════╗                         │
│                           ║    gojira-tmux    ║                         │
│                           ╚═══════════════════╝                         │
│                                                                         │
│                         First-Time Setup                                │
│                                                                         │
│         ┌─────────────────────────────────────────────┐                 │
│         │                                             │                 │
│         │   Jira API Token not configured.            │                 │
│         │                                             │                 │
│         │   To generate a token:                      │                 │
│         │   1. Visit id.atlassian.com/manage-profile  │                 │
│         │   2. Go to Security > API tokens            │                 │
│         │   3. Create and copy your token             │                 │
│         │                                             │                 │
│         │   Enter Jira API Token: _______________     │                 │
│         │                                             │                 │
│         └─────────────────────────────────────────────┘                 │
│                                                                         │
│         ─────────────────────────────────────────────                   │
│         Token will be stored securely in system keychain                │
│         Press Esc to cancel                                             │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Security Considerations

| Concern | Mitigation |
|---------|------------|
| API token exposure | Stored in OS keyring, never in config files |
| Team access control | Email whitelist in config |
| Token rotation | Users can regenerate tokens in Atlassian account |
| Token validation | Token verified against `/rest/api/3/myself` on setup |

## UI Layout Specification

### Screen Layout

```
┌─────────────────────────────────────────────────────────────────────────┐
│  [Project ▼]          [Member ▼]           [Status ▼]                   │
│   -All-                -All-                All                         │
├─────────────────────────────────────────────────────────────────────────┤
│ ● │ Key      │ Summary            │ Status  │ Assignee │ Pri  │ Due    │
│───┼──────────┼────────────────────┼─────────┼──────────┼──────┼────────│
│ 🔴│ PROJ-123 │ Fix login bug      │ Open    │ J.Doe    │ High │ 01/15  │
│ 🟡│ PROJ-124 │ Add dark mode      │ Open    │ J.Smith  │ Med  │  --    │
│   │ PROJ-125 │ Update docs        │ Ready   │ B.Wilson │ Low  │ 01/20  │
│   │ PROJ-126 │ API refactor       │ In Test │ J.Doe    │ High │ 01/18  │
├─────────────────────────────────────────────────────────────────────────┤
│ PROPERTIES                        │ COMMENTS                            │
│───────────────────────────────────│─────────────────────────────────────│
│ Reporter:    Admin User           │ [2024-01-10] J.Smith:               │
│ Created:     2024-01-05           │   "Updated the PR, ready for..."    │
│ Description:                      │                                     │
│   The login page has an issue     │ [2024-01-08] J.Doe:                 │
│   when users attempt to...        │   "Found the root cause..."         │
│ Sprint:      Sprint 23            │                                     │
│ Epic:        Authentication       │ [2024-01-05] Admin:                 │
│ Labels:      bug, urgent          │   "Created from support ticket"     │
│ Story Points: 5                   │                                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### Filter Bar Components

| Component | Options | Behavior |
|-----------|---------|----------|
| Project | `-All-` + configured projects | Filters tickets by project key |
| Member | `-All-` + configured team members | Filters by assignee email |
| Status | All, Open, Ready, In Test, Done | Filters by ticket status |

### Main Tickets Table

**Columns**:
| Column | Description |
|--------|-------------|
| Indicator | Attention flag (see below) |
| Key | Jira issue key (e.g., PROJ-123) |
| Summary | Issue title/summary |
| Status | Current workflow status |
| Assignee | Assigned team member name |
| Priority | Issue priority level |
| Due Date | Expected completion date |
| Updated | Last update timestamp |

**Attention Indicators**:
| Indicator | Condition |
|-----------|-----------|
| 🔴 Red dot | Issue is Open AND no comment by assignee in 14+ days |
| 🟡 Yellow dot | Issue is Open AND no due date set |
| (empty) | No attention required |

### Detail Panels

**Properties Panel (Bottom-Left)**:
Displays fields NOT shown in main table, one per row:
- Reporter
- Created date
- Description (multi-line)
- Sprint
- Epic Link
- Labels
- Story Points

**Comments Panel (Bottom-Right)**:
- All comments for selected issue
- Sorted newest to oldest
- Format: `[YYYY-MM-DD] Author: "comment text..."`

### Interaction Behaviors

| Action | Result |
|--------|--------|
| Change any filter | Refresh main table with new filter criteria |
| Select row in main table | Update both Properties and Comments panels |
| `↑`/`k` | Move selection up one row |
| `↓`/`j` | Move selection down one row |
| `Tab` | Cycle focus between panels |
| `q` | Quit application |
| `r` | Refresh current view |

## JQL Query Patterns

### Status Filter Mappings

| UI Filter | JQL Clause |
|-----------|------------|
| All | (no status filter) |
| Open | `status = "Open"` |
| Ready | `status = "Ready for Development"` |
| In Test | `status = "In Test"` |
| Done | `status = "Done"` |

### Query Examples

Filtering tickets by team member email:
```
assignee = "john.doe@company.com" AND project = "PROJ1" ORDER BY updated DESC
```

Filtering by project only:
```
project = "PROJ1" ORDER BY updated DESC
```

Combined filters (member + project + status):
```
assignee = "john.doe@company.com" AND project = "PROJ1" AND status = "Open" ORDER BY updated DESC
```

All projects, specific member, excluding Done:
```
assignee = "john.doe@company.com" AND status != "Done" ORDER BY updated DESC
```

## BubbleTea Component Architecture

```
internal/infrastructure/tui/
  app.go              # Main application model (router between screens)
  setup_screen.go     # First-time email + API token setup
  main_screen.go      # Main tickets view (after auth)
  filter_bar.go       # Project/Member/Status dropdowns
  tickets_table.go    # Main tickets list table
  properties_panel.go # Selected issue properties
  comments_panel.go   # Selected issue comments
  styles.go           # lipgloss styling definitions
  keys.go             # Keyboard binding definitions
```

### Component Hierarchy

```
App (root model)
├── SetupScreen (first-time only)
│   ├── EmailInput
│   └── TokenInput
└── MainScreen (after authentication)
    ├── FilterBar
    │   ├── ProjectSelect
    │   ├── MemberSelect
    │   └── StatusSelect
    ├── TicketsTable
    └── DetailView
        ├── PropertiesPanel
        └── CommentsPanel
```

### Screen Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Startup   │────>│   Setup     │────>│    Main     │
│   (check    │     │  (if no     │     │   Screen    │
│   keyring)  │     │  API token) │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       │ API token exists
       └────────────────────────────────>│    Main     │
                                         │   Screen    │
                                         └─────────────┘
```
