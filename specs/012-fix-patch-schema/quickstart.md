# Quickstart: Fix PATCH Endpoint OpenAPI Schema (012) — Dual Patch Format Support

## What this fix does

Extends the `PATCH /pets/{id}` endpoint to accept **both** formats simultaneously:
- **JSON Merge Patch** (RFC 7396): `{"name": "Rex"}` — existing behavior, preserved
- **RFC 6902 JSON Patch**: `[{"op": "replace", "path": "/name", "value": "Rex"}]` — new support

Uses OpenAPI `oneOf` in `tests/testdata/petstore.yaml` to declare both schemas, extends the mock server handler to dispatch on body shape, and adds a new integration test for the RFC 6902 path.

## Files to change

| File | What changes |
|------|-------------|
| [tests/testdata/petstore.yaml](tests/testdata/petstore.yaml) | Add `JsonPatchDocument` + `JsonPatchOp` schemas; change PATCH requestBody to `oneOf` |
| [tests/mockservers/petstore/main.go](tests/mockservers/petstore/main.go) | Extend `patchPet` to detect body shape and dispatch to merge-patch or RFC 6902 handler |
| [tests/integration/http_tools_test.go](tests/integration/http_tools_test.go) | Keep existing object-body test; add RFC 6902 array-body test |

## Before/After — PATCH requestBody in petstore.yaml

**Before**:
```yaml
requestBody:
  required: false
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/PetPatch"
```

**After**:
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

## New schemas to add to petstore.yaml

```yaml
JsonPatchDocument:
  type: array
  items:
    $ref: "#/components/schemas/JsonPatchOp"
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

## Mock server handler — dispatch logic

```go
// Add to imports: "bytes", "encoding/json"

type jsonPatchOp struct {
    Op    string          `json:"op"`
    Path  string          `json:"path"`
    Value json.RawMessage `json:"value,omitempty"`
    From  string          `json:"from,omitempty"`
}

func (s *petstoreStore) patchPet(w http.ResponseWriter, r *http.Request) {
    id, ok := pathID(w, r)
    if !ok { return }

    s.mu.Lock()
    defer s.mu.Unlock()
    p, exists := s.pets[id]
    if !exists {
        writeError(w, http.StatusNotFound, "pet not found")
        return
    }

    if r.ContentLength != 0 {
        var raw json.RawMessage
        if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
            writeError(w, http.StatusBadRequest, "invalid request body")
            return
        }
        trimmed := bytes.TrimLeft(raw, " \t\r\n")
        if len(trimmed) > 0 && trimmed[0] == '[' {
            // RFC 6902 JSON Patch
            var ops []jsonPatchOp
            if err := json.Unmarshal(raw, &ops); err != nil {
                writeError(w, http.StatusBadRequest, "invalid JSON Patch body")
                return
            }
            for _, op := range ops {
                if op.Op == "replace" {
                    switch op.Path {
                    case "/name":
                        var v string
                        if err := json.Unmarshal(op.Value, &v); err == nil {
                            p.Name = v
                        }
                    case "/age":
                        var v int
                        if err := json.Unmarshal(op.Value, &v); err == nil {
                            p.Age = &v
                        }
                    }
                }
            }
        } else {
            // JSON Merge Patch
            var patch petPatch
            if err := json.Unmarshal(raw, &patch); err != nil {
                writeError(w, http.StatusBadRequest, "invalid request body")
                return
            }
            if patch.Name != nil { p.Name = *patch.Name }
            if patch.Age != nil  { p.Age = patch.Age }
        }
    }

    s.pets[id] = p
    writeJSON(w, http.StatusOK, p)
}
```

## New integration test sub-test

```go
t.Run("http_patch rfc6902 array body", func(t *testing.T) {
    text := callTool(t, c, "http_patch", map[string]any{
        "path": "/pets/1",
        "body": []any{map[string]any{"op": "replace", "path": "/name", "value": "Rex"}},
    })
    var result map[string]any
    require.NoError(t, json.Unmarshal([]byte(text), &result))
    assert.Equal(t, float64(200), result["status_code"])
})
```

## Verification

```bash
go test ./tests/integration/... -run TestHTTPToolsEndToEnd/http_patch
go test ./...
golangci-lint run
```
