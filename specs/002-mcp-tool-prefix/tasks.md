---
description: "Task list for MCP Tool Name Prefix"
---

# Tasks: MCP Tool Name Prefix

**Input**: Design documents from `/specs/002-mcp-tool-prefix/`
**Prerequisites**: plan.md ✅ | spec.md ✅ | research.md ✅ | data-model.md ✅ | contracts/ ✅

**Tests**: Included — TDD is mandatory per the project constitution (Principle II).
Tests MUST be written and confirmed failing before implementation begins.

**Scope**: Amendment to `001-openapi-mcp-server`. No new source files; changes confined to
`pkg/config/config.go`, `pkg/openapimcp/server.go`, `pkg/tools/` (Register* signatures),
and `cmd/prograpimcp/main.go`.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US2)

---

## Phase 1: Setup

**Purpose**: No new project setup needed — this is an amendment. One documentation update.

- [X] T001 Update `config.yaml.example` to document the new `tool_prefix` field with comment at repository root

---

## Phase 2: Foundational

**Purpose**: Add `ToolPrefix` to `ServerConfig` with validation. All story phases depend on this.

**⚠️ CRITICAL**: Story phases cannot begin until this phase is complete.

### Tests ⚠️ Write first — confirm they FAIL before implementing

- [X] T002 Add `ToolPrefix` validation tests to `tests/unit/config_test.go`: valid prefix `"myapi"`, prefix with trailing `_` → stripped silently, empty string → valid (no prefix), starts with digit → error, contains hyphen → error, contains space → error, purely numeric `"123"` → error, starts with letter+digits `"v2svc"` → valid, starts with underscore `"_internal"` → valid

### Implementation

- [X] T003 Add `ToolPrefix string` field with YAML tag `tool_prefix` to `ServerConfig` struct in `pkg/config/config.go`
- [X] T004 Add `ToolPrefix` validation block to `Config.Validate()` in `pkg/config/config.go`: strip trailing `_`, if non-empty must match `^[a-zA-Z_][a-zA-Z0-9_]*$`, return descriptive error identifying invalid characters on failure

**Checkpoint**: `go test ./tests/unit/...` MUST pass before proceeding to story phases.

---

## Phase 3: User Story 1 — Configure Tool Name Prefix at Startup (Priority: P1) 🎯 MVP

**Goal**: When `tool_prefix` is set in config, all 6 MCP tools are registered with the prefix
prepended. When unset, tools use original names. Prefix is applied at `Start()` time.

**Independent Test**: Start server with `tool_prefix: "myapi"` in config, list registered
tools via MCP, verify all 6 names start with `myapi_` and calling `myapi_http_get` succeeds.

### Tests for US1 ⚠️ Write first — confirm they FAIL before implementing

- [X] T005 [US1] Add integration test `TestToolPrefix` to `tests/integration/http_tools_test.go`: load Petstore with `ToolPrefix: "test"`, call `test_explore_api` via InProcess transport, verify `PathInfo` array returned; also verify `explore_api` (unprefixed) returns tool-not-found error
- [X] T006 [P] [US1] Add integration test `TestNoPrefixDefaultBehavior` to `tests/integration/http_tools_test.go`: load Petstore with empty `ToolPrefix`, verify original tool names (`explore_api`, `http_get`) still work
- [X] T007 [P] [US1] Add integration test `TestTrailingUnderscoreStripped` to `tests/integration/http_tools_test.go`: configure `ToolPrefix: "myapi_"`, verify tools registered as `myapi_http_get` (single `_`), NOT `myapi__http_get`

### Implementation

- [X] T008 [US1] Add `prefix string` parameter to `tools.RegisterHTTPTools` signature in `pkg/tools/http.go` and update call sites in `pkg/openapimcp/server.go`; inside `RegisterHTTPTools`, prepend `applyPrefix(prefix, "http_get")` etc. to the name passed to `mcp.NewTool`
- [X] T009 [US1] Add `prefix string` parameter to `tools.RegisterExploreTools` in `pkg/tools/explore.go`; apply prefix to `"explore_api"` tool name
- [X] T010 [US1] Add `prefix string` parameter to `tools.RegisterSchemaTools` in `pkg/tools/schema.go`; apply prefix to `"get_schema"` tool name
- [X] T011 [US1] Implement helper `applyPrefix(prefix, name string) string` in `pkg/openapimcp/server.go`: strips trailing `_` from prefix, returns `prefix + "_" + name` if non-empty, else returns `name`
- [X] T012 [US1] Update `Server.Start()` in `pkg/openapimcp/server.go` to compute `effectivePrefix := applyPrefix(s.cfg.Server.ToolPrefix, "")` (or inline), then pass it to all three `Register*` calls
- [X] T013 [US1] Update startup log line in `pkg/openapimcp/server.go` (or `cmd/prograpimcp/main.go`) to print effective prefix: `"[prograpimcp] tool prefix: myapi (6 tools registered)"` or `"[prograpimcp] tool prefix: none"` when empty

