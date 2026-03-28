# Feature Specification: OpenAPI MCP Server

**Feature Branch**: `001-openapi-mcp-server`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "Build an MCP server that loads one or more openapi 3 API definitions and exposes them using MCP tools."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Validated HTTP Calls with Schema Feedback (Priority: P1)

An AI agent wants to call an API endpoint. It invokes one of the general-purpose HTTP tools
(GET, POST, PUT, or PATCH) with a target path and optional request data. Before any network call
is made, the server checks that the path exists in the loaded API definition and that the provided
data matches the schema for that endpoint. If either check fails, the agent receives a structured
error identifying exactly what is wrong (e.g., missing required field, wrong type, path not found).
If both checks pass, the HTTP call is executed and the full response is returned to the agent.

**Why this priority**: This is the core value proposition. Without accurate validation feedback,
the agent cannot self-correct and the server provides no benefit over a plain HTTP client.

**Independent Test**: Can be fully tested by loading a single OpenAPI definition, calling an HTTP
tool with both valid and invalid inputs, and verifying that validation errors are returned before
any network call is made, while valid calls produce actual HTTP responses.

**Acceptance Scenarios**:

1. **Given** a loaded OpenAPI definition with a `GET /pets` endpoint, **When** the agent calls
   the GET tool with path `/pets`, **Then** the server validates the path as existing, executes
   the HTTP GET request, and returns the API response.
2. **Given** a loaded OpenAPI definition with a `POST /pets` endpoint requiring a `name` field,
   **When** the agent calls the POST tool with path `/pets` and a body missing `name`, **Then**
   the server returns a validation error identifying `name` as a required missing field — without
   making any network call.
3. **Given** a loaded OpenAPI definition, **When** the agent calls any HTTP tool with a path that
   does not exist in the definition (e.g., `/nonexistent`), **Then** the server returns a
   path-not-found error with the list of closest matching paths as a hint.
4. **Given** a loaded OpenAPI definition with a required query parameter, **When** the agent calls
   the GET tool omitting that parameter, **Then** the server returns a validation error specifying
   the missing parameter before making any network call.

---

### User Story 2 - Progressive API Path Discovery (Priority: P2)

An AI agent has no prior knowledge of an API. It uses the exploration tool to iteratively discover
what paths are available. It can start with a broad query (returning all top-level paths), then
narrow by path prefix to drill into a specific resource area without being overwhelmed by the
entire API surface at once.

**Why this priority**: Discovery is the prerequisite to making any valid call. An agent that
cannot navigate the API structure cannot use the HTTP tools effectively.

**Independent Test**: Can be fully tested by loading an OpenAPI definition with at least 10 paths,
calling the explore tool without a filter (all paths returned), then calling it with a prefix
filter (only matching paths returned), and verifying the output includes method and summary for
each path.

**Acceptance Scenarios**:

1. **Given** a loaded API definition with 20 paths, **When** the agent calls the explore tool with
   no filter, **Then** all 20 paths are returned, each with their supported HTTP methods and a
   short description.
2. **Given** a loaded API definition with paths `/pets`, `/pets/{id}`, `/owners`, `/owners/{id}`,
   **When** the agent calls the explore tool filtered by prefix `/pets`, **Then** only `/pets`
   and `/pets/{id}` are returned.
3. **Given** multiple loaded API definitions, **When** the agent calls the explore tool specifying
   a particular API identifier, **Then** only the paths for that API definition are returned.

---

### User Story 3 - Schema Retrieval for a Specific Endpoint (Priority: P2)

An AI agent has identified an endpoint it wants to call. Before constructing a request, it
retrieves the full schema for that endpoint — including all required and optional parameters, the
request body structure, and the expected response format. This allows the agent to construct a
valid request on the first attempt.

**Why this priority**: Retrieving the schema before calling reduces trial-and-error loops and
allows the agent to be proactive rather than reactive about validation errors.

**Independent Test**: Can be fully tested by calling the schema tool with a known path and method,
then verifying the returned structure includes parameter definitions, request body schema, and at
least the success response schema.

