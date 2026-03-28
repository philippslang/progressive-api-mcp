# Contract: GET /health

**Version**: 1.0 | **Date**: 2026-03-28

## Endpoint

```
GET /health
```

Served on the same host and port as the MCP endpoint. No path prefix, no authentication.

## Request

No request body, query parameters, or headers required.

| Field   | Value |
|---------|-------|
| Method  | GET   |
| Path    | /health |
| Auth    | None  |
| Body    | None  |

## Response — Healthy (200 OK)

```
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{"status":"ok"}
```

## Response — Wrong Method (405 Method Not Allowed)

Requests using any method other than GET receive a 405 response (standard Go `http.ServeMux` behavior).

```
HTTP/1.1 405 Method Not Allowed
```

## Stability

This endpoint is part of the server's operational interface. The path `/health` and the 200 status code on healthy server are stable guarantees. The response body schema (`{"status":"ok"}`) may gain additional fields in future versions; consumers should not fail on unknown fields.

## Out of Scope

- Dependency health (database, downstream APIs) — not checked.
- Readiness vs. liveness distinction — this is a liveness check only.
- Authentication or rate limiting — not applied.
