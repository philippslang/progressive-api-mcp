# Tasks: Mock API Test Servers

**Input**: Design documents from `/specs/008-mock-api-servers/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-contracts.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the directory scaffolding for all three mock servers

- [x] T001 Create directory structure: `tests/mockservers/petstore/`, `tests/mockservers/bookstore/`, `tests/mockservers/malformed/`

**Checkpoint**: Directory structure exists — all three user story phases can begin in parallel

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No shared foundational infrastructure required — each mock server is fully independent. All three user story phases may start immediately after Phase 1.

---

## Phase 3: User Story 1 — Petstore Mock Server (Priority: P1) 🎯 MVP

**Goal**: A runnable CLI that fully implements the Petstore API (`tests/testdata/petstore.yaml`) with an in-memory store. Accepts `--port` flag, prints `Listening on http://localhost:<port>` on startup, and handles SIGINT/SIGTERM gracefully.

**Independent Test**: Start the server with `--port 18080`. Use curl to: create a pet (POST /pets), retrieve it (GET /pets/1), update it (PUT /pets/1), patch it (PATCH /pets/1), delete it (DELETE /pets/1), list pets with limit (GET /pets?limit=2), list owners (GET /owners), get owner by ID (GET /owners/1). All responses must match the contracts in `specs/008-mock-api-servers/contracts/cli-contracts.md`.

- [x] T002 [US1] Define `Pet`, `Owner`, and `PetstoreStore` types (with `sync.RWMutex`, `pets map[int]Pet`, `owners map[int]Owner`, `nextPetID int`) and JSON encode/decode helpers in `tests/mockservers/petstore/main.go`
- [x] T003 [US1] Implement `GET /pets` handler (returns JSON array, honors `?limit=` param) and `POST /pets` handler (assigns next ID, returns 201) on `PetstoreStore` in `tests/mockservers/petstore/main.go`
- [x] T004 [US1] Implement `GET /pets/{id}`, `PUT /pets/{id}`, `PATCH /pets/{id}` (partial update — only non-zero fields overwritten), and `DELETE /pets/{id}` (returns 204) handlers on `PetstoreStore` in `tests/mockservers/petstore/main.go`
- [x] T005 [US1] Implement `GET /owners` and `GET /owners/{id}` handlers; pre-seed owners `{1,"Alice"}` and `{2,"Bob"}` at store construction in `tests/mockservers/petstore/main.go`
- [x] T006 [US1] Wire `cobra` root command with `--port` flag (default 8080); use `net.Listen("tcp", ...)` first, register all routes on `http.ServeMux` using Go 1.22 pattern syntax (`GET /pets/{id}`), print `Listening on http://localhost:<port>` to stdout, call `http.Serve(listener, mux)` in a goroutine, block on SIGINT/SIGTERM, call `server.Shutdown` with a 5-second timeout in `tests/mockservers/petstore/main.go`

**Checkpoint**: `go build ./tests/mockservers/petstore` succeeds. Petstore mock fully functional and independently testable.

---

## Phase 4: User Story 2 — Bookstore Mock Server (Priority: P2)

**Goal**: A runnable CLI that fully implements the Bookstore API (`tests/testdata/bookstore.yaml`) with an in-memory store. Accepts `--port` flag (default 9090), prints listening address on startup.

**Independent Test**: Start with `--port 19090`. Use curl to: create a book (POST /books), retrieve it (GET /books/1), list all books (GET /books), attempt to get a missing book (GET /books/99 → 404). All responses must match contracts.

- [x] T007 [P] [US2] Define `Book` and `BookstoreStore` types (with `sync.RWMutex`, `books map[int]Book`, `nextBookID int`) and JSON helpers in `tests/mockservers/bookstore/main.go`
- [x] T008 [US2] Implement `GET /books` (returns JSON array), `POST /books` (assigns next ID, returns 201), and `GET /books/{id}` (returns 200 or 404) handlers on `BookstoreStore` in `tests/mockservers/bookstore/main.go`
- [x] T009 [US2] Wire `cobra` root command with `--port` flag (default 9090); `net.Listen`, register routes, print `Listening on http://localhost:<port>`, `http.Serve` in goroutine, graceful shutdown on SIGINT/SIGTERM in `tests/mockservers/bookstore/main.go`

**Checkpoint**: `go build ./tests/mockservers/bookstore` succeeds. Bookstore mock fully functional and independently testable.

