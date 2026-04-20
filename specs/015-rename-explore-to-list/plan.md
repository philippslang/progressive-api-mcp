# Implementation Plan: Rename `explore_api` to `list_api`

**Branch**: `015-rename-explore-to-list` | **Date**: 2026-04-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/015-rename-explore-to-list/spec.md`

## Summary

Rename the MCP tool `explore_api` to `list_api` throughout live code, tests, user-facing documentation, and the config-validation whitelist. No behavioural change: inputs, outputs, error codes, allow-list semantics, and tool-prefix handling stay identical. Hard rename — the old identifier is removed rather than aliased, so misconfigured `allow_list.tools: [explore_api]` entries fail loudly at startup.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `github.com/mark3labs/mcp-go` v0.46.0, `github.com/pb33f/libopenapi` — no new dependencies
**Storage**: N/A (in-memory registry, no persistence)
**Testing**: `go test ./...`, `testify` assertions, in-process MCP client transport
**Target Platform**: Linux / macOS / Windows (Go cross-compiled)
**Project Type**: Library + CLI (`pkg/openapimcp` library, `cmd/prograpimcp` binary)
**Performance Goals**: N/A — no runtime-performance-sensitive changes; this is a static rename.
**Constraints**: Must leave historical spec artefacts under `specs/001-…` through `specs/014-…` untouched (they are frozen records).
**Scale/Scope**: 4 live code files, 3 test files, 2 user-facing docs (`README.md`, `config.yaml.example`). Historical spec files are out of scope.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution at `.specify/memory/constitution.md` is the unpopulated template (placeholder principles only). No enforceable gates are defined. Proceeding.

**Initial check**: PASS (no applicable gates).
**Post-design check**: PASS (no applicable gates).

## Project Structure

### Documentation (this feature)

```text
specs/015-rename-explore-to-list/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── list-tool.md     # Tool contract (post-rename)
└── tasks.md             # Phase 2 output (/speckit.tasks command — NOT created by /speckit.plan)
```

### Source Code (repository root)

Existing layout (unchanged by this feature except for the one file rename):

```text
pkg/
├── openapimcp/         # server.go — update RegisterExploreTools call
├── config/             # config.go — update knownToolNames + hint + doc comment
└── tools/
    ├── explore.go      # RENAMED → list.go
    ├── http.go
    ├── schema.go
    ├── search.go
    └── ...

tests/
├── integration/
│   └── http_tools_test.go    # 29 references to update (tool calls, helper registration)
├── contract/
│   └── mcp_tools_contract_test.go  # 1 reference
└── unit/
    └── config_test.go        # 4 references (allow-list validation tests)

README.md                # 5 references — update
config.yaml.example      # 1 reference — update
```

**Structure Decision**: No structural change. One file rename (`pkg/tools/explore.go` → `pkg/tools/list.go`) with corresponding identifier renames inside.

### Files touched (live code + tests + user-facing docs)

| File | Change |
|------|--------|
| `pkg/tools/explore.go` → `pkg/tools/list.go` | File rename; `RegisterExploreTools` → `RegisterListTools`; tool name literal `"explore_api"` → `"list_api"`; update doc comment. `PathInfo` type name is generic and stays. |
| `pkg/openapimcp/server.go` | Update single call site to `tools.RegisterListTools(...)`. |
| `pkg/config/config.go` | `knownToolNames`: replace `"explore_api"` entry with `"list_api"`. `knownToolNamesHint`: replace `explore_api` with `list_api`. `APIAllowList.Tools` doc comment: replace `"explore_api"` with `"list_api"`. |
| `tests/integration/http_tools_test.go` | Replace `RegisterExploreTools` call, all 29 `"explore_api"` tool-name literals, and any sub-test titles that quote `explore_api`. |
| `tests/contract/mcp_tools_contract_test.go` | Replace single occurrence (comment/test identifier). |
| `tests/unit/config_test.go` | Replace 4 occurrences in allow-list validation test cases. |
| `README.md` | Replace 5 occurrences (table row, prose, code block, two tool-prefix examples). |
| `config.yaml.example` | Replace single occurrence in the `allow_list.tools` example comment. |

### Files deliberately NOT touched

- All files under `specs/001-…` through `specs/014-…` — historical records; per spec Assumptions, frozen.
- `pkg/tools/search.go`, `pkg/tools/schema.go`, `pkg/tools/http.go` — unrelated to the rename.
- Mock server code under `tests/mockservers/` — does not reference the tool name.

## Complexity Tracking

> Fill ONLY if Constitution Check has violations that must be justified.

No constitution violations. Section left empty intentionally.
