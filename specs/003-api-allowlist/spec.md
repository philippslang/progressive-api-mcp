# Feature Specification: API Tool & Path Allow List

**Feature Branch**: `003-api-allowlist`
**Created**: 2026-03-28
**Status**: Draft
**Input**: User description: "Extend the configuration by creating an allow list for every openapi spec. This allow list should contain what mcp tools are allowed, and for every mcp tool what paths are allowed, where paths is the same as in the openapi doc, e.g. '/pets'. By default, all paths and tools are allowed."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Restrict Which Tools Are Exposed Per API (Priority: P1)

An operator configuring the server wants to limit which MCP tools are available for a given API. For example, they want agents to only be able to read data (explore and GET), not write it (POST, PUT, PATCH). They add an allow list to the API entry in the config file listing only the tools they want exposed.

**Why this priority**: The most fundamental control — deciding which categories of action are permitted at all. Blocks an entire class of risk (writes, mutations) with a single config line.

**Independent Test**: Configure a Petstore API entry with `allowed_tools: [explore_api, get_schema, http_get]`. Verify that calling `http_post` returns a tool-not-found error, while `http_get` succeeds.

**Acceptance Scenarios**:

1. **Given** an API configured with `allowed_tools: [explore_api, http_get]`, **When** an agent calls `http_post`, **Then** the tool does not exist in the MCP session and the call fails.
2. **Given** an API configured with `allowed_tools: [explore_api, http_get]`, **When** an agent calls `explore_api`, **Then** it succeeds and returns path information.
3. **Given** an API configured with no `allowed_tools` field, **When** any tool is called, **Then** all 6 tools are available (default allow-all behaviour).

---

### User Story 2 — Restrict Which Paths Are Allowed Per Tool (Priority: P2)

An operator wants to allow `http_get` but only against certain paths — for instance, only `/pets` and `/pets/{id}`, not `/admin` or `/owners`. They configure per-tool path allow lists inside the API entry.

**Why this priority**: Path-level restriction is the fine-grained complement to tool-level restriction. Together they provide complete surface-area control without touching the OpenAPI document itself.

**Independent Test**: Configure `http_get` with `allowed_paths: ["/pets", "/pets/{id}"]`. Verify that `GET /owners` is rejected with a `PATH_NOT_PERMITTED` response while `GET /pets` succeeds.

**Acceptance Scenarios**:

1. **Given** `http_get` restricted to `allowed_paths: ["/pets"]`, **When** the agent calls `http_get` with `path: "/pets"`, **Then** the request is executed normally.
2. **Given** `http_get` restricted to `allowed_paths: ["/pets"]`, **When** the agent calls `http_get` with `path: "/owners"`, **Then** the server returns a `PATH_NOT_PERMITTED` error identifying the denied path.
3. **Given** a tool with no `allowed_paths` entry but the tool itself is allowed, **When** the agent calls any valid path, **Then** all paths defined in the OpenAPI document are accessible.
4. **Given** `http_get` with `allowed_paths: ["/pets"]`, **When** the agent calls `http_get` with `path: "/pets/42"` (concrete path matching template `/pets/{id}`), **Then** the call is rejected because `/pets/{id}` is not in the allow list.

---

### User Story 3 — Combine Tool and Path Restrictions (Priority: P3)

An operator applies both tool-level and path-level restrictions in the same API entry: only `explore_api`, `http_get`, and `http_post` are allowed, and `http_post` is further limited to `/pets` only.

**Why this priority**: Real-world configurations will mix both levels. This story validates that they compose correctly.

**Independent Test**: Configure the combined restrictions. Verify `http_post /pets` succeeds, `http_post /owners` is denied, and `http_put /pets` returns tool-not-found.

**Acceptance Scenarios**:

1. **Given** combined restrictions (tools: explore_api, http_get, http_post; http_post paths: /pets), **When** `http_post /pets` is called, **Then** it executes normally.
2. **Given** the same combined restrictions, **When** `http_post /owners` is called, **Then** a `PATH_NOT_PERMITTED` error is returned.
3. **Given** the same combined restrictions, **When** `http_put /pets` is called, **Then** the tool does not exist in the MCP session.

