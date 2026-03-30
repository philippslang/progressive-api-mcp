# Data Model: Fix PATCH Endpoint OpenAPI Schema (012) — Dual Patch Format Support

## Entities

### PetPatch (existing — JSON Merge Patch)

A partial update object for a `Pet`. All fields optional.

| Field | Type    | Required | Constraint |
|-------|---------|----------|------------|
| `name`| string  | no       | Replaces Pet.Name if present |
| `age` | integer | no       | Replaces Pet.Age if present |

**OpenAPI schema** (unchanged):
```yaml
PetPatch:
  type: object
  properties:
    name:
      type: string
    age:
      type: integer
```

**Go struct** (unchanged):
```go
type petPatch struct {
    Name *string `json:"name,omitempty"`
    Age  *int    `json:"age,omitempty"`
}
```

---

### JsonPatchOp (new — RFC 6902)

A single RFC 6902 JSON Patch operation.

| Field  | Type   | Required | Constraint |
|--------|--------|----------|------------|
| `op`   | string | yes      | enum: `add`, `remove`, `replace`, `move`, `copy`, `test` |
| `path` | string | yes      | JSON Pointer (RFC 6901) |
| `value`| any    | no       | Any JSON value; absent for `remove` ops |
| `from` | string | no       | JSON Pointer; used by `move` and `copy` (app-level) |

**OpenAPI schema** (new):
```yaml
JsonPatchOp:
  type: object
  required:
    - op
    - path
  properties:
    op:
      type: string
      enum: [add, remove, replace, move, copy, test]
    path:
      type: string
    value: {}
    from:
      type: string
```

**Go struct** (new):
```go
type jsonPatchOp struct {
    Op    string          `json:"op"`
    Path  string          `json:"path"`
    Value json.RawMessage `json:"value,omitempty"`
    From  string          `json:"from,omitempty"`
}
```

---

### JsonPatchDocument (new)

An ordered array of `JsonPatchOp` items. Empty array `[]` is valid (no-op).

**OpenAPI schema** (new):
```yaml
JsonPatchDocument:
  type: array
  items:
    $ref: "#/components/schemas/JsonPatchOp"
```

---

### PATCH Request Body — Combined Schema (new)

The requestBody for `PATCH /pets/{id}` accepts either format via `oneOf`:

```yaml
requestBody:
  required: false
  content:
    application/json:
      schema:
        oneOf:
          - $ref: "#/components/schemas/JsonPatchDocument"
          - $ref: "#/components/schemas/PetPatch"
```

---

## Validation Rules

| Payload | Schema result | Handler path |
|---------|--------------|-------------|
| `[{"op":"replace","path":"/name","value":"Rex"}]` | PASS (JsonPatchDocument) | RFC 6902 apply loop |
| `{"name":"Rex"}` | PASS (PetPatch) | Merge patch |
| `[{"path":"/name","value":"Rex"}]` — missing `op` | FAIL (JsonPatchOp requires `op`) | Rejected by MCP validator |
| `[{"op":"merge","path":"/name","value":"Rex"}]` — bad enum | FAIL (`op` not in enum) | Rejected by MCP validator |
| `"just a string"` | FAIL (neither object nor array) | Rejected by MCP validator |
| `[]` — empty array | PASS (JsonPatchDocument allows 0 items) | RFC 6902: no-op |

---

## Mock Server State Transitions

```
Pet in store
  ├── PATCH {"name":"Rex"}  (merge patch)
  │     → Pet.Name = "Rex", 200 OK
  ├── PATCH [{"op":"replace","path":"/name","value":"Rex"}]  (RFC 6902)
  │     → Pet.Name = "Rex", 200 OK
  ├── PATCH [{"op":"replace","path":"/age","value":5}]  (RFC 6902)
  │     → Pet.Age = 5, 200 OK
  ├── PATCH []  (empty RFC 6902)
  │     → Pet unchanged, 200 OK
  └── PATCH <neither valid object nor valid RFC 6902 array>
        → MCP validation error, never reaches server
```

---

## Mock Server Dispatch Logic

```go
// patchPet — body shape detection
var raw json.RawMessage
if err := json.NewDecoder(r.Body).Decode(&raw); err != nil { /* 400 */ }

trimmed := bytes.TrimLeft(raw, " \t\r\n")
if len(trimmed) > 0 && trimmed[0] == '[' {
    // RFC 6902 path
    var ops []jsonPatchOp
    json.Unmarshal(raw, &ops)
    applyJsonPatch(p, ops)
} else {
    // Merge patch path
    var mp petPatch
    json.Unmarshal(raw, &mp)
    applyMergePatch(&p, mp)
}
```
