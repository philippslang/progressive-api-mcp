# Implementation Plan: search_api Tool

**Branch**: `014-search-api-tool` | **Date**: 2026-04-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/014-search-api-tool/spec.md`

## Summary

Add a new MCP tool `search_api` that greps across the OpenAPI documents of all registered APIs (or a single API via optional `api` filter) and returns endpoints whose path template or operation description/summary contains a given substring (case-insensitive). Each hit returns `{api, method, path, schema}`. The tool follows the same registration pattern as `explore_api`/`get_schema`: one global registration, optional `api` parameter, honours `allow_list.tools` and `allow_list.paths`. Schema extraction reuses the existing `schemaToMap` helper from `pkg/tools/schema.go`.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi — no new dependencies
**Storage**: In-memory only (uses already-parsed OpenAPI documents from registry)
**Testing**: `go test` + testify; table-driven unit tests, integration tests via in-process MCP client
**Target Platform**: Linux server (HTTP and stdio transports)
**Project Type**: Library (`pkg/openapimcp`) + CLI thin wrapper (`cmd/prograpimcp`)
**Performance Goals**: Linear scan over all operations per search call is acceptable — typical configurations hold at most a few APIs with hundreds of paths. No indexing required for initial version.
**Constraints**: No new dependencies; tool must register via the existing `knownToolNames` mechanism so `allow_list.tools` can reference `"search_api"`.
**Scale/Scope**: New file `pkg/tools/search.go`, plus small edits to `pkg/openapimcp/server.go` (registration) and `pkg/config/config.go` (known tool name).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution file (`.specify/memory/constitution.md`) contains only the unfilled template. Applying the CLAUDE.md guidelines as gates:

| Rule | Status |
|------|--------|
| `pkg/` packages MUST NOT import `cobra` or `viper` | PASS — new file uses only mcp-go, libopenapi, stdlib |
| No new dependencies | PASS |
| Exported types/functions in stable packages MUST have doc comments | PASS — planned |
| No global state; dependencies passed via constructors | PASS — follows existing `Register*Tools(s, reg, prefix, allowedTools)` signature |
| Table-driven tests | PASS — planned |

No violations; no Complexity Tracking entries required.

## Project Structure

### Documentation (this feature)

```text
specs/014-search-api-tool/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── search-tool.md
└── tasks.md             # Phase 2 output (/speckit.tasks — not created here)
```

### Source Code (repository root)

```text
pkg/tools/
├── search.go            # NEW — SearchResult struct + RegisterSearchTools
├── explore.go           # unchanged (pattern reference)
├── schema.go            # unchanged (schemaToMap helper reused via package-internal access)
└── http.go              # unchanged

pkg/config/
└── config.go            # EDIT — add "search_api" to knownToolNames + hint string

pkg/openapimcp/
└── server.go            # EDIT — call tools.RegisterSearchTools alongside other Register* calls

tests/unit/
└── search_test.go       # NEW — table-driven tests for the matcher helper

tests/integration/
└── http_tools_test.go   # EDIT — add TestSearchAPI covering US1, US2, edge cases

tests/contract/
└── mcp_tools_contract_test.go   # EDIT — add contract test for SearchResult shape
```

**Structure Decision**: Single-project layout (existing). One new source file under `pkg/tools/` plus small edits to two existing files. No new packages.
