# Research: search_api Tool

**Branch**: `014-search-api-tool` | **Date**: 2026-04-18

## Decision 1: Tool registration shape — global (single registration) or per-API

**Decision**: One global registration via `RegisterSearchTools(s, reg, prefix, allowedTools)`, matching the existing `RegisterExploreTools`/`RegisterSchemaTools` pattern. The tool accepts an optional `api` parameter.

**Rationale**: `explore_api` and `get_schema` are also cross-API tools registered once, using an optional/required `api` argument to select which API the caller wants. `search_api` extends that pattern: it naturally spans APIs (US1), with the `api` filter narrowing to one (US2).

**Alternatives considered**:
- Register one `search_api` per API (same as how some MCP servers expose operation-scoped tools) — rejected because US1 explicitly requires a single call that searches across all APIs.

---

## Decision 2: Which operation fields to match against

**Decision**: Match against the path template AND, for each HTTP method (GET/POST/PUT/PATCH/DELETE), the operation's `Summary` and `Description`. If both summary and description are empty, fall back to the path-item-level description (if any).

**Rationale**: OpenAPI conventions use `summary` as a short human label and `description` as a longer prose field. Both are "description" in the user's spec language. Matching both maximises recall without introducing false positives.

**Alternatives considered**:
- Match only `Summary` — rejected; many OpenAPI documents put the meaningful text only in `Description`.
- Match against operation `OperationID` too — rejected for v1; the user spec says "path and description" explicitly, and operation IDs are often opaque identifiers.

---

## Decision 3: Match semantics — substring, case-insensitive, no regex

**Decision**: `strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))`. The needle is the raw `query` string after trimming leading/trailing whitespace.

**Rationale**: The spec says "like grep would do" + "partial name match". Default grep-like behaviour is substring. Regex would require syntax documentation and error handling; users who want regex can post-process.

**Alternatives considered**:
- Regex — rejected for v1 (expanded scope, more error paths).
- Word-boundary match — rejected; the spec explicitly cites "partial" match and the edge case `"pet" matches "/carpets"` is called out as acceptable.

---

## Decision 4: Schema field contents

**Decision**: `schema` is the resolved request-body schema (same shape that `get_schema` returns under `request_body.schema`), or `null` when the operation has no request body. Reuse the existing `schemaToMap` helper in `pkg/tools/schema.go` (same package — no export change required).

**Rationale**: The user said "Returns api name and endpoint and schema." For an endpoint, "schema" most naturally means the request-body schema — that's what callers need to construct a call. Reusing `schemaToMap` keeps output shape consistent with `get_schema`.

**Alternatives considered**:
- Return the whole SchemaResult envelope from `get_schema` — rejected; too verbose for a search listing. Callers can follow up with `get_schema` on a chosen hit.
- Return only the schema *type* (e.g., `"object"`) — rejected; insufficient for LLM downstream use.

---

## Decision 5: Allow-list interaction

**Decision**: Per-API `allow_list.paths["search_api"]` filters which path templates may appear in results for that API. Per-API `allow_list.tools` must include `"search_api"` (via the union across APIs, as `computeAllowedTools` already does) for the tool to be registered.

**Rationale**: Consistent with how `explore_api` filters against `allow_list.paths["explore_api"]`. Requires adding `"search_api"` to `knownToolNames` so config validation accepts the name.

**Alternatives considered**:
- Hard-code search to honour the union of all per-tool path allow-lists — rejected; it would surprise users who expect per-tool isolation (e.g., "I allowed search_api but not http_get; search still reveals GET paths").

---

## Decision 6: Error handling

**Decision**: Two error cases, both using the existing `toolErrorResult` helper:
- Empty/whitespace `query` → `INVALID_INPUT` with message "query must not be empty"
- `api` filter names an unregistered API → `API_NOT_FOUND` with the offending name

Other edge cases (zero matches) return an empty JSON array with no error.

**Rationale**: Matches the error-envelope pattern used across existing tools (see `toolErrorResult` usages in `http.go`, `explore.go`, `schema.go`).

**Alternatives considered**:
- Empty query → return all endpoints — rejected; the spec explicitly requires an error (duplicates `explore_api`, and large result sets waste tokens).

---

## Unknowns resolved

No NEEDS CLARIFICATION markers remain. All decisions above are derived from the existing tool implementations (`pkg/tools/explore.go`, `pkg/tools/schema.go`) and the spec.