---

## Phase 5: User Story 3 — Malformed Mock Server (Priority: P3)

**Goal**: A runnable CLI that starts an HTTP server returning `500 Internal Server Error` with `{"message":"upstream error"}` for every request, simulating a broken upstream. Accepts `--port` flag (default 9999).

**Independent Test**: Start with `--port 19999`. Send `curl http://localhost:19999/anything` and any other path — every response must be HTTP 500 with `Content-Type: application/json` and body `{"message":"upstream error"}`.

- [ ] T010 [P] [US3] Implement a catch-all `http.HandlerFunc` that writes `Content-Type: application/json`, status 500, and body `{"message":"upstream error"}` for every request in `tests/mockservers/malformed/main.go`
- [ ] T011 [US3] Wire `cobra` root command with `--port` flag (default 9999); `net.Listen`, register catch-all handler on `http.NewServeMux()`, print `Listening on http://localhost:<port>`, `http.Serve` in goroutine, graceful shutdown on SIGINT/SIGTERM in `tests/mockservers/malformed/main.go`

**Checkpoint**: `go build ./tests/mockservers/malformed` succeeds. Malformed mock always returns 500 and is independently testable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify the full build, vet, and quickstart validation across all three servers

- [ ] T012 [P] Run `go build ./tests/mockservers/...` and confirm all three binaries compile without errors
- [ ] T013 [P] Run `go vet ./tests/mockservers/...` and resolve any reported issues
- [ ] T014 Run quickstart validation per `specs/008-mock-api-servers/quickstart.md`: start each server, send the curl requests listed, confirm expected status codes and response bodies

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: No blocking work — skip directly to user stories after Phase 1
- **User Stories (Phases 3–5)**: All depend only on Phase 1 (directory creation); all three can proceed in parallel after T001
- **Polish (Phase 6)**: Depends on all desired user story phases being complete

### User Story Dependencies

- **US1 (P1)**: Starts after T001 — no dependencies on US2 or US3
- **US2 (P2)**: Starts after T001 — no dependencies on US1 or US3; T007 is marked [P] and can run in parallel with US1 work
- **US3 (P3)**: Starts after T001 — no dependencies on US1 or US2; T010 is marked [P] and can run in parallel with US1 and US2 work

### Within Each User Story

- US1: T002 → T003 → T004 → T005 → T006 (sequential, all edit the same file)
- US2: T007 → T008 → T009 (sequential, all edit the same file)
- US3: T010 → T011 (sequential, same file)

### Parallel Opportunities

- After T001: T002 (US1), T007 (US2), T010 (US3) can all start in parallel
- T012 and T013 (Polish) can run in parallel once all story phases complete

---

## Parallel Example: After Phase 1

```
# All three stories can begin simultaneously after T001:
Agent A → T002, T003, T004, T005, T006  (Petstore server)
Agent B → T007, T008, T009              (Bookstore server)
Agent C → T010, T011                    (Malformed server)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: T001
2. Complete Phase 3: T002–T006 (Petstore mock)
3. **STOP and VALIDATE**: Start server, run curl commands, verify all petstore endpoints
4. Ship Petstore mock as MVP

### Incremental Delivery

1. T001 → foundation ready
2. T002–T006 → Petstore mock working → validate independently
3. T007–T009 → Bookstore mock working → validate independently
4. T010–T011 → Malformed mock working → validate independently
5. T012–T014 → Full build/vet/quickstart pass → feature complete

### Parallel Team Strategy

With multiple developers (after T001):
- Developer A: Phase 3 (Petstore — US1)
- Developer B: Phase 4 (Bookstore — US2)
- Developer C: Phase 5 (Malformed — US3)

All work in separate files under `tests/mockservers/` with no shared code, so zero merge conflicts.

---

## Notes

- [P] tasks operate on different files and have no dependencies on incomplete tasks in the same phase
- [Story] labels map each task to a specific user story for traceability
- Each user story is independently completable and testable without the others
- PATCH handler (T004): only fields present in the JSON body should overwrite existing values — use a partial-decode approach (decode into a map or a struct with pointer fields)
- Go 1.22+ path parameters: use `mux.HandleFunc("GET /pets/{id}", ...)` and `r.PathValue("id")` — no third-party router needed
- Use `net.Listen` before starting `http.Serve` to guarantee the port is bound before printing the listening address
- Avoid any global state; pass the store into handlers via closure or method receiver
