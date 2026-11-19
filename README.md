# Symmate Runtime

A privacy-first agent coordination runtime, designed for edge deployment.
One binary (`symmated`) serves one host's Symmate.

## Building

```bash
make build          # Build binary to bin/symmated
make test           # Run tests
make lint           # Run linter
make fmt            # Format code
make clean          # Clean build artifacts
```

Web apps (SvelteKit + Bun) live in `web/` — see `web/package.json` for `dev:*` / `build:*` scripts.

## Servers

| Port | Surface | Notes |
|---|---|---|
| 8080 | `/health`, `/version` | `SYMMATE_PORT` |
| 7842 | Visitor API: `/chat` (SSE), `/id`, `/llms.txt` | `SYMMATE_VISITOR_PORT` |
| 7841 | Owner API (loopback only) | used by `web/apps/operator` |
| 7843 | A2A endpoint + `/.well-known/agent-card.json` | |
| 7844 | MCP server (SSE + JSON-RPC) | |

## Required files

| Path | Purpose |
|---|---|
| Constitution (default `/mnt/KB/claude/_protocol/symmate-prompts/symmate-v1.md`, override with `SYMMATE_CONSTITUTION`) | §A of the system prompt. **Startup fails without it.** |
| `<host_dir>/context.md` | Host profile + §C of the system prompt |
| `<host_dir>/vertical.md` | Optional §B |

`host_dir` defaults to `/var/lib/symmate/default` (`SYMMATE_HOST_DIR`).

## Tool access

Tools come in two tiers:

- **Public** — `get_datetime`, `fetch_url`, `web_search`. Available to every caller.
- **Restricted** — integrations (GitHub today). **Owner only by default.**

Who gets what:

| Caller | Tools |
|---|---|
| Verified owner (phrase / device recognition) | everything |
| Registered user (sends a `user_token`) | public + the groups the owner toggled on for them |
| Anonymous visitor | public only |

Registered users are managed in the operator dashboard (**Users** page) or the owner API:

```
GET    /users                 master switch, available groups, users
PUT    /users/settings        {"enabled": true|false}   master switch (default off)
POST   /users                 {"name", "groups"}  → returns the token once
PATCH  /users/{id}            {"name"?, "enabled"?, "groups"?}
POST   /users/{id}/token      rotate token
DELETE /users/{id}
```

A user sends their token as `user_token` in the `/chat` body or as the `X-Symmate-User` header.
Each tool belongs to exactly one group (the prefix of its first tag), so every toggle unlocks a
concrete set of tools. Only groups that would unlock something are offered: a group with no
registered tools, or one fully blocked by the `tool_tags` allowlist (Policy page), doesn't appear.
The allowlist narrows everyone, including the owner, and is applied live.

## Integrations

### GitHub (read-only)

Enabled when `GITHUB_TOKEN` or `GITHUB_REPO` is set. Group: `engineering`.

| Env | Meaning |
|---|---|
| `GITHUB_TOKEN` | Fine-grained PAT; read-only access to contents, issues, pull requests is enough. Optional for public repos (60 req/h). |
| `GITHUB_REPO` | Default repo, `owner/name` |
| `GITHUB_API_URL` | API base for GitHub Enterprise (default `https://api.github.com`) |

Tools: `analyze_pr_risk`, `get_deploy_risk`, `get_cycle_time`, `list_pull_requests`, `list_issues`.
Risk scores are a transparent heuristic (sensitive paths, CI/dependency changes, size, test changes);
every signal behind a score is returned.

### Not yet connected

The other department tools (sales, finance, HR, legal, ...) return sample data and are **not registered**.
Set `SYMMATE_ENABLE_MOCK_TOOLS=1` to register them for demos; their output is labelled as demo data.

## Other configuration

| Env | Meaning |
|---|---|
| `SYMMATE_SLUG`, `SYMMATE_BASE_URL` | Host identity / base URL |
| `SYMMATE_{VISITOR,A2A,MCP}_PUBLIC_URL` | Public URLs advertised in the agent card |
| `HERMES_ENDPOINT`, `STAK_API_KEY` | Local LiteLLM (default `http://stak:4000/v1`) |
| `SYMMATE_SEARCH_URL` | Enables `web_search` |
| `SYGIL_URL` | A2A reputation checks |
| `POKER_ENDPOINT` | Visitor forcefield |
| `OWNER_CORS_ORIGIN` | CORS origin for the owner API |
| `LOG_FORMAT=pretty` | Console logs instead of JSON |
