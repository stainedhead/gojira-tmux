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
- **Team member aliases** - Short aliases for quick filtering (e.g., "JohnA" instead of "John Anderson")
- **Attention indicators** - Red/yellow dots flag issues needing attention
- **Keyboard navigation** - Full TUI with vim-style keys
- **Atlassian API tokens** - Simple, secure authentication via Atlassian API tokens
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

atlassian:
  email: "your-email@company.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
    alias: "JohnD"  # optional short alias for filtering
```

2. **Generate an Atlassian API token** at https://id.atlassian.com/manage/api-tokens

3. **Run** (the app will prompt for your token on first launch):

```bash
gojira
```

   Or set the token via environment variable:

```bash
export JIRA_API_TOKEN="your-jira-api-token"
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
