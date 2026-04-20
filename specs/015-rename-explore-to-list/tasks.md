# Tasks: Rename `explore_api` to `list_api`

**Input**: Design documents from `/specs/015-rename-explore-to-list/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓, quickstart.md ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: `US1` or `US2` for user-story phase tasks only
- Exact file paths are included in every description

---

## Phase 1: Setup

**Purpose**: Confirm baseline is green before performing the rename. No new packages or dependencies.

- [X] T001 Run `go test ./...` to confirm the baseline is green before making any changes

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: Flip the config-validation whitelist from `explore_api` to `list_api`. Without this, both stories fail at startup (`list_api` in `allow_list.tools` is rejected; `explore_api` continues to be accepted), so this must land before user-story work.

- [X] T002 In `pkg/config/config.go`, replace the `"explore_api": {}` entry in `knownToolNames` with `"list_api": {}` so `list_api` is the valid identifier and `explore_api` is rejected
- [X] T003 In `pkg/config/config.go`, update `knownToolNamesHint` to `"list_api, get_schema, search_api, http_get, http_post, http_put, http_patch"`
- [X] T004 In `pkg/config/config.go`, update the doc comment on `APIAllowList.Tools` (the `Valid values: ...` line) to list `"list_api"` instead of `"explore_api"`

**Checkpoint**: `go build ./...` still passes. Configs that reference `list_api` validate; configs that reference `explore_api` are rejected with the new hint.

---

## Phase 3: User Story 1 - Call the Tool by Its New Name (Priority: P1) 🎯 MVP

**Goal**: The MCP server advertises and handles the tool under the name `list_api`. The old name is gone from the wire and from Go identifiers. Behaviour is unchanged.

**Independent Test**: Start a server with a single registered API, list tools via MCP, and confirm `list_api` appears and `explore_api` does not. Call `list_api` with no args (single-API case) and with `api: "petstore"` (multi-API case) and confirm identical behaviour to the old `explore_api`. Apply a `tool_prefix: "store"` and confirm the name is `store_list_api`, not `store_explore_api`.

### Implementation for User Story 1

- [X] T005 [US1] Rename the file `pkg/tools/explore.go` to `pkg/tools/list.go` (via `git mv`); no content changes yet in this task
- [X] T006 [US1] In `pkg/tools/list.go`, rename the exported function `RegisterExploreTools` to `RegisterListTools` (signature unchanged)
- [X] T007 [US1] In `pkg/tools/list.go`, replace the tool name literal `"explore_api"` with `"list_api"` in three places: the `allowedTools` lookup, the `mcp.NewTool(applyPrefix(prefix, ...), ...)` argument, and the `entry.AllowList.Paths["explore_api"]` lookup
- [X] T008 [US1] In `pkg/tools/list.go`, update the function doc comment from `// RegisterExploreTools registers the explore_api MCP tool.` to `// RegisterListTools registers the list_api MCP tool.`
- [X] T009 [US1] In `pkg/openapimcp/server.go`, replace the single call `tools.RegisterExploreTools(s.mcpSrv, s.registry, effectivePrefix, allowedTools)` with `tools.RegisterListTools(s.mcpSrv, s.registry, effectivePrefix, allowedTools)`
- [X] T010 [US1] In `tests/integration/http_tools_test.go`, update the helper `makeTestClientFull`: replace `tools.RegisterExploreTools(srv, reg, prefix, allowedTools)` with `tools.RegisterListTools(srv, reg, prefix, allowedTools)`
- [X] T011 [US1] In `tests/integration/http_tools_test.go`, replace every remaining occurrence of the string literal `"explore_api"` with `"list_api"` (tool-name arguments in `callTool`/`callToolRaw`, prefixed names such as `"test_explore_api"` → `"test_list_api"`, and any sub-test names or comments that quote the old name)
- [X] T012 [US1] Run `go build ./...` and `go test ./tests/integration/ -count=1` to confirm User Story 1 compiles and all existing integration tests pass with the new name

**Checkpoint**: User Story 1 is fully functional. The server exposes `list_api` (and only `list_api`); every integration test that previously exercised `explore_api` now exercises `list_api` and passes. Tool-prefix handling works with the new base name.

---

## Phase 4: User Story 2 - Restrict the Renamed Tool via Allow-List (Priority: P2)

**Goal**: `allow_list.tools` and `allow_list.paths` accept `list_api` and reject `explore_api` at startup with an error that points operators to the new name.

**Independent Test**: Build a `config.Config` with `allow_list.tools: ["list_api"]` and confirm `Validate()` returns nil. Build one with `allow_list.tools: ["explore_api"]` and confirm `Validate()` returns an error whose message contains `list_api` in the hint list. Do the same for `allow_list.paths` keys. This story does not depend on the server being running — it is a pure config-layer test.

### Implementation for User Story 2

