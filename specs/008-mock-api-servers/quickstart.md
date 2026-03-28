# Quickstart: Mock API Test Servers

**Feature**: 008-mock-api-servers
**Date**: 2026-03-28

## Build

```bash
# Build all three mock servers
go build -o petstore-mock  ./tests/mockservers/petstore
go build -o bookstore-mock ./tests/mockservers/bookstore
go build -o malformed-mock ./tests/mockservers/malformed

# Or build all at once
go build ./tests/mockservers/...
```

## Run

```bash
# Petstore mock on default port 8080
./petstore-mock

# Petstore mock on a custom port
./petstore-mock --port 18080

# Bookstore mock on default port 9090
./bookstore-mock

# Malformed mock on default port 9999
./malformed-mock
```

Each server prints its address when ready:

```
Listening on http://localhost:8080
```

## Verify

```bash
# Petstore: list pets (empty on startup)
curl http://localhost:8080/pets

# Petstore: create a pet
curl -X POST http://localhost:8080/pets \
  -H "Content-Type: application/json" \
  -d '{"name":"Fido","species":"dog","age":3}'

# Petstore: get the pet back
curl http://localhost:8080/pets/1

# Bookstore: list books
curl http://localhost:9090/books

# Malformed: any request returns 500
curl http://localhost:9999/anything
```

## Use in Integration Tests

```go
// Start a petstore mock on a free port
cmd := exec.CommandContext(ctx, "petstore-mock", "--port", "0")
// Wait for "Listening on ..." line on stdout
// Extract port from the output line
// Run test requests against that port
// Cancel ctx to shut down
```

Or use `tests/mockservers/petstore` directly as a library if the test imports it in-process (no subprocess needed for unit-style testing).

## Stop

Send SIGINT (Ctrl-C) or SIGTERM to the process. The server exits cleanly.
