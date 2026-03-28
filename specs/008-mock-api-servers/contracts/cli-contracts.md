# CLI Contracts: Mock API Test Servers

**Feature**: 008-mock-api-servers
**Date**: 2026-03-28

## Common Contract (all three servers)

### Invocation

```
<server-binary> [--port <port>]
```

| Argument | Type    | Default | Description                             |
|----------|---------|---------|-----------------------------------------|
| --port   | integer | varies  | TCP port to listen on; 0 = OS-assigned  |

**Defaults by server**:
- `petstore-mock`: `--port 8080`
- `bookstore-mock`: `--port 9090`
- `malformed-mock`: `--port 9999`

### Startup Output (stdout)

On successful start, the server prints exactly one line to stdout:

```
Listening on http://localhost:<port>
```

Callers (including integration tests) MUST wait for this line before sending requests.

### Error Output (stderr)

Errors (e.g., port already in use, invalid port) are written to stderr and the process exits with a non-zero exit code.

### Shutdown

The server shuts down gracefully on SIGINT or SIGTERM. No special output on shutdown.

---

## Petstore Mock — Endpoint Contract

**Base URL**: `http://localhost:<port>`

### `GET /pets`

| Aspect         | Value                                                  |
|----------------|--------------------------------------------------------|
| Query params   | `limit` (integer, optional) — max results to return    |
| Success status | 200                                                    |
| Response body  | `application/json` array of Pet objects                |
| Empty store    | 200 with `[]`                                          |

### `POST /pets`

| Aspect         | Value                                      |
|----------------|--------------------------------------------|
| Request body   | `application/json` — `{name, species, age?}` |
| Success status | 201                                        |
| Response body  | `application/json` Pet object with assigned `id` |

### `GET /pets/{id}`

| Aspect         | Value                                              |
|----------------|----------------------------------------------------|
| Path param     | `id` (integer)                                     |
| Success status | 200 with Pet object                                |
| Not found      | 404 with `{"message": "pet not found"}`            |

### `PUT /pets/{id}`

| Aspect         | Value                                              |
|----------------|----------------------------------------------------|
| Path param     | `id` (integer)                                     |
| Request body   | `application/json` — `{name, species, age?}`       |
| Success status | 200 with updated Pet object                        |
| Not found      | 404 with `{"message": "pet not found"}`            |

### `PATCH /pets/{id}`

| Aspect         | Value                                              |
|----------------|----------------------------------------------------|
| Path param     | `id` (integer)                                     |
| Request body   | `application/json` — `{name?, age?}` (partial)     |
| Success status | 200 with updated Pet object                        |
| Not found      | 404 with `{"message": "pet not found"}`            |

### `DELETE /pets/{id}`

| Aspect         | Value                                              |
|----------------|----------------------------------------------------|
| Path param     | `id` (integer)                                     |
| Success status | 204 (no body)                                      |
| Not found      | 404 with `{"message": "pet not found"}`            |

### `GET /owners`

| Aspect         | Value                              |
|----------------|------------------------------------|
| Success status | 200                                |
| Response body  | `application/json` array of Owner objects |

### `GET /owners/{id}`

| Aspect         | Value                                                |
|----------------|------------------------------------------------------|
| Path param     | `id` (integer)                                       |
| Success status | 200 with Owner object                                |
| Not found      | 404 with `{"message": "owner not found"}`            |

---

## Bookstore Mock — Endpoint Contract

**Base URL**: `http://localhost:<port>`

### `GET /books`

| Aspect         | Value                                            |
|----------------|--------------------------------------------------|
| Success status | 200                                              |
| Response body  | `application/json` array of Book objects         |
| Empty store    | 200 with `[]`                                    |

### `POST /books`

| Aspect         | Value                                            |
|----------------|--------------------------------------------------|
| Request body   | `application/json` — `{title, author}`           |
| Success status | 201                                              |
| Response body  | `application/json` Book object with assigned `id` |

### `GET /books/{id}`

| Aspect         | Value                                              |
|----------------|----------------------------------------------------|
| Path param     | `id` (integer)                                     |
| Success status | 200 with Book object                               |
| Not found      | 404 with `{"message": "book not found"}`           |

---

## Malformed Mock — Endpoint Contract

**Base URL**: `http://localhost:<port>`

### All routes (`/*`)

| Aspect         | Value                                            |
|----------------|--------------------------------------------------|
| Any path/method | 500 Internal Server Error                       |
| Response body  | `application/json` — `{"message": "upstream error"}` |
| Content-Type   | `application/json`                               |
