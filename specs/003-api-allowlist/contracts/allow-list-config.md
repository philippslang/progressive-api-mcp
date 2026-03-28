# Contract: APIAllowList Configuration Shape

**Feature**: 003-api-allowlist
**Date**: 2026-03-28

## Config File Contract

The `allow_list` key is optional inside each `apis` entry. When absent, behaviour is identical to all tools allowed and all paths allowed.

### Schema

```yaml
apis:
  - name: string               # required
    definition: string         # required
    host: string               # optional
    base_path: string          # optional
    allow_list:                # optional; omit for allow-all
      tools:                   # optional list of base tool names
        - string               # one of: explore_api get_schema http_get http_post http_put http_patch
      paths:                   # optional map: tool name → list of OpenAPI path templates
        <tool_name>:           # must be one of the 6 known tool names
          - string             # OpenAPI path template, e.g. /pets or /pets/{id}
```

### Validation Errors

| Condition | Error Message Pattern |
|-----------|----------------------|
| Unknown tool name in `tools` | `apis[N].allow_list.tools: unknown tool name "X"; valid names are: explore_api, get_schema, http_get, http_post, http_put, http_patch` |
| Unknown key in `paths` | `apis[N].allow_list.paths: unknown tool name "X"; valid names are: ...` |

## MCP Tool Error Response Contract

### `PATH_NOT_PERMITTED`

Returned when the resolved path is valid in the OpenAPI spec but blocked by the allow list.

```json
{
  "code": "PATH_NOT_PERMITTED",
  "message": "path \"/owners\" is not in the allow list for http_get on API \"petstore\""
}
```

- `code`: always the string `"PATH_NOT_PERMITTED"`
- `message`: human-readable, includes path, tool, and API name
- `details`: absent
- `hints`: absent

### `TOOL_NOT_PERMITTED`

Returned when a multi-API tool call targets an API whose allow list excludes that tool.

```json
{
  "code": "TOOL_NOT_PERMITTED",
  "message": "tool \"http_post\" is not permitted for API \"petstore\""
}
```

- `code`: always the string `"TOOL_NOT_PERMITTED"`
- `message`: human-readable, includes tool name and API name

## Backward Compatibility Contract

- Any existing config file that does not include `allow_list` MUST load and behave identically to pre-feature behaviour.
- All 6 tools MUST be registered when no `allow_list` is present.
- All OpenAPI paths MUST be accessible when no `paths` restriction is present.
- Existing tool response shapes (`HTTPResult`, `ToolError` with existing codes, `PathInfo`, `SchemaResult`) are unchanged.
