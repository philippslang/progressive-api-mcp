---
description: "Task list for Fix PATCH Endpoint OpenAPI Schema — Dual Patch Format Support"
---

# Tasks: Fix PATCH Endpoint OpenAPI Schema — Dual Patch Format Support

**Input**: Design documents from `/specs/012-fix-patch-schema/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/patch-schema.yaml ✓, quickstart.md ✓

**Organization**: Tasks are grouped by user story. This is a focused 3-file bug fix — no new dependencies required.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to
- All tasks include exact file paths

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm working environment before editing; no new project scaffolding required

- [x] T001 Verify `go test ./...` baseline passes and note any pre-existing failures before making changes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: YAML schema fix — prerequisite for both user stories. Must complete before US1 or US2 can be validated independently.

**⚠️ CRITICAL**: Both user stories depend on this phase. US1 requires the new `oneOf` schema to accept RFC 6902 arrays; US2 requires `PetPatch` to remain present in the `oneOf` so merge-patch still validates.

- [x] T002 Add `JsonPatchOp` and `JsonPatchDocument` component schemas to `tests/testdata/petstore.yaml` under `components.schemas` (per contracts/patch-schema.yaml and data-model.md)
- [x] T003 Change `PATCH /pets/{id}` `requestBody` schema in `tests/testdata/petstore.yaml` from `$ref: "#/components/schemas/PetPatch"` to `oneOf: [JsonPatchDocument, PetPatch]` (depends on T002)

**Checkpoint**: `tests/testdata/petstore.yaml` now declares both patch formats via `oneOf`. Schema validation by libopenapi-validator will accept both RFC 6902 arrays and merge-patch objects.

---

## Phase 3: User Story 1 — Send a valid JSON Patch request via MCP tool (Priority: P1) 🎯 MVP

**Goal**: An RFC 6902 JSON Patch array `[{"op":"replace","path":"/name","value":"Rex"}]` sent through the MCP tool reaches the mock server, is processed, and returns 200 OK.

**Independent Test**: Run `go test ./tests/integration/... -run TestHTTPToolsEndToEnd/http_patch_rfc6902_array_body` and confirm 200 response.

### Implementation for User Story 1

- [x] T004 [US1] Add `jsonPatchOp` struct and ensure `"bytes"` is imported in `tests/mockservers/petstore/main.go` (per quickstart.md handler spec)
- [x] T005 [US1] Rewrite `patchPet` handler in `tests/mockservers/petstore/main.go` to detect body shape: decode to `json.RawMessage`, inspect first non-whitespace byte — `[` → RFC 6902 apply loop; `{` → existing merge-patch path (per quickstart.md full handler code)
- [x] T006 [P] [US1] Add `http_patch rfc6902 array body` sub-test after the existing `http_patch valid body` sub-test in `tests/integration/http_tools_test.go` (per plan.md Phase C and quickstart.md)

**Checkpoint**: `go test ./tests/integration/... -run TestHTTPToolsEndToEnd/http_patch_rfc6902_array_body` passes with status 200. User Story 1 is fully functional.

---

## Phase 4: User Story 2 — Existing valid requests continue to work (Priority: P2)

**Goal**: Non-PATCH endpoints and the merge-patch path on PATCH are unaffected. The existing `http_patch valid body` test (sends `{"name": "Rex"}`) continues to pass.

**Independent Test**: Run `go test ./tests/integration/... -run TestHTTPToolsEndToEnd/http_patch_valid_body` and confirm it still returns 200.

### Implementation for User Story 2

No code changes required for US2. The `oneOf` schema (T003) preserves `PetPatch`, and the rewritten handler (T005) preserves the merge-patch branch. This phase is verification only.

- [x] T007 [US2] Run `go test ./tests/integration/... -run TestHTTPToolsEndToEnd` and confirm both `http_patch valid body` (merge-patch) and `http_patch rfc6902 array body` (RFC 6902) sub-tests pass
- [x] T008 [P] [US2] Run `go test ./tests/integration/... -run TestHTTPToolsEndToEnd` excluding PATCH sub-tests (GET, POST, DELETE groups) and confirm all pass unchanged

**Checkpoint**: All previously-passing integration tests still pass. No regressions introduced.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Full test suite validation and lint check

- [x] T009 Run `go test ./...` and confirm zero failures across all packages
- [x] T010 [P] Run `golangci-lint run` and confirm no new lint issues introduced

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — run immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — **BLOCKS both user stories**
  - T002 → T003 (sequential; T003 references schemas added by T002)
- **US1 (Phase 3)**: Depends on Phase 2 completion
  - T004 → T005 (struct must exist before handler rewrite)
  - T006 is [P] — can be written in parallel with T004/T005 (different file)
- **US2 (Phase 4)**: Depends on Phase 2 + Phase 3 completion (verification requires full implementation)
  - T007 and T008 are [P] — can run simultaneously
- **Polish (Phase 5)**: Depends on all story phases complete
  - T009 and T010 are [P] — can run simultaneously

### User Story Dependencies

- **US1 (P1)**: Depends on Foundational (Phase 2). No dependency on US2.
- **US2 (P2)**: Depends on Foundational (Phase 2) and US1 (Phase 3) — US2 is verification that US1's handler change preserves merge-patch behavior.

### Within Each Phase

- T002 → T003 (same file; T003 references schemas from T002)
- T004 → T005 (same file; struct from T004 used in T005's handler)
- T006 is independent of T004/T005 (different file: `http_tools_test.go` vs `main.go`)

---

## Parallel Example: User Story 1

```bash
# T006 can run in parallel with T004+T005 (different files):
Task T004+T005: "Extend patchPet handler in tests/mockservers/petstore/main.go"
Task T006:      "Add rfc6902 integration sub-test in tests/integration/http_tools_test.go"

# After both complete, verify:
go test ./tests/integration/... -run TestHTTPToolsEndToEnd/http_patch
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Verify baseline
2. Complete Phase 2: YAML schema fix (T002 → T003)
3. Complete Phase 3: Handler extension + new test (T004 → T005, T006 in parallel)
4. **STOP and VALIDATE**: `go test ./tests/integration/... -run TestHTTPToolsEndToEnd/http_patch_rfc6902_array_body`
5. US1 is fully delivered — RFC 6902 path works end-to-end

### Full Delivery (Both Stories)

1. MVP above
2. Phase 4: Run verification tests — confirm merge-patch and other endpoints unchanged
3. Phase 5: `go test ./...` + lint — zero failures, zero new issues

---

## Notes

- All 3 changed files are in `tests/` — no `pkg/` packages touched (CLAUDE.md constraint satisfied)
- No new Go module dependencies — `bytes` is stdlib (CLAUDE.md constraint satisfied)
- The handler rewrite in T005 uses the exact code from `quickstart.md` — follow it precisely
- The `oneOf` in T003 must list `JsonPatchDocument` first (array) then `PetPatch` (object) to match libopenapi-validator discrimination order per research.md Decision 1
- Empty array `[]` is a valid `JsonPatchDocument` — handler must treat it as a no-op (no mutations, return 200)
