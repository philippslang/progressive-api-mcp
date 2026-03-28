# Data Model: API Tool & Path Allow List

**Feature**: 003-api-allowlist
**Date**: 2026-03-28

## New Type: `APIAllowList`

**Location**: `pkg/config/config.go`

```go
// APIAllowList restricts which MCP tools and paths are available for one API entry.
// Zero value means no restrictions: all tools registered, all paths accessible.
type APIAllowList struct {
    // Tools is the list of base tool names to register for this API.
    // Nil or empty means all 6 tools are allowed.
    // Valid values: "explore_api", "get_schema", "http_get", "http_post", "http_put", "http_patch"
    Tools []string `yaml:"tools"`

    // Paths maps each base tool name to the OpenAPI path templates it may access.
    // Nil or empty map means all paths are allowed for all tools.
    // A tool name present in Paths but absent from Tools is valid (restriction is dormant).
    // An empty slice value for a tool means: tool is allowed but NO paths are accessible.
    Paths map[string][]string `yaml:"paths"`
}
```

**Validation rules** (enforced in `Config.Validate()`):
- Each name in `Tools` must be one of: `explore_api`, `get_schema`, `http_get`, `http_post`, `http_put`, `http_patch`
- Each key in `Paths` must be one of the same 6 names
- Unknown names return a descriptive error identifying the bad value and listing valid options

## Modified Type: `APIConfig`

**Location**: `pkg/config/config.go`

```go
type APIConfig struct {
    Name       string       `yaml:"name"`
    Definition string       `yaml:"definition"`
    Host       string       `yaml:"host"`
    BasePath   string       `yaml:"base_path"`
    AllowList  APIAllowList `yaml:"allow_list"` // NEW — zero value = allow all
}
```

## Modified Type: `APIEntry`

**Location**: `pkg/registry/registry.go`

```go
type APIEntry struct {
    Name      string
    Config    config.APIConfig
    BaseURL   string
    Validator *validator.Validator
    doc       libopenapi.Document
    AllowList config.APIAllowList // NEW — copied from Config.AllowList in Load()
}
```

The `AllowList` is a value copy (not a pointer), so tool handlers can read it without nil-checks.

## New Error Code: `PATH_NOT_PERMITTED`

Used in `ToolError.Code` when a path is known to the OpenAPI spec but blocked by the allow list.

```json
{
  "code": "PATH_NOT_PERMITTED",
  "message": "path \"/owners\" is not in the allow list for http_get on API \"petstore\""
}
```

## New Error Code: `TOOL_NOT_PERMITTED`

Used when a multi-API server has the tool registered (because another API allows it), but the resolved API's allow list excludes it.

```json
{
  "code": "TOOL_NOT_PERMITTED",
  "message": "tool \"http_post\" is not permitted for API \"petstore\""
}
```

## Config YAML Example

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    allow_list:
      tools:
        - explore_api
        - get_schema
        - http_get
        - http_post
      paths:
        http_get:
          - /pets
          - /pets/{id}
        http_post:
          - /pets

  - name: bookstore
    definition: ./bookstore.yaml
    host: https://books.example.com
    # No allow_list — all tools and paths accessible
```

## Allowed Tool Set — Union Computation

`server.go` computes `allowedForSession map[string]bool` before calling `Register*`:

```
allowedForSession = union of AllowList.Tools across all APIConfigs
                    where empty/nil Tools means all 6 are included
```

This map is passed to `RegisterHTTPTools`, `RegisterExploreTools`, `RegisterSchemaTools`. Each function checks the map before calling `s.AddTool`.

## State Transitions / Enforcement Flow

```
Request arrives at tool handler
  ↓
resolveAPI(reg, apiName) → APIEntry (with .AllowList)
  ↓
[if multi-API and this API's Tools list is non-empty]
  isToolAllowed(toolBase, entry.AllowList.Tools)?
    NO  → return TOOL_NOT_PERMITTED
    YES → continue
  ↓
[if entry.AllowList.Paths[toolBase] is non-empty]
  resolveTemplate(path) → template
  isPathAllowed(template, entry.AllowList.Paths[toolBase])?
    NO  → return PATH_NOT_PERMITTED
    YES → continue
  ↓
proceed with normal validation and execution
```
