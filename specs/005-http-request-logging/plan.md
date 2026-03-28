# Implementation Plan: HTTP Request Logging

**Branch**: `005-http-request-logging` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-http-request-logging/spec.md`

## Summary

Wrap the existing `http.ServeMux` in `pkg/openapimcp/server.go` with a logging middleware that records one line per request to stderr — containing method, full request URI, response status code, and elapsed time. No new dependencies; the middleware requires a small `responseWriter` wrapper to capture the status code written by downstream handlers.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Standard library `net/http`, `fmt`, `os`, `time` — no new dependencies
**Storage**: N/A
**Testing**: `go test ./...` — integration test via existing `httptest`-style setup
**Target Platform**: Linux server (same as existing binary)
**Project Type**: CLI binary wrapping a library (`pkg/openapimcp`)
**Performance Goals**: Logging overhead < 1 ms per request (trivially satisfied by a single `fmt.Fprintf` to stderr)
**Constraints**: No new external dependencies; output to stderr only; no log levels or filtering
**Scale/Scope**: One middleware function, one wrapper type — all in `server.go`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution file is an unfilled template — no codified gates. Plan proceeds under `CLAUDE.md` guidelines:

- No new dependencies (stdlib only) ✓
- No global state ✓
- `pkg/` packages must not import `cobra`/`viper` ✓
- Exported types/functions must have doc comments ✓ (new unexported helpers need no doc comment)

**Gate result**: PASS

## Project Structure

### Documentation (this feature)

```text
specs/005-http-request-logging/
├── plan.md              # This file
├── research.md          # Phase 0 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

No `contracts/` or `data-model.md` needed — no new external interface, no entities.

### Source Code (repository root)

```text
pkg/openapimcp/
└── server.go            # Add responseWriter wrapper + loggingMiddleware, apply to httpServer.Handler

tests/integration/
└── health_test.go       # Extend existing test or add a new test to verify log output
```

**Structure Decision**: All changes in `pkg/openapimcp/server.go`. The middleware and wrapper type are unexported helpers — no new file or package warranted.

## Complexity Tracking

No constitution violations — table left empty.
