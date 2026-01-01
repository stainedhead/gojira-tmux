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
| `GET /rest/api/2/search` | JQL queries filtered by assignee email |
| `GET /rest/api/2/project` | List and validate projects |
| `GET /rest/api/2/issue/{issueKey}` | Retrieve issue details |

### Configuration Management

**Format**: YAML
**Package**: gopkg.in/yaml.v3

Configuration file stores all non-sensitive settings. Sensitive credentials (Okta client secret, Jira API token) stored via environment variables and OS-native secure storage.

## Configuration Structure

```yaml
# config.yaml

jira:
  url: "https://company.atlassian.net"
  username: "service-account@company.com"
  # API token via JIRA_API_TOKEN env var or secure storage

okta:
  issuer: "https://company.okta.com/oauth2/default"
  client_id: "YOUR_OKTA_CLIENT_ID"
  # client_secret via OKTA_CLIENT_SECRET env var
  callback_port: 8080
  scopes:
    - "openid"
    - "profile"
    - "email"

projects:
  - key: "PROJ1"
    name: "Project One"
  - key: "PROJ2"
    name: "Project Two"
  # Minimum 1 project required

team:
  - name: "John Doe"
    email: "john.doe@company.com"
  - name: "Jane Smith"
    email: "jane.smith@company.com"
  - name: "Bob Wilson"
    email: "bob.wilson@company.com"
```

### Configuration Fields

| Section | Field | Description |
|---------|-------|-------------|
| `jira.url` | string | Jira instance base URL |
| `jira.username` | string | Jira service account username/email |
| `okta.issuer` | string | Okta authorization server URL |
| `okta.client_id` | string | Okta OAuth application client ID |
| `okta.callback_port` | int | Local port for OAuth callback (default: 8080) |
| `okta.scopes` | []string | OIDC scopes for user authentication |
| `projects[].key` | string | Jira project key (e.g., "PROJ") |
| `projects[].name` | string | Display name for TUI |
| `team[].name` | string | Team member display name (used in TUI selection) |
| `team[].email` | string | Email for Jira assignee filtering |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `JIRA_API_TOKEN` | Jira API token for REST API authentication |
| `OKTA_CLIENT_SECRET` | Okta OAuth client secret (required for OIDC flow) |

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
      okta.go         # Okta OIDC flow implementation
      token_store.go  # Secure credential storage (keyring)
      callback.go     # Local HTTP server for OAuth callback
    config/
      config.go       # YAML config loading
      types.go        # Config struct definitions
    jira/
      client.go       # go-jira wrapper implementing JiraClient interface
      search.go       # JQL query building
```

## Authentication

### Dual Authentication Architecture

gojira-tmux uses a dual authentication approach:

1. **User Authentication**: OAuth 2.0/OIDC via Okta to verify user identity
2. **API Authentication**: Jira API tokens for REST API access

This approach avoids registering the application with Jira while leveraging enterprise SSO through Okta.

### Why This Approach

| Concern | Solution |
|---------|----------|
| No Jira app registration required | Use Jira API tokens instead of Jira OAuth |
| Enterprise SSO integration | Okta OIDC for user authentication |
| Secure credential management | API tokens stored in OS keyring |
| User access control | Okta validates user identity; Jira permissions based on token |

### Authentication Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   gojira    │     │   Browser   │     │    Okta     │     │    Jira     │
│    TUI      │     │             │     │             │     │    API      │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │                   │
       │ 1. Start login    │                   │                   │
       ├──────────────────>│                   │                   │
       │   (open browser)  │                   │                   │
       │                   │ 2. OIDC auth      │                   │
       │                   ├──────────────────>│                   │
       │                   │                   │                   │
       │                   │ 3. User login     │                   │
       │                   │   (SSO/MFA)       │                   │
       │                   │<──────────────────┤                   │
       │                   │                   │                   │
       │                   │ 4. Redirect +code │                   │
       │ 5. Callback       │<──────────────────┤                   │
       │<──────────────────┤                   │                   │
       │  (localhost:8080) │                   │                   │
       │                   │                   │                   │
       │ 6. Exchange code  │                   │                   │
       ├───────────────────────────────────────>│                   │
       │                   │                   │                   │
       │ 7. ID token +     │                   │                   │
       │    user info      │                   │                   │
       │<───────────────────────────────────────┤                   │
       │                   │                   │                   │
       │ 8. Validate user email against team config                │
       │                   │                   │                   │
       │ 9. Load Jira API token from secure storage                │
       │                   │                   │                   │
       │ 10. API calls with Basic Auth (username + API token)      │
       ├───────────────────────────────────────────────────────────>│
       │                   │                   │                   │
       │ 11. Jira data     │                   │                   │
       │<───────────────────────────────────────────────────────────┤
       │                   │                   │                   │
```

