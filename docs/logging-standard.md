# Logging Standard

This document defines structured logging conventions for the project. Every log line should be useful to an operator reading it in production or to a developer debugging an issue. A log line that does not help someone understand what happened, where it happened, why it was logged, and how the system reached that state is noise.

## Core Principle: What, Where, Why, How

Every log call should answer four questions through the combination of its message and structured fields:

| Question | Answered By | Example |
|----------|-------------|---------|
| **What** happened? | The message string | `"Record created"`, `"Failed to persist record"` |
| **Where** in the system? | Fields identifying the component, endpoint, or resource | `resource_id`, `request_id`, `component` |
| **Why** was this logged? | The log level and contextual fields that explain the trigger | Level WARN + `error` field tells you something failed |
| **How** did we get here? | State fields showing the path or inputs that led to this point | `item_count`, `client_count`, `addr`, `method`, `path` |

The message says **what**. The fields say **where**, **why**, and **how**.

### Applying the Principle

```
// Good - answers all four questions
// What:  Record persistence failed
// Where: Record 42 in the engine component
// Why:   An error occurred during database write (WARN level)
// How:   The record had 20 items ready to persist
log.warn("Failed to persist record",
    resource_id: 42,
    component: "engine",
    item_count: 20,
    error: err,
)

// Bad - answers only "what", nothing else
log.warn("Failed to persist record", error: err)

// Bad - stuffs context into the message instead of fields
log.warn("Failed to persist record 42 with 20 items")
```

### Field Selection Guide

Not every log line needs every possible field. Include fields that help an operator act on this specific log line:

**Always include** the identity of the resource being operated on (`resource_id`, `request_id`). Without this, the log line cannot be correlated with anything.

**Include state fields when they explain how we got here.** If a resource fails to create, the item count and configuration explain what we were trying to do. If an HTTP request fails, the method and path explain what was requested.

```
// Resource failure - item_count and resource_id explain what we were processing
log.warn("Failed to create resource",
    resource_id: nextID,
    item_count: cfg.ItemCount,
    error: err,
)

// Server startup - addr and timeouts explain configuration
log.info("HTTP server started",
    addr: cfg.ListenAddr,
    read_timeout: cfg.ReadTimeout,
)
```

**Omit fields that add no diagnostic value.** If a field is always the same value within a code path, or if it cannot help narrow down a problem, leave it out. Prefer fewer meaningful fields over many redundant ones.

## Logging Library

Use a single structured logging library for all runtime output. Do not use `print`, `printf`, `console.log`, or any unstructured output mechanism for operational logging.

Choose a library that supports:
- Structured key-value fields (not just string interpolation)
- Log levels (ERROR, WARN, INFO, DEBUG at minimum)
- JSON or other machine-parseable output formats
- Typed field constructors where available

Once chosen, use that library exclusively throughout the project. Document the chosen library in the project's main configuration or contributing guide.

## Shared Helpers

Create shared helper utilities for common logging patterns such as error wrapping and duration formatting. These helpers ensure consistent field keys across the codebase.

Helpers should provide:

- **Error helper** -- wraps an error value under the standard `"error"` key
- **Duration helper** -- wraps a duration value under the standard `"duration"` key
- **Context propagation** -- retrieves or enriches a logger from a request context or scope

```
// Example helper usage
logger = loghelper.from_context(ctx)
ctx = loghelper.with(ctx, component: "engine")
```

Add additional helpers as needed for common field types. Keep them in a single shared package or module so every component uses the same field keys.

## Message Format

### Casing

Messages must start with an **uppercase letter**.

```
// Good
log.info("Record created", ...)
log.warn("Failed to broadcast event", ...)

// Bad
log.info("record created", ...)
log.warn("failed to broadcast event", ...)
```

### Punctuation

- No trailing period
- Use `...` only for messages indicating an ongoing action before a corresponding completion message

```
log.info("Shutting down server...")
// ... shutdown logic ...
log.info("Server shutdown complete")
```

### Phrasing

The message answers **what happened**. Keep it short and let the fields carry the rest.

- **Events:** Past tense or noun phrases -- `"Record created"`, `"Item revealed"`, `"Client connected"`
- **Failures:** `"Failed to <action>"` -- `"Failed to persist record"`, `"Failed to broadcast event"`
- **State changes:** Present participle only with `...` pairs -- `"Shutting down server..."` / `"Server shutdown complete"`

