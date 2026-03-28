# Tasks: Health Endpoint

**Input**: Design documents from `/specs/004-health-endpoint/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, contracts/ ✓

**Organization**: Tasks grouped by user story. Single story (P1) covers the entire feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to ([US1])

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new project structure or dependencies needed — feature adds to existing `pkg/openapimcp/server.go`.

*(No setup tasks — project already exists with all required dependencies)*

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational changes needed — no new shared packages, no schema changes, no new config fields.

*(No foundational tasks — all changes are self-contained in one file)*

---

## Phase 3: User Story 1 - Automated Health Check (Priority: P1) 🎯 MVP

**Goal**: Expose `GET /health` on the same host/port as `/mcp`, returning `{"status":"ok"}` with HTTP 200, accessible to any monitoring or orchestration tool without configuration or authentication.

**Independent Test**: Start the server (`go run ./cmd/prograpimcp --spec <any-valid-spec>`) and `curl http://127.0.0.1:8080/health` — expect `200 {"status":"ok"}`.

### Implementation for User Story 1

- [x] T001 [US1] Refactor HTTP transport in `pkg/openapimcp/server.go`: replace `httpSrv.Start(addr)` with a custom `http.ServeMux` + `http.Server`; register the `StreamableHTTPServer` as an `http.Handler` at `/mcp` and add a `/health` handler returning `{"status":"ok"}` with `Content-Type: application/json`; update the shutdown path to call `Shutdown` on the custom `http.Server`
- [x] T002 [P] [US1] Write integration test for `GET /health` in `tests/integration/health_test.go`: start a server via the library entry point and assert the health endpoint returns HTTP 200 with a JSON body containing `"status":"ok"`

**Checkpoint**: After T001 and T002, `GET /health` returns 200, existing `GET /mcp` MCP traffic continues to work, and all tests pass.

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Confirm no regressions and the binary builds cleanly.

- [x] T003 Run `go build -o /dev/null ./cmd/prograpimcp` and `go test ./...` to verify the build succeeds and all existing tests remain green

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: N/A
- **Foundational (Phase 2)**: N/A
- **User Story 1 (Phase 3)**: No dependencies — can start immediately
  - T001 and T002 are independent (different files) and can run in parallel
- **Polish (Phase 4)**: Depends on T001 and T002 completion

### Within User Story 1

- T001 and T002 are fully independent — different files, no shared state
- T003 must run after both T001 and T002

### Parallel Opportunities

- T001 and T002 can be worked simultaneously (different files: `server.go` vs `health_test.go`)

---

## Parallel Example: User Story 1

```bash
# T001 and T002 can run simultaneously:
Task T001: Modify pkg/openapimcp/server.go (HTTP transport + health handler)
Task T002: Write tests/integration/health_test.go (integration test)

# Then sequentially:
Task T003: go build + go test ./...
```

---

## Implementation Strategy

### MVP (User Story 1 Only — entire feature)

1. T001 — Refactor transport in `server.go`, add `/health` handler
2. T002 — Integration test
3. T003 — Build + test pass

**Total**: 3 tasks. Feature complete after T003.

---

## Notes

- [P] tasks = different files, no dependencies between them
- [US1] label maps all tasks to the single user story
- T001 is the only file that changes existing code; all other tasks are additive
- The `StreamableHTTPServer` already implements `http.Handler` (confirmed in research.md) — no monkey-patching or private API access needed
