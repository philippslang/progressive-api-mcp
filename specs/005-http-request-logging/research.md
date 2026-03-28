# Research: HTTP Request Logging

**Branch**: `005-http-request-logging` | **Date**: 2026-03-28

## Decision 1: Middleware approach vs. per-handler logging

**Decision**: Implement a single HTTP middleware function that wraps the entire `http.ServeMux`, logging all routes uniformly.

**Rationale**: We already own the `http.ServeMux` and `http.Server` (introduced in feature 004). Wrapping the mux's handler in one middleware is the standard Go idiom and guarantees all routes — including future ones — are covered automatically. Per-handler logging would require touching every handler and would miss routes added later.

**Alternatives considered**:
- Per-handler logging — rejected: brittle, misses future routes, violates DRY.
- Third-party logging library (e.g., `chi`, `gorilla/handlers`) — rejected: no new dependencies allowed; the stdlib is sufficient.

---

## Decision 2: Capturing the response status code

**Decision**: Introduce an unexported `responseWriter` struct that embeds `http.ResponseWriter`, overrides `WriteHeader`, and records the status code. Default to 200 if `WriteHeader` is never called (Go's implicit default).

**Rationale**: Go's standard `http.ResponseWriter` does not expose the status code after it has been sent. The minimal wrapper pattern is universally used in the Go ecosystem for this purpose and requires no dependencies.

**Alternatives considered**:
- Reading the status from response headers after the fact — not possible; headers are sent, not buffered.
- Using `httptest.ResponseRecorder` — only appropriate in tests, not production code.

---

## Decision 3: Log format

**Decision**: `[prograpimcp] METHOD /path?query STATUS Xms` written to `os.Stderr` via `fmt.Fprintf`.

Example: `[prograpimcp] GET /health 200 0ms`

**Rationale**: Matches the prefix style already used by existing startup log lines (`[prograpimcp] MCP endpoint: ...`). Human-readable, grep-friendly. Including elapsed time in milliseconds is sufficient precision for operational use.

**Alternatives considered**:
- Apache Combined Log Format — more standard for web servers but heavier and less readable in a terminal context.
- Structured JSON — out of scope per spec assumptions (v1 is human-readable only).
- Nanosecond or microsecond precision for elapsed time — overkill; milliseconds are what operators care about.

---

## Decision 4: Where to add the code

**Decision**: Add the `responseWriter` wrapper type and `loggingMiddleware` function directly in `pkg/openapimcp/server.go`, applied at the point where `httpServer.Handler` is set.

**Rationale**: Both are small (< 20 lines combined), unexported, and tightly coupled to the HTTP transport setup that already lives in `server.go`. Creating a new file for two unexported helpers would be premature.

**Alternatives considered**:
- New file `pkg/openapimcp/middleware.go` — rejected: adds file count for < 20 lines of unexported code.
- New package `pkg/middleware/` — rejected: over-engineering a one-use helper.

---

## All NEEDS CLARIFICATION markers resolved

None were raised in the spec. No open questions remain.
