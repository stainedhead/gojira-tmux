# gojira-tmux

A terminal-based Jira viewer for development teams. View and filter tickets directly from your terminal.

```
┌─────────────────────────────────────────────────────────────────────────┐
│  [Project ▼]          [Member ▼]           [Status ▼]                   │
├─────────────────────────────────────────────────────────────────────────┤
│ 🔴│ PROJ-123 │ Fix login bug      │ Open    │ J.Doe    │ High │ 01/15  │
│ 🟡│ PROJ-124 │ Add dark mode      │ Open    │ J.Smith  │ Med  │  --    │
│   │ PROJ-125 │ Update docs        │ Ready   │ B.Wilson │ Low  │ 01/20  │
├─────────────────────────────────────────────────────────────────────────┤
│ PROPERTIES                        │ COMMENTS                            │
│ Reporter: Admin User              │ [2024-01-10] J.Smith: "Updated..."  │
│ Sprint:   Sprint 23               │ [2024-01-08] J.Doe: "Found the..."  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Features

- **Team filtering** - Filter tickets by team member, project, or status
- **Attention indicators** - Red/yellow dots flag issues needing attention
- **Keyboard navigation** - Full TUI with vim-style keys
- **Okta SSO** - Enterprise authentication via Okta OIDC
- **Secure storage** - API tokens stored in OS keychain

## Installation

### From GitHub Releases

Download the latest binary for your platform from [Releases](https://github.com/stainedhead/gojira-tmux/releases).

### From Source

```bash
go install github.com/stainedhead/gojira-tmux/cmd/gojira@latest
```

## Quick Start

1. **Create configuration file** (`config.yaml`):

```yaml
jira:
  url: "https://your-company.atlassian.net"
  username: "your-email@company.com"

okta:
  issuer: "https://your-company.okta.com/oauth2/default"
  client_id: "your-okta-client-id"
  callback_port: 8080
  scopes: ["openid", "profile", "email"]

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
```

2. **Set environment variables**:

```bash
export JIRA_API_TOKEN="your-jira-api-token"
export OKTA_CLIENT_SECRET="your-okta-client-secret"
```

3. **Run**:

```bash
gojira
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `Tab` | Switch focus between panels |
| `Shift+Tab` | Reverse switch focus |
| `Enter` | View issue details |
| `f` | Focus filter bar |
| `r` | Refresh |
| `Esc` | Close details/cancel |
| `q` | Quit |

## Documentation

| Document | Description |
|----------|-------------|
| [Product Summary](./documentation/product-summary.md) | Vision and features |
| [Product Details](./documentation/product-details.md) | UI and workflows |
| [Technical Details](./documentation/technical-details.md) | Architecture and APIs |

## Development

```bash
# Build
go build -o bin/gojira ./cmd/gojira

# Test
go test ./...

# Lint
golangci-lint run
```

See [AGENTS.md](./AGENTS.md) for development practices and guidelines.

## License

MIT
