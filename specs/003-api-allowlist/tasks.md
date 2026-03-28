---
description: "Task list for API Tool & Path Allow List"
---

# Tasks: API Tool & Path Allow List

**Input**: Design documents from `/specs/003-api-allowlist/`
**Prerequisites**: plan.md ✅ | spec.md ✅ | research.md ✅ | data-model.md ✅ | contracts/ ✅

**Tests**: Included — TDD is mandatory per the project constitution (Principle II).
Tests MUST be written and confirmed failing before implementation begins.

**Scope**: Amendment to `001-openapi-mcp-server`. No new source files; changes confined to
`pkg/config/config.go`, `pkg/registry/registry.go`, `pkg/tools/` (three handler files),
`pkg/openapimcp/server.go`, and test files.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US3)

---

## Phase 1: Setup

**Purpose**: No new project setup needed — this is an amendment. One documentation update.

- [X] T001Update `config.yaml.example` at repository root to document the new `allow_list` block with commented-out example showing `tools` and `paths` sub-keys

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add `APIAllowList` struct and wire it into `APIConfig` and `APIEntry`. All user story phases depend on this.

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete.

### Tests ⚠️ Write first — confirm they FAIL before implementing

- [X] T002 Add `TestAPIAllowListValidation` table-driven tests to `tests/unit/config_test.go`: valid tool names (all 6 individually), empty `Tools` slice valid, unknown tool name in `Tools` returns error, unknown key in `Paths` map returns error, tool in `Paths` but not in `Tools` is valid (dormant restriction)

### Implementation

- [X] T003 Add `APIAllowList` struct to `pkg/config/config.go` with fields `Tools []string \`yaml:"tools"\`` and `Paths map[string][]string \`yaml:"paths"\`` and doc comment per Principle I
- [X] T004 Add `AllowList APIAllowList \`yaml:"allow_list"\`` field to `APIConfig` struct in `pkg/config/config.go` (zero value = allow all)
- [X] T005 Add `AllowList` validation block to `Config.Validate()` in `pkg/config/config.go`: iterate each API entry's `AllowList.Tools` and `AllowList.Paths` keys; reject unknown names with message `"apis[N].allow_list.tools: unknown tool name \"X\"; valid names are: explore_api, get_schema, http_get, http_post, http_put, http_patch"`
- [X] T006 Add `AllowList config.APIAllowList` field to `APIEntry` struct in `pkg/registry/registry.go`; assign `AllowList: cfg.AllowList` inside `Load()` when constructing the `APIEntry`

**Checkpoint**: `go test ./tests/unit/...` MUST pass before proceeding to story phases.

---

## Phase 3: User Story 1 — Restrict Which Tools Are Exposed Per API (Priority: P1) 🎯 MVP

**Goal**: When `allow_list.tools` is configured, only the listed tools are registered in the MCP session; others are entirely absent.

**Independent Test**: Configure a Petstore API entry with `allowed_tools: [explore_api, http_get]`, verify `http_post` is absent from the MCP session (call returns tool-not-found error), verify `http_get` and `explore_api` work normally.

### Tests for US1 ⚠️ Write first — confirm they FAIL before implementing

- [X] T007 [US1] Add integration test `TestToolAllowList` to `tests/integration/http_tools_test.go`: add `makeTestClientWithAllowedTools(t, reg, httpClient, allowedTools map[string]bool)` helper; use it to build a client restricted to `{explore_api: true, http_get: true}`; verify `explore_api` succeeds; verify calling `http_post` returns an error (tool absent from session)
- [X] T008 [P] [US1] Add integration test `TestDefaultAllToolsAllowed` to `tests/integration/http_tools_test.go`: verify that passing `nil` as `allowedTools` (or an empty map) still registers all 6 tools and they all respond normally

### Implementation for US1