Do not put field values in the message string. The message is a static category; the fields are the instance data.

```
// Good - message is a static category, fields identify the instance
log.info("Processing complete",
    resource_id: resource.ID,
    item_count: 20,
)

// Bad - instance data baked into the message
log.info("Resource 42 complete with 20 items")
```

### Character Set

Messages, field keys, and string field values must use **ASCII only**. No unicode characters, emoji, or special symbols.

## Structured Fields

### Use Typed Constructors

Where the logging library provides typed field constructors, always use them. Do not use bare key-value positional arguments or untyped maps when a typed API is available.

```
// Good - typed, explicit
log.info("HTTP server started",
    addr: string("0.0.0.0:8080"),
)

// Bad - positional args, no type safety, easy to misorder
log.info("HTTP server started", "addr", "0.0.0.0:8080")
```

If the logging library does not provide typed constructors, use structured key-value pairs consistently and document the convention.

### Key Naming

Use **snake_case** for all field keys.

```
// Good
resource_id: 42
request_id: "abc-123"
item_count: 20
client_count: 5

// Bad
resourceId: 42
requestID: "abc-123"
itemCount: 20
```

Keys should be descriptive enough that they make sense when read without the surrounding code. Prefer `"listen_addr"` over `"addr"`, `"draw_duration"` over `"duration"` when the bare word is ambiguous.

**Exception:** The shared `error` and `duration` helpers produce the keys `"error"` and `"duration"` respectively. These are the only single-word keys permitted.

### Standard Field Keys

Use these universal keys consistently across the codebase. Standardised keys make it possible to search and filter logs across components.

#### Universal Keys

| Key | Type | Answers | Description |
|-----|------|---------|-------------|
| `request_id` | string | Where | Request correlation ID -- set by middleware |
| `component` | string | Where | System component: `engine`, `server`, `store` |
| `method` | string | How | HTTP method: `GET`, `POST`, etc. |
| `path` | string | How | HTTP request path |
| `status` | integer | How | HTTP response status code |
| `addr` | string | How | Listen address as `host:port` |
| `client_count` | integer | How | Number of connected clients |
| `bytes` | integer | How | Response body size |
| `version` | string | How | Application version string |
| `error` | error/string | Why | Error value (set via shared helper, never manually) |
| `duration` | duration | How | Elapsed time (set via shared helper, never manually) |

#### Domain Keys

Define domain-specific keys per project for your primary resources and workflows. Document them alongside the universal keys above. For example, a project might define:

- `resource_id` -- primary correlation ID for the domain entity
- `item_count` -- number of items being processed
- Fields specific to your domain model (e.g., `order_id`, `user_id`, `batch_size`)

Follow the same naming and typing conventions as the universal keys.

## Log Levels

The level answers **why this was logged** -- how urgent and what kind of event it represents.

### ERROR

A component cannot function. The system is degraded or shutting down. An operator must act.

Fields should include enough context to identify the failing component and what it was trying to do.

```
log.error("Failed to start HTTP server",
    addr: cfg.ListenAddr,
    error: err,
)
```

### WARN

Something failed but the system continues. A recoverable error, degraded feature, or unexpected condition that an operator should be aware of.

Fields should include the identity of the affected resource and the error. Add state fields that help diagnose why it failed.

```
log.warn("Failed to persist record",
    resource_id: record.ID,
    component: "engine",
    error: err,
)

log.warn("Client disconnected unexpectedly",
    request_id: reqID,
    error: err,
)
```

### INFO

Important state transitions and operational milestones. These form the narrative of what the system is doing in production.

Keep INFO volume low. Every INFO line should be something an operator would want to see when tailing logs. Resource start/complete pairs are acceptable because they represent user-visible events and carry the resource ID needed for tracing.

```
log.info("Application initialized",
    version: Version,
    log_level: cfg.LogLevel,
)

log.info("Processing started",
    resource_id: resource.ID,
    item_count: 20,
)

log.info("Processing complete",
    resource_id: resource.ID,
)

log.info("Request completed",
    request_id: reqID,
    status: 200,
    duration: elapsed,
)
```

### DEBUG

Detailed operational information for troubleshooting. Disabled in production by default. Volume is not a concern here -- be generous with context.

