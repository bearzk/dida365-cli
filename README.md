# Dida365 CLI

A command-line interface for [Dida365](https://dida365.com) (滴答清单 / TickTick) task management, designed for automation and scripting workflows.

## Features

- **Task Management**: Full CRUD operations for tasks (create, read, update, delete, complete)
- **Kanban Support**: List columns per project and move tasks between columns
- **Project Access**: List and view projects
- **JSON Output**: All commands output JSON for easy parsing with `jq` or other tools
- **Scripting-Friendly**: Clear exit codes and structured error messages
- **Secure**: Config file stored with user-only permissions

## Installation

### From Source

```bash
git clone https://github.com/bearzk/dida365-cli.git
cd dida365-cli
go build -o dida365 .
sudo mv dida365 /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/bearzk/dida365-cli@latest
```

## Getting Started

### 1. Obtain API Credentials

1. Visit the [Dida365 Developer Portal](https://developer.dida365.com) (or [TickTick Developer Portal](https://developer.ticktick.com))
2. Register an application to get `client_id` and `client_secret`
3. Set redirect URL to: `http://localhost:8080/callback`

### 2. Authenticate via OAuth2

The CLI handles the complete OAuth2 flow automatically:

```bash
# For Dida365 (China)
dida365 auth login \
  --client-id "your_client_id" \
  --client-secret "your_client_secret" \
  --service dida365

# For TickTick (International)
dida365 auth login \
  --client-id "your_client_id" \
  --client-secret "your_client_secret" \
  --service ticktick

# Use custom port if 8080 is occupied
dida365 auth login \
  --client-id "your_client_id" \
  --client-secret "your_client_secret" \
  --service dida365 \
  --port 9000
```

The CLI will:
1. Start a local callback server
2. Open your browser to the authorization page
3. Handle the callback and exchange the code for tokens
4. Save credentials securely to `~/.dida365/config.json`

### 3. Verify Authentication

```bash
dida365 auth status
```

## Usage

### Projects

**List all projects:**
```bash
dida365 project list
```

**Get project details:**
```bash
dida365 project get <project-id>
```

**List Kanban columns in a project:**
```bash
dida365 project columns <project-id>
```

**Get full project data (tasks, columns):**
```bash
dida365 project data <project-id>
```

### Tasks

**Create a task:**
```bash
dida365 task create \
  --title "Deploy to production" \
  --project-id "proj123" \
  --content "Deploy v2.0.0" \
  --due-date "2026-04-30" \
  --start-date "2026-05-01"
```

**List tasks in a project:**
```bash
dida365 task list proj123
```

**Get task details:**
```bash
dida365 task get task456 --project-id proj123
```

**Update a task:**
```bash
dida365 task update task456 \
  --project-id proj123 \
  --title "New title" \
  --content "Updated content" \
  --due-date "2026-05-01 18:30" \
  --start-date "2026-05-01 18:30"
```

**Complete a task:**
```bash
dida365 task complete task456 --project-id proj123
```

**Delete a task:**
```bash
dida365 task delete task456 --project-id proj123
```

**Move a task to a different Kanban column:**
```bash
# First, find the column IDs
dida365 project columns <project-id>

# Then move the task
dida365 task move task456 --project-id proj123 --column-id <column-id>
```

**Accepted `--start-date` and `--due-date` formats:**
```bash
# All-day (date only)
dida365 task create --title "Monthly summary" --project-id "proj123" \
  --start-date "2026-04-28" --due-date "2026-04-30"

# Specific local time (space separator)
dida365 task create --title "Release window" --project-id "proj123" \
  --start-date "2026-04-30 10:00" --due-date "2026-04-30 18:30"

# Specific local time (T separator)
dida365 task create --title "Release window" --project-id "proj123" \
  --start-date "2026-04-30T10:00" --due-date "2026-04-30T18:30"

# RFC3339 with timezone
dida365 task update task456 --project-id proj123 \
  --start-date "2026-04-30T09:00:00+08:00" --due-date "2026-04-30T23:59:59+08:00"
```

## Scripting Examples

### Extract specific fields with jq

```bash
# Get task ID from created task
TASK_ID=$(dida365 task create --title "Test" --project-id "proj123" | jq -r .id)

# Get all project names
dida365 project list | jq '.[].name'

# Filter incomplete tasks
dida365 task list proj123 | jq 'map(select(.status == 0))'
```

### Error handling in scripts

```bash
#!/bin/bash
set -e

# Check auth status
if ! dida365 auth status >/dev/null 2>&1; then
  echo "Error: Not authenticated. Run 'dida365 auth login' to authenticate."
  exit 1
fi

# Create task
TASK_JSON=$(dida365 task create --title "Automated task" --project-id "proj123")
if [ $? -ne 0 ]; then
  echo "Failed to create task"
  exit 1
fi

echo "Task created successfully"
echo "$TASK_JSON" | jq .
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Create deployment task
  run: |
    dida365 task create \
      --title "Deployment to ${{ github.ref_name }}" \
      --project-id "${{ secrets.DIDA365_PROJECT_ID }}" \
      --content "SHA: ${{ github.sha }}"
  env:
    DIDA365_CONFIG: ${{ secrets.DIDA365_CONFIG }}
```

## Exit Codes

- `0` - Success
- `1` - Configuration error (missing or invalid config)
- `2` - Authentication error (invalid token, OAuth2 flow failed)
- `3` - API error (resource not found, bad request, server error)
- `4` - Network error (connection timeout, DNS failure)
- `5` - Validation error (invalid arguments, missing flags)
- `6` - OAuth2 server error (callback server failed to start)

## JSON Output Format

### Success Response

Commands output the resource or array directly:

```json
{
  "id": "task123",
  "projectId": "proj456",
  "title": "Task title",
  "dueDate": "2026-04-30T15:59:59Z",
  "isAllDay": false,
  "status": 0,
  "sortOrder": 1
}
```

### Error Response

Errors are output to stderr:

```json
{
  "error": "human-readable message",
  "code": "ERROR_CODE"
}
```

## Configuration File

Location: `~/.dida365/config.json`

```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "access_token": "your_access_token",
  "token_expiry": "2026-08-21T14:30:00Z",
  "base_url": "https://dida365.com"
}
```

**Token Management:**
- Access tokens are long-lived (~6 months)
- Re-authenticate with `dida365 auth login` when your token expires

**Security:** The config file is created with `0600` permissions (user read/write only).

## Development

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/client -v
```

### Building

```bash
go build -o dida365 .
```

### Project Structure

```
dida365-cli/
├── cmd/              # Cobra commands
├── internal/
│   ├── client/       # API client
│   ├── config/       # Config management
│   └── models/       # Data models
├── main.go           # Entry point
└── README.md
```

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `go test ./...`
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Related Projects

- [Dida365 Official Website](https://dida365.com)
- [TickTick (International)](https://ticktick.com)
- [Dida365 API Documentation](https://developer.dida365.com)