- [X] T009 [US1] Update `RegisterHTTPTools` signature in `pkg/tools/http.go` to add `allowedTools map[string]bool` parameter; wrap each `s.AddTool` call with `if allowedTools == nil || allowedTools["http_get"]` (and `http_post`, `http_put`, `http_patch`); update `makeTestClient` and `makeTestClientWithPrefix` integration test helpers to pass `nil`
- [X] T010 [P] [US1] Update `RegisterExploreTools` signature in `pkg/tools/explore.go` to add `allowedTools map[string]bool` parameter; guard `s.AddTool` with `if allowedTools == nil || allowedTools["explore_api"]`; update integration test helpers
- [X] T011 [P] [US1] Update `RegisterSchemaTools` signature in `pkg/tools/schema.go` to add `allowedTools map[string]bool` parameter; guard `s.AddTool` with `if allowedTools == nil || allowedTools["get_schema"]`; update integration test helpers
- [X] T012 [US1] Add `computeAllowedTools(apis []config.APIConfig) map[string]bool` in `pkg/openapimcp/server.go`: returns `nil` when all APIs have empty `AllowList.Tools` (allow-all), otherwise returns a union map of every tool name listed across all APIs; pass result to all three `Register*` calls in `Start()`
- [X] T013 [US1] Add per-request tool check in the MCP tool handlers in `pkg/tools/http.go`: after `resolveAPI`, if `entry.AllowList.Tools` is non-empty and the current tool base name is not in the list, return `TOOL_NOT_PERMITTED` error; add `toolNotPermitted(toolName, apiName string)` helper returning the standard `ToolError` shape

**Checkpoint**: `go test ./...` MUST pass. US1 independently testable.

---

## Phase 4: User Story 2 — Restrict Which Paths Are Allowed Per Tool (Priority: P2)

**Goal**: When `allow_list.paths` maps a tool to a list of path templates, any request to a path not in that list is rejected with `PATH_NOT_PERMITTED` before validation or execution.

**Independent Test**: Configure `http_get` with `allowed_paths: ["/pets", "/pets/{id}"]`; verify `GET /pets` succeeds; verify `GET /owners` returns `PATH_NOT_PERMITTED`; verify `GET /pets/42` (matches `/pets/{id}`) is accepted.

### Tests for US2 ⚠️ Write first — confirm they FAIL before implementing

- [X] T014 [US2] Add integration test `TestPathAllowList` to `tests/integration/http_tools_test.go`: build a registry with Petstore and an `AllowList.Paths` restricting `http_get` to `["/pets"]`; call `http_get /pets` → expect success; call `http_get /owners` → expect `PATH_NOT_PERMITTED`
- [X] T015 [P] [US2] Add integration test `TestExplorePathAllowList` to `tests/integration/http_tools_test.go`: restrict `explore_api` paths to `["/pets", "/pets/{id}"]`; call `explore_api` with no prefix → verify returned list contains only those two paths
- [X] T016 [P] [US2] Add integration test `TestDefaultAllPathsAllowed` to `tests/integration/http_tools_test.go`: no `AllowList.Paths` set → all paths accessible for all tools (regression guard)

### Implementation for US2

- [X] T017 [US2] Add `isPathPermitted(template string, allowed []string) bool` helper in `pkg/tools/http.go`: returns `true` if `allowed` is nil/empty, otherwise `true` iff `template` is in the slice (exact string match); this is the single-responsibility path-check function
- [X] T018 [US2] Add path allow-list check in `validateAndExecute` in `pkg/tools/http.go`: after `resolveAPI` and after `matchPath` resolves the template, if `entry.AllowList.Paths[toolBase]` is non-empty and the resolved template is not permitted, return `toolErrorResult("PATH_NOT_PERMITTED", ...)` with message identifying path, tool, and API name
- [X] T019 [US2] Add path allow-list filter in `RegisterExploreTools` handler in `pkg/tools/explore.go`: after building the `paths []PathInfo` slice, if `entry.AllowList.Paths["explore_api"]` is non-empty, filter the slice to only paths whose template is in the allowed list; apply this filter before the existing `prefix` filter
- [X] T020 [US2] Add path allow-list check in `RegisterSchemaTools` handler in `pkg/tools/schema.go`: after resolving the matched template, if `entry.AllowList.Paths["get_schema"]` is non-empty and the template is not permitted, return `PATH_NOT_PERMITTED`

**Checkpoint**: `go test ./...` MUST pass. US2 independently testable.

---

## Phase 5: User Story 3 — Combined Tool and Path Restrictions (Priority: P3)

**Goal**: Tool-level and path-level restrictions compose correctly in the same API entry — no interaction bugs between the two enforcement layers.

