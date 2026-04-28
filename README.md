# Symmate Runtime

A privacy-first agent coordination runtime, designed for edge deployment.

## Directory Structure

```
cmd/symmated/       Entry point and HTTP server
internal/
  config/           Configuration loading
  logger/           Logging setup (zerolog)
pkg/
  version/          Version information
```

## Building

```bash
make build          # Build binary to bin/symmated
make test           # Run tests
make lint           # Run linter
make fmt            # Format code
make clean          # Clean build artifacts
```

## Running

```bash
./bin/symmated      # Start server on default port 8080
```

Configure via environment:
- `SYMMATE_PORT` — HTTP server port (default: 8080)
- `SYMMATE_HOST` — HTTP server host (default: 0.0.0.0)
- `LOG_FORMAT` — Set to "pretty" for console output (default: JSON)

Or via `config.toml`:
```toml
[server]
port = 8080
host = "0.0.0.0"
```

## Endpoints

- `GET /health` — Health check (returns `ok`)
- `GET /version` — Version info

## Next Steps

See `/mnt/KB/symmate-workers/` for the full project plan. Worker B handles storage and encryption.
