# Tasks: API-Level Skip Validation Configuration

**Input**: Design documents from `/specs/013-api-skip-validation/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Exact file paths are included in all descriptions

---

## Phase 1: Setup

**Purpose**: No new project setup required — this feature adds fields to an existing Go project with no new dependencies or files. Phase 1 is a single baseline verification step.

- [x] T001 Run `go test ./...` to confirm the baseline is green before making any changes

---

## Phase 2: Foundational (Blocking Prerequisite)

**Purpose**: Add the `SkipValidation` field to `APIConfig`. This is the source of truth for the feature and MUST exist before any other task can be started.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T002 Add `SkipValidation bool \`yaml:"skip_validation"\`` field to the `APIConfig` struct in `pkg/config/config.go`

**Checkpoint**: `APIConfig` now has the `skip_validation` field; `go build ./...` must pass.

---

## Phase 3: User Story 1 - Enable Skip Validation for a Specific API (Priority: P1) 🎯 MVP

**Goal**: When `skip_validation: true` is set for an API, tool calls with schema-invalid payloads bypass validation and reach the upstream API.

**Independent Test**: Configure a single API with `skip_validation: true`, start the MCP server against a test HTTP server, send a payload that violates the OpenAPI schema, and confirm no validation error is returned and the upstream server receives the request.

### Implementation for User Story 1

- [x] T003 [P] [US1] Add `SkipValidation bool` field to the `APIEntry` struct in `pkg/registry/registry.go`
- [x] T004 [P] [US1] Add unit test table cases for `skip_validation` YAML parsing (true/false/absent → correct bool) in `tests/unit/config_test.go`
- [x] T005 [US1] Set `entry.SkipValidation = cfg.SkipValidation` during `Registry.Load()` in `pkg/registry/registry.go` (depends on T003)
- [x] T006 [US1] Add conditional guard in `validateAndExecute()` in `pkg/tools/http.go`: skip `entry.Validator.Validate(req)` when `entry.SkipValidation` is `true` (depends on T005)
- [x] T007 [US1] Add integration test in `tests/integration/http_tools_test.go` verifying that an invalid payload is forwarded to the upstream when `skip_validation: true` and that a validation error is still returned when `skip_validation: false` (depends on T006)

**Checkpoint**: User Story 1 is fully functional. `go test ./...` passes. Invalid payloads pass through for the skip-enabled API; valid behavior unchanged for non-skip APIs.

---

## Phase 4: User Story 2 - Per-API Independent Configuration (Priority: P2)

**Goal**: `skip_validation` on one API does not affect validation behaviour on any other API configured in the same server instance.

**Independent Test**: Configure two APIs in one server — one with `skip_validation: true`, one without. Send an invalid payload to each. Confirm the first forwards and the second rejects.

### Implementation for User Story 2

- [x] T008 [US2] Add integration test in `tests/integration/http_tools_test.go` verifying per-API isolation: invalid payload forwarded for skip-enabled API, rejected for non-skip API in the same server instance (depends on T006)

**Checkpoint**: Both user stories are independently functional. `go test ./...` passes.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Verify code quality and backward compatibility across the whole change set.

- [x] T009 [P] Add doc comment to the `SkipValidation` field in `APIConfig` in `pkg/config/config.go`
- [x] T010 [P] Add doc comment to the `SkipValidation` field in `APIEntry` in `pkg/registry/registry.go`
- [x] T011 Run `go test ./...` and confirm all tests pass (no regressions)
- [x] T012 Run `golangci-lint run` and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Baseline)**: No dependencies — run immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user story work
- **Phase 3 (US1)**: Depends on Phase 2 completion
- **Phase 4 (US2)**: Depends on Phase 3 T006 completion (needs the skip guard in place)
- **Phase 5 (Polish)**: Depends on all user story phases complete

### Within Phase 3

```
T002 (Foundational)
  └── T003 [P] ─── T005 ─── T006 ─── T007
  └── T004 [P]
```

- T003 and T004 can start in parallel immediately after T002 (different files: `registry.go` vs `config_test.go`)
- T005 depends on T003
- T006 depends on T005
- T007 depends on T006

### Within Phase 4

- T008 depends on T006 (integration test requires the skip guard to exist)

### Within Phase 5

- T009 and T010 are parallel (different files)
- T011 depends on T009 and T010
- T012 depends on T011

---

## Parallel Example: Phase 3

```bash
# After T002 completes, launch T003 and T004 in parallel:
Task T003: "Add SkipValidation bool to APIEntry in pkg/registry/registry.go"
Task T004: "Add unit test table cases for skip_validation in tests/unit/config_test.go"

# After T003 completes:
Task T005: "Set entry.SkipValidation in Registry.Load() in pkg/registry/registry.go"

# After T005 completes:
Task T006: "Add conditional guard in validateAndExecute() in pkg/tools/http.go"

# After T006 completes, launch T007 and T008 in parallel (different test functions in same file — do sequentially to avoid conflicts):
Task T007: "Add integration test for skip bypass in tests/integration/http_tools_test.go"
Task T008: "Add integration test for per-API isolation in tests/integration/http_tools_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Baseline check
2. Complete Phase 2: Add `SkipValidation` to `APIConfig` (T002)
3. Complete Phase 3: T003 → T004 (parallel) → T005 → T006 → T007
4. **STOP and VALIDATE**: Run `go test ./...` — US1 is fully working
5. Ship if sufficient

### Incremental Delivery

1. MVP (Phase 1–3): Single-API skip validation working ✓
2. Isolation proof (Phase 4): Multi-API independence verified ✓
3. Polish (Phase 5): Lint-clean, doc-commented code ✓

---

## Notes

- [P] tasks operate on different files — no merge conflicts
- T007 and T008 both modify `tests/integration/http_tools_test.go` — do NOT run in parallel
- The skip guard is a **one-line conditional** wrapping the existing `entry.Validator.Validate(req)` call — keep it minimal
- Backward compatibility is guaranteed by Go's zero-value bool (`false`) — no migration needed
- No new packages, no new dependencies, no new files required
