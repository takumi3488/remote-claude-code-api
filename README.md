# remote-claude-code-api

A Connect RPC server that wraps Claude Code's `--remote` mode. Accepts requests over HTTP/1.1 and h2c (HTTP/2 cleartext), passes the given prompt to the Claude CLI, and returns the result. Optionally accepts a GitHub repository, clones it into a temporary directory, and runs Claude inside it.

## Usage

### Run the server

```bash
PORT=8080 GITHUB_TOKEN=ghp_xxx go run .
```

### Run with Docker

```bash
docker build -t remote-claude-code-api .
docker run -p 8080:8080 \
  -e GITHUB_TOKEN=ghp_xxx \
  remote-claude-code-api
```

### Request examples

**Prompt only (no repository):**

```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello, Claude!"}'
```

**With a GitHub repository:**

```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Summarize this repository", "repository": "owner/repo"}'
```

**Using the Connect RPC procedure path directly:**

```bash
curl -X POST http://localhost:8080/claude.v1.ClaudeService/Run \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "repository": "owner/repo"}'
```

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Listening port | `8080` |
| `GITHUB_TOKEN` | GitHub Personal Access Token (required for private repositories) | none |

## API

### RunRequest

| Field | Type | Description |
|-------|------|-------------|
| `prompt` | string | Prompt passed to Claude |
| `repository` | string | GitHub repository in `owner/repo` format (optional) |

### RunResponse

| Field | Type | Description |
|-------|------|-------------|
| `output` | string | Claude's output |

## Error codes

| Code | Cause |
|------|-------|
| `invalid_argument` | `repository` is not in `owner/repo` format |
| `not_found` | `git clone` failed (repository not found or insufficient permissions) |
| `internal` | Claude CLI execution error |
