# Implementation Plan: Fix PATCH Endpoint OpenAPI Schema — Dual Patch Format Support

**Branch**: `012-fix-patch-schema` | **Date**: 2026-03-30 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-fix-patch-schema/spec.md`

## Summary

Extend the `PATCH /pets/{id}` endpoint in `tests/testdata/petstore.yaml` to accept **both** JSON Merge Patch (RFC 7396, plain object) and RFC 6902 JSON Patch (array of operations) via an OpenAPI `oneOf` schema. Update the petstore mock server handler to dispatch on body shape, and add an integration test for the RFC 6902 path while preserving the existing merge-patch test.

**User clarification** (supersedes original spec FR-004): Both patch formats must be supported simultaneously — plain JSON objects are valid, not rejected.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator — no new dependencies
**Storage**: N/A (in-memory mock server)
**Testing**: `go test ./...` (table-driven, integration tests in `tests/integration/`)
**Target Platform**: Linux / any Go platform
**Project Type**: library + CLI (MCP server proxy for OpenAPI)
**Performance Goals**: N/A
**Constraints**: No new Go module dependencies; must not break existing tests
**Scale/Scope**: 3 files changed

## Constitution Check

Constitution template is unpopulated. General CLAUDE.md constraints:

- [x] No new dependencies — `bytes` package used in mock server is stdlib, already available
- [x] `pkg/` packages not touched — changes confined to `tests/`
- [x] No global state introduced
- [x] Table-driven test style preserved

**GATE: PASS** — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/012-fix-patch-schema/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── patch-schema.yaml  # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (affected files only)

```text
tests/
├── testdata/
│   └── petstore.yaml          # Add JsonPatchDocument+JsonPatchOp; PATCH body → oneOf
├── mockservers/petstore/
│   └── main.go                # Extend patchPet: body-shape dispatch + RFC 6902 apply loop
└── integration/
    └── http_tools_test.go     # Keep object test; add RFC 6902 array test
```

## Implementation Phases

### Phase A — YAML Schema Fix (`tests/testdata/petstore.yaml`)

1. **Change the PATCH requestBody schema** from `$ref: PetPatch` to:
   ```yaml
   schema:
     oneOf:
       - $ref: "#/components/schemas/JsonPatchDocument"
       - $ref: "#/components/schemas/PetPatch"
   ```

2. **Add two new schemas** to `components/schemas` (keep existing `PetPatch` unchanged):
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

### Phase B — Mock Server Handler Extension (`tests/mockservers/petstore/main.go`)

1. **Add `bytes` import** (stdlib, already in module).

2. **Add `jsonPatchOp` struct**:
   ```go
   type jsonPatchOp struct {
       Op    string          `json:"op"`
       Path  string          `json:"path"`
       Value json.RawMessage `json:"value,omitempty"`
       From  string          `json:"from,omitempty"`
   }
   ```

3. **Rewrite `patchPet`** to:
   - If `ContentLength == 0` → skip body processing, return 200 with unchanged pet.
   - Decode body into `json.RawMessage`.
   - Trim leading whitespace; if first byte is `[` → RFC 6902 path; else → merge-patch path.
   - RFC 6902 path: unmarshal as `[]jsonPatchOp`, iterate, apply `replace` ops for `/name` and `/age`.
   - Merge-patch path: unmarshal as `petPatch`, apply non-nil fields (existing logic).
   - Save updated pet, return 200.

   See [quickstart.md](quickstart.md) for the full handler code.

### Phase C — Integration Test Update (`tests/integration/http_tools_test.go`)

1. **Keep** existing `http_patch valid body` sub-test (sends `{"name": "Rex"}` as object — merge-patch path). This test must continue to pass.

2. **Add** new sub-test `http_patch rfc6902 array body`:
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

   Place after the existing `http_patch valid body` sub-test.

## Verification Checklist

- [ ] `go test ./tests/integration/... -run TestHTTPToolsEndToEnd` — all sub-tests green
- [ ] `http_patch valid body` (merge-patch object) → 200 OK
- [ ] `http_patch rfc6902 array body` (RFC 6902 array) → 200 OK
- [ ] All other test groups unchanged and passing
- [ ] `go test ./...` — zero failures
- [ ] `golangci-lint run` — no new issues

## Complexity Tracking

No constitution violations requiring justification.
