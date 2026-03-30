# Feature Specification: Fix PATCH Endpoint OpenAPI Schema

**Feature Branch**: `012-fix-patch-schema`
**Created**: 2026-03-30
**Status**: Draft
**Input**: User description: "Bug: OpenAPI schema mismatch on scenario PATCH endpoints — request body declared as type: object but server expects a JSON Patch array (RFC 6902)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Send a valid JSON Patch request via MCP tool (Priority: P1)

A developer uses the MCP tool to call a PATCH endpoint on a mock API server. They provide a JSON Patch array (RFC 6902 format) as the request body. The MCP server validates the body against the OpenAPI spec and, because the spec now correctly declares an array schema, forwards the request to the Go server, which processes it successfully.

**Why this priority**: This is the core scenario that is completely broken today — no body format satisfies both the MCP schema validator and the Go server simultaneously.

**Independent Test**: Submit a valid JSON Patch array `[{"op":"replace","path":"/name","value":"foo"}]` through the MCP tool to a PATCH endpoint and confirm it returns a 200-level response without validation errors.

**Acceptance Scenarios**:

1. **Given** the OpenAPI spec has a PATCH endpoint whose requestBody schema is declared as an array of JSON Patch operations, **When** a client sends `[{"op":"replace","path":"/name","value":"test"}]`, **Then** the MCP server accepts the body, forwards it, and the Go server responds with 2xx.
2. **Given** the same endpoint, **When** a client sends an object `{"op":"replace","path":"/name","value":"test"}` (not wrapped in an array), **Then** the MCP server rejects it with a schema validation error describing the array requirement.
3. **Given** the same endpoint, **When** a client sends a JSON Patch operation missing the required `op` field, **Then** validation fails with a clear error identifying the missing field.

---

### User Story 2 - Existing valid requests continue to work (Priority: P2)

Non-PATCH endpoints on mock API servers are unaffected by this schema correction. Developers using GET, POST, PUT, or DELETE tools see no change in behavior.

**Why this priority**: Schema changes must be surgical — breaking previously-working endpoints would degrade the overall tool.

**Independent Test**: Exercise GET and POST endpoints on any mock server and confirm they still return their previous responses without errors.

**Acceptance Scenarios**:

1. **Given** a GET endpoint is registered in the mock server, **When** called via MCP tool, **Then** it returns the same response as before this fix.
2. **Given** a POST endpoint with an object requestBody schema, **When** called with a valid object body, **Then** it passes validation and returns the expected response.

---

### Edge Cases

- What happens when the JSON Patch array is empty (`[]`)? The request should pass schema validation (an empty array is valid) and the Go server determines whether an empty patch is a no-op or an error.
- What happens when a Patch operation includes an `op` value not in the enum (e.g., `"op":"merge"`)? The MCP schema validator should reject it before forwarding.
- What happens when `value` is absent on a `remove` operation? `value` is optional by RFC 6902 — schema validation should pass.
- What happens when `from` is absent on a `move` or `copy` operation? This is an application-level constraint; schema validation passes, the Go server handles the error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The OpenAPI spec for the mock scenario server MUST declare the requestBody of all affected PATCH endpoints as `type: array` with items conforming to RFC 6902 JSON Patch operation schema.
- **FR-002**: Each JSON Patch operation item MUST require `op` (enum: add, remove, replace, move, copy, test) and `path` (string), and optionally allow `value` (any type) and `from` (string).
- **FR-003**: The MCP server MUST accept a well-formed JSON Patch array body and forward it to the backend without modification.
- **FR-004**: The MCP server MUST reject a request body that is a plain JSON object (not an array) with an informative validation error.
- **FR-005**: The corrected schema MUST NOT affect any non-PATCH endpoint schemas.

### Key Entities

- **JSON Patch Operation**: A single RFC 6902 patch step with fields `op`, `path`, `value` (optional), and `from` (optional).
- **PATCH Request Body**: An ordered array of one or more JSON Patch operations sent as the HTTP request body.
- **OpenAPI RequestBody Schema**: The schema declaration in the OpenAPI spec file used by the MCP server to validate incoming request bodies before forwarding them.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of well-formed JSON Patch array payloads sent to PATCH endpoints via the MCP tool reach the backend without a schema validation error.
- **SC-002**: 100% of malformed payloads (plain objects, missing required `op`/`path` fields, invalid `op` enum values) are rejected by schema validation before reaching the backend.
- **SC-003**: All previously passing tests for non-PATCH endpoints continue to pass after the schema fix is applied.
- **SC-004**: Both affected PATCH endpoints become reachable end-to-end; a round-trip test (send valid patch → receive success response) succeeds for each.

## Assumptions

- The bug is confined to the OpenAPI spec YAML file(s) used by the mock scenario server; no changes to the Go server deserialization logic are required.
- Both affected PATCH endpoints share the same incorrect `type: object` requestBody declaration and can be fixed with the same corrective schema block.
- The `value` field in a JSON Patch operation may hold any JSON type (string, number, object, array, null) and is declared as an empty schema `{}` with no type constraint.
- There are no other PATCH endpoints in the codebase that intentionally use a non-RFC-6902 array body; if discovered, they are out of scope for this fix.
- The fix is applied to the spec fixture file(s) in `tests/testdata/` or equivalent; no runtime code generation or compilation changes are required.
