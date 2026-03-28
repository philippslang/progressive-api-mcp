# Research: Mock API Test Servers

**Feature**: 008-mock-api-servers
**Date**: 2026-03-28

## Decisions

### 1. Directory Layout for Mock Server CLIs

**Decision**: Place each mock server as a standalone Go `main` package under `tests/mockservers/{name}/main.go`.

**Rationale**: The project already has `cmd/` for production CLIs. Test-only CLIs belong under `tests/` as requested. A subdirectory per server keeps each binary isolated and buildable independently (`go build ./tests/mockservers/petstore`).

**Alternatives considered**:
- Single binary with subcommands: rejected — harder to start each server independently in tests, more complex signal handling.
- Inside `tests/integration/`: rejected — mixing test helper binaries with `_test.go` files creates confusion and Go's test toolchain doesn't build `main` packages in test files.

---

### 2. HTTP Routing Approach

**Decision**: Use Go standard library `net/http` with `http.ServeMux` for routing.

**Rationale**: The project explicitly avoids new dependencies (CLAUDE.md). The existing `tests/integration` tests use `httptest.NewServer` with stdlib handlers. The petstore and bookstore endpoint sets are small and fixed; a mux with pattern matching is sufficient.

**Alternatives considered**:
- Third-party router (chi, gorilla/mux): rejected — adds dependencies.
- Single handler with manual path parsing: rejected — harder to read and maintain.

---

### 3. CLI Framework for Mock Servers

**Decision**: Use `cobra` (already in `go.mod`) for each mock server CLI, following the same pattern as `cmd/prograpimcp/main.go`.

**Rationale**: `cobra` is already a direct dependency. Using it for mock CLIs is consistent with the rest of the project. The restriction in CLAUDE.md that `pkg/` must not import cobra does not apply to CLIs under `tests/` or `cmd/`.

**Alternatives considered**:
- Bare `flag` package: simpler but inconsistent with project conventions and harder to extend.

---

### 4. Port Argument Convention

**Decision**: Expose `--port` as a required positional argument (first arg) rather than a flag, matching how integration tests typically invoke test servers (`./petstore-mock 8080`).

**Rationale**: The existing integration tests (e.g., `health_test.go`) start servers with a known port. Using positional args makes shell invocation cleaner. The existing project CLIs use flags but those are production binaries; test servers have a single required parameter.

**Correction**: After reviewing `cmd/prograpimcp/main.go`, using a `--port` flag is more consistent with the project pattern. Go positional args are less idiomatic here.

**Final decision**: Use `--port` flag with a sensible default (8080 for petstore, 9090 for bookstore, 9999 for malformed). Accept 0 to let the OS pick a free port.

---

### 5. In-Memory Store Concurrency

**Decision**: Protect the in-memory store with a `sync.RWMutex` — reads lock shared, writes lock exclusive.

**Rationale**: While the spec notes single-client sequential use is the primary case, standard library HTTP servers dispatch requests on goroutines. A mutex is the minimal safe approach and adds negligible complexity.

**Alternatives considered**:
- No synchronization: unsafe even for tests due to Go's race detector.
- Channels: over-engineered for a simple key-value store.

---

### 6. Owner Pre-Seeding

**Decision**: Pre-seed two owners (`{id:1, name:"Alice"}`, `{id:2, name:"Bob"}`) at Petstore server startup.

**Rationale**: The Petstore OpenAPI spec has no `POST /owners` endpoint, so owners cannot be created at runtime. Pre-seeding two entries provides enough data for meaningful tests without hard-coding a large fixture set.

---

### 7. Malformed Server Behavior

**Decision**: The malformed mock server starts an HTTP listener and returns `500 Internal Server Error` with a JSON `{"message": "upstream error"}` body for all requests.

**Rationale**: This simulates a broken upstream API reliably. The server must start successfully (exit code 0, listening address printed) — the "malformed" aspect is only in the responses, not the server lifecycle.

---

### 8. Build Integration

**Decision**: Mock server binaries are buildable with `go build ./tests/mockservers/...` and are covered by `go vet ./...` and `go test ./...` (compilation check only, no test functions in the server packages themselves).

**Rationale**: Consistent with how `cmd/` binaries are treated. No separate Makefile targets needed.