---

### Edge Cases

- What happens when `allowed_tools` is an empty list? No tools are registered for that API — every tool call fails with tool-not-found.
- What if a path in `allowed_paths` does not exist in the OpenAPI document? It is silently ignored at startup; it will simply never match a real request.
- What if `allowed_paths` is an empty list for a tool? The tool is registered but every path call returns `PATH_NOT_PERMITTED`.
- What if the same path is listed twice in `allowed_paths`? Duplicates are deduplicated silently.
- Path matching in `allowed_paths` uses exact OpenAPI template strings (e.g. `/pets/{id}`), not concrete runtime values (e.g. `/pets/42`).
- When multiple APIs are loaded, each API's allow list is independent — restricting one API has no effect on another.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Each API entry in the configuration MUST support an optional `allow_list` section containing tool and path restrictions.
- **FR-002**: `allow_list` MUST support an optional `tools` field: a list of tool names drawn from the fixed set `explore_api`, `get_schema`, `http_get`, `http_post`, `http_put`, `http_patch`.
- **FR-003**: When `allow_list.tools` is absent or null, the system MUST behave as if all 6 tools are allowed (default allow-all).
- **FR-004**: `allow_list` MUST support an optional `paths` field: a map from tool name to a list of OpenAPI path template strings (e.g. `/pets`, `/pets/{id}`).
- **FR-005**: When a tool has no entry in `allow_list.paths`, the system MUST allow all paths for that tool.
- **FR-006**: When a tool has an entry in `allow_list.paths`, the system MUST reject requests to any path not in that list with a `PATH_NOT_PERMITTED` error that identifies the denied path.
- **FR-007**: Path matching in `allow_list.paths` MUST use exact OpenAPI template string comparison — not concrete path matching.
- **FR-008**: The system MUST only register MCP tools that appear in `allow_list.tools` for that API (or all tools when `allow_list.tools` is absent).
- **FR-009**: The configuration loader MUST validate that any tool name in `allow_list.tools` or `allow_list.paths` keys is one of the 6 known names, returning a descriptive error for unknown names.
- **FR-010**: Allow list restrictions MUST be scoped per API entry; different APIs in the same server instance can have different restrictions.
- **FR-011**: Existing API config entries without an `allow_list` field MUST continue to work identically to before — zero behaviour change.

### Key Entities

- **APIAllowList**: The restriction block inside an API config entry. Contains an optional list of allowed tool names and an optional map of tool name → allowed path templates.
- **Tool Name**: One of the 6 fixed base MCP tool names: `explore_api`, `get_schema`, `http_get`, `http_post`, `http_put`, `http_patch`. Always expressed without prefix in the allow list.
- **Allowed Path**: An OpenAPI path template string exactly as it appears in the spec document (e.g. `/pets`, `/pets/{id}`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator can restrict an API to read-only access by adding a 3-line config block — zero changes to the OpenAPI document required.
- **SC-002**: A denied tool is absent from the MCP session entirely (not merely returning errors) — 100% of the time.
- **SC-003**: A denied path returns `PATH_NOT_PERMITTED` in under the same latency as a normal `PATH_NOT_FOUND` response.
- **SC-004**: An unknown tool name in `allow_list.tools` is caught at server startup with a clear error, before any agent connects — 100% of the time.
- **SC-005**: All existing behaviour is preserved when no `allow_list` is configured — zero regressions on the existing test suite.

## Assumptions

- The allow list uses base tool names (without any configured `tool_prefix`). If a `tool_prefix` is also set, the prefix is applied after allow list filtering — operators do not need to include the prefix in their allow list.
- Path matching is exact template-string comparison: `/pets/{id}` is distinct from `/pets/42`. This keeps the allow list readable and predictable.
- When multiple APIs are loaded, the allow list is scoped per API entry. There is no global allow list.
- The `allow_list` field is optional and purely additive to the existing `APIConfig` structure — existing configurations without it continue to work unchanged.
- Enforcement happens at two levels: tool registration time (tools not in the allow list are never registered) and request time (paths not in the allow list are rejected before validation or execution).
