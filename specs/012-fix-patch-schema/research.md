# Research: Fix PATCH Endpoint OpenAPI Schema (012) — Dual Patch Format Support

## Decision 1: Support both formats simultaneously via `oneOf`

**Decision**: The `PATCH /pets/{id}` requestBody schema must accept **either** a JSON Merge Patch object **or** an RFC 6902 JSON Patch array. This is achieved with an OpenAPI `oneOf` discriminated by type:

```yaml
requestBody:
  required: false
  content:
    application/json:
      schema:
        oneOf:
          - $ref: "#/components/schemas/JsonPatchDocument"   # RFC 6902 array
          - $ref: "#/components/schemas/PetPatch"            # JSON Merge Patch object
```

`oneOf` is appropriate because a JSON value cannot simultaneously be an array and an object — the two branches are mutually exclusive, so exactly one will match.

**Rationale**: User explicitly requires both formats to be supported. The original spec's FR-004 ("reject a plain JSON object") is superseded.

**Alternatives considered**:
- `anyOf` — semantically valid but `oneOf` better communicates mutual exclusivity.
- Content-type discrimination (`application/merge-patch+json` vs `application/json-patch+json`) — cleaner RFC approach but requires changes to the MCP tool layer, which doesn't currently support content-type negotiation per request. Ruled out as over-engineering.

---

## Decision 2: Mock server handler must branch on body shape

**Decision**: The petstore mock server's `patchPet` handler must detect which format was sent and dispatch accordingly:

- Peek at the first non-whitespace byte of the JSON body:
  - `[` → decode as `[]jsonPatchOp` (RFC 6902)
  - `{` → decode as `petPatch` (Merge Patch)

Implementation: buffer the body into `json.RawMessage`, then branch on `bytes.TrimLeft(raw, " \t\r\n")[0]`.

**Rationale**: Both formats must reach the backend and produce a correct response. The mock server currently only handles the object format; it must be extended rather than replaced.

**Alternatives considered**: Separate endpoints (`/pets/{id}/patch-rfc6902`) — rejected; the goal is a single PATCH endpoint accepting either format.

---

## Decision 3: RFC 6902 schema components

**Decision**: Add `JsonPatchDocument` (array) and `JsonPatchOp` (object) schemas to `components/schemas` in petstore.yaml. Keep the existing `PetPatch` (object) schema.

`JsonPatchOp` fields:
- `op` (string, required): enum `[add, remove, replace, move, copy, test]`
- `path` (string, required): JSON Pointer
- `value` (any, optional): unconstrained type — declared as `{}` (no type constraint per RFC 6902)
- `from` (string, optional): required by RFC for `move`/`copy` (app-level, not schema-enforced)

---

## Decision 4: Integration test update

**Decision**: The existing `http_patch valid body` integration test sends `{"name": "Rex"}` (merge-patch object) and must continue to pass — this validates the merge-patch path. A new sub-test sends a valid RFC 6902 array and must also pass — this validates the RFC 6902 path.

The previous plan's "add rejection test for object body" is **cancelled** because object bodies are now valid.

---

## Decision 5: libopenapi-validator and `oneOf` arrays vs objects

**Decision**: No special handling needed in MCP layer. The pb33f/libopenapi-validator already handles `oneOf` validation correctly for OpenAPI 3.1 schemas — it will validate the body against each branch and pass if exactly one matches.

**Research note**: `oneOf` with array vs object branches is unambiguous — the JSON parser's type system makes this discrimination lossless. No edge cases identified.

---

## Decision 6: `value: {}` schema in OpenAPI 3.1

**Decision**: Declare `value` as an empty schema `{}` (no `type` key) rather than `type: null` or `type: [string, number, ...]`. In OpenAPI 3.1 / JSON Schema, `{}` matches any value including null, which is what RFC 6902 requires.

---

## Implementation File Map

| File | Change |
|------|--------|
| `tests/testdata/petstore.yaml` | Add `JsonPatchDocument` + `JsonPatchOp` schemas; change PATCH requestBody to `oneOf[JsonPatchDocument, PetPatch]` |
| `tests/mockservers/petstore/main.go` | Extend `patchPet` to branch on body shape: array → RFC 6902 apply loop; object → existing merge-patch logic |
| `tests/integration/http_tools_test.go` | Keep existing object-body test; add RFC 6902 array-body test |
