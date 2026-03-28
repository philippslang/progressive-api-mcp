# Implementation Plan: MCP Test Client CLI

**Branch**: `006-mcp-test-client` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-mcp-test-client/spec.md`

## Summary

Add a second CLI binary `mcpls` at `cmd/mcpls/` that connects to a running MCP server over HTTP, initializes the MCP session, retrieves the tool list, and prints one tool name per line to stdout. Uses the mcp-go client library already present in go.mod — no new dependencies. No dependency on the OpenAPI/registry packages.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `github.com/mark3labs/mcp-go/client`, `github.com/mark3labs/mcp-go/client/transport`, `github.com/mark3labs/mcp-go/mcp` — all already in go.mod; no new dependencies
**Storage**: N/A
**Testing**: `go test ./...` — existing test suite must remain green; no new test file required for this CLI (it is a thin wrapper)
**Target Platform**: Linux (same as existing binary)
**Project Type**: Standalone CLI binary (`cmd/mcpls/`)
**Performance Goals**: Full tool list retrieved and printed in under 5 seconds
**Constraints**: No dependency on `pkg/openapimcp`, `pkg/registry`, `pkg/loader`, or `pkg/validator`; no new external dependencies; binary must build independently with `go build ./cmd/mcpls`
**Scale/Scope**: Single `main.go` file, ~60 lines

## Constitution Check

Constitution file is an unfilled template — no codified gates. Plan proceeds under `CLAUDE.md` guidelines:

- No new external dependencies ✓
- No global state ✓
- `pkg/` packages not imported by this CLI (it imports only mcp-go client packages) ✓
- New `cmd/mcpls/` follows the existing `cmd/prograpimcp/` thin-wrapper pattern ✓

**Gate result**: PASS

## Project Structure

### Documentation (this feature)

```text
specs/006-mcp-test-client/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── contracts/
│   └── mcpls-cli.md     # CLI contract (args, exit codes, output format)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/mcpls/
└── main.go              # New binary: connect → initialize → list tools → print → exit
```

**Structure Decision**: Single `main.go` in `cmd/mcpls/`. No new `pkg/` package needed — the logic is a linear 60-line script. Cobra is available but unnecessary for a single positional argument; `os.Args` or `flag` is sufficient and keeps the binary lean.

## Complexity Tracking

No constitution violations — table left empty.
