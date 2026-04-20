# Data Model: Rename `explore_api` to `list_api`

**Feature**: 015-rename-explore-to-list | **Date**: 2026-04-18

This feature is a pure rename; no new entities, fields, relationships, or state transitions are introduced. For completeness, the existing entities that touch the renamed tool are listed below with their pre- and post-rename identifiers.

## Affected identifiers

| Scope | Before | After | Notes |
|-------|--------|-------|-------|
| MCP tool name (wire) | `explore_api` | `list_api` | Subject to `server.tool_prefix` — e.g. `store_list_api` when prefix is `store`. |
| `allow_list.tools` value (config) | `"explore_api"` | `"list_api"` | Accepted by `Config.Validate()`; old name rejected with `knownToolNamesHint`. |
| `allow_list.paths` key (config) | `"explore_api"` | `"list_api"` | Same whitelist applies. |
| Go file | `pkg/tools/explore.go` | `pkg/tools/list.go` | File rename. |
| Go function | `tools.RegisterExploreTools` | `tools.RegisterListTools` | Same signature. |
| Go result-element type | `tools.PathInfo` | `tools.PathInfo` | Unchanged — the name describes the type, not the tool. |

## Unchanged entities

- `APIEntry`, `APIConfig`, `APIAllowList`, `Registry` — no field changes.
- `PathInfo` struct (`{Path, Methods, Description}`) — structural shape identical.
- Tool parameter schema: `api` (optional string), `prefix` (optional string).
- Tool response JSON: array of `PathInfo`, sorted by `path`.
- Error envelope (`{code, message, details, hints}`) — reused verbatim.

## Validation rules

Post-rename, `Config.Validate()` MUST:

1. Accept `"list_api"` anywhere `"explore_api"` was previously accepted (`allow_list.tools` list entries, `allow_list.paths` keys).
2. Reject `"explore_api"` in those same positions with the error `unknown tool name "explore_api"; valid names are: <knownToolNamesHint>` where `<knownToolNamesHint>` includes `list_api` and not `explore_api`.

No state transitions — the tool is stateless, the config is loaded once at startup.
