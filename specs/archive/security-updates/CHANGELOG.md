# Changelog

## [Unreleased] - 2026-02-12

### Breaking Changes

- **Configuration format**: The `jira:` and `atlassian:` config sections have been merged into a single `atlassian:` section. See [MIGRATION-v3.md](./MIGRATION-v3.md) for upgrade instructions.

### Changed

- **Jira REST API v3**: Migrated all API calls from deprecated v2 endpoints to v3:
  - Search: `/rest/api/2/search` → `/rest/api/3/search/jql`
  - Auth: `/rest/api/2/myself` → `/rest/api/3/myself`
  - Issue details: `/rest/api/2/issue/{key}` → `/rest/api/3/issue/{key}`
- **Cursor-based pagination**: Search results now use `nextPageToken` instead of offset-based `startAt` pagination
- **ADF support**: Issue descriptions and comment bodies are now parsed from Atlassian Document Format (ADF) and converted to plain text for TUI display

### Removed

- `jira:` config section (consolidated into `atlassian:`)
- `jira.username` config field (was already unused)

### Fixed

- HTTP 410 "API has been removed" errors caused by Atlassian's removal of v2 search endpoint

## [0.1.0] - 2026-02-10

### Added

- Atlassian API token authentication (replacing Okta OAuth)
- Team member aliases for disambiguation
- Setup screen for first-time token entry
- Secure token storage via OS keychain
- Issue search with JQL filtering by project, assignee, and status
- Issue detail view with properties and comments panels
- Keyboard-driven TUI with vim-style navigation
- Attention indicators (red/yellow dots) for issues needing action
- Multi-project support
- YAML-based configuration
