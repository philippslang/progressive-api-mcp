# Implementation Plan: Mock API Test Servers

**Branch**: `008-mock-api-servers` | **Date**: 2026-03-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-mock-api-servers/spec.md`

## Summary

Add three standalone CLI mock servers under `tests/mockservers/` — one each for the Petstore, Bookstore, and Malformed API fixtures in `tests/testdata/`. Each server uses an in-memory store, a `--port` flag, and prints `Listening on http://localhost:<port>` on startup. The malformed server always returns HTTP 500 to simulate a broken upstream. No new dependencies are required.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: `cobra` (CLI), `net/http` stdlib (HTTP server), `encoding/json` stdlib (JSON responses), `sync` stdlib (RWMutex for in-memory store)
**Storage**: In-memory only (`sync.RWMutex`-protected maps); no disk persistence
**Testing**: `go test ./...` (compilation check for mock server packages; functional testing via integration tests)
**Target Platform**: Linux/macOS developer workstation; same environment as CI
**Project Type**: CLI tools (test infrastructure)
**Performance Goals**: Start and first response within 2 seconds; single-client sequential access
**Constraints**: No new external dependencies; mock CLIs must reside under `tests/`
**Scale/Scope**: Three small CLI binaries; <300 LOC each

## Constitution Check

The project constitution template is unpopulated with project-specific rules. Using CLAUDE.md as the governing document:

| Rule | Status |
|------|--------|
| No new dependencies | PASS — stdlib + cobra (already in go.mod) only |
| `pkg/` must not import cobra/viper | PASS — mock servers live under `tests/`, not `pkg/` |
| Standard Go idioms, gofmt enforced | PASS — will follow |
| No global state; dependencies via constructors | PASS — store passed into handler struct |
| All exported types in stable packages have doc comments | N/A — these are `main` packages |

No violations. No complexity justification required.

## Project Structure

### Documentation (this feature)

```text
specs/008-mock-api-servers/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── cli-contracts.md # Phase 1 output
└── tasks.md             # Phase 2 output (from /speckit.tasks)
```

### Source Code (repository root)

```text
tests/
├── mockservers/
│   ├── petstore/
│   │   └── main.go      # Petstore mock CLI (pets + owners endpoints)
│   ├── bookstore/
│   │   └── main.go      # Bookstore mock CLI (books endpoints)
│   └── malformed/
│       └── main.go      # Malformed mock CLI (always returns 500)
├── testdata/
│   ├── petstore.yaml    # (existing) Petstore OpenAPI spec
│   ├── bookstore.yaml   # (existing) Bookstore OpenAPI spec
│   └── malformed.yaml   # (existing) Malformed OpenAPI spec
└── integration/         # (existing) Integration tests that will use the mocks
```

**Structure Decision**: Single subtree under `tests/mockservers/` with one `main` package per server. This keeps mock servers co-located with the testdata they implement and separate from production code under `cmd/` and `pkg/`.

## Implementation Notes

### Petstore Mock (`tests/mockservers/petstore/main.go`)

- Store struct holds `pets map[int]Pet`, `owners map[int]Owner`, `nextPetID int`, protected by `sync.RWMutex`
- Owners pre-seeded at startup: `{1, "Alice"}`, `{2, "Bob"}`
- Routes registered on `http.NewServeMux()`:
  - `GET /pets` — list with optional `?limit=` param
  - `POST /pets` — create, assign next ID
  - `GET /pets/{id}` — requires Go 1.22+ pattern `GET /pets/{id}` with `r.PathValue("id")`
  - `PUT /pets/{id}` — full replace
  - `PATCH /pets/{id}` — partial update (only non-zero/non-nil fields in request)
  - `DELETE /pets/{id}` — remove, return 204
  - `GET /owners` — list all
  - `GET /owners/{id}` — get by ID

### Bookstore Mock (`tests/mockservers/bookstore/main.go`)

- Store struct holds `books map[int]Book`, `nextBookID int`, protected by `sync.RWMutex`
- Routes: `GET /books`, `POST /books`, `GET /books/{id}`

### Malformed Mock (`tests/mockservers/malformed/main.go`)

- No store
- Single catch-all handler: writes `Content-Type: application/json`, status 500, body `{"message":"upstream error"}`

### Shared Patterns (each server)

1. `cobra.Command` root command with `--port` flag
2. `http.Server` created with mux, started in goroutine
3. Print `Listening on http://localhost:<port>\n` to stdout after `Listener.Accept()` is ready (use `net.Listen` first, then start `http.Serve(l, mux)` to avoid race between print and readiness)
4. Block on `os.Signal` channel (SIGINT, SIGTERM); call `server.Shutdown(ctx)` on signal

### Go 1.22 Path Parameters

Use the new `http.ServeMux` pattern matching introduced in Go 1.22:

```go
mux.HandleFunc("GET /pets/{id}", handler)
// extract with: id := r.PathValue("id")
```

This avoids any third-party router dependency.