DEBUG is where you answer **how** in full detail: configuration values, intermediate state, timing intervals, client state.

```
log.debug("Engine configuration",
    draw_duration: cfg.DrawDuration,
    wait_duration: cfg.WaitDuration,
    item_count: cfg.ItemCount,
    max_value: cfg.MaxValue,
)

log.debug("Item revealed",
    resource_id: resource.ID,
    item_value: item,
    revealed_count: revealedCount,
)

log.debug("Event broadcast",
    event_type: eventType,
    client_count: clientCount,
)
```

## Resource Lifecycle Pattern

Resources (jobs, orders, tasks, or any primary unit of work) should be traceable from creation to completion. Follow this pattern:

```
// 1. Resource created (DEBUG) - configuration and initial state
log.debug("Creating resource",
    resource_id: nextID,
    item_count: len(items),
)

// 2. Resource started (INFO) - resource is now active
log.info("Processing started",
    resource_id: resource.ID,
    item_count: len(items),
)

// 3. Progress updates (DEBUG) - detailed processing progress
log.debug("Item processed",
    resource_id: resource.ID,
    item_value: item,
    processed: i + 1,
    remaining: total - i - 1,
)

// 4. Resource complete (INFO) - resource finished successfully
log.info("Processing complete",
    resource_id: resource.ID,
)
```

Every failure within a resource lifecycle should include the resource's identity so it can be correlated:

```
log.warn("Failed to broadcast item",
    resource_id: resource.ID,
    item_value: item,
    error: err,
)
```

## Request Lifecycle Pattern

HTTP requests use the middleware-provided `request_id` for correlation:

```
// Middleware automatically logs request start with request_id, method, path

// Failures during request processing
log.warn("Failed to fetch resource",
    request_id: reqID,
    resource_id: resourceID,
    error: err,
)

// Middleware automatically logs completion with status and duration
```

## Deprecation Logging

When a config field, feature, code path, or API is deprecated, the system must log clearly that it was encountered, what the replacement is, and when it will be removed.

### Deprecated Config Fields

Log at **WARN** on startup when a deprecated config field is present:

```
log.warn("Deprecated config field used",
    field: "interval",
    replacement: "draw_duration",
    removed_in: "v2.0.0",
)
```

### Deprecated Functionality

Log at **WARN** when deprecated functionality is invoked at runtime:

```
log.warn("Deprecated API endpoint called",
    endpoint: "/api/v1/current",
    replacement: "/api/v1/resources/latest",
    removed_in: "v2.0.0",
    request_id: reqID,
)
```

### Standard Deprecation Field Keys

| Key | Type | Description |
|-----|------|-------------|
| `field` | string | Name of the deprecated config field or parameter |
| `endpoint` | string | Deprecated API endpoint path |
| `feature` | string | Name of the deprecated feature or behaviour |
| `replacement` | string | What to use instead (empty string if no replacement) |
| `removed_in` | string | Version where the deprecated item will be removed |

### Rules

1. **Always log at WARN for user-facing deprecations** (config fields, API endpoints). Operators need to see these in production logs.
2. **Use DEBUG for internal code paths** that are only relevant to developers.
3. **Log once, not repeatedly.** Config deprecations log during startup. Runtime deprecations log once per session or use a guard to prevent repeated logging.
4. **Always include `replacement`**, even if the value is `"none"` -- the operator needs to know whether to migrate or simply stop using it.
5. **Always include `removed_in`** with the target version so operators can plan.

## Anti-Patterns

### Orphan logs

A log line with no identifying fields cannot be correlated with anything. Always include at least one identity field.

```
// Bad - which resource? which request?
log.warn("Failed to broadcast event", error: err)

// Good
log.warn("Failed to broadcast event",
    resource_id: resource.ID,
    error: err,
)
```

### Echo logs

Do not log just to confirm a code path was reached. The log should carry information that is not obvious from the code alone.

```
// Bad - adds no information an operator can use
log.debug("Entering processResource")
log.debug("About to generate items")

// Good - carries state that helps with debugging
log.debug("Generated items for resource",
    resource_id: resource.ID,
    items: items,
)
```

### Message-stuffed context

Do not encode structured data into the message string. It breaks filtering, searching, and JSON parsing.

```
// Bad
log.info("Starting server on " + addr)

// Good
log.info("HTTP server started",
    addr: addr,
)
```

## Summary

