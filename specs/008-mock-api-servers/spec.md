# Feature Specification: Mock API Test Servers

**Feature Branch**: `008-mock-api-servers`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "write three test servers in the form of clis under the /tests directory. each server mocks one of the apis in /tests/testdata with in-memory store"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run Petstore Mock Server (Priority: P1)

A developer running integration tests needs a local server that faithfully implements the Petstore API so they can exercise full request/response cycles without calling a real service. They launch the Petstore mock server from the command line, specifying which port to listen on, and the server starts up immediately with an empty in-memory store ready to accept requests.

**Why this priority**: The Petstore API is the primary fixture in the test suite and is referenced by the most integration tests. A working Petstore mock unblocks the majority of testing scenarios.

**Independent Test**: Can be fully tested by starting the server and sending HTTP requests to list, create, read, update, delete pets and owners — all responses must match the Petstore OpenAPI contract.

**Acceptance Scenarios**:

1. **Given** the mock server is not running, **When** a developer runs the Petstore CLI with a port argument, **Then** the server starts and prints the listening address to stdout.
2. **Given** the Petstore mock server is running with an empty store, **When** a client sends a create-pet request with valid pet data, **Then** the server responds with status 201 and returns the created pet with an auto-assigned numeric ID.
3. **Given** a pet has been created, **When** a client requests that pet by ID, **Then** the server returns the pet with status 200.
4. **Given** no pet exists with the requested ID, **When** a client requests a pet by that ID, **Then** the server returns status 404 with an error message body.
5. **Given** a pet exists, **When** a client sends a full-update request for that pet, **Then** the server returns the updated pet with status 200.
6. **Given** a pet exists, **When** a client sends a partial-update request with only some fields, **Then** only the provided fields are updated and the server returns the full updated pet with status 200.
7. **Given** a pet exists, **When** a client sends a delete request for that pet, **Then** the server returns status 204 and the pet is no longer retrievable.
8. **Given** multiple pets exist, **When** a client requests the pet list with a limit parameter, **Then** the server returns at most that many pets.
9. **Given** the server is running, **When** a client requests the owner list, **Then** the server returns all owners in the in-memory store.
10. **Given** the server is running, **When** a client requests an owner by ID, **Then** the server returns the matching owner or status 404 if not found.

---

### User Story 2 - Run Bookstore Mock Server (Priority: P2)

A developer testing multi-API scenarios needs a local Bookstore API server running on a separate port to verify that the system correctly handles multiple concurrent backends. They launch the Bookstore mock server independently of the Petstore mock.

**Why this priority**: The Bookstore fixture is used for multi-API tests. Having a dedicated mock server enables tests that verify concurrent use of multiple APIs.

**Independent Test**: Can be fully tested by starting the server and sending HTTP requests to list, create, and read books — all responses must match the Bookstore OpenAPI contract.

**Acceptance Scenarios**:

1. **Given** the Bookstore mock is not running, **When** a developer starts it with a port argument, **Then** the server starts and prints the listening address.
2. **Given** the server is running, **When** a client creates a book with title and author, **Then** the server returns status 201 with the created book including an auto-assigned ID.
3. **Given** a book exists, **When** a client requests it by ID, **Then** the server returns the book with status 200.
4. **Given** no book exists for the requested ID, **When** a client requests it, **Then** the server returns status 404.
5. **Given** multiple books exist, **When** a client requests the book list, **Then** all books are returned as a list with status 200.

---

### User Story 3 - Run Malformed API Mock Server (Priority: P3)

A developer testing error-handling behavior needs a server that simulates a broken or unreachable upstream API. They launch the third mock CLI, which starts a minimal server that responds with server-error status codes, enabling reliable negative-path test scenarios.

**Why this priority**: The malformed fixture exists specifically to test how the system handles invalid configurations. A dedicated server enables reliable, repeatable negative-path testing.

**Independent Test**: Can be fully tested by starting the server and confirming that every request receives an HTTP error response, allowing tests to verify graceful degradation.

**Acceptance Scenarios**:

1. **Given** the malformed mock is not running, **When** a developer starts it with a port argument, **Then** the server starts and prints the listening address.
2. **Given** the server is running, **When** a client sends any HTTP request to any path, **Then** the server returns an HTTP 5xx error response.

