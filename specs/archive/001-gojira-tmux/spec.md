# Feature Specification: gojira-tmux TUI Application

**Feature Branch**: `001-gojira-tmux`
**Created**: 2026-01-01
**Status**: Draft
**Input**: User description: "Team-based Jira TUI viewer with Okta SSO and API token authentication"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - First-Time API Token Setup (Priority: P1)

A new user launches the application for the first time. They need to configure their Jira API token before they can access any Jira data. The application guides them through generating and storing their token securely.

**Why this priority**: Without a stored API token, no Jira data can be retrieved. This is a prerequisite for all other functionality.

**Independent Test**: Can be fully tested by launching the app with no stored credentials and verifying the setup flow completes successfully with token stored in system keychain.

**Acceptance Scenarios**:

1. **Given** no Jira API token exists in secure storage, **When** user launches the application, **Then** the setup screen is displayed with instructions for generating a token
2. **Given** user is on the setup screen, **When** user enters a valid API token and presses Enter, **Then** the token is stored securely in the system keychain
3. **Given** user enters an empty token, **When** user presses Enter, **Then** an error message is displayed and token is not stored
4. **Given** user is on the setup screen, **When** user presses Esc, **Then** the application exits without storing any token

---

### User Story 2 - Okta SSO Authentication (Priority: P1)

A user with a stored API token needs to authenticate via corporate Okta SSO. The application opens their browser for authentication and waits for the callback to complete the login flow.

**Why this priority**: User identity verification is required before accessing team-specific Jira data. Users cannot proceed to the main screen without valid Okta authentication.

**Independent Test**: Can be fully tested by launching the app with stored API token and completing Okta login flow, verifying user email is validated against team list.

**Acceptance Scenarios**:

1. **Given** API token exists and user is not authenticated, **When** user presses Enter on the login screen, **Then** browser opens to Okta authorization URL
2. **Given** browser is open to Okta, **When** user completes SSO login successfully, **Then** application receives callback and extracts user email from ID token
3. **Given** user email is extracted from Okta, **When** email matches an entry in configured team list, **Then** user is granted access to main screen
4. **Given** user email is extracted from Okta, **When** email does NOT match any entry in configured team list, **Then** error is displayed and access is denied
5. **Given** user is on awaiting callback screen, **When** user presses 'c', **Then** authentication is cancelled and user returns to login screen

---

### User Story 3 - View Ticket List (Priority: P1)

An authenticated user sees a list of Jira tickets from configured projects. The list displays key information at a glance and updates when filters change.

**Why this priority**: Viewing tickets is the core value proposition. Without this, the application provides no utility.

**Independent Test**: Can be fully tested by authenticating and verifying tickets from configured projects appear in the main table with correct columns.

**Acceptance Scenarios**:

1. **Given** user is authenticated, **When** main screen loads, **Then** tickets from all configured projects are displayed in the table
2. **Given** tickets are displayed, **When** user views the table, **Then** each row shows: indicator, key, summary, status, assignee, priority, due date, updated timestamp
3. **Given** tickets are displayed, **When** user presses 'r', **Then** ticket list is refreshed from Jira API
4. **Given** tickets are loading, **When** data is being fetched, **Then** loading indicator is displayed

---

### User Story 4 - Filter by Team Member (Priority: P2)

A user wants to see only tickets assigned to a specific team member. They select a team member from a dropdown and the ticket list updates accordingly.

**Why this priority**: Team-based filtering is a core differentiator but requires basic ticket viewing to work first.

**Independent Test**: Can be fully tested by selecting different team members from dropdown and verifying only their assigned tickets appear.

**Acceptance Scenarios**:

1. **Given** user is on main screen, **When** user opens the Member dropdown, **Then** "-All-" and all configured team member names are listed
2. **Given** Member dropdown is open, **When** user selects a team member, **Then** ticket list filters to show only tickets assigned to that member's email
3. **Given** Member filter is set, **When** user selects "-All-", **Then** ticket list shows tickets for all team members

---

### User Story 5 - Filter by Project (Priority: P2)

A user wants to see only tickets from a specific Jira project. They select a project from a dropdown and the ticket list updates accordingly.

**Why this priority**: Multi-project support enables use across different teams but requires basic ticket viewing to work first.

**Independent Test**: Can be fully tested by selecting different projects from dropdown and verifying only tickets from that project appear.

**Acceptance Scenarios**:

1. **Given** user is on main screen, **When** user opens the Project dropdown, **Then** "-All-" and all configured project names are listed
2. **Given** Project dropdown is open, **When** user selects a project, **Then** ticket list filters to show only tickets from that project
3. **Given** Project filter is set, **When** user selects "-All-", **Then** ticket list shows tickets from all configured projects

---

### User Story 6 - Filter by Status (Priority: P2)

A user wants to see tickets in a specific workflow status. They select a status from a dropdown and the ticket list updates accordingly.

**Why this priority**: Status filtering helps users focus on actionable items but requires basic ticket viewing to work first.

**Independent Test**: Can be fully tested by selecting different statuses and verifying only tickets in that status appear.

**Acceptance Scenarios**:

