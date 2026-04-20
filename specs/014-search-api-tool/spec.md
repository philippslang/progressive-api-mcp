# Feature Specification: search_api Tool

**Feature Branch**: `014-search-api-tool`
**Created**: 2026-04-18
**Status**: Draft
**Input**: User description: "Add a search_api tool. It has an optional filter for a single api. I searches all endpoints for a partial name match like grep would do, the path and the description. Returns api name and endpoint and schema."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Find Endpoints by Keyword Across All APIs (Priority: P1)

A developer or LLM wants to find every endpoint across every registered API whose path or description mentions a given keyword (e.g., "pet", "order", "user"), without having to enumerate paths one-by-one via `explore_api`.

**Why this priority**: This is the core value of the feature. Without global keyword search, users must guess which API holds the endpoint they need and enumerate manually — the main pain the feature is meant to remove.

**Independent Test**: Configure two or more APIs, call `search_api` with a keyword that appears in some endpoint paths or descriptions, and confirm the returned list contains every matching endpoint (from any registered API) with its API name, endpoint path+method, and schema.

**Acceptance Scenarios**:

1. **Given** two APIs are registered and the query "pet" appears in paths in both APIs, **When** `search_api` is called with `query: "pet"` and no API filter, **Then** the response contains all matching endpoints from both APIs, each identified by API name, path, method, and schema.
2. **Given** the query matches text that appears in an endpoint's description but not its path, **When** `search_api` is called, **Then** that endpoint is included in the results.
3. **Given** the query matches no endpoint anywhere, **When** `search_api` is called, **Then** the response is an empty result list (not an error).

---

### User Story 2 - Narrow Search to a Single API (Priority: P2)

The user knows the target API already and wants to search only its endpoints without noise from other APIs.

**Why this priority**: Refinement on top of US1. Adds focus when the caller already knows the relevant API, but the tool is already useful without this.

**Independent Test**: Configure two APIs containing overlapping keywords. Call `search_api` with both `query` and an `api` filter set to one of them; confirm only that API's matching endpoints are returned.

**Acceptance Scenarios**:

1. **Given** APIs "petstore" and "zoo" both contain the word "pet" in some paths, **When** `search_api` is called with `query: "pet"` and `api: "petstore"`, **Then** only petstore endpoints are returned.
2. **Given** the `api` filter names an API that does not exist, **When** `search_api` is called, **Then** the response is an error indicating the API is not registered (consistent with other tools' error handling).

---

### Edge Cases

- **Empty query**: If `query` is empty or whitespace-only, the tool returns an error (no all-endpoint dump is possible — `explore_api` already serves that use case).
- **Case sensitivity**: Matches are case-insensitive (e.g., `query: "PET"` matches path `/pets` and description `Returns a pet`).
- **Partial/substring matches**: The match is a substring match, similar to `grep` without flags — no regex, no word boundaries. `query: "pet"` matches `/pets/{id}` and also `/carpets`.
- **Path-only match AND description-only match**: Both qualify — either field matching is enough.
- **Large result sets**: If the match count is very large, the tool returns all matches (no truncation) — calibration of result-set size is the caller's responsibility via a more specific query.
- **Schema unavailable for an endpoint**: If an endpoint has no request schema defined in the OpenAPI document, the `schema` field for that result is null or omitted (the result is still returned).
- **Allow-list interaction**: Endpoints filtered out by the per-API `allow_list.paths` restriction MUST NOT appear in search results (consistent with `explore_api` behaviour).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST expose a new MCP tool named `search_api` that accepts a required `query` string parameter and an optional `api` string parameter.
- **FR-002**: When called without the `api` parameter, `search_api` MUST search across every registered API.
- **FR-003**: When called with the `api` parameter, `search_api` MUST restrict the search to endpoints of that single registered API.
- **FR-004**: The search MUST match if the `query` substring appears (case-insensitively) anywhere in the endpoint's path template OR its description.
- **FR-005**: Each result item MUST include the API name, HTTP method, path template, and the request schema (or null if none defined).
- **FR-006**: If `query` is empty or whitespace-only, the tool MUST return an error explaining that a non-empty query is required.
- **FR-007**: If `api` is provided and does not match any registered API, the tool MUST return an error indicating the API is not registered.
- **FR-008**: Results MUST exclude any endpoints that are blocked by the per-API `allow_list.paths` restriction for the relevant HTTP method.
- **FR-009**: If no endpoints match the query, the tool MUST return an empty result list successfully (not an error).
- **FR-010**: The `search_api` tool MUST be subject to the existing tool-prefix and tool-allow-list mechanisms, consistent with other MCP tools in the server.

### Key Entities

- **SearchResult**: A single matching endpoint. Fields: `api` (string — the API name), `method` (string — HTTP method, e.g., "GET"), `path` (string — OpenAPI path template), `schema` (object or null — request-body schema for the endpoint, or null when none is defined).
- **SearchRequest**: The tool input. Fields: `query` (required string), `api` (optional string — restricts to one registered API).
- **SearchResponse**: An array of `SearchResult` items (possibly empty).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can find every endpoint across all registered APIs whose path or description contains a keyword in a single tool call, without needing to enumerate APIs manually.
- **SC-002**: 100% of matches returned actually contain the query substring (case-insensitive) in either path or description; 0% false positives.
- **SC-003**: 100% of endpoints that contain the query substring AND pass the allow-list are present in results; 0% false negatives for allowed endpoints.
- **SC-004**: Searches with an unknown `api` filter return an actionable error identifying the unknown name within a single tool call.
- **SC-005**: Searches with an empty query return an actionable error explaining the required-input condition within a single tool call.

## Assumptions

- "Endpoints" refers to operations defined in each API's OpenAPI document — one entry per (path, method) pair.
- "Description" means the operation-level description/summary in the OpenAPI document (the field that describes what the endpoint does). Path-item-level descriptions are included if they are the only description present.
- "Schema" means the request-body schema for the endpoint, resolved from the OpenAPI document. If the endpoint has no request body, the schema is null.
- The tool is purely a discovery aid — it performs no upstream HTTP calls and does not trigger validation.
- Result ordering is stable but unspecified to the caller (implementation-chosen, e.g., iteration order of APIs then paths).
- Consistent with existing tools (`explore_api`, `get_schema`), allow-list restrictions apply. Tools blocked entirely by the tool allow-list do not register `search_api` for that API.
- No pagination or result limits are imposed; callers refine via more specific query strings.
