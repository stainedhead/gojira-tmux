# Jira REST API v3 Migration Guide

If you're upgrading gojira-tmux from a version that used the old `jira:` + `atlassian:` config format, this guide will walk you through the changes. The migration takes about 2 minutes.

## Why This Migration Is Needed

Atlassian removed the Jira REST API v2 search endpoint (`/rest/api/2/search`) in August 2025. If you're seeing errors like this:

```
Jira API error (status 410): {"errorMessages":["The requested API has been removed.
Please migrate to the /rest/api/3/search/jql API."]}
```

You need to update gojira-tmux to the latest version **and** update your config file.

## What Changed

### Configuration (You Need to Update This)

The `jira:` and `atlassian:` config sections have been merged into a single `atlassian:` section.

**Before:**
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
```

**After:**
```yaml
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"

projects:
  - key: "PROJ"
    name: "My Project"

team:
  - name: "John Doe"
    email: "john.doe@company.com"
```

The `projects:` and `team:` sections are unchanged.

### API Endpoints (Transparent to You)

Under the hood, gojira-tmux now uses Jira REST API v3 endpoints. This happens automatically — no action needed from you:

| What | Old (v2) | New (v3) |
|------|----------|----------|
| Issue search | `/rest/api/2/search` | `/rest/api/3/search/jql` |
| Token validation | `/rest/api/2/myself` | `/rest/api/3/myself` |
| Issue details | `/rest/api/2/issue/{key}` | `/rest/api/3/issue/{key}` |

### Other Improvements

- **ADF support**: Issue descriptions and comments now use Atlassian Document Format, automatically converted to plain text for the TUI
- **Cursor-based pagination**: Search results use the newer `nextPageToken` model instead of offset-based pagination

## Migration Steps

### 1. Back Up Your Config

```bash
cp config.yaml config.yaml.backup
```

### 2. Update Your Config File

Open `config.yaml` and make these changes:

1. Move the `url` from the `jira:` section into the `atlassian:` section
2. Remove the entire `jira:` section
3. If you had `custom_fields` under `jira:`, move them under `atlassian:`

**Example:**
```yaml
# Remove these lines:
# jira:
#   url: "https://your-company.atlassian.net"
#   custom_fields:
#     sprint: "customfield_10020"

# Your atlassian section should now look like this:
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"
  custom_fields:         # optional, move from old jira: section
    sprint: "customfield_10020"
    epic: "customfield_10014"
```

See `config.example.yaml` for a complete example.

### 3. Rebuild the Application

```bash
go build -o bin/gojira ./cmd/gojira
```

### 4. Run and Verify

```bash
./bin/gojira
```

You should see your tickets load without any 410 errors.

## Troubleshooting

### "atlassian.url is required"

Your config file still uses the old format. Move the URL from `jira:` to `atlassian:` as described above.

### "atlassian.email is required"

Add your email to the `atlassian:` section:
```yaml
atlassian:
  url: "https://your-company.atlassian.net"
  email: "your-email@company.com"
```

### "atlassian.url must start with https://"

Make sure your URL uses `https://`, not `http://`:
```yaml
atlassian:
  url: "https://your-company.atlassian.net"  # correct
  # url: "http://your-company.atlassian.net" # wrong
```

### HTTP 410 errors persist

You're running an old version of gojira-tmux. Rebuild from the latest source:
```bash
git pull
go build -o bin/gojira ./cmd/gojira
```

### No search results returned

- Check that your JQL queries are valid
- Verify your API token hasn't expired
- Confirm your Atlassian URL is correct

### Authentication fails (401/403)

- Regenerate your API token at https://id.atlassian.com/manage/api-tokens
- Make sure the email in your config matches the email on your Atlassian account
- Delete your stored token and re-enter it: restart gojira and follow the setup prompts

## FAQ

**Q: Do I need a new API token?**
No. Your existing API token works with the v3 API. Only your config file format needs to change.

**Q: Will my team members need to do anything?**
Each team member needs to update their own `config.yaml` file. The `team:` section format hasn't changed.

**Q: What if I have custom fields configured?**
Move them from `jira.custom_fields` to `atlassian.custom_fields`. The field IDs stay the same.

**Q: Can I keep the old config format?**
No. The old `jira:` section is no longer supported. The application will show a validation error on startup if it finds the old format.

**Q: What happened to the `jira.username` field?**
It was already unused and has been removed. Authentication uses `atlassian.email` with your API token.
