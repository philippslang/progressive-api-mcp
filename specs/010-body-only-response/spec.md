# Feature Specification: Body-Only HTTP Responses

**Feature Branch**: `010-body-only-response`
**Created**: 2026-03-29
**Status**: Draft
**Input**: User description: "Add an API level option to the configuration for http methods to only return the body, not the header"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Body-Only Response Per API (Priority: P1)

A user adds a `response_body_only` flag to an API entry in the config file. When this flag is enabled, all HTTP tool calls for that API return only the response body instead of the full response object (which currently includes status code, headers, and body).

**Why this priority**: This is the foundational capability. It reduces response noise for callers who only care about the returned data and don't need to inspect status codes or response headers — a common scenario when the upstream API is well-behaved and errors are indicated in the body.

**Independent Test**: Can be fully tested by configuring `response_body_only: true` for a test API, calling any HTTP tool, and verifying the response contains only the body content rather than a structured object with `status_code`, `headers`, and `body` fields.

**Acceptance Scenarios**:

1. **Given** `response_body_only: true` is set for API "petstore", **When** `http_get /pets` is called, **Then** the result contains only the raw body content (e.g. the JSON array of pets), with no `status_code` or `headers` fields in the output.
2. **Given** `response_body_only: false` (or the field is absent) for an API, **When** any HTTP tool is called, **Then** the result contains the full response object with `status_code`, `headers`, and `body` fields — identical to the current behavior.
3. **Given** `response_body_only: true` and the upstream API returns a non-2xx status, **When** an HTTP tool is called, **Then** the result still contains only the body (the caller receives whatever the upstream returned as body, without status code wrapping).

---

### User Story 2 - Body-Only Works Across All HTTP Tools (Priority: P2)

The `response_body_only` setting applies uniformly to `http_get`, `http_post`, `http_put`, and `http_patch`. A user does not need to configure it per-tool — one API-level flag covers all four.

**Why this priority**: Without uniform coverage, users would get inconsistent behavior across HTTP methods, making the feature unreliable and confusing.

**Independent Test**: Can be fully tested by enabling `response_body_only` for an API and calling each of the four HTTP tools in turn, verifying each returns body-only output.

**Acceptance Scenarios**:

1. **Given** `response_body_only: true` for an API, **When** `http_post`, `http_put`, or `http_patch` is called with a valid request, **Then** each returns only the response body.
2. **Given** `response_body_only: true` is set on API "petstore" but not on API "bookstore", **When** `http_get` is called for bookstore, **Then** bookstore returns the full response object; petstore returns body only.

---

### Edge Cases

- What happens when the upstream API returns an empty body? The result is an empty string or null value — not an error.
- What happens when the upstream API returns a non-JSON body with `response_body_only: true`? The raw body text is returned as-is (same as the current body field behavior).
- What if the caller needs to detect errors when `response_body_only: true` is set? This is the caller's responsibility; the feature makes no special provision for error detection. This is by design — the feature is for well-behaved APIs where the body alone carries sufficient information.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The config format MUST support an optional `response_body_only` boolean field per API entry.
- **FR-002**: When `response_body_only` is `true` for an API, the response of all HTTP tool calls for that API MUST contain only the response body content.
- **FR-003**: When `response_body_only` is `false` or absent for an API, the response of HTTP tool calls MUST be the full response object containing `status_code`, `headers`, and `body` — identical to current behavior.
- **FR-004**: The `response_body_only` setting MUST apply uniformly to `http_get`, `http_post`, `http_put`, and `http_patch`.
- **FR-005**: When `response_body_only` is `true` and the upstream returns a JSON body, the result MUST be the parsed JSON value directly (not wrapped in an outer object).
- **FR-006**: When `response_body_only` is `true` and the upstream returns a non-JSON body, the result MUST be the raw body text directly.
- **FR-007**: APIs with no `response_body_only` field MUST behave identically to current behavior (full response object returned).

### Key Entities

- **API Config Entry**: Per-API configuration block; gains an optional `response_body_only` boolean field (default: `false`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of HTTP tool responses for APIs with `response_body_only: true` contain only the body content — no `status_code` or `headers` fields in the output.
- **SC-002**: 100% of HTTP tool responses for APIs without `response_body_only` (or with it set to `false`) are identical in shape to the current behavior — zero regressions.
- **SC-003**: A user can enable body-only responses for an API by adding a single boolean field to one config section; no additional steps or files are required.
- **SC-004**: The setting applies to all four HTTP tools without any per-tool configuration.

## Assumptions

- "Body only" means the response value is exactly what the upstream server returned as its body — no additional wrapping, envelope, or metadata.
- When the body is JSON, it is returned as a parsed value (same representation as the current `body` field in the full response).
- When the body is non-JSON (e.g. plain text, HTML), it is returned as a string.
- The feature does not affect error detection or error handling; it is the caller's responsibility to handle upstream error responses when operating in body-only mode.
- The `response_body_only` flag applies to the MCP response only; the actual HTTP request and all validation behavior is unchanged.
- Existing API configs without `response_body_only` continue to work without modification.