**Checkpoint**: `go test ./...` MUST pass. US1 independently testable.

---

## Phase 4: User Story 2 — Override Prefix via CLI Argument (Priority: P2)

**Goal**: `--tool-prefix` CLI flag and `PROGRAPIMCP_SERVER_TOOL_PREFIX` env var override the
config file value, following the same precedence as `--host`, `--port`, `--transport`.

**Independent Test**: Start with config `tool_prefix: "fromfile"` and flag
`--tool-prefix fromcli`, verify tools registered as `fromcli_*`.

### Tests for US2 ⚠️ Write first — confirm they FAIL before implementing

- [X] T014 [US2] Add unit test `TestToolPrefixPrecedence` to `tests/unit/config_test.go`: construct `Config` with `ToolPrefix: "fromfile"`, override via viper simulation to `"fromcli"`, verify effective prefix is `"fromcli"` (this tests the wiring logic, not the CLI itself)

### Implementation

- [X] T015 [US2] Add `--tool-prefix` flag to cobra root command in `cmd/prograpimcp/main.go`: `cmd.Flags().String("tool-prefix", "", "MCP tool name prefix (overrides config)")`
- [X] T016 [US2] Bind `--tool-prefix` flag to viper in `cmd/prograpimcp/main.go`: `viper.BindPFlag("server.tool_prefix", cmd.Flags().Lookup("tool-prefix"))` — placed alongside existing `BindPFlag` calls for `host`, `port`, `transport`
- [X] T017 [US2] Apply viper override in the CLI's config-loading block in `cmd/prograpimcp/main.go`: after `config.LoadFile`, if `cmd.Flags().Changed("tool-prefix")`, set `cfg.Server.ToolPrefix = viper.GetString("server.tool_prefix")` — same pattern as `host`, `port`, `transport` overrides

**Checkpoint**: `go test ./...` MUST pass. Both user stories complete and independently testable.

---

## Phase 5: Polish

- [X] T018 [P] Run `go build ./...` to confirm all packages compile cleanly after signature changes to `Register*` functions
- [X] T019 [P] Verify all exported symbols touched by this feature have godoc comments: `ServerConfig.ToolPrefix` field comment in `pkg/config/config.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS story phases
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 3 (CLI wiring calls the same `cfg.Server.ToolPrefix` set by Phase 3)
- **Phase 5 (Polish)**: Depends on Phase 4

### Within Each Phase

- Tests marked ⚠️ MUST be written and confirmed FAILING before implementation
- `pkg/config` changes (T003–T004) before `pkg/openapimcp` changes (T008–T013)
- `Register*` signature changes (T008–T010) before server wiring (T011–T012)

---

## Parallel Opportunities

```
Phase 2 tests: T002 only (single test file)

Phase 3 tests: T005, T006, T007 — all can be written in parallel (same file, non-conflicting)

Phase 3 implementation:
  T008, T009, T010 — parallel (different files: http.go, explore.go, schema.go)
  T011 → T012 → T013 — sequential (helper before caller, caller before log)

Phase 4: T015, T016, T017 — sequential (same file: main.go)

Phase 5: T018, T019 — parallel
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1 (T001)
2. Complete Phase 2 (T002–T004) — foundational, BLOCKING
3. Complete Phase 3 (T005–T013) — prefix in config + applied at registration
4. **STOP and VALIDATE**: `go test ./...` passes; smoke test with `config.yaml` + prefix

### Full Delivery

1. MVP above
2. Phase 4 (T014–T017) — CLI flag + env var override
3. Phase 5 (T018–T019) — polish

---

## Notes

- `[P]` tasks touch different files — safe to parallelize
- Tests marked ⚠️ follow Red-Green-Refactor: write → confirm fail → implement → confirm pass
- `Register*` function signature changes affect 3 files simultaneously — coordinate carefully
- The `applyPrefix` helper belongs in `pkg/openapimcp/server.go` (not in `pkg/tools/`) because
  it is a server-level concern, not a tool-level concern
