# Implementation Plan: Health Endpoint

**Branch**: `004-health-endpoint` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-health-endpoint/spec.md`

## Summary

Add a `/health` HTTP endpoint to the MCP server that returns HTTP 200 with a minimal status body. The endpoint must be served on the same host/port as the MCP server so operators and monitoring tools have a single address to check. No new dependencies are required.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0 (StreamableHTTPServer), standard library `net/http`
**Storage**: N/A
**Testing**: `go test ./...` — table-driven unit tests + integration tests via `httptest`
**Target Platform**: Linux server (same as existing binary)
**Project Type**: CLI binary wrapping a library (`pkg/openapimcp`)
**Performance Goals**: Sub-millisecond response; health check overhead is negligible
**Constraints**: No new external dependencies; endpoint must be on the same port as MCP (`/mcp`)
**Scale/Scope**: Single handler, no state

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution file is an unfilled template — no codified gates exist. The plan proceeds under the project guidelines in `CLAUDE.md`:

- No new dependencies (standard library `net/http` is already in use) ✓
- No global state; dependencies passed via constructors ✓
- `pkg/` packages must not import `cobra`/`viper` ✓
- All exported types/functions must have doc comments ✓

**Gate result**: PASS

## Project Structure

### Documentation (this feature)

```text
specs/004-health-endpoint/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── contracts/
│   └── health-endpoint.md   # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
pkg/openapimcp/
└── server.go            # Add health handler registration here

tests/integration/
└── health_test.go       # New integration test

tests/unit/
└── (no new unit test file needed — handler logic is trivial)
```

**Structure Decision**: Single-project layout, extending `pkg/openapimcp/server.go`. The health handler is a one-liner and does not warrant a new package or file.

## Complexity Tracking

No constitution violations — table left empty.