| Rule | Standard |
|------|----------|
| Core principle | Every log answers what, where, why, how |
| Logging library | One structured logger, no print/printf |
| Error attributes | Via shared helper producing `"error"` key |
| Duration attributes | Via shared helper producing `"duration"` key |
| Message casing | Uppercase first letter |
| Message punctuation | No trailing period |
| Message content | Static event category only, no instance data |
| Character set | ASCII only |
| Field style | Typed constructors where available |
| Field key casing | `snake_case` |
| Field selection | Include identity + state fields that aid diagnosis |
| Deprecation (user-facing) | WARN, log once, include `replacement` and `removed_in` |
| Deprecation (internal) | DEBUG, include `replacement` and `removed_in` |

---

## Appendix A: Go -- `log/slog` and `slogx`

The Go backend uses `log/slog` as its structured logger and the shared `pkg/slogx` package for helpers and middleware.

### Package

```go
import "log/slog"
import "github.com/aussiebroadwan/taboo/pkg/slogx"
```

Do not use `fmt.Println`, `log.Printf`, or any other mechanism for runtime output.

### Creating a Logger

Use `slogx.New` with functional options:

```go
logger := slogx.New(
    slogx.WithFormat(slogx.FormatJSON),   // or slogx.FormatText
    slogx.WithLevel(slog.LevelInfo),
    slogx.WithOutput(os.Stdout),
    slogx.WithService("taboo"),
    slogx.WithVersion(version),
)
```

### Context Propagation

Attach and retrieve loggers from `context.Context`:

```go
// Attach logger to context
ctx := slogx.NewContext(ctx, logger)

// Retrieve logger from context (falls back to slog.Default)
logger := slogx.FromContext(ctx)

// Add fields to the context-scoped logger
ctx = slogx.With(ctx, slog.String("component", "engine"))
```

### Attribute Helpers

Use `slogx.Error` for consistent error field keys:

```go
slog.Warn("Failed to persist game",
    slog.Int64("game_id", game.ID),
    slogx.Error(err),   // produces key "error"
)
```

### Typed Constructors

Always use explicit constructors -- never bare positional args:

```go
// Good
slog.Info("Game started",
    slog.Int64("game_id", game.ID),
    slog.Int("picks", len(picks)),
    slog.String("component", "engine"),
    slog.Duration("draw_duration", cfg.DrawDuration),
)

// Bad
slog.Info("Game started", "game_id", game.ID, "picks", len(picks))
```

### HTTP Middleware

`slogx.Middleware` handles request logging automatically. It generates a `request_id` (UUID), logs request start and completion with `method`, `path`, `status`, `duration`, and `bytes`, and attaches a request-scoped logger to the context.

```go
mux.Use(slogx.Middleware(logger, "/health"))  // "/health" logged at DEBUG
```

Quiet paths (e.g., `/health`, `/metrics`) are logged at DEBUG instead of INFO to reduce noise.

---

## Appendix B: JavaScript -- `logger.js`

The frontend uses a lightweight structured logger at `frontend/src/logger.js`.

### Import

```js
import logger from './logger.js';
```

Do not use bare `console.log` for operational logging.

### Basic Usage

```js
logger.info("Game started", { game_id: 42, picks: 20 });
logger.warn("Failed to broadcast event", { game_id: 42, error: err.message });
logger.debug("Pick revealed", { game_id: 42, pick: 7 });
logger.error("Connection lost", { component: "sse", error: err.message });
```

### Child Loggers

Use `.with()` to create a child logger with default fields -- equivalent to `slogx.With` in Go:

```js
const log = logger.with({ component: "sse" });

log.info("Connected");                        // includes component: "sse"
log.warn("Reconnecting", { attempt: 3 });     // includes component: "sse"
log.error("Connection failed", { error: err.message });
```

### API

| Method | Maps To | Level |
|--------|---------|-------|
| `logger.debug(msg, fields)` | `console.debug` | DEBUG |
| `logger.info(msg, fields)` | `console.info` | INFO |
| `logger.warn(msg, fields)` | `console.warn` | WARN |
| `logger.error(msg, fields)` | `console.error` | ERROR |
| `logger.with(fields)` | -- | Returns child logger |

All the same message format and field naming rules from the main standard apply. Messages are uppercase, no trailing period, ASCII only, fields use `snake_case`.
