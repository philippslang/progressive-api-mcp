# Feature Specification: API-Level Skip Validation Configuration

**Feature Branch**: `013-api-skip-validation`
**Created**: 2026-03-31
**Status**: Draft
**Input**: User description: "Add a configuration at the api level to skip validation of payload in the mcp server and forward it straight to the api instead."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enable Skip Validation for a Specific API (Priority: P1)

A developer configuring the MCP server wants to disable request payload validation for a specific API definition so that requests are forwarded directly to the upstream API without being rejected by the MCP server's validator.

**Why this priority**: Core feature requirement. Without this, the skip-validation flag serves no purpose. This is the primary use case that delivers value.

**Independent Test**: Can be fully tested by setting `skip_validation: true` on a configured API, sending a request with a payload that would normally fail schema validation, and confirming the request reaches the upstream API successfully and is not rejected by the MCP server.

**Acceptance Scenarios**:

1. **Given** a configured API with `skip_validation: true`, **When** a tool call is made with a request payload that violates the API schema, **Then** the MCP server forwards the payload to the upstream API without returning a validation error.
2. **Given** a configured API with `skip_validation: true`, **When** a tool call is made with a valid payload, **Then** the request is forwarded to the upstream API as normal.
3. **Given** a configured API without `skip_validation` set (or set to `false`), **When** a tool call is made with an invalid payload, **Then** the MCP server returns a validation error and does not forward the request.

---

### User Story 2 - Per-API Independent Configuration (Priority: P2)

A developer has multiple APIs configured in the same MCP server instance and wants to skip validation on one API while keeping validation enabled on others.

**Why this priority**: Without per-API isolation, users would be forced into an all-or-nothing choice. Per-API control enables selective trust of upstream APIs.

**Independent Test**: Can be tested by configuring two APIs — one with `skip_validation: true` and one without — then verifying that invalid payloads are forwarded only for the first API while the second API still rejects them.

**Acceptance Scenarios**:

1. **Given** two APIs configured — API-A with `skip_validation: true` and API-B with `skip_validation: false`, **When** an invalid payload is sent to API-A, **Then** it is forwarded to API-A without error.
2. **Given** the same configuration, **When** an invalid payload is sent to API-B, **Then** the MCP server rejects it with a validation error.

---

### Edge Cases

- What happens when `skip_validation: true` is set but the upstream API itself rejects the payload? The upstream API response (including error) is returned to the caller as-is.
- What happens when the API URL is unreachable with `skip_validation: true`? Network/connectivity errors are still surfaced normally.
- What happens if `skip_validation` is omitted from the config? Default behavior is to validate (backward-compatible default: `false`).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The API configuration MUST support an optional `skip_validation` boolean field at the API level.
- **FR-002**: When `skip_validation` is `true` for an API, the MCP server MUST bypass request payload schema validation for all tool calls targeting that API.
- **FR-003**: When `skip_validation` is `true`, the MCP server MUST forward the payload to the upstream API unchanged.
- **FR-004**: When `skip_validation` is `false` or absent, the MCP server MUST continue to validate request payloads before forwarding (existing behavior preserved).
- **FR-005**: The `skip_validation` flag MUST be independently configurable per API; changing it for one API MUST NOT affect other APIs.
- **FR-006**: The configuration loading system MUST correctly parse and apply the `skip_validation` field without requiring changes to other config fields.

### Key Entities

- **API Config**: Represents a single upstream API registration. Gains a new optional `skip_validation` boolean attribute.
- **Payload Validation Step**: The conditional validation gate applied per tool call — executed when `skip_validation` is false, bypassed when true.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can enable skip-validation for any configured API by changing a single config field, with no code changes required.
- **SC-002**: 100% of tool calls to a skip-validation-enabled API that carry schema-invalid payloads are forwarded to the upstream without validation rejection.
- **SC-003**: 100% of tool calls to APIs with validation enabled continue to be validated (no regressions in existing behavior).
- **SC-004**: A missing or unset `skip_validation` value defaults to validation-enabled behavior, causing zero unintended bypass of validation.

## Assumptions

- The existing per-API configuration structure supports adding new optional boolean fields without breaking existing configs that omit them.
- "API level" means the configuration block for a single upstream API, not a global server-wide setting.
- Skipping validation applies to outbound request payloads only (not response validation or header validation).
- The feature does not change how the MCP server constructs tool input schemas for the model — schema exposure remains unchanged regardless of `skip_validation`.
- Backward compatibility is required: existing configs without `skip_validation` must continue to work as before with validation enabled.
