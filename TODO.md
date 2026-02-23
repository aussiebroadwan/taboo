# TODO

This project hasn't been touched in a while. I'm more experienced with
certain things and have a different mentality in project structure and
design. This was made when I wanted to play with websockets, protobufs
and building my own HTML5 game/render engine. It was fun at the time
but I now want to build a version 2 which is as minimal setup.

## Version 2 Scope

### In Scope

- [x] New project structure (cmd/, internal/, pkg/, sdk/) *(feat/project-structure)*
- [x] Convert WebSockets + Protobufs to Server Side Events (SSE) *(feat/sse-api)*
- [x] SSE heartbeat events (15s default, configurable) *(feat/sse-api)*
- [x] CLI subcommands: `taboo serve`, `taboo version` *(feat/project-structure)*
- [x] CLI subcommands: `taboo migrate`, `taboo verify` *(feat/middleware-cli-sdk)*
- [x] Backend serves frontend static files *(feat/static-frontend)*
- [x] SQLite database with store interface pattern (for future driver support) *(feat/project-structure)*
- [x] Justfile for task running *(feat/project-structure)*
- [x] Config file support (YAML, startup-only) *(feat/project-structure)*
- [x] Environment variable overrides for config (TABOO_* prefix) *(feat/project-structure)*
- [x] Logging levels (config + CLI flag) *(feat/project-structure)*
- [x] Health endpoints: `GET /livez`, `GET /readyz` (DB check) *(feat/project-structure)*
- [x] Health endpoints: game engine check in /readyz *(feat/static-frontend)*
- [x] API versioning (`/api/v1/` prefix) *(feat/sse-api)*
- [x] Standard error response format (SDK DTOs + httpx error helpers) *(feat/sse-api)*
- [x] CORS middleware (env-aware defaults, configurable origins) *(feat/middleware-cli-sdk)*
- [x] X-Request-ID in responses *(feat/project-structure)*
- [x] Graceful shutdown (SIGTERM/SIGINT, drain connections) *(feat/project-structure)*
- [x] Request timeouts (context deadlines) *(feat/middleware-cli-sdk)*
- [x] gzip response compression middleware *(feat/middleware-cli-sdk)*
- [x] Per-IP rate limiting middleware *(feat/middleware-cli-sdk)*
- [x] Go SDK for API consumers (REST + SSE client with typed events and auto-reconnection) *(feat/middleware-cli-sdk)*
- [x] GitHub workflows: Docker image build, Linux amd64/arm64 binary builds *(feat/ci-workflows)*
- [x] GitHub workflows: tests, golangci-lint *(feat/static-frontend)*
- [x] Unit and integration tests for API and SDK *(feat/middleware-cli-sdk)*
- [x] Minimal frontend changes for SSE compatibility *(feat/static-frontend)*
- [ ] Documentation: systemd deployment, Caddy reverse proxy, config.example.yaml

### Out of Scope (Deferred)

- [ ] Postgres driver (deferred until Kubernetes deployment needed)
- [x] Frontend redesign: TypeScript + DOM/CSS rewrite *(feat/frontend-redesign)*
- [ ] Docker Stack removal (evaluate after v2 backend is stable)

## Project Structure

```
.
+--- cmd/taboo/main.go          # Entrypoint
+--- internal/
|    +--- app/                  # Application start, CLI flags, subcommands
|    +--- domain/               # Domain models for database interaction
|    +--- config/               # Config definition, parsing, validation
|    +--- http/                 # API handlers using SDK DTOs, calls services
|    +--- service/              # Business logic, invoked by HTTP, uses domain models
|    +--- store/
|         +--- store.go         # Interface for database operations
|         +--- drivers/
|              +--- sqlite/
|                   +--- gen/        # sqlc generated code
|                   +--- migrations/ # up/down migration files
|                   +--- queries/    # SQL query files
|                   +--- store.go    # Interface implementation
+--- sdk/                       # Go client library with DTOs
+--- pkg/
|    +--- httpx/                # Middleware (CORS, gzip, rate limit, timeout), SSE, error helpers
|    +--- slogx/                # Logger creation, HTTP logging middleware, attribute helpers
+--- frontend/
```

## SSE Event Design

