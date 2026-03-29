# Feature Specification: Ignore Headers for APIs

**Feature Branch**: `009-ignore-api-headers`
**Created**: 2026-03-29
**Status**: Draft
**Input**: User description: "Add an option to ignore headers for APIs. For example, a user declares to ignore headers for the petstore API. They declare this in the config. Then the required headers will not be shown in explore_api of get_schema calls, and also not required on schema checks in any of the http functions."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Header Suppression Per API (Priority: P1)

A user adds an `ignore_headers` list to an API entry in the config file. Headers on that list are treated as suppressed for all subsequent interactions with that API — they are neither shown in schema exploration nor required during request validation.

**Why this priority**: This is the foundational capability. Without a config-level declaration, none of the downstream behaviors (hidden in explore, skipped in validation) are possible.

**Independent Test**: Can be fully tested by adding an `ignore_headers` entry to the config for a test API and verifying the server starts without error with the setting in effect.

**Acceptance Scenarios**:

1. **Given** a config file with an `ignore_headers` list for API "petstore", **When** the server loads the config, **Then** the server starts successfully and the suppressed headers are associated with the petstore API.
2. **Given** a config file with no `ignore_headers` entry for an API, **When** the server loads the config, **Then** that API behaves identically to the current behavior (full schema shown, all headers validated).
3. **Given** an `ignore_headers` list containing a header name that does not appear in the API schema, **When** the server loads the config, **Then** the server starts without error (unknown names are silently ignored).

---

### User Story 2 - Hidden Headers in Schema Exploration (Priority: P2)

A user asks the server to describe or explore an API's endpoints via `explore_api` or `get_schema`. Headers listed in `ignore_headers` for that API do not appear in the output — they are invisible to the caller as if they were never declared in the schema.

**Why this priority**: Reduces noise for the caller. The primary use case is that infrastructure-managed headers (e.g., internal auth tokens injected by a gateway) should be invisible to MCP consumers.

**Independent Test**: Can be fully tested by calling `explore_api` / `get_schema` on a configured API and confirming the suppressed headers are absent from the response.

**Acceptance Scenarios**:

1. **Given** "X-Internal-Token" is in `ignore_headers` for "petstore", **When** `explore_api` is called for a petstore endpoint that declares "X-Internal-Token" as a required header, **Then** "X-Internal-Token" does not appear in the returned description.
2. **Given** "X-Internal-Token" is in `ignore_headers` for "petstore", **When** `get_schema` is called for a petstore endpoint, **Then** the schema does not list "X-Internal-Token" as a parameter.
3. **Given** "X-Internal-Token" is in `ignore_headers` for "petstore" but NOT for "bookstore", **When** `get_schema` is called for a bookstore endpoint that declares "X-Internal-Token", **Then** "X-Internal-Token" is still present in bookstore's schema output.

---

### User Story 3 - Validation Skips Ignored Headers (Priority: P3)

A user calls an HTTP tool (`http_get`, `http_post`, etc.) without supplying a header that would normally be required by the API schema. Because that header is in `ignore_headers`, validation passes and the request is forwarded to the upstream API.

**Why this priority**: Without this, the feature is incomplete — schema exploration hides the header but validation would still reject requests that omit it, making endpoints effectively unreachable through the MCP tool.

**Independent Test**: Can be fully tested by invoking `http_get` against an endpoint that has a schema-required header, omitting that header, and verifying the call succeeds rather than returning a validation error.

**Acceptance Scenarios**:

1. **Given** "Authorization" is in `ignore_headers` for "petstore", **When** `http_get /pets` is called without an "Authorization" header, **Then** the request passes validation and is forwarded to the upstream API.
2. **Given** "Authorization" is NOT in `ignore_headers`, **When** `http_get /pets` is called without the required "Authorization" header, **Then** validation fails with a clear error indicating the missing required header.
3. **Given** a header is in `ignore_headers`, **When** a caller explicitly provides that header in their `headers` argument, **Then** the provided value is still forwarded to the upstream API (suppression only waives the requirement; it does not drop a supplied value).

---

### Edge Cases

- What happens when the same header name appears in `ignore_headers` multiple times? Duplicate entries are treated as a single entry; no error is raised.
- What happens when `ignore_headers` is an empty list? Equivalent to omitting the field — all headers are shown and validated normally.
- What happens when a header is ignored but the upstream API enforces it server-side? The upstream API returns its own error response, which is forwarded to the caller unchanged. The MCP layer takes no special action.
- What happens if header name casing differs between the config and the schema (e.g., `authorization` vs `Authorization`)? Matching is case-insensitive; the header is suppressed regardless of casing.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The config format MUST support an optional `ignore_headers` field per API entry, accepting a list of header name strings.
- **FR-002**: Header names in `ignore_headers` MUST be matched case-insensitively against header names in the API schema.
- **FR-003**: The `explore_api` output MUST omit any header whose name matches an entry in the API's `ignore_headers` list.
- **FR-004**: The `get_schema` output MUST omit any header whose name matches an entry in the API's `ignore_headers` list.
- **FR-005**: Request validation for all HTTP tools (`http_get`, `http_post`, `http_put`, `http_patch`) MUST NOT treat ignored headers as required, even if the API schema marks them as required.
- **FR-006**: When a caller explicitly supplies an ignored header in their request arguments, that header value MUST be forwarded to the upstream API unchanged.
- **FR-007**: APIs with no `ignore_headers` config entry MUST behave identically to current behavior — no change to schema display or validation.
- **FR-008**: An `ignore_headers` field set to an empty list MUST be treated the same as if the field is absent.

### Key Entities

- **API Config Entry**: Per-API configuration block; gains an optional `ignore_headers` field (list of case-insensitive header name strings).
- **Ignored Header**: A header name declared in `ignore_headers` for a given API; its suppression is scoped to that API only and affects both schema display and required-field validation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of headers listed in `ignore_headers` for an API are absent from that API's `explore_api` and `get_schema` outputs.
- **SC-002**: Requests to HTTP tools that omit ignored required headers pass validation in 100% of cases (zero false validation failures due to suppressed headers).
- **SC-003**: Headers not in `ignore_headers` continue to appear in schema outputs and are enforced during validation — zero regressions on existing behavior verified by the full test suite passing.
- **SC-004**: A user can activate header suppression for an API by editing a single config section; no additional files or steps are required.

## Assumptions

- Header name matching is case-insensitive, consistent with HTTP/1.1 standards (`authorization` and `Authorization` refer to the same header).
- Suppression is scoped per-API and per-header-name. There is no global "ignore all headers" option in this feature.
- The suppression applies only within the MCP layer (display and validation). Actual HTTP forwarding to the upstream server is unaffected, except that omitted ignored headers are simply not sent.
- All existing API configs without an `ignore_headers` field continue to work without modification.
- The feature applies uniformly to all four HTTP tools; per-method suppression is out of scope.
