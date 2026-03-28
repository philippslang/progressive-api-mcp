# Data Model: Mock API Test Servers

**Feature**: 008-mock-api-servers
**Date**: 2026-03-28

## Petstore Server

### Pet

| Field   | Type    | Required | Constraints                         |
|---------|---------|----------|-------------------------------------|
| id      | integer | yes      | Auto-assigned; starts at 1; unique  |
| name    | string  | yes      | Non-empty                           |
| species | string  | yes      | Non-empty                           |
| age     | integer | no       | Positive integer when present       |

**State transitions**: Created → (Updated | Patched)* → Deleted

**PATCH semantics**: Only fields present in the request body are updated. Missing fields retain their existing values.

### Owner

| Field | Type    | Required | Constraints                        |
|-------|---------|----------|------------------------------------|
| id    | integer | yes      | Auto-assigned; pre-seeded at start |
| name  | string  | yes      | Non-empty                          |

**Pre-seeded records at startup**:
- `{id: 1, name: "Alice"}`
- `{id: 2, name: "Bob"}`

Owners are read-only at runtime (no create/update/delete endpoints in the OpenAPI spec).

### In-Memory Store (Petstore)

```
PetstoreStore {
    mu      sync.RWMutex
    pets    map[int]Pet
    nextPetID  int          // starts at 1, incremented on each creation
    owners  map[int]Owner   // pre-seeded, immutable after startup
}
```

---

## Bookstore Server

### Book

| Field  | Type    | Required | Constraints                        |
|--------|---------|----------|------------------------------------|
| id     | integer | yes      | Auto-assigned; starts at 1; unique |
| title  | string  | yes      | Non-empty                          |
| author | string  | yes      | Non-empty                          |

### In-Memory Store (Bookstore)

```
BookstoreStore {
    mu         sync.RWMutex
    books      map[int]Book
    nextBookID int             // starts at 1, incremented on each creation
}
```

---

## Malformed Server

No data model. The malformed server holds no state and serves no real resources.

```
MalformedServer {
    // stateless — always returns 500
}
```

---

## Error Response (all servers)

All 404 and error responses use the same JSON structure:

```json
{
  "message": "<human-readable description>"
}
```

This matches the `Error` schema defined in `petstore.yaml` and is used consistently across all servers for simplicity.