Replace WebSocket + Protobuf with typed SSE events:

```
event: game:state
data: {"game_id": 123, "picks": [1, 5, 12], "next_game": "2024-01-01T12:00:00Z"}

event: game:pick
data: {"pick": 42}

event: game:complete
data: {"game_id": 123}

event: game:heartbeat
data: {}
```

## Config File

Supports YAML (preferred) or JSON. Detected by file extension. Startup-only (no hot-reload).
Environment variables can override config values (prefixed with `TABOO_`).

Options:
- Environment (dev/prod) - affects CORS defaults and logging
- Listen host/port
- SSL/TLS files (optional)
- CORS (allowed_origins, max_age)
- Request timeout
- Rate limiting
- Pick timing, wait timing
- Database selection (sqlite for now)
- Log level, format, output

See `config.example.yaml` for reference.

## Backend Data Flow

1. DTO models in API endpoint
2. API endpoint converts DTOs to service function params
3. Services operate on domain models (same as database/store)
4. Services process data, update state, return errors or models
5. API endpoint converts domain models/errors back to DTOs for response

## SSE Handling

Game service manages SSE sessions. Clients register to receive events.
The service broadcasts picks and state changes to all registered sessions.

## Health Endpoints

- `GET /livez` - Liveness probe, returns 200 if process is running
- `GET /readyz` - Readiness probe, checks:
  - Database connectivity (ping)
  - Game engine goroutine is running

## Justfile Targets

```
just build      # Build the binary
just test       # Run tests
just lint       # Run golangci-lint
just generate   # Run sqlc generate
just fmt        # Format code (go fmt)
just dev        # Run server with dev config (text logs, debug level)
just verify     # Validate config file (alias for taboo verify)
```

## CLI Subcommands

```
taboo serve      # Start the server
taboo migrate    # Database migration commands (up, down, status)
taboo verify     # Validate config file against rules (lint-style output)
taboo version    # Print version and exit
```

### Config Verification Rules

The `verify` subcommand validates config files with lint-style output (error, warn, info):

**Error** (invalid config, will fail to start):
- Missing required fields (e.g., `server.host`)
- Invalid duration format (e.g., `timeout: "abc"`)
- Invalid enum values (e.g., `env: "staging"` when only dev/prod allowed)
- Database path not writable (sqlite)
- TLS cert/key files don't exist when specified

**Warn** (will start but may cause issues):
- `rate_limit: 0` (disabled) in prod
- CORS `allowed_origins: ["*"]` in prod
- `logging.level: debug` in prod (performance impact)
- Very short timeouts (`timeout < 5s`)
- Very long SSE heartbeat (`sse_heartbeat > 60s`)

**Info** (suggestions and best practices):
- Using default values (explicitly set recommended)
- `logging.format: text` in prod (json recommended for log aggregation)
- Discord credentials not configured (Activity support disabled)

Example output:
```
$ taboo verify -c config.yaml
config.yaml:7  error  env must be "dev" or "prod", got "staging"
config.yaml:31 warn   rate_limit is disabled in prod environment
config.yaml:54 info   logging.format using default value (json)

✗ 1 error, 1 warning, 1 info
```

### CLI Flags

Common flags available on all subcommands:

- `--config` / `-c` - Config file path (default: `./config.yaml`)
- `--log-level` - Override log level
- `-v` / `--verbose` - Shorthand for `--log-level=debug`

## pkg/slogx

Structured logging package providing logger creation, helpers, and HTTP middleware.

### Logger Creation

```go
logger := slogx.New(slogx.Config{
    Service: "taboo",
    Version: "1.0.0",
    Env:     "prod",      // "dev" or "prod" - dev adds source info
    Level:   "info",      // "debug", "info", "warn", "error", "silent"
    Format:  "json",      // "json" or "text"
    Output:  "stderr",    // "stdout" or "stderr"
})

// Simple alternative for basic usage
logger := slogx.NewSimple("info")
```

Features:
- Service metadata attached as default attributes (service, version, env)
- Source information in dev mode or debug level
- RFC3339 timestamp formatting
- Sets as default logger via `slog.SetDefault()`

### Attribute Helpers

