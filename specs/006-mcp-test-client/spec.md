# Feature Specification: MCP Test Client CLI

**Feature Branch**: `006-mcp-test-client`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "create a test client for this. it should be another cli that only depends on mcp and lists all files given the address of the mcp server."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List All MCP Tools (Priority: P1)

A developer or operator wants to verify what tools an MCP server is exposing by running a lightweight diagnostic CLI. They provide the server's HTTP address on the command line, and the tool connects, discovers all registered tools, and prints their names to standard output — one per line.

**Why this priority**: This is the sole purpose of the tool. Without it, the only way to inspect what a running server exposes is to use a full MCP client or read the server's source code. A dedicated CLI makes this instant and scriptable.

**Independent Test**: Run `mcpls http://127.0.0.1:8080/mcp` against a running server and verify the output lists the expected tool names, one per line, and exits with code 0.

**Acceptance Scenarios**:

1. **Given** an MCP server is running at a known address, **When** the CLI is invoked with that address, **Then** it prints each registered tool name on its own line and exits successfully.
2. **Given** an MCP server is running, **When** the CLI is invoked, **Then** the output is machine-parseable (one tool name per line, no decorations) so it can be piped to other tools (e.g., `grep`, `wc -l`).
3. **Given** an invalid or unreachable address is provided, **When** the CLI is invoked, **Then** it prints a clear error message to standard error and exits with a non-zero exit code.
4. **Given** the server is reachable but returns no tools, **When** the CLI is invoked, **Then** it prints nothing to standard output and exits successfully.

---

### Edge Cases

- What if the server address is missing the `/mcp` path suffix? The CLI should accept a bare host:port and append the default path, or clearly document that the full endpoint URL is required.
- What if the server takes too long to respond? The CLI should time out after a reasonable interval and exit with an error.
- What if the MCP handshake succeeds but the tool list response is malformed? The CLI should surface a clear error rather than silently producing incorrect output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST accept the MCP server's endpoint URL as a required positional argument.
- **FR-002**: The CLI MUST connect to the server, perform the MCP initialization handshake, and request the list of registered tools.
- **FR-003**: The CLI MUST print each tool name to standard output, one name per line, in the order returned by the server.
- **FR-004**: The CLI MUST exit with code 0 on success (including when the server returns an empty tool list).
- **FR-005**: The CLI MUST print a human-readable error message to standard error and exit with a non-zero code when the server is unreachable, the handshake fails, or the tool list cannot be retrieved.
- **FR-006**: The CLI MUST time out if the server does not respond within a reasonable interval (default: 10 seconds) and report the timeout as an error.
- **FR-007**: The CLI MUST have no dependency on the OpenAPI parsing or registry components of this project — it must function as a standalone MCP protocol client.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can discover all tools on a running server with a single command in under 5 seconds.
- **SC-002**: The output can be consumed by standard shell tools (`grep`, `wc`, `sort`) without post-processing.
- **SC-003**: 100% of connection or protocol errors produce a non-zero exit code and a message on stderr — no silent failures.
- **SC-004**: The binary can be built and run independently of the main server binary with no shared runtime dependencies beyond the MCP protocol library.

## Assumptions

- "Lists all files" is interpreted as listing all MCP tools registered on the server (the set of capabilities the server advertises via the MCP `tools/list` response).
- The CLI is a diagnostic/developer tool, not a production client; no authentication or TLS is required for v1.
- The server address is provided as a full URL (e.g., `http://127.0.0.1:8080/mcp`). Appending a default path from a bare host:port is out of scope for v1.
- The CLI binary is named `mcpls` (an analogy to `ls` — lists what the server has).
- The tool only uses the same MCP protocol library already used by the main server; no other new dependencies are introduced.