---

### Edge Cases

- What happens when the requested port is already in use? The server exits with a clear error message identifying the port conflict.
- What happens when a client requests a list endpoint with an empty store? The server returns status 200 with an empty array.
- What happens when a create request is sent twice with the same data? Each request produces a distinct record with a unique auto-assigned ID.
- What happens when a partial-update request is sent with no fields or an empty body? The resource is returned unchanged with status 200.
- What happens when a client requests a path not defined in the API contract? The server returns status 404.
- What happens when the `limit` parameter on `GET /pets` is zero or negative? The server returns all results (limit is ignored or treated as unset).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each of the three mock servers MUST be launchable as a standalone command-line program accepting a port argument.
- **FR-002**: Each server MUST start and be ready to accept requests within 2 seconds of launch on a standard developer machine.
- **FR-003**: Each server MUST print its listening address to stdout on startup so callers can confirm readiness.
- **FR-004**: The Petstore mock MUST implement all endpoints from `tests/testdata/petstore.yaml`: list pets, create pet, get pet by ID, update pet, partially update pet, delete pet, list owners, get owner by ID.
- **FR-005**: The Bookstore mock MUST implement all endpoints from `tests/testdata/bookstore.yaml`: list books, create book, get book by ID.
- **FR-006**: The malformed mock MUST start successfully and return HTTP 5xx responses to all requests regardless of path or method.
- **FR-007**: All mock servers MUST use an in-memory store; data persists only for the lifetime of the process and is never written to disk.
- **FR-008**: The Petstore and Bookstore mocks MUST auto-assign unique integer IDs to created resources, starting from 1 and incrementing per creation.
- **FR-009**: Mock servers MUST return `Content-Type: application/json` for all endpoints that include a response body.
- **FR-010**: Mock servers MUST return HTTP status codes that match the OpenAPI contract (200, 201, 204, 404 as appropriate).
- **FR-011**: All three CLIs MUST reside under the `tests/` directory and be buildable with the standard project build process without additional setup.
- **FR-012**: When a resource is not found by ID, the server MUST return status 404 with a JSON body containing a `message` field.
- **FR-013**: The `GET /pets` endpoint MUST honor the optional `limit` query parameter, returning at most that many results when provided.
- **FR-014**: Owners in the Petstore mock MUST be pre-seeded with at least two sample records, since the Petstore API contract does not include a create-owner endpoint.

### Key Entities

- **Pet**: A record with a unique integer ID, name, species, and optional age. Lives only in the server's in-memory store for the process lifetime.
- **Owner**: A record with a unique integer ID and name. Pre-seeded at server startup; lives in-memory for the process lifetime.
- **Book**: A record with a unique integer ID, title, and author. Lives only in the server's in-memory store for the process lifetime.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All three mock servers start and respond to their first request within 2 seconds of process launch on a standard developer machine.
- **SC-002**: 100% of endpoints defined in `petstore.yaml` and `bookstore.yaml` are reachable via the respective mocks and return the correct HTTP status codes.
- **SC-003**: A full create-read-update-delete cycle on a Petstore pet completes successfully in a single test run without any external dependencies or manual setup beyond launching the server.
- **SC-004**: Integration tests that previously required external services can be run fully offline using only the three mock servers.
- **SC-005**: Each mock server can be started, exercised, and shut down independently without affecting the behavior of the other two servers.

## Assumptions

- The three API fixtures to mock are `petstore.yaml`, `bookstore.yaml`, and `malformed.yaml` from `tests/testdata/`.
- The malformed server does not implement a real data model; its sole purpose is to simulate a broken upstream for negative-path test coverage.
- Owners in the Petstore mock are pre-seeded with a small fixed set of sample data since the OpenAPI spec does not define a create-owner endpoint.
- Mock servers require no authentication; all endpoints are open.
- The port is the only required CLI argument; host binding defaults to `localhost`.
- The in-memory store does not need to be safe for concurrent writes from multiple simultaneous clients; sequential single-client test access is the primary use case.
- The mock servers are test tooling only and are not intended for production use.
- No new external dependencies are required beyond what is already present in the project.
