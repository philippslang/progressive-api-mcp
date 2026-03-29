# Tasks: Body-Only HTTP Responses

**Input**: Design documents from `/specs/010-body-only-response/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)

---

## Phase 1: Setup

- [x] T001 Confirm branch `010-body-only-response` is checked out and `go build ./...` passes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add `response_body_only` to the config struct. US1 and US2 both depend on this.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T002 Add `ResponseBodyOnly bool` field (yaml: `response_body_only`) to `config.APIConfig` in `pkg/config/config.go`; no validation beyond YAML bool parsing is required

**Checkpoint**: `go build ./...` passes. Config struct accepts `response_body_only` from YAML.

---

## Phase 3: User Story 1 — Configure Body-Only Response Per API (Priority: P1) 🎯 MVP

**Goal**: When `response_body_only: true` is set for an API, all HTTP tool calls return only the response body instead of the full `{ status_code, headers, body }` envelope.

**Independent Test**: Configure a test API with `response_body_only: true`, call `http_get` on any endpoint, and verify the result is the raw body value — not an object with `status_code` or `headers` fields.

### Implementation for User Story 1

- [x] T003 [US1] In `executeHTTP` in `pkg/tools/http.go`, after `bodyVal` is computed and before `HTTPResult` is assembled, add a branch: if `entry.Config.ResponseBodyOnly` is `true`, marshal `bodyVal` directly and return without constructing `HTTPResult`

**Checkpoint**: `go build ./...` and `go test ./...` pass. HTTP tools return body-only output for APIs with `response_body_only: true`.

---

## Phase 4: User Story 2 — Body-Only Works Across All HTTP Tools (Priority: P2)

**Goal**: Confirm the US1 implementation already covers all four HTTP tools (`http_get`, `http_post`, `http_put`, `http_patch`) uniformly, since all four call through `executeHTTP`.

**Independent Test**: Enable `response_body_only` for an API and call each of the four HTTP tools; each must return body-only output.

### Implementation for User Story 2

- [x] T004 [US2] Verify in `pkg/tools/http.go` that all four tool handlers (`http_get`, `http_post`, `http_put`, `http_patch`) route through `validateAndExecute` → `executeHTTP`; no per-tool changes should be needed — document the verification as a code comment if needed

**Checkpoint**: `go test ./...` passes for all tools. No per-tool code changes needed if `executeHTTP` is the single dispatch point (expected).

---

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T005 [P] Update `config.yaml` (root example config) to add a `response_body_only` commented example line showing the field syntax
- [x] T006 Run `go test ./...` to confirm the full test suite passes with no regressions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user story phases
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 3 (verification that all tools use `executeHTTP`)
- **Phase 5 (Polish)**: Depends on Phases 3, 4

### User Story Dependencies

- **US1 (P1)**: Depends only on Foundational phase
- **US2 (P2)**: Logically depends on US1 completion for confidence; in practice it is a verification step

### Parallel Opportunities

- T005 (config.yaml update) can run in parallel with T006 (test run) — different files

---

## Parallel Example: Polish Phase

```
# T005 and T006 touch different files:
Task T005: Update config.yaml with response_body_only comment
Task T006: Run go test ./...
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. T001: Verify branch
2. T002: Add `ResponseBodyOnly` to `APIConfig`
3. T003: Add branch in `executeHTTP`
4. **STOP and VALIDATE**: `go build ./...` and `go test ./...` pass; body-only mode works for `http_get`
5. Proceed to T004–T006

### Incremental Delivery

1. T001–T002: Config field in place → Foundation ready
2. T003: Response shape conditional in `executeHTTP` → Feature complete (US1 = US2 for this feature)
3. T004: Verification pass → Confidence confirmed
4. T005–T006: Polish → Ship

---

## Notes

- The entire feature is ~5 lines: one `bool` field in `config.go` and one early-return branch in `executeHTTP`
- US2 is a verification story, not a new implementation — if T003 is correct, T004 requires no code changes
- `bodyVal` in `executeHTTP` is already the right value to marshal in body-only mode (parsed JSON or raw string)
- No new packages, no new go.mod entries, no new binaries