**Acceptance Scenarios**:

1. **Given** a loaded API with a `POST /pets` endpoint, **When** the agent calls the schema tool
   for `POST /pets`, **Then** the response includes the request body schema (required fields,
   types, constraints) and the expected response schema.
2. **Given** a loaded API with a `GET /pets/{id}` endpoint, **When** the agent calls the schema
   tool for `GET /pets/{id}`, **Then** the response includes the path parameter definition for
   `id` and the response schema.
3. **Given** a path that does not exist in the loaded definition, **When** the agent calls the
   schema tool, **Then** a clear error is returned identifying the path as unknown.

---

### User Story 4 - Operator Configures Multiple APIs at Startup (Priority: P3)

An operator (developer or system administrator) wants to expose multiple different REST APIs
through a single MCP server instance. They provide one or more OpenAPI 3 definition files, each
paired with a configuration file that specifies the target host and optional base path. The server
loads all definitions at startup and makes all APIs available to agents simultaneously under
distinct identifiers.

**Why this priority**: Multi-API support extends the utility of the server, but a single-API
instance already delivers complete value. This story adds breadth without blocking core usage.

**Independent Test**: Can be tested by starting the server with two distinct OpenAPI definitions
and config files, then verifying that the explore tool returns paths from each API when queried
by their respective identifiers.

**Acceptance Scenarios**:

1. **Given** two OpenAPI definitions (`petstore` and `bookstore`) each with their own config file,
   **When** the server starts, **Then** both APIs are available and independently queryable by
   their configured names.
2. **Given** a config file specifying `host: https://api.example.com` and `basePath: /v2`,
   **When** the agent makes a call to `/pets`, **Then** the HTTP request is sent to
   `https://api.example.com/v2/pets`.
3. **Given** a config file with only a host (no base path override), **When** the agent makes a
   call, **Then** the base path from the OpenAPI definition itself is used as the fallback.
4. **Given** the server is started with a malformed OpenAPI definition file, **When** startup
   occurs, **Then** the server reports a clear error and refuses to start.

---

### Edge Cases

- What happens when the same path exists in two loaded API definitions and no API identifier
  is provided?
- What happens when the target host is unreachable during an HTTP call?
- What happens when the target API returns a non-2xx response — is that treated as a tool error
  or returned as data to the agent?
- What happens when an OpenAPI path uses templating (e.g., `/pets/{id}`) and the agent provides
  a literal `{id}` without substitution?
- What happens when a request body contains additional properties not defined in the schema —
  strict or lenient validation?
- What happens when an OpenAPI definition references external `$ref` schemas?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The server MUST load one or more OpenAPI 3.x definition files at startup, specified
  via a server configuration.
- **FR-002**: Each OpenAPI definition MUST be paired with an accompanying configuration file that
  specifies at minimum a target host URL and an API identifier (name).
- **FR-003**: Configuration files MUST support an optional base path override; if omitted, the
  base path declared in the OpenAPI definition is used.
- **FR-004**: The server MUST expose a general-purpose HTTP GET tool accepting a path, optional
  query parameters, optional headers, and an API identifier when multiple APIs are loaded.
- **FR-005**: The server MUST expose a general-purpose HTTP POST tool accepting a path, optional
  query parameters, optional headers, a request body, and an API identifier when multiple APIs
  are loaded.
- **FR-006**: The server MUST expose a general-purpose HTTP PUT tool accepting a path, optional
  query parameters, optional headers, a request body, and an API identifier when multiple APIs
  are loaded.
- **FR-007**: The server MUST expose a general-purpose HTTP PATCH tool accepting a path, optional
  query parameters, optional headers, a request body, and an API identifier when multiple APIs
  are loaded.
- **FR-008**: Each HTTP tool MUST validate the requested path against the loaded OpenAPI definition
  before executing any network call; if the path does not exist, a structured error MUST be
  returned including a list of similar valid paths.
- **FR-009**: Each HTTP tool MUST validate request parameters and request body against the OpenAPI
  schema for the target endpoint before executing any network call; validation errors MUST identify
  the specific field and the nature of the violation.