### User Authentication (Okta OIDC)

**Purpose**: Verify user identity via corporate SSO

**Go Package**: [coreos/go-oidc](https://github.com/coreos/go-oidc)

| Component | Description |
|-----------|-------------|
| Protocol | OpenID Connect (OAuth 2.0 + identity layer) |
| Flow | Authorization Code Flow with PKCE |
| Callback | Local HTTP server on configurable port |
| Validation | ID token signature + claims verification |

**Required OIDC Scopes**:
| Scope | Purpose |
|-------|---------|
| `openid` | Required for OIDC |
| `profile` | User name information |
| `email` | User email for team validation |

### API Authentication (Jira API Token)

**Purpose**: Authenticate REST API calls to Jira

Jira API tokens use HTTP Basic Authentication:
- Username: Jira account email
- Password: API token (generated in Atlassian account settings)

**Request Header**:
```
Authorization: Basic base64(username:api_token)
```

**Generating Jira API Token**:
1. Log in to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Name the token (e.g., "gojira-tmux")
4. Copy and store securely

### Login Screen Layout

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│                           ╔═══════════════════╗                         │
│                           ║    gojira-tmux    ║                         │
│                           ╚═══════════════════╝                         │
│                                                                         │
│                      Jira Team Ticket Viewer                            │
│                                                                         │
│         ┌─────────────────────────────────────────────┐                 │
│         │                                             │                 │
│         │   Status: Not authenticated                 │                 │
│         │                                             │                 │
│         │   Okta:  https://company.okta.com           │                 │
│         │   Jira:  https://company.atlassian.net      │                 │
│         │                                             │                 │
│         └─────────────────────────────────────────────┘                 │
│                                                                         │
│                      [ Login with Okta ]                                │
│                                                                         │
│         ─────────────────────────────────────────────                   │
│         Press Enter to open browser for authentication                  │
│         Press 'q' to quit                                               │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Awaiting Callback Screen

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│                           ╔═══════════════════╗                         │
│                           ║    gojira-tmux    ║                         │
│                           ╚═══════════════════╝                         │
│                                                                         │
│                      Authenticating via Okta...                         │
│                                                                         │
│         ┌─────────────────────────────────────────────┐                 │
│         │                                             │                 │
│         │   ◐ Waiting for authentication...           │                 │
│         │                                             │                 │
│         │   A browser window has been opened.         │                 │
│         │   Please complete the Okta login process.   │                 │
│         │                                             │                 │
│         │   Listening on: http://localhost:8080       │                 │
│         │                                             │                 │
│         └─────────────────────────────────────────────┘                 │
│                                                                         │
│         ─────────────────────────────────────────────                   │
│         Press 'c' to cancel authentication                              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Authentication States

```
┌──────────────┐    Login     ┌──────────────┐   Okta OK   ┌──────────────┐
│    Logged    │   Initiated  │   Awaiting   │  + Jira    │  Authenticated│
│     Out      ├─────────────>│   Callback   ├────────────>│    (Main UI) │
└──────────────┘              └───────┬──────┘             └───────┬──────┘
       ▲                              │                            │
       │                              │ Timeout/                   │ Session
       │                              │ Cancel/                    │ Expired
       │                              │ Invalid User               │
       │                              ▼                            │
       │                       ┌──────────────┐                    │
       │                       │    Error     │                    │
       │                       │   Display    │                    │
       └───────────────────────┴──────────────┘<───────────────────┘
```

### User Validation

After successful Okta authentication:
1. Extract email from ID token claims
2. Verify email exists in configured `team` list
3. If not in team list, display error and deny access

This ensures only configured team members can use the application.

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
| `gojira-okta-refresh` | Okta refresh token (optional, for silent re-auth) |

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
| Unauthorized access | Okta SSO validates corporate identity |
| Team access control | Email whitelist in config |
| Token rotation | Users can regenerate tokens in Atlassian account |
| Session hijacking | Okta handles session security with MFA support |

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
internal/adapter/tui/
  app.go              # Main application model (router between screens)
  setup_screen.go     # First-time Jira API token setup
  login_screen.go     # Okta login UI
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
│   └── TokenInput
├── LoginScreen
│   ├── StatusDisplay
│   └── OktaLoginButton
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
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Startup   │────>│   Setup     │────>│   Login     │────>│    Main     │
│   (check    │     │  (if no     │     │   (Okta)    │     │   Screen    │
│   keyring)  │     │  API token) │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                                       │                   │
       │ API token exists                      │                   │ Logout
       └───────────────────────────────────────┘                   │
                                                                   │
       ┌───────────────────────────────────────────────────────────┘
       ▼
┌─────────────┐
│   Login     │
│   Screen    │
└─────────────┘
```