```go
// Error returns an Attr for logging errors
slog.Info("operation failed", slogx.Error(err))

// Duration returns an Attr for logging durations
slog.Info("request completed", slogx.Duration(elapsed))
```

### HTTP Middleware

Logs HTTP requests with request tracking and contextual logging:

```go
mux := http.NewServeMux()
handler := slogx.HTTPMiddleware(logger)(mux)
http.ListenAndServe(":8080", handler)
```

Features:
- Generates or uses `X-Request-ID` header for request tracking
- Creates logger with request metadata (method, path, remote_addr)
- Attaches logger to context via `slogx.WithContext()` for downstream handlers
- Logs request completion with status code and duration_ms

### Context Logger

```go
// Attach logger to context
ctx := slogx.WithContext(ctx, logger)

// Retrieve logger from context in handlers
logger := slogx.FromContext(ctx)
logger.Info("handling request")
```

## pkg/httpx

HTTP utilities, middleware, and SSE helpers.

### Middleware

```go
// Middleware chain
handler := httpx.Chain(
    httpx.Recoverer,
    httpx.CORS(cfg.Server.CORS),
    httpx.Gzip(httpx.GzipConfig{
        SkipPaths: []string{"/api/v1/events"},  // Skip SSE
    }),
    httpx.RateLimit(100),       // per-IP requests per second
    httpx.Timeout(30*time.Second),
    slogx.HTTPMiddleware(logger),
)(mux)
```

### Gzip

Compresses responses for bandwidth savings. **Skips SSE endpoints** because:
- Gzip buffers data, adding latency to real-time events
- Small event payloads (~30-50 bytes) have more overhead than savings
- Some proxies buffer gzipped streams, breaking real-time delivery

Useful for: game list API, static assets.

### CORS

Environment-aware defaults:
- **Dev mode (`env: dev`)**: Allows all origins (`*`)
- **Prod mode (`env: prod`)**: Requires explicit `allowed_origins` in config

Methods, headers, and exposed headers are fixed based on actual API endpoints.
Only `allowed_origins` and `max_age` are configurable.

### Error Responses

Standard error format (used by SDK DTOs):

```json
{
  "error": {
    "code": "GAME_NOT_FOUND",
    "message": "Game 123 not found"
  }
}
```

Error helper:

```go
// Returns JSON error response with appropriate status code
httpx.Error(w, httpx.ErrNotFound("Game", gameID))
httpx.Error(w, httpx.ErrBadRequest("invalid game_id format"))
httpx.Error(w, httpx.ErrInternal(err))
```

### SSE

```go
// SSE stream helper with heartbeat support
func (h *Handler) StreamEvents(w http.ResponseWriter, r *http.Request) {
    stream := httpx.NewSSEStream(w, httpx.SSEConfig{
        Heartbeat: 15 * time.Second,  // Send game:heartbeat every 15s (0 to disable)
    })
    defer stream.Close()

    for event := range h.gameService.Subscribe(r.Context()) {
        stream.Send(event.Type, event.Data)
    }
}
```

## API Versioning

All API endpoints use `/api/v1/` prefix:

```
GET  /api/v1/games              # List games (cursor-based pagination)
GET  /api/v1/games?cursor=abc&limit=20
GET  /api/v1/games/:id          # Get game by ID
GET  /api/v1/events             # SSE stream

GET  /livez                     # Liveness probe
GET  /readyz                    # Readiness probe
```

## Branch Strategy

Work on `v2` branch with feature branches (`feat/<part_name>`) merging into it.
Once complete, merge `v2` into `main`. Not backwards compatible with v1 as most
people have forgotten about this project so I dont really care.

---

## Notes

A big change is removing the websockets and protobufs. There were plans on
supporting the user to make requests to the api and interact and make bets. But
this has been scrapped. So now the endpoint's whole purpose is single direction
communications updates from the server. Because of this we can just use
server side events (SSE) which will drastically reduce the code. There was
no point in using protobufs, it was only used to learn how they work with a
server and client implementation. Now we have had our fun we can remove this.

This will need to be done in a new branch. Probably multiple where there is
a version 2 branch and then `feat/<part_name>` which will merge into the
version 2 branch before merging version 2 into main. I'm not looking for the
version 2 to be fully backwards compatible with the current main as there are
going to be alot of core fundamental changes.
