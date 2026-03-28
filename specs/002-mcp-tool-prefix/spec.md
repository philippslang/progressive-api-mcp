# Feature Specification: MCP Tool Name Prefix

**Feature Branch**: `002-mcp-tool-prefix`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "The configuration file or command line argument needs to contain a prefix for all mcp tool names."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Tool Name Prefix at Startup (Priority: P1)

An operator deploying the MCP server wants all exposed tool names to carry a custom prefix so
that tool names are namespaced and cannot conflict with other MCP servers in the same agent
environment. They add a single prefix setting to the configuration file (or pass it as a CLI
argument). When the server starts, every tool it exposes — `http_get`, `http_post`, `http_put`,
`http_patch`, `explore_api`, `get_schema` — is registered under the prefixed name (e.g.,
`myapi_http_get`, `myapi_explore_api`). The agent discovers and calls the prefixed names; the
server handles them correctly.

**Why this priority**: Without this, operators cannot safely run multiple MCP servers in the
same agent context because tool names will collide. This is the core value of the feature.

**Independent Test**: Can be fully tested by starting the server with `tool_prefix: "myapi"` in
the config and verifying that all six tools are registered as `myapi_http_get`,
`myapi_http_post`, `myapi_http_put`, `myapi_http_patch`, `myapi_explore_api`,
`myapi_get_schema` — and that calling each prefixed name works correctly.

**Acceptance Scenarios**:

1. **Given** a config file with `tool_prefix: "myapi"`, **When** the server starts, **Then**
   all six MCP tools are registered with names starting with `myapi_` and the original
   unprefixed names are NOT registered.
2. **Given** a prefixed server, **When** an agent calls `myapi_http_get` with a valid path,
   **Then** the call succeeds exactly as `http_get` would without a prefix.
3. **Given** a config file with no `tool_prefix` set, **When** the server starts, **Then** all
   tools are registered under their original names (`http_get`, `explore_api`, etc.) — no
   prefix applied.

---

### User Story 2 - Override Prefix via CLI Argument (Priority: P2)

An operator wants to override the prefix at runtime without editing the configuration file —
for example, to run two instances of the server side by side with different prefixes, or to
use a deployment-specific prefix set in a startup script. They pass `--tool-prefix myapi` on
the command line (or set the equivalent environment variable). This value takes precedence over
any `tool_prefix` value in the config file, following the same precedence rules as all other
server settings.

**Why this priority**: Config-file-only configuration is sufficient for most deployments, but
CLI override is needed for scripted multi-instance deployments. It also completes the
conventional precedence contract already established for `host`, `port`, and `transport`.

**Independent Test**: Can be tested by starting the server with `tool_prefix: "fromfile"` in
the config and `--tool-prefix fromcli` on the command line, then verifying all tools are
registered as `fromcli_*` and NOT `fromfile_*`.

**Acceptance Scenarios**:

1. **Given** a config file with `tool_prefix: "fromfile"` and CLI flag `--tool-prefix fromcli`,
   **When** the server starts, **Then** all tools are registered as `fromcli_*`.
2. **Given** a config file with `tool_prefix: "fromfile"` and environment variable
   `PROGRAPIMCP_SERVER_TOOL_PREFIX=fromenv`, **When** the server starts without a CLI flag,
   **Then** all tools are registered as `fromenv_*`.
3. **Given** no prefix in config file and no CLI flag and no environment variable, **When** the
   server starts, **Then** tools use their original unprefixed names.

---

### Edge Cases

- What happens when the prefix is an empty string — is it treated as "no prefix"?
- What happens when the prefix contains characters that are invalid in MCP tool names
  (spaces, special characters)?
- What happens when the prefix ends with `_` — does the system add another `_`, resulting in
  double underscores like `myapi__http_get`?
- What happens when two running server instances register tools with the same prefix in the
  same agent context?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The server configuration MUST support a `tool_prefix` field in the server
  settings section; when set, its value is prepended to all MCP tool names at registration
  time.
- **FR-002**: The prefix MUST be joined to the tool name with a single underscore separator
  (e.g., prefix `"myapi"` + tool `"http_get"` = `"myapi_http_get"`).
- **FR-003**: When `tool_prefix` is absent or an empty string, tools MUST be registered under
  their original names with no modification.
- **FR-004**: The server MUST accept a `--tool-prefix` CLI flag that overrides the config file
  value, following the same precedence rules as `--host`, `--port`, and `--transport`
  (CLI flag > environment variable > config file > default/none).
- **FR-005**: The equivalent environment variable MUST follow the existing naming convention:
  `PROGRAPIMCP_SERVER_TOOL_PREFIX`.
- **FR-006**: The prefix MUST be validated at startup; it MUST contain only alphanumeric
  characters and underscores (`[a-zA-Z0-9_]`), and MUST NOT be purely numeric. If invalid, the
  server MUST report a clear error and refuse to start.
- **FR-007**: A trailing underscore in the prefix value MUST NOT cause a double underscore in
  the resulting tool name; the system MUST strip a trailing `_` from the prefix before
  concatenating.
- **FR-008**: The prefix MUST apply uniformly to all tools registered by the server:
  `http_get`, `http_post`, `http_put`, `http_patch`, `explore_api`, `get_schema`, and any
  future tools added to the server.
- **FR-009**: The server startup log MUST include the effective prefix (or indicate "no prefix")
  so operators can verify the configuration at a glance.

### Key Entities

- **Tool Prefix**: A short alphanumeric+underscore string that is prepended to all MCP tool
  names. It is part of the server configuration and applies globally to all tools for a given
  server instance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of registered tool names carry the configured prefix — no tool is
  accessible under its unprefixed name when a prefix is set.
- **SC-002**: An operator can change the active prefix solely by editing one field (in the
  config file or CLI argument), with no other changes required.
- **SC-003**: Two server instances with different prefixes can be connected to the same agent
  simultaneously without any tool name collision.
- **SC-004**: A server started with an invalid prefix value fails within 1 second with a
  human-readable error message that identifies the invalid characters.

## Assumptions

- The tool name prefix is a server-wide setting; per-API or per-tool prefix granularity is
  out of scope.
- The separator between prefix and tool name is always `_` (underscore); it is not
  configurable.
- The prefix is applied only at registration time; the internal names used within the server
  code are not affected.
- Agents interacting with the server are expected to discover available tool names via the MCP
  protocol's tool-listing capability; they do not hard-code tool names.
- This feature builds on the existing configuration system established in feature
  `001-openapi-mcp-server`; no new configuration file format or CLI framework is introduced.
