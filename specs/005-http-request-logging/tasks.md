# Tasks: HTTP Request Logging

**Input**: Design documents from `/specs/005-http-request-logging/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓

**Organization**: Tasks grouped by user story. Single story (P1) covers the entire feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to ([US1])

---

## Phase 1: Setup (Shared Infrastructure)

*(No setup tasks — project exists, no new dependencies or packages needed)*

---

## Phase 2: Foundational (Blocking Prerequisites)

*(No foundational tasks — all changes are self-contained in one file)*

---

## Phase 3: User Story 1 - Request/Response Visibility (Priority: P1) 🎯 MVP

**Goal**: Every inbound HTTP request produces one log line on stderr containing method, full request URI, response status code, and elapsed time.

**Independent Test**: Start the server and send `GET /health`, `POST /mcp`, and `GET /nonexistent`. Verify three log lines appear on stderr, each showing the correct method, path, status code, and a duration value.

### Implementation for User Story 1

- [x] T001 [US1] Add unexported `responseWriter` wrapper struct and `loggingMiddleware` function to `pkg/openapimcp/server.go`; wrap the existing `mux` with `loggingMiddleware` before assigning to `httpServer.Handler`
- [x] T002 [P] [US1] Write integration test in `tests/integration/logging_test.go` that starts the server, sends a request, and asserts a log line with method, path, status code, and duration appears on stderr

**Checkpoint**: After T001 and T002, every request to any route produces a log line, existing `/health` and MCP tests still pass.

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T003 Run `go build -o /dev/null ./cmd/prograpimcp` and `go test ./...` to verify build succeeds and all tests pass

---

## Dependencies & Execution Order

- T001 and T002 are independent (different files) — can run in parallel
- T003 must run after both T001 and T002

---

## Parallel Example: User Story 1

```bash
# T001 and T002 simultaneously:
Task T001: Modify pkg/openapimcp/server.go
Task T002: Write tests/integration/logging_test.go

# Then:
Task T003: go build + go test ./...
```

---

## Implementation Strategy

1. T001 — Add `responseWriter` wrapper + `loggingMiddleware`, apply to `httpServer.Handler`
2. T002 — Integration test verifying log output
3. T003 — Full build + test pass

**Total**: 3 tasks. Feature complete after T003.

---

## Notes

- `responseWriter` wraps `http.ResponseWriter`, overrides `WriteHeader` to capture status, defaults to 200 if `WriteHeader` is never called
- Log format: `[prograpimcp] METHOD /path?query STATUS Xms` to `os.Stderr`
- All changes in `pkg/openapimcp/server.go` — no new files in `pkg/`
