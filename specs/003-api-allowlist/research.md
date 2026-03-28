# Research: API Tool & Path Allow List

**Feature**: 003-api-allowlist
**Date**: 2026-03-28

## Decision 1: Where to enforce tool-level restrictions

**Decision**: Compute a union of allowed tools across all loaded APIs at `Start()` time; pass the union as `map[string]bool` to each `Register*` function; skip `s.AddTool` for tools absent from the map.

**Rationale**: Tool registration happens once at startup for the whole MCP session. The MCP protocol has no per-call tool-registration concept — a tool either exists in the session or it doesn't. The union approach means: if any API allows `http_post`, it is registered (needed for multi-API servers). If no API allows it, it is never registered and is genuinely absent from the session.

**Alternatives considered**:
- Wrap tool handlers to return `TOOL_NOT_PERMITTED` rather than skipping registration → rejected because SC-002 requires the tool to be absent, not just error-returning.
- Pass full `APIAllowList` into register functions per-API → impossible since tools are registered once for all APIs.
- Register separate tool instances per API (e.g. `petstore_http_get`) → rejected: collides with the existing `tool_prefix` feature; adds complexity; not what the spec describes.

---

## Decision 2: Where to enforce path-level restrictions

**Decision**: At request time inside each tool handler, immediately after `resolveAPI` resolves the target API entry. The `APIEntry` carries a copy of `AllowList` (populated by `registry.Load`). Each tool handler reads `entry.AllowList.Paths[toolBase]` and rejects non-listed paths with `PATH_NOT_PERMITTED`.

**Rationale**: Path restrictions are per-API. Only after `resolveAPI` do we know which API — and therefore which path list — applies. Registering separate per-API handlers is unnecessary complexity. The path check is a simple slice scan (O(n), n ≤ ~20 paths in practice) and adds negligible latency.

**Alternatives considered**:
- Middleware layer wrapping tool handlers → rejected: adds an indirection layer not justified by a single use case (Rule of Three).
- Filter paths at OpenAPI document level (strip disallowed paths from the parsed model) → rejected: too invasive, hard to reverse, breaks `explore_api`'s ability to show "you asked for this but it's restricted".

---

## Decision 3: `explore_api` path filtering semantics

**Decision**: When `AllowList.Paths["explore_api"]` is non-empty, filter the returned path list to only paths that (a) exist in the OpenAPI document AND (b) are in the allow list. The existing `prefix` filter argument still applies on top.

**Rationale**: `explore_api` is the agent's discovery tool. Showing paths the agent cannot call (because they're restricted) would generate confusion and wasted calls. Filtering to allowed paths gives agents an accurate, actionable view of the surface area they can work with.

**Alternatives considered**:
- Show all paths, annotate restricted ones → rejected: adds complexity to the response format; breaks the simple `PathInfo` array contract.
- Treat `explore_api` path allow list as same as HTTP tools → same decision, just made explicit.

---

## Decision 4: New error code `PATH_NOT_PERMITTED`

**Decision**: Use a distinct `PATH_NOT_PERMITTED` error code (not reuse `PATH_NOT_FOUND`) when a path is known but restricted.

**Rationale**: Agents benefit from knowing the distinction: `PATH_NOT_FOUND` means "this path doesn't exist in the API — try `explore_api`"; `PATH_NOT_PERMITTED` means "this path exists but is blocked by configuration — don't retry". Different agent strategies apply to each case.

**Alternatives considered**:
- Reuse `PATH_NOT_FOUND` → rejected: loses diagnostic information; agent would wastefully explore paths that can never succeed.
- Reuse `VALIDATION_FAILED` → rejected: semantically incorrect.

---

## Decision 5: `allow_list` as value type (not pointer) on `APIConfig`

**Decision**: `APIConfig.AllowList` is `APIAllowList` (value), not `*APIAllowList` (pointer). Zero value of `APIAllowList` (empty slices/map) means allow-all.

**Rationale**: Pointer would require nil-checks everywhere. Value type with zero-value semantics means existing code that doesn't set `AllowList` automatically gets allow-all behaviour with no code changes. Cleaner and more idiomatic Go.

**Alternatives considered**:
- Pointer field → rejected: nil-checks proliferate; zero value is more ergonomic.

---

## Performance

Path check is a linear scan of `AllowList.Paths[toolBase]` (a `[]string`). In practice this list will have fewer than 20 entries. At 1 ns/op per comparison, 20 entries = 20 ns — well within the ≤ 1 µs budget. No caching or trie-based lookup needed.

Benchmark to be added: `BenchmarkPathAllowCheck` in `tests/unit/` verifying the check completes in < 1 µs for lists of 10, 50, and 100 entries.
