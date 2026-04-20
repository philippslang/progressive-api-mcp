# Feature Specification: Rename `explore_api` to `list_api`

**Feature Branch**: `015-rename-explore-to-list`
**Created**: 2026-04-18
**Status**: Draft
**Input**: User description: "Change the name of explore_api to list_api"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Call the Tool by Its New Name (Priority: P1)

A user of the MCP server (human operator or an AI client) wants to see the endpoints offered by a registered API. Today they call a tool named `explore_api`; going forward, they call `list_api` instead — the behaviour, inputs, and outputs are identical, only the name changes.

**Why this priority**: This is the entire feature. Without the rename being usable end-to-end, there is nothing to ship.

**Independent Test**: Start the server with a configured API, issue a `tools/list` MCP request, and confirm the tool is advertised as `list_api` (or `<prefix>_list_api` when a tool prefix is configured) and no tool named `explore_api` is present. Call `list_api` with and without the `api` parameter and confirm the same behaviour that `explore_api` previously exhibited.

**Acceptance Scenarios**:

1. **Given** a server with one registered API, **When** the client lists tools, **Then** `list_api` appears and `explore_api` does not.
2. **Given** `list_api` is called with no arguments and a single API is registered, **Then** the response contains that API's path list — identical to the previous `explore_api` output.
3. **Given** `list_api` is called with `api: "petstore"`, **Then** the response contains petstore's path list.
4. **Given** a server configured with `tool_prefix: "myapi"`, **When** the client lists tools, **Then** the tool is named `myapi_list_api` and no tool named `myapi_explore_api` exists.

---

### User Story 2 - Restrict the Renamed Tool via Allow-List (Priority: P2)

An operator who restricts which tools are exposed per API via `allow_list.tools` needs to reference the tool by its new name.

**Why this priority**: Existing deployments that use `allow_list.tools: [explore_api, ...]` must be updated; the config must validate the new name and reject the old one so misconfigurations are surfaced at load time rather than silently ignored.

**Independent Test**: Load a config whose `allow_list.tools` contains `list_api` — the server starts and the tool is registered. Load a config whose `allow_list.tools` contains `explore_api` — the server refuses to start with an error that names `list_api` in the valid-values hint.

**Acceptance Scenarios**:

1. **Given** a config with `allow_list.tools: [list_api]`, **When** the server starts, **Then** startup succeeds and only `list_api` is registered for that API.
2. **Given** a config with `allow_list.tools: [explore_api]`, **When** the server starts, **Then** startup fails with an error indicating `explore_api` is not a known tool and listing `list_api` as a valid name.
3. **Given** a config with `allow_list.paths: { list_api: ["/pets"] }`, **When** `list_api` is called for that API, **Then** only `/pets` appears in the output.

---

### Edge Cases

- A config that used the old `allow_list.tools: [explore_api]` must not silently drop the tool; it must fail validation so operators notice and update their config.
- Tool-prefix handling: the prefix (e.g. `myapi_`) is applied to the new base name (`list_api`), not to any residual `explore_api` identifier.
- Documentation and example configs that reference `explore_api` must be updated so new users are not led astray.
- Any test, fixture, or doc that calls `explore_api` by name must be updated; otherwise regression tests would continue to exercise a tool that no longer exists.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The MCP tool that today is named `explore_api` MUST be advertised under the name `list_api` (subject to the configured tool prefix).
- **FR-002**: No tool named `explore_api` MUST be registered, exposed, or callable after this change.
- **FR-003**: The tool's inputs, outputs, error codes, and per-API allow-list semantics MUST remain identical — only the name changes.
- **FR-004**: `allow_list.tools` MUST accept `"list_api"` as a valid value and MUST reject `"explore_api"` with an error message that lists `list_api` among the accepted tool names.
- **FR-005**: `allow_list.paths` MUST accept `"list_api"` as a key and MUST reject `"explore_api"` as a key with the same error behaviour.
- **FR-006**: The example configuration file and any human-facing documentation that mentions `explore_api` MUST be updated to say `list_api`.
- **FR-007**: All automated tests that reference `explore_api` by name MUST be updated so the test suite continues to pass after the rename.

### Key Entities

- **Tool name**: The identifier the MCP client uses to invoke the operation. Changes from `explore_api` to `list_api`.
- **Allow-list tool identifier**: The string users write in `allow_list.tools` and `allow_list.paths` to reference this tool. Changes in lockstep with the tool name.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After the change, 100% of MCP clients that call `list_api` receive the same response they would previously have received from `explore_api`.
- **SC-002**: After the change, 0 MCP tools are registered under the old name `explore_api` — verifiable via a `tools/list` response containing no matching entry.
- **SC-003**: A configuration file that references `explore_api` in `allow_list.tools` or `allow_list.paths` fails validation with a clear error message in under 1 second of load time.
- **SC-004**: The project's example configuration and any user-facing documentation contain 0 occurrences of the identifier `explore_api`.
- **SC-005**: The full automated test suite passes with 0 references to `explore_api` remaining in test code.

## Assumptions

- **Hard rename, no alias**: The old name `explore_api` is removed outright rather than kept as a deprecated alias. The user asked to "change the name," not to "add an alias." Operators with existing configs will see an explicit validation error on startup and update the name — this is preferable to silent acceptance that masks config drift.
- **No behavioural change**: The handler logic, inputs, outputs, error codes, and allow-list semantics are unchanged. This is purely a rename.
- **Tool prefix mechanism is unchanged**: The existing `server.tool_prefix` setting continues to apply to the new base name.
- **In-repo scope only**: "Documentation" here means files inside this repository (example configs, README, spec artefacts). External consumers' documentation is their responsibility.
- **Spec artefact preservation**: Historical feature specs under `specs/001-…` through `specs/014-…` that reference `explore_api` are frozen historical records and are out of scope for renaming; only live code, tests, and user-facing docs are updated.