- **FR-010**: When validation passes, each HTTP tool MUST execute the HTTP call against the
  configured host and return the full response (status code, headers, and body) to the agent.
- **FR-011**: The server MUST expose an exploration tool that lists available API paths with their
  supported HTTP methods and descriptions. When multiple APIs are loaded, the tool MUST require an
  explicit API identifier parameter; if omitted, the server MUST return an error listing the
  available API names — consistent with the behavior of the HTTP tools (FR-014).
- **FR-012**: The exploration tool MUST support filtering results by path prefix.
- **FR-013**: The server MUST expose a schema tool that accepts a path and HTTP method and returns
  the full schema for that endpoint: path parameters, query parameters, request body schema, and
  response schemas.
- **FR-014**: When multiple API definitions are loaded and a tool call omits the API identifier,
  the server MUST return an error listing the available API names rather than guessing.
- **FR-015**: The server MUST return a structured, machine-readable error format for all validation
  failures, including: error type, affected field or path segment, and a human-readable
  description.
- **FR-016**: If any loaded OpenAPI definition file is malformed or fails OpenAPI 3 structure
  validation at startup, the server MUST report the specific error and refuse to start.

### Key Entities

- **API Definition**: An OpenAPI 3.x document describing the endpoints, schemas, and parameters
  of a single REST API.
- **API Configuration**: A companion file to an API Definition specifying the deployment target
  (host URL, optional base path override, required API identifier/name).
- **MCP Tool**: A callable function exposed by the MCP server to the AI agent (GET, POST, PUT,
  PATCH, explore, schema).
- **Validation Result**: A structured object indicating pass or fail, with field-level error
  details on failure.
- **HTTP Response**: The raw response from the target API, including status code, headers, and
  body.
- **Path Match**: The result of resolving an agent-provided concrete path against the templated
  path patterns in the OpenAPI definition (e.g., `/pets/42` → `/pets/{id}`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An AI agent with zero prior knowledge can discover all available paths for a loaded
  API within 3 tool invocations.
- **SC-002**: 100% of path and schema validation decisions are deterministic — the same input and
  API definition always produce the same validation result.
- **SC-003**: Validation error messages are self-sufficient: an AI agent can self-correct and
  produce a valid follow-up request without human intervention, measurable by a successful call
  within the same session.
- **SC-004**: An AI agent successfully navigates from zero knowledge to a correct, executed HTTP
  call within 5 tool interactions for any endpoint in a loaded definition.
- **SC-005**: A validated HTTP call (path check + schema check + network request + response)
  completes within 2 seconds under normal conditions, excluding network latency to the target host.
- **SC-006**: The server starts successfully within 3 seconds when loading up to 10 OpenAPI
  definitions at startup.

## Assumptions

- The primary consumer of the MCP tools is an AI agent, not a human end-user; tool interfaces
  are optimized for machine consumption.
- Authentication to the target API (e.g., API keys, bearer tokens) is the agent's responsibility
  and is passed as request headers. The server does not manage or inject credentials.
- The OpenAPI definition files provided at startup are syntactically valid OpenAPI 3.x documents;
  the server validates structure but does not repair malformed definitions.
- HTTP DELETE is out of scope for the initial version, as a conservative default for AI-operated
  tooling where destructive operations carry higher risk.
- Response body validation (verifying the API's response body against the spec) is out of scope
  for the initial version; the raw response is returned as-is.
- The server operates in a trusted network environment. It does not enforce rate limiting or
  access control on the MCP tools themselves.
- Path template parameters (e.g., `/pets/{id}`) are resolved by the agent providing concrete
  values; the server matches the agent's provided path to the correct template.
- Request body additional properties not defined in the schema are rejected by default
  (strict validation), consistent with the project constitution's preference for explicit
  and unambiguous data contracts.
- External `$ref` schema references within OpenAPI definitions must resolve locally or via
  accessible URLs at startup; unresolvable references cause a startup failure.
