# Tasks: search_api Tool

**Input**: Design documents from `/specs/014-search-api-tool/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: `US1` or `US2` for user-story phase tasks only
- Exact file paths are included in all descriptions

---

## Phase 1: Setup

**Purpose**: Verify baseline before any code changes. This feature is additive — no new packages or dependencies are required.

- [x] T001 Run `go test ./...` to confirm the baseline is green before making changes

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: Register the tool name in the config-validation whitelist. Without this, any config that references `"search_api"` in `allow_list` is rejected at load time, so this must exist before the tool can be usefully deployed — even though nothing else blocks on it for raw functionality, user stories should be shippable end-to-end.

- [x] T002 Add `"search_api": {}` to the `knownToolNames` map and append `, search_api` to the `knownToolNamesHint` string in `pkg/config/config.go`
- [x] T003 Update the YAML tag comment on `APIAllowList.Tools` in `pkg/config/config.go` to list `search_api` among the valid tool names

**Checkpoint**: `go build ./...` still passes. Configs that mention `"search_api"` in `allow_list` now validate instead of being rejected.

---

## Phase 3: User Story 1 - Find Endpoints by Keyword Across All APIs (Priority: P1) 🎯 MVP

**Goal**: A single `search_api` call returns every endpoint across every registered API whose path, summary, or description matches the query substring.

**Independent Test**: Configure two APIs (petstore + bookstore) with overlapping keywords. Call `search_api` with `query: "pet"` and no `api` filter. Verify the response contains matching endpoints from both APIs, each including `{api, method, path, schema}`. Also verify that an empty query returns `INVALID_INPUT`, and that no-match returns an empty array (not an error).

### Implementation for User Story 1

- [x] T004 [US1] Create `pkg/tools/search.go` with: package declaration; `SearchResult` struct (`api`, `method`, `path`, `schema` with `omitempty`); `RegisterSearchTools(s *server.MCPServer, reg *registry.Registry, prefix string, allowedTools map[string]bool)` signature that mirrors `RegisterExploreTools`; and a stub handler that returns an empty JSON array.
- [x] T005 [US1] In `pkg/tools/search.go`, add the input-validation branch: trim `query`; if empty, return `toolErrorResult("INVALID_INPUT", "query must not be empty", nil, nil)`.
- [x] T006 [US1] In `pkg/tools/search.go`, implement the core matcher: iterate `reg.ListNames()`, call `reg.Lookup(name)`, build v3 model, iterate `model.Paths.PathItems.FromOldest()`, for each of GET/POST/PUT/PATCH/DELETE check if the operation exists and whether the lowercased concat of path + summary + description contains the lowercased query; on hit, append a `SearchResult` (use `schemaToMap` for request body when present).
- [x] T007 [US1] In `pkg/tools/search.go`, apply the per-API allow-list filter: before appending results for a given path, check `IsPathPermitted(path, entry.AllowList.Paths["search_api"])` and skip when not permitted.
- [x] T008 [US1] In `pkg/openapimcp/server.go`, add `tools.RegisterSearchTools(s.mcpSrv, s.registry, effectivePrefix, allowedTools)` on the line after `RegisterSchemaTools`.
- [x] T009 [US1] Add integration test `TestSearchAPIAcrossAllAPIs` to `tests/integration/http_tools_test.go` covering: (a) two APIs registered, query matches paths in both, returns hits from both; (b) query matches a description-only field and that endpoint is returned; (c) no-match returns empty array with no error code; (d) empty query returns `INVALID_INPUT`.

**Checkpoint**: User Story 1 is fully functional. `go test ./...` passes. `search_api` is visible via MCP and returns correct cross-API results.

---

## Phase 4: User Story 2 - Narrow Search to a Single API (Priority: P2)

**Goal**: When `api` is set, the search is limited to that API and unknown names produce a clear error.

**Independent Test**: Configure two APIs with overlapping keywords. Call `search_api` with `query: "pet"` and `api: "petstore"`. Verify only petstore results are returned. Call again with `api: "nosuch"` and verify `API_NOT_FOUND`.

### Implementation for User Story 2

- [x] T010 [US2] In `pkg/tools/search.go`, extend the handler: when `req.GetString("api", "")` is non-empty, call `reg.Lookup(name)`; on miss return `toolErrorResult("API_NOT_FOUND", fmt.Sprintf("API %q is not registered", name), nil, nil)`; on hit, scan only that entry.
- [x] T011 [US2] Add integration test `TestSearchAPISingleAPIFilter` to `tests/integration/http_tools_test.go` covering: (a) two APIs, `api: "petstore"` returns only petstore hits; (b) `api: "nosuch"` returns `API_NOT_FOUND`.

**Checkpoint**: Both user stories are independently functional. `go test ./...` passes.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Doc comments, contract test for response shape, and config example update.

- [x] T012 [P] Add contract test `TestSearchResultShape` to `tests/contract/mcp_tools_contract_test.go` verifying `SearchResult` JSON marshals to an object with `api`, `method`, `path` always present, and `schema` omitted when nil.
- [x] T013 [P] Add unit test `TestSearchMatcher` in `tests/unit/search_test.go` with table cases: case-insensitive, substring, path-only match, description-only match, summary-only match, no match (pure function extracted from the handler if reasonable, otherwise skip this task and rely on integration tests).
- [x] T014 [P] Add an example comment block in `config.yaml.example` showing `search_api` usage in `allow_list.tools` and `allow_list.paths`.
- [x] T015 [P] Add doc comments to the exported `SearchResult` type and `RegisterSearchTools` function in `pkg/tools/search.go`.
- [x] T016 Run `go vet ./...` and `go test ./...` and confirm clean output.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Baseline)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1
- **Phase 3 (US1)**: Depends on Phase 2 (the tool registers regardless, but configs that restrict tool lists require the name to be recognised)
- **Phase 4 (US2)**: Depends on T006 in Phase 3 (needs the core matcher + handler wiring in place)
- **Phase 5 (Polish)**: Depends on US1 and US2 complete

### Within Phase 3 (US1)

```
T004 ─── T005 ─── T006 ─── T007 ─── T008 ─── T009
```

All sequential — they modify the same file (`search.go`) or depend on earlier steps. T009 (integration test) depends on T008 (server wire-up).

### Within Phase 4 (US2)

```
T010 ─── T011
```

T010 modifies `search.go` (conflicts with US1 tasks if run in parallel); T011 is the integration test.

### Within Phase 5

- T012, T013, T014, T015 operate on different files — parallel-eligible
- T016 depends on all prior tasks (final green-run check)

---

## Parallel Example: Phase 5

```bash
# After all user story work is done, launch polish tasks in parallel:
Task T012: "Add SearchResult contract test to tests/contract/mcp_tools_contract_test.go"
Task T013: "Add matcher unit tests to tests/unit/search_test.go"
Task T014: "Add search_api example to config.yaml.example"
Task T015: "Add doc comments in pkg/tools/search.go"

# Then finally:
Task T016: "Run go vet + go test"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 → Phase 2 → Phase 3 (T004–T009)
2. **STOP and VALIDATE**: cross-API substring search works; error for empty query; empty array for no match
3. Ship MVP

### Incremental Delivery

1. MVP (Phases 1–3): cross-API search ✓
2. Phase 4: single-API filter ✓
3. Phase 5: polish (doc comments, contract test, example config) ✓

---

## Notes

- Only T004 creates a new file (`pkg/tools/search.go`); all other code edits modify existing files.
- The MVP intentionally omits a pure-function `match()` extraction — integration tests cover the matcher thoroughly. Extract in T013 only if it simplifies unit testing.
- `schemaToMap` is package-internal in `pkg/tools` — no export change required since `search.go` is the same package.
- No new dependencies; no changes to `go.mod`.
- Backward-compatible: configs that previously didn't mention `search_api` continue to work; the new tool becomes available to all APIs with no explicit opt-in.
