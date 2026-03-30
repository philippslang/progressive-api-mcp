# Feature Specification: Fully Resolved Schema Objects in get_schema

**Feature Branch**: `011-resolve-schema-objects`
**Created**: 2026-03-29
**Status**: Draft
**Input**: User description: "At the moment get_schema doesn't resolve objects. All it shows for a POST /pets is that an object is required in the body. The user wants to see the fully resolved schema."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See Full Request Body Schema (Priority: P1)

When a user calls `get_schema` for an endpoint with an object request body, the response includes the complete schema: all property names, their types, which properties are required, descriptions, formats, and any nested object structures — not just `{ "type": "object" }`.

**Why this priority**: This is the primary gap. Without resolved schemas, callers cannot know what fields to send in a request body. The feature has no value until this is fixed.

**Independent Test**: Call `get_schema` for `POST /pets` on a petstore-style API. The `request_body` field in the response must include a `properties` map listing each field (e.g. `name`, `age`, `tag`) with their types and required status.

**Acceptance Scenarios**:

1. **Given** an endpoint with an object request body that has named properties, **When** `get_schema` is called, **Then** the response `request_body` includes a `properties` map with each property's name, type, description, and required status.
2. **Given** an endpoint with a required request body, **When** `get_schema` is called, **Then** `request_body.required` is `true` and the body schema is fully expanded.
3. **Given** an endpoint with no request body, **When** `get_schema` is called, **Then** the `request_body` field is absent from the response (current behavior, unchanged).

---

### User Story 2 - See Full Response Schema (Priority: P2)

When a user calls `get_schema` for an endpoint, response schemas (for each status code) are also fully resolved — showing property names, types, required fields, and nested objects — instead of just `{ "type": "object" }`.

**Why this priority**: Response schema resolution is equally important for callers to understand what they will receive. Without it, callers cannot parse or validate responses correctly.

**Independent Test**: Call `get_schema` for `POST /pets`. The `responses["201"].schema` field must include a `properties` map, not just `{ "type": "object" }`.

**Acceptance Scenarios**:

1. **Given** an endpoint that returns an object response, **When** `get_schema` is called, **Then** the relevant response entry includes a resolved schema with `properties`, `type`, and `required` fields.
2. **Given** an endpoint with multiple response codes (200, 400, etc.), **When** `get_schema` is called, **Then** each response code's schema is independently resolved.
3. **Given** a response schema that is an array of objects, **When** `get_schema` is called, **Then** the response schema shows the array type and the resolved item schema.

---

### User Story 3 - Nested Object Resolution (Priority: P3)

Schemas that contain nested objects (properties whose type is another object with their own properties) are recursively resolved, so callers see the complete structure without having to make additional tool calls.

**Why this priority**: Real-world APIs routinely have nested objects (e.g. a `Pet` with an `Owner` embedded). Without recursive resolution the feature is incomplete for these cases.

**Independent Test**: Call `get_schema` for an endpoint whose schema has a property of type `object` with its own sub-properties. The response must show those sub-properties expanded, not just `{ "type": "object" }`.

**Acceptance Scenarios**:

1. **Given** a schema property whose type is an object with named sub-properties, **When** `get_schema` is called, **Then** the sub-properties are included in the output under the parent property.
2. **Given** deeply nested schemas (object within object), **When** `get_schema` is called, **Then** the full nesting is resolved up to a reasonable depth without causing infinite loops for circular references.
3. **Given** a schema with a circular reference (schema A references schema B which references schema A), **When** `get_schema` is called, **Then** the system terminates gracefully and returns what it has resolved so far without crashing.

---

### Edge Cases

- What if a schema uses `$ref` to reference another schema definition? The reference must be resolved to its target before returning the schema.
- What if a property has no type declared? The property is included with whatever attributes are declared; the missing type is not fabricated.
- What if `properties` is empty (an object with no declared properties)? The `properties` map is omitted or returned as empty; no error is raised.
- What if a schema uses `allOf`, `oneOf`, or `anyOf` combiners? These are included in the resolved output as-is; deep resolution of combiners is out of scope for this feature.
- What if a schema has array items that are primitives? The item type is included (e.g. `"items": { "type": "string" }`).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `get_schema` response for request bodies MUST include a `properties` map when the schema declares named properties, with each property's name, type, description, format, and required status.
- **FR-002**: The `get_schema` response for each response code MUST include a resolved schema with `properties` when the schema declares named properties.
- **FR-003**: Schema resolution MUST follow `$ref` references and return the resolved content in place of the reference.
- **FR-004**: Nested objects (properties whose type is `object`) MUST be recursively resolved to show their sub-properties.
- **FR-005**: The `required` list at the schema level MUST be represented in the output so callers can identify which properties are mandatory.
- **FR-006**: Schema resolution MUST handle circular `$ref` chains without crashing — it MUST terminate and return partial results.
- **FR-007**: Schema resolution for `allOf`, `oneOf`, `anyOf` combiners is out of scope; these nodes are included in the output as-is without deep expansion.
- **FR-008**: Endpoints with no request body or no declared response schema MUST continue to behave as currently (field absent or minimal).
- **FR-009**: Array schemas MUST include the resolved item schema (e.g. `"items": { "type": "string" }` or a resolved object for arrays of objects).

### Key Entities

- **Resolved Schema**: A schema representation that expands object types to include their `properties`, `required` list, and recursively resolved nested schemas. Replaces the current shallow `{ "type": "object" }` output.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of `get_schema` responses for endpoints with object request bodies include a `properties` map when properties are declared in the schema.
- **SC-002**: 100% of `get_schema` responses for endpoints with object response schemas include a `properties` map when properties are declared.
- **SC-003**: Circular reference schemas resolve without crashing in 100% of cases — the tool returns a result (partial if needed) rather than an error or hang.
- **SC-004**: Callers can identify all required fields for a request body from a single `get_schema` call, without needing any additional tool calls.
- **SC-005**: Zero regressions on existing `get_schema` behavior for endpoints with primitive, array, or absent schemas.

## Assumptions

- The OpenAPI documents loaded into the server are valid and have machine-resolvable `$ref` chains (no external URL references that require HTTP fetches).
- Schema resolution depth is bounded at a reasonable limit (e.g. 10 levels) to prevent runaway recursion on pathological schemas; this limit is not user-configurable in this feature.
- `allOf`, `oneOf`, `anyOf` are included in output as opaque nodes; deep resolution of composition keywords is a separate, future feature.
- The `required` list is represented at the property level in the output (each property indicates whether it is required) rather than as a separate top-level array, for easier consumption by callers.
- The feature applies to the `get_schema` tool only; `explore_api` output is unchanged.
