# Feature Specification: HTTP Request Logging

**Feature Branch**: `005-http-request-logging`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "Add logging to http requests for any route, including response status code."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Request/Response Visibility (Priority: P1)

An operator running the server wants to see a log line for every inbound HTTP request — including the method, path, response status code, and elapsed time — written to standard error so it can be captured by any log aggregator or terminal session.

**Why this priority**: Without request logs, diagnosing problems (wrong status codes, unexpected traffic, slow responses) requires guessing. This is the core operational requirement of the feature.

**Independent Test**: Start the server and send several HTTP requests (to `/mcp`, `/health`, and a non-existent path). Verify that a log line appears for each request and includes the method, path, status code, and duration.

**Acceptance Scenarios**:

1. **Given** the server is running, **When** a client sends `GET /health`, **Then** a log line is written containing the method (`GET`), the path (`/health`), the response status code (`200`), and the time taken to serve the request.
2. **Given** the server is running, **When** a client sends a request that results in a non-200 status (e.g., a request to an unknown path returning 404), **Then** the log line correctly reflects the actual status code returned.
3. **Given** the server is running, **When** multiple requests arrive in sequence, **Then** a separate log line is written for each request.

---

### Edge Cases

- What happens when the server receives a request that panics or errors internally? The log line should still be written with the appropriate error status code.
- What if the request path contains query parameters? The logged path should include the full request URI (path + query string) so operators can diagnose parameter-related issues.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The server MUST log one line per inbound HTTP request across all routes (including `/health`, `/mcp`, and any unrecognised paths).
- **FR-002**: Each log line MUST include the HTTP method (e.g., GET, POST).
- **FR-003**: Each log line MUST include the request path (including query string if present).
- **FR-004**: Each log line MUST include the HTTP response status code actually sent to the client.
- **FR-005**: Each log line MUST include the elapsed time from receiving the request to sending the response.
- **FR-006**: Log output MUST be written to standard error so it is captured alongside other server diagnostic messages and does not contaminate standard output.
- **FR-007**: Logging MUST add no perceptible latency to request handling from an operator's perspective.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of requests to any route produce exactly one log line each.
- **SC-002**: Every log line includes method, path, status code, and elapsed time — verifiable by inspection for any request.
- **SC-003**: The logging overhead adds less than 1 millisecond to request round-trip time under normal operating conditions.
- **SC-004**: Operators can determine the outcome of any past request (success or error) purely from log output, without additional instrumentation.

## Assumptions

- Log lines are written in a simple, human-readable text format. Structured JSON logging is out of scope for v1.
- No log rotation, log levels, or log filtering controls are required for v1 — all requests are always logged.
- The server's existing output to standard error (startup messages, tool prefix info) is preserved unchanged; request logging is additive.
- No personally identifiable information (PII) is expected in request paths for this server's use case; path logging is safe by default.
- Log output is consumed by the operator running the process (terminal or log aggregator attached to stderr) — no persistence layer is required.
