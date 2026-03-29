# Tasks: Ignore Headers for APIs

**Input**: Design documents from `/specs/009-ignore-api-headers/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup

**Purpose**: No new project or dependencies needed. Verify the branch and working tree are clean before starting.

- [x] T001 Confirm branch `009-ignore-api-headers` is checked out and `go build ./...` passes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add `ignore_headers` to the config struct and validate it. All three user stories depend on this.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T002 Add `IgnoreHeaders []string` field (yaml: `ignore_headers`) to `config.APIConfig` in `pkg/config/config.go`
- [x] T003 Update `config.Validate()` in `pkg/config/config.go` to return an error when any entry in `IgnoreHeaders` is an empty string, with message format `apis[%d].ignore_headers[%d]: header name must not be empty`

**Checkpoint**: `go build ./...` and `go test ./pkg/config/...` pass. Config accepts and rejects `ignore_headers` entries correctly.

---

## Phase 3: User Story 1 — Configure Header Suppression Per API (Priority: P1) 🎯 MVP

**Goal**: At runtime, each `APIEntry` holds a case-insensitive set of ignored header names built from the config. This is the data foundation for schema filtering and validation bypass.

**Independent Test**: Start the server with a config containing `ignore_headers: [Authorization]` for an API. Server starts without error. No observable behaviour change in tool output yet — that is US2/US3.

### Implementation for User Story 1

- [x] T004 [US1] Add `IgnoreHeaders map[string]struct{}` field to `registry.APIEntry` in `pkg/registry/registry.go`
- [x] T005 [US1] In `registry.Load()` in `pkg/registry/registry.go`, build the `IgnoreHeaders` map by iterating `cfg.IgnoreHeaders`, lowercasing each name with `strings.ToLower`, and storing in the map; assign to `entry.IgnoreHeaders`

**Checkpoint**: `go build ./...` passes. `registry.APIEntry.IgnoreHeaders` is populated after `Load` when config carries `ignore_headers` values.

---

## Phase 4: User Story 2 — Hidden Headers in Schema Exploration (Priority: P2)

**Goal**: `get_schema` exposes header parameters for non-ignored headers under a new `header_parameters` field. Ignored headers are absent from the response entirely.

**Independent Test**: Call `get_schema` on an endpoint that declares a required `Authorization` header, with `Authorization` in `ignore_headers`. The response must not contain `Authorization` in any parameter field. Call `get_schema` on an endpoint with a non-ignored header `X-Request-ID`; it must appear under `header_parameters`.

### Implementation for User Story 2

- [x] T006 [US2] Add `HeaderParameters map[string]any` field (json: `header_parameters,omitempty`) to `tools.SchemaResult` in `pkg/tools/schema.go`
- [x] T007 [US2] Extend the parameter extraction loop in `RegisterSchemaTools` in `pkg/tools/schema.go` to handle `param.In == "header"`: build a `paramInfo` map (same structure as path/query params) and collect into a local `headerParams` map
- [x] T008 [US2] After collecting `headerParams`, filter out any key where `strings.ToLower(key)` is present in `entry.IgnoreHeaders`, then assign the remaining non-empty map to `result.HeaderParameters` in `pkg/tools/schema.go`

**Checkpoint**: `go test ./pkg/tools/...` passes. `get_schema` on an endpoint with mixed ignored/non-ignored headers returns only non-ignored ones under `header_parameters`.

---

## Phase 5: User Story 3 — Validation Skips Ignored Headers (Priority: P3)

**Goal**: HTTP tool calls that omit an ignored required header pass validation. Calls that supply an ignored header still forward it to upstream.

**Independent Test**: Call `http_get` (or any HTTP tool) on an endpoint that declares a required header that is in `ignore_headers`, without supplying that header. Validation must pass and the request must reach `executeHTTP`. Verify a header that is NOT in `ignore_headers` and is required still fails validation when omitted.

### Implementation for User Story 3

- [x] T009 [US3] In `validateAndExecute` in `pkg/tools/http.go`, after the synthetic `req` is built and its caller-supplied headers are set, iterate over `entry.IgnoreHeaders`; for each name not already present in `req.Header` (checked case-insensitively), call `req.Header.Set(name, "*")` so the validator treats it as supplied

**Checkpoint**: `go test ./...` passes. `validateAndExecute` no longer returns `VALIDATION_FAILED` for missing ignored required headers.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Ensure the feature is observable, documented, and leaves no rough edges.

- [x] T010 [P] Update `config.yaml` (root example config) to add an `ignore_headers` comment block showing the field syntax
- [x] T011 [P] Add `ignore_headers` field description to `CLAUDE.md` under Active Technologies / recent changes (run `.specify/scripts/bash/update-agent-context.sh claude` after)
- [x] T012 Run `go test ./...` to confirm the full test suite passes with no regressions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user story phases
- **Phase 3 (US1)**: Depends on Phase 2 — no dependency on US2 or US3
- **Phase 4 (US2)**: Depends on Phase 3 (needs `entry.IgnoreHeaders` on `APIEntry`)
- **Phase 5 (US3)**: Depends on Phase 3 (needs `entry.IgnoreHeaders` on `APIEntry`)
- **Phase 6 (Polish)**: Depends on Phases 3, 4, 5

### User Story Dependencies

- **US1 (P1)**: Only depends on Foundational phase
- **US2 (P2)**: Depends on US1 (reads `entry.IgnoreHeaders`)
- **US3 (P3)**: Depends on US1 (reads `entry.IgnoreHeaders`); US2 and US3 are independent of each other and can proceed in parallel after US1

### Within Each User Story

- Struct field additions before logic that uses them
- Config/registry changes before tool-layer changes (US1 before US2/US3)

### Parallel Opportunities

- T006, T007, T008 (US2 schema) and T009 (US3 validation) can run in parallel once T005 (US1) is complete — they touch different files (`schema.go` vs `http.go`)
- T010 and T011 (Polish) can run in parallel

---

## Parallel Example: After US1 completes

```
# US2 (schema.go) and US3 (http.go) can proceed simultaneously:
Task T006: Add HeaderParameters field to SchemaResult in pkg/tools/schema.go
Task T009: Inject placeholder headers in validateAndExecute in pkg/tools/http.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational — T002, T003
3. Complete Phase 3: US1 — T004, T005
4. **STOP and VALIDATE**: Server starts with `ignore_headers` config; build passes
5. Proceed to US2/US3

### Incremental Delivery

1. T001–T003: Config accepts `ignore_headers` → Foundation ready
2. T004–T005: Registry carries the set at runtime → US1 complete (MVP)
3. T006–T008: `get_schema` hides ignored headers → US2 complete
4. T009: HTTP tools skip ignored required headers → US3 complete (feature fully functional)
5. T010–T012: Polish → Ship

---

## Notes

- [P] tasks touch different files and have no shared incomplete dependencies
- US2 (`schema.go`) and US3 (`http.go`) are fully independent after US1; parallelize freely
- The `entry.IgnoreHeaders` map uses lowercase keys — always `strings.ToLower` before lookup
- Placeholder value `"*"` in T009 is injected into the synthetic validation request only; it is never forwarded to the upstream API
- No new packages, no new go.mod entries, no new binaries