- [X] T013 [US2] In `tests/unit/config_test.go`, update the three existing allow-list test cases that use `"explore_api"` to use `"list_api"` instead, and add a new table case titled `"legacy explore_api in allow_list.tools rejected"` that passes `config.APIAllowList{Tools: []string{"explore_api"}}` and asserts the returned error message contains both the string `explore_api` (the offending value) and `list_api` (the replacement hint)
- [X] T014 [US2] In `tests/unit/config_test.go`, add a table case titled `"legacy explore_api in allow_list.paths rejected"` that passes `config.APIAllowList{Tools: []string{"list_api"}, Paths: map[string][]string{"explore_api": {"/pets"}}}` and asserts the returned error message contains both `explore_api` and `list_api`
- [X] T015 [US2] Run `go test ./tests/unit/ -run TestConfigValidate -count=1` to confirm the unit tests pass; the startup rejection flow and the new hint are verified end-to-end

**Checkpoint**: Both user stories are independently functional. `go test ./...` passes. An operator migrating a stale config sees a clear error naming `list_api` as the replacement.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: User-facing documentation and the contract test comment — everything a new reader sees.

- [X] T016 [P] In `README.md`, replace all five occurrences of `explore_api` with `list_api` (the tools table row, the progressive-discovery prose, the example flow diagram line, and the two `tool_prefix` example comments such as `store_list_api`, `pay_list_api`)
- [X] T017 [P] In `config.yaml.example`, replace the single occurrence of `explore_api` in the `allow_list.tools` example comment with `list_api`
- [X] T018 [P] In `tests/contract/mcp_tools_contract_test.go`, update the `TestPathInfoShape` doc comment from `// TestPathInfoShape verifies the PathInfo JSON shape used by explore_api.` to `// TestPathInfoShape verifies the PathInfo JSON shape used by list_api.`
- [X] T019 Run `go vet ./...` and `go test ./...` and confirm clean output; then grep the live code, tests, and user-facing docs for any remaining `explore_api` occurrences (`grep -rn explore_api pkg/ cmd/ tests/ README.md config.yaml.example`) and confirm the only matches are inside historical spec files under `specs/001-…` through `specs/014-…`, which are deliberately frozen per `research.md` Decision 2

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Baseline)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1
- **Phase 3 (US1)**: Depends on Phase 2 (config must accept `list_api` before the tool can be registered with that key at startup)
- **Phase 4 (US2)**: Depends on Phase 2 (pure config-layer change; can run in parallel with Phase 3 if desired — see below)
- **Phase 5 (Polish)**: Depends on US1 and US2 complete

### Within Phase 3 (US1)

```
T005 ─── T006 ─── T007 ─── T008 ─── T009 ─── T010 ─── T011 ─── T012
```

All sequential: T005–T008 modify the same file (`pkg/tools/list.go` after the rename); T009 depends on the new function name existing; T010–T011 depend on the new function name being imported; T012 is the phase gate.

### Within Phase 4 (US2)

```
T013 ─── T014 ─── T015
```

T013 and T014 both modify `tests/unit/config_test.go`, so they are sequential. T015 is the phase gate.

### Phase 3 and Phase 4 in parallel

Phase 4 touches only `tests/unit/config_test.go`; Phase 3 touches `pkg/tools/`, `pkg/openapimcp/`, and `tests/integration/`. Once Phase 2 lands, an implementer can run Phase 3 and Phase 4 concurrently on separate branches/worktrees and merge them independently.

### Within Phase 5

- T016, T017, T018 operate on different files — parallel-eligible
- T019 depends on all prior tasks (final green-run + grep sweep)

---

## Parallel Example: Phase 5

```bash
# After all user-story work is done, launch polish tasks in parallel:
Task T016: "Rename explore_api → list_api in README.md (5 occurrences)"
Task T017: "Rename explore_api → list_api in config.yaml.example"
Task T018: "Update PathInfoShape contract-test comment"

# Then finally:
Task T019: "Run go vet + go test, grep-sweep the repo for stragglers"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 → Phase 2 → Phase 3 (T005–T012)
2. **STOP and VALIDATE**: server advertises `list_api`, not `explore_api`; all integration tests pass; `tool_prefix` applies to `list_api`
3. Ship MVP

### Incremental Delivery

1. MVP (Phases 1–3): wire-level rename works ✓
2. Phase 4: config-validation rejects the old name with a clear hint ✓
3. Phase 5: user-facing docs and polish ✓

---

## Notes

- No new files are created except the renamed `pkg/tools/list.go` (via `git mv` from `pkg/tools/explore.go`).
- No new packages or dependencies; `go.mod` is unchanged.
- Historical spec files under `specs/001-…` through `specs/014-…` are deliberately left referencing `explore_api` — they are frozen records. Only `specs/015-rename-explore-to-list/` and the spec-checklist file may mention the old name for migration/context purposes.
- The `PathInfo` Go type keeps its name: it describes what the value is, not which tool returns it.
- The tool's inputs, outputs, error envelope, and allow-list semantics are byte-identical to the pre-rename behaviour — any behavioural test that passed before must pass after.