**Independent Test**: Configure `allowed_tools: [explore_api, http_get, http_post]` and `http_post.allowed_paths: ["/pets"]`; verify `http_post /pets` succeeds, `http_post /owners` returns `PATH_NOT_PERMITTED`, and `http_put /pets` returns tool-not-found.

### Tests for US3 ⚠️ Write first — confirm they FAIL before implementing

- [X] T021 [US3] Add integration test `TestCombinedAllowList` to `tests/integration/http_tools_test.go`: configure combined restrictions (tools: explore_api, http_get, http_post; http_post paths: /pets); verify `http_post /pets` executes; verify `http_post /owners` returns `PATH_NOT_PERMITTED`; verify `http_put /pets` is absent from session (tool-not-found error)

### Implementation for US3

No new implementation code is needed — US3 exercises the already-implemented US1 and US2 enforcement layers together. The test validates they compose correctly.

**Checkpoint**: `go test ./...` MUST pass. All three user stories independently testable.

---

## Phase 6: Polish

- [X] T022 [P] Add contract shape test for `APIAllowList` to `tests/contract/mcp_tools_contract_test.go`: verify `config.APIAllowList{Tools: []string{"http_get"}, Paths: map[string][]string{"http_get": {"/pets"}}}` compiles and has correct field names and types
- [X] T023 [P] Add `BenchmarkPathAllowCheck` to `tests/unit/` (new file `allowlist_bench_test.go`): benchmark `isPathPermitted` with allow lists of 10, 50, and 100 entries; assert each completes in < 1 µs
- [X] T024 [P] Run `go build ./...` to confirm all packages compile cleanly after signature changes to `Register*` functions
- [X] T025 [P] Verify all exported symbols added by this feature have godoc comments: `APIAllowList`, `APIAllowList.Tools`, `APIAllowList.Paths`, and the `AllowList` field on `APIConfig` and `APIEntry`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all story phases
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 2 (independent of US1 at config/registry level; US2 tests use US1 helpers — run after US1 passes)
- **Phase 5 (US3)**: Depends on Phase 3 and Phase 4 both passing
- **Phase 6 (Polish)**: Depends on Phase 5

### Within Each Phase

- Tests marked ⚠️ MUST be written and confirmed FAILING before implementation begins
- `pkg/config` and `pkg/registry` changes (T003–T006) before tool handler changes (T009–T013, T017–T020)
- `Register*` signature changes (T009–T011) before server wiring (T012)
- `isPathPermitted` helper (T017) before its callers (T018–T020)

---

## Parallel Opportunities

```
Phase 2 tests:   T002 only (single test file)
Phase 2 impl:    T003, T004 parallel (same file, non-conflicting structs) → T005 → T006

Phase 3 tests:   T007, T008 parallel (same file, different test functions)
Phase 3 impl:    T009, T010, T011 parallel (different files: http.go, explore.go, schema.go)
                 T012 → after T009-T011 (server.go wires them all)
                 T013 → after T012 (adds per-request check on top)

Phase 4 tests:   T014, T015, T016 parallel (same file, different test functions)
Phase 4 impl:    T017 → T018, T019, T020 parallel (T017 helper used by all three)

Phase 5 tests:   T021 only
Phase 5 impl:    None (composition test)

Phase 6:         T022, T023, T024, T025 all parallel
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002–T006)
3. Complete Phase 3: US1 (T007–T013)
4. **STOP and VALIDATE**: `go test ./...` green; `http_post` absent when not in `allow_list.tools`
5. Ship: operators can restrict which tool categories are available

### Incremental Delivery

1. Setup + Foundational → config + registry carry the allow list data
2. US1 → tool-level restriction enforced at registration time
3. US2 → path-level restriction enforced at request time
4. US3 → validated that both levels compose correctly
5. Each story delivers independent value and leaves test suite green

---

## Notes

- `[P]` tasks operate on different files or independent test functions — no write conflicts
- Story labels map directly to user stories in `spec.md`
- The `Register*` signature changes (T009–T011) will require updating `makeTestClient` and `makeTestClientWithPrefix` integration helpers — include those updates in the same task
- `isPathPermitted` (T017) is the single path-check function used by all three HTTP handlers — no duplication (Rule of Three satisfied)
- The `TOOL_NOT_PERMITTED` error (T013) is only reachable in multi-API servers where the tool is registered for one API but not another — single-API tool restrictions are enforced by not registering the tool at all
