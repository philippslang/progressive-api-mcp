# Implementation Plan: API Tool & Path Allow List

**Branch**: `003-api-allowlist` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-api-allowlist/spec.md`

## Summary

Add a per-API `allow_list` configuration block that gates (a) which of the 6 MCP tools are registered for a given API and (b) which OpenAPI path templates each allowed tool may access. Default behaviour (no `allow_list` present) is unchanged: all tools registered, all paths accessible. Enforcement is split between server startup (tool registration) and request time (path check). This is an amendment to `001-openapi-mcp-server`; no new files are required beyond config, registry, tools, and wiring.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator, spf13/cobra + viper — no new dependencies
**Storage**: N/A (in-process, stateless config load)
**Testing**: `go test ./...` — table-driven unit tests, InProcessTransport integration tests, contract shape tests
**Target Platform**: Linux/macOS server binary + embeddable library
**Project Type**: CLI binary + embeddable Go library
**Performance Goals**: Allow-list path check adds ≤ 1 µs overhead per request (simple slice scan on a small list)
**Constraints**: No new dependencies; no global state; no breaking changes to existing config files
**Scale/Scope**: Amendment — changes confined to `pkg/config`, `pkg/registry`, `pkg/tools`, `pkg/openapimcp`

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Code Quality | ✅ Pass | All new exported symbols will have doc comments; `APIAllowList` is a named type with clear responsibility |
| II. Test-First (NON-NEGOTIABLE) | ✅ Pass | Tests written and confirmed failing before each implementation task |
| III. Integration & Contract Tests | ✅ Pass | Integration tests use InProcessTransport against real registry; contract test verifies `APIAllowList` shape |
| IV. Performance | ✅ Pass | SC-003 defines latency parity; benchmark for path check in hot loop included in tasks |
| V. Simplicity | ✅ Pass | One new struct, one map lookup per request — no abstractions beyond the Rule of Three |

No gate violations. Complexity Tracking table left empty.

## Project Structure

### Documentation (this feature)

```text
specs/003-api-allowlist/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (amendment — changed files only)

```text
pkg/config/config.go          # APIAllowList struct + AllowList field on APIConfig + validation
pkg/registry/registry.go      # APIEntry gains AllowList field; Load() copies it from config
pkg/tools/http.go             # validateAndExecute checks path allow list; RegisterHTTPTools filters tools
pkg/tools/explore.go          # explore_api filters returned paths to allowed set
pkg/tools/schema.go           # get_schema rejects disallowed paths
pkg/openapimcp/server.go      # computes union of allowed tools across APIs; passes to Register*

tests/unit/config_test.go     # APIAllowList validation tests (written first — TDD Red)
tests/integration/http_tools_test.go   # allow-list integration tests (written first — TDD Red)
tests/contract/mcp_tools_contract_test.go  # contract shape test for APIAllowList
```

**Structure Decision**: Single-project layout unchanged. All changes are amendments to existing packages; no new source files needed.

## Design Decisions

### 1. `APIAllowList` config struct

```yaml
# config.yaml — new allow_list block inside an api entry
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    allow_list:
      tools:          # omit for all-tools-allowed
        - explore_api
        - http_get
      paths:          # omit for all-paths-allowed (per tool)
        http_get:
          - /pets
          - /pets/{id}
```

```go
// pkg/config/config.go
type APIAllowList struct {
    Tools []string            `yaml:"tools"`  // nil/empty = all 6 tools allowed
    Paths map[string][]string `yaml:"paths"`  // nil/empty = all paths allowed per tool
}
```

`AllowList` is a value (not pointer) on `APIConfig` — zero value is naturally "allow all".

### 2. Tool registration — union across APIs

A tool is registered in the MCP session if **at least one loaded API** allows it. In the common single-API case this produces exact "absent" semantics. In multi-API servers, a tool that is blocked by API A but allowed by API B is still registered; if called against API A it returns `TOOL_NOT_PERMITTED`.

`server.go` computes a `map[string]bool` union of allowed tools from all `APIConfig.AllowList.Tools`, then passes it to each `Register*` call. Each `Register*` function receives `allowedTools map[string]bool` and skips `s.AddTool` for tools absent from the map.

### 3. Path enforcement — at request time in tool handlers

Path restrictions live in `APIEntry.AllowList.Paths` (copied from config by `registry.Load`). Enforcement runs after `resolveAPI` — when the specific API is known — in:
- `validateAndExecute` (HTTP tools): after `resolveAPI`, before `matchPath`. Checks the resolved template against the allow list.
- `explore_api` handler: filters the returned path list to paths present in `AllowList.Paths["explore_api"]` (or all paths when nil).
- `get_schema` handler: rejects the path if it is not in `AllowList.Paths["get_schema"]`.

New error code: `PATH_NOT_PERMITTED` — distinct from `PATH_NOT_FOUND` so agents can distinguish "unknown path" from "known but restricted path".

### 4. Config validation

`Config.Validate()` checks:
- Each name in `AllowList.Tools` is one of the 6 known base tool names — error if unknown.
- Each key in `AllowList.Paths` is one of the 6 known base tool names — error if unknown.
- A tool name in `AllowList.Paths` that is NOT in `AllowList.Tools` (when Tools is non-empty) is not an error — the path restriction is dormant but valid.

### 5. `explore_api` path filtering semantics

When `explore_api` has an `allowed_paths` restriction, the returned list is filtered to only paths in the allow list that also exist in the OpenAPI document. The `prefix` filter argument (existing feature) is applied on top of the allow-list filter. This gives agents accurate information about what they can actually call.

## Complexity Tracking

*(No violations — no entries required)*