1. **Given** user is on main screen, **When** user opens the Status dropdown, **Then** options are: All, Open, Ready, In Test, Done
2. **Given** Status dropdown is open, **When** user selects "Open", **Then** ticket list filters to show only Open tickets
3. **Given** multiple filters are set, **When** user changes any filter, **Then** all active filters are combined (AND logic)

---

### User Story 7 - View Ticket Details (Priority: P3)

A user selects a ticket to see its full details including properties not shown in the main table and all comments.

**Why this priority**: Detailed view enhances usefulness but basic list viewing must work first.

**Independent Test**: Can be fully tested by selecting a ticket and verifying properties panel and comments panel populate correctly.

**Acceptance Scenarios**:

1. **Given** ticket list is displayed, **When** user selects a row using arrow keys or j/k, **Then** the row is highlighted
2. **Given** a ticket is selected, **When** selection changes, **Then** Properties panel updates to show: Reporter, Created, Description, Sprint, Epic, Labels, Story Points
3. **Given** a ticket is selected, **When** selection changes, **Then** Comments panel updates to show all comments sorted newest to oldest
4. **Given** user is viewing details, **When** user presses Tab, **Then** focus cycles between table, properties, and comments panels

---

### User Story 8 - Attention Indicators (Priority: P3)

A user can quickly identify tickets that need attention through visual indicators in the ticket list.

**Why this priority**: Indicators are a quality-of-life feature that requires core viewing functionality first.

**Independent Test**: Can be fully tested by verifying correct indicator colors appear for tickets matching the stale/missing-date criteria.

**Acceptance Scenarios**:

1. **Given** a ticket is Open AND assignee has not commented in 14+ days, **When** ticket is displayed, **Then** red dot indicator appears
2. **Given** a ticket is Open AND has no due date set, **When** ticket is displayed, **Then** yellow dot indicator appears
3. **Given** a ticket meets both conditions (no comment and no due date), **When** ticket is displayed, **Then** red dot takes precedence (more urgent)
4. **Given** a ticket is not Open OR does not meet any attention criteria, **When** ticket is displayed, **Then** no indicator appears

---

### Edge Cases

- What happens when Jira API returns an error (401 unauthorized, 500 server error)?
  - Display user-friendly error message and offer retry option
- What happens when no tickets match the current filters?
  - Display "No tickets found" message in table area
- What happens when Okta authentication times out?
  - Display timeout error after 5 minutes and return to login screen
- What happens when system keyring is unavailable (e.g., headless Linux)?
  - Fall back to secure file-based storage with user warning
- What happens when config.yaml is missing or malformed?
  - Display configuration error with guidance on required format

## Clarifications

### Session 2026-01-01

- Q: When a project has thousands of tickets, how many should the application display at once? → A: Display most recent 100 tickets, sorted by last updated
- Q: When a user closes and reopens the application, should they need to re-authenticate with Okta? → A: Persist Okta session for 8 hours (workday), then require re-auth

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST authenticate users via Okta OpenID Connect using Authorization Code Flow with PKCE
- **FR-001a**: System MUST persist Okta session for 8 hours; after expiration, user must re-authenticate
- **FR-002**: System MUST store Jira API tokens in OS-native secure storage (macOS Keychain, Linux Secret Service, Windows Credential Manager)
- **FR-003**: System MUST validate authenticated user email against configured team member list and deny access if not found
- **FR-004**: System MUST load and parse YAML configuration file containing Jira URL, Okta settings, projects, and team members
- **FR-005**: System MUST retrieve tickets from Jira using JQL queries filtered by project, assignee email, and status, limited to the most recent 100 tickets sorted by last updated
- **FR-006**: System MUST display ticket list with columns: indicator, key, summary, status, assignee, priority, due date, updated
- **FR-007**: System MUST display ticket properties panel showing: reporter, created, description, sprint, epic, labels, story points
- **FR-008**: System MUST display ticket comments panel showing all comments sorted newest to oldest
- **FR-009**: System MUST show red dot indicator for Open tickets with no assignee comment in 14+ days
- **FR-010**: System MUST show yellow dot indicator for Open tickets with no due date set
- **FR-011**: System MUST support keyboard navigation using arrow keys, j/k, Tab, Enter, Esc, q, r, and c

### Key Entities

- **Issue**: A Jira ticket with key, summary, description, status, priority, assignee, reporter, due date, created date, updated date, sprint, epic, labels, story points, and comments
- **TeamMember**: A team member with display name (for UI selection) and email (for Jira assignee filtering)
- **Project**: A Jira project with key (for JQL queries) and display name (for UI)
- **Comment**: A ticket comment with author name, body text, and created timestamp
- **User**: An authenticated user with email (from Okta ID token) used for team validation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users complete first-time API token setup in under 2 minutes
- **SC-002**: Users complete Okta authentication flow in under 30 seconds (excluding browser login time)
- **SC-003**: Ticket list loads and displays within 3 seconds of authentication completing
- **SC-004**: Users can filter to a specific team member's tickets in under 5 seconds
- **SC-005**: Ticket details (properties and comments) update within 1 second of row selection
- **SC-006**: 95% of authenticated users successfully view their team's tickets on first attempt
- **SC-007**: Users can identify tickets needing attention (red/yellow indicators) without reading full details
