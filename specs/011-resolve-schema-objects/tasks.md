# Tasks: Fully Resolved Schema Objects in get_schema

**Input**: Design documents from `/specs/011-resolve-schema-objects/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, contracts/ ✓

**Organization**: Tasks are grouped by user story. US1–US3 all touch the same function, so they are implemented sequentially in a single pass through `schemaToMap`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup

- [x] T001 Confirm branch `011-resolve-schema-objects` is checked out and `go build ./...` passes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Fix the wrong request body key name and verify the existing contract test to understand what needs updating before touching `schemaToMap`.

**⚠️ CRITICAL**: Read the existing contract test before modifying `schema.go` to understand assertions that will break.

- [x] T002 Read `tests/contract/mcp_tools_contract_test.go` to identify all assertions on `get_schema` response shape, specifically any references to `request_body["type"]`
- [x] T003 In `pkg/tools/schema.go`, change the request body schema key from `bodySchema["type"] = schemaToMap(schema)` to `bodySchema["schema"] = schemaToMap(schema)`
- [x] T004 Update `tests/contract/mcp_tools_contract_test.go` to expect `request_body["schema"]` instead of `request_body["type"]` wherever the contract test checks this field

**Checkpoint**: `go test ./tests/contract/...` passes after T003 and T004 are both done.

---

## Phase 3: User Story 1 — Full Request Body Schema (Priority: P1) 🎯 MVP

**Goal**: `get_schema` returns fully resolved object schemas for request bodies, including `properties`, per-property `required` flags, `description`, `format`, and `items` for arrays.

**Independent Test**: Call `get_schema` for `POST /pets`. The `request_body.schema` field must include a `properties` map listing each declared field with its type and required status.

### Implementation for User Story 1

- [x] T005 [US1] In `pkg/tools/schema.go`, add a new private function `schemaToMapDepth(schema *base.Schema, depth int) map[string]any` that: (1) returns an empty map if `schema == nil` or `depth > 10`; (2) extracts `type` (from `schema.Type[0]` if non-empty), `format` (if non-empty), `description` (if non-empty); (3) if `schema.Properties` is non-nil, iterates `schema.Properties.FromOldest()` to build a `requiredSet` from `schema.Required []string`, then for each property pair calls `proxy.Schema()` and recurses with `schemaToMapDepth(propSchema, depth+1)`, adding `"required": requiredSet[name]` to the property map, collecting into `"properties"`; (4) if `schema.Items != nil && schema.Items.N == 0 && schema.Items.A != nil`, recurses into `schema.Items.A.Schema()` and adds result under `"items"`; (5) returns the assembled map
- [x] T006 [US1] In `pkg/tools/schema.go`, update `schemaToMap(schema *base.Schema) map[string]any` to be a one-line wrapper: `return schemaToMapDepth(schema, 0)`

**Checkpoint**: `go build ./...` passes. `go test ./...` passes. `get_schema` on a petstore POST endpoint now shows properties in `request_body.schema`.

---

## Phase 4: User Story 2 — Full Response Schemas (Priority: P2)

**Goal**: Response schemas in `get_schema` are also fully resolved. Since `schemaToMap` is already called for response schemas, completing US1 automatically delivers US2.

**Independent Test**: Call `get_schema` for `POST /pets`. The `responses["201"].schema` field must include a `properties` map.

### Implementation for User Story 2

- [x] T007 [US2] Verify in `pkg/tools/schema.go` that the response extraction loop already calls `schemaToMap(schema)` (it does — at line ~189); since `schemaToMap` was updated in T006, no code change is needed. Confirm `go test ./...` still passes.

**Checkpoint**: Response schemas are resolved. No additional code changes required.

---

## Phase 5: User Story 3 — Nested Object Resolution (Priority: P3)

**Goal**: Schemas with nested objects are recursively resolved. Since `schemaToMapDepth` already recurses into properties, US3 is delivered by T005. This phase verifies the depth limit and circular-reference safety.

**Independent Test**: If the petstore schema has no nested objects, manually trace through `schemaToMapDepth` with a nested schema to confirm recursion is bounded. Check that depth > 10 returns an empty map rather than recursing indefinitely.

### Implementation for User Story 3

- [x] T008 [US3] Review `schemaToMapDepth` written in T005 to confirm: (a) depth guard is `depth > 10` (not `>=`), allowing 11 levels (0–10) which is sufficient; (b) the nil guard on each `proxy.Schema()` call prevents panics on unresolved proxies; no code change needed if correct

**Checkpoint**: Nested schemas resolve correctly. `go test ./...` passes.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T009 Run `go test ./...` to confirm all tests pass with no regressions
- [x] T010 [P] Review the `schemaToMapDepth` implementation for any missing nil guards (e.g. `schema.Items.A` nil check before calling `.Schema()`) and fix any found in `pkg/tools/schema.go`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user story phases
- **Phase 3 (US1)**: Depends on Phase 2 (key rename done, contract test updated)
- **Phase 4 (US2)**: Depends on Phase 3 (schemaToMap already recursive)
- **Phase 5 (US3)**: Depends on Phase 3 (same recursive function covers nesting)
- **Phase 6 (Polish)**: Depends on Phases 3, 4, 5

### User Story Dependencies

- **US1 (P1)**: Core implementation — all other stories depend on it
- **US2 (P2)**: Automatically satisfied by US1 (response extraction already uses `schemaToMap`)
- **US3 (P3)**: Automatically satisfied by US1 (recursion already handles nesting)

### Within Each Story

- T003 and T004 must be done together before running tests
- T005 before T006 (helper before wrapper)

---

## Parallel Opportunities

- T009 and T010 in the Polish phase touch different concerns and can run together

---

## Implementation Strategy

### MVP First (User Story 1)

1. T001: Verify branch
2. T002: Read contract test
3. T003–T004: Fix key name + update contract test
4. T005–T006: Implement recursive `schemaToMapDepth` and update wrapper
5. **STOP and VALIDATE**: `go test ./...` passes; `get_schema POST /pets` shows properties
6. Proceed to T007–T010

### Incremental Delivery

1. T001–T004: Foundational key fix + test alignment
2. T005–T006: Recursive schema resolution → US1 complete (and US2/US3 come for free)
3. T007–T008: Verification passes → All 3 stories confirmed
4. T009–T010: Polish → Ship

---

## Notes

- US2 and US3 require no additional code — they are delivered automatically by T005
- The key rename (T003) is a breaking change: `request_body["type"]` → `request_body["schema"]`
- `schemaToMapDepth` must handle nil returns from `.Schema()` on each SchemaProxy — always nil-check before recursing
- The `orderedmap` iterator uses `FromOldest()` consistent with other libopenapi usage in this codebase
- `schema.Items` is `*base.DynamicValue[*base.SchemaProxy, bool]`: check `Items.N == 0` before accessing `Items.A`
