# Progressive API MCP Server

An MCP server that exposes OpenAPI-defined REST APIs to AI agents — without turning every endpoint into a separate tool.

## The problem with naive OpenAPI → MCP conversion

The obvious approach is to register one MCP tool per API endpoint. A moderately-sized API with 50 endpoints becomes 50 tools. The agent must load all of them into context upfront, burning tokens before a single call is made, and the tool list tells it almost nothing useful about how the API actually works.

## Progressive disclosure instead

`prograpimcp` exposes **6 tools regardless of how many endpoints the API has**:

| Tool | What it does |
|------|-------------|
| `search_api` | Search matching paths and their HTTP methods. Supports prefix filtering (e.g. `/pets`) to narrow scope. |
| `list_api` | List available paths and their HTTP methods. Supports prefix filtering (e.g. `/pets`) to narrow scope. |
| `get_schema` | Return the full request/response schema for one specific endpoint. |
| `http_get` | Execute a validated GET request. |
| `http_post` | Execute a validated POST request. |
| `http_put` | Execute a validated PUT request. |
| `http_patch` | Execute a validated PATCH request. |

The agent starts with `list_api` to get a fast, cheap overview. It then calls `get_schema` on whichever endpoint it cares about, and only then makes the actual HTTP call. Context grows on demand; nothing is loaded eagerly.

```
list_api          →  "here are the paths, pick one"
get_schema /pets POST →  "here is the shape of that request"
http_post /pets      →  actual call, validated against schema
```

Each step gives immediate feedback. The agent learns the shape of the API incrementally rather than drowning in a wall of tool definitions.

## Configuration

```yaml
# config.yaml
server:
  host: "127.0.0.1"
  port: 8080
  transport: "http"       # or "stdio"
  # tool_prefix: "myapi"  # optional — prefixes all tool names: myapi_http_get, etc.

apis:
  - name: petstore
    definition: "./petstore.yaml"
    host: "https://api.example.com"   # optional: overrides servers[0].url
    base_path: "/v2"                  # optional: overrides servers[0].url path
```

Multiple APIs can be loaded simultaneously. When more than one is loaded, every tool accepts an `api` parameter to disambiguate. Omitting it on a multi-API server returns an `AMBIGUOUS_API` response listing the available names — another step in the disclosure ladder rather than a hard failure.

## Running

```bash
# Build
go build -o prograpimcp ./cmd/prograpimcp

# Start (Streamable HTTP transport)
./prograpimcp --config config.yaml

# Start (stdio transport, for Claude Desktop / MCP clients that use stdio)
./prograpimcp --config config.yaml --transport stdio

# Override config values via flags or environment variables
./prograpimcp --tool-prefix myapi --port 9090
PROGRAPIMCP_SERVER_TOOL_PREFIX=myapi ./prograpimcp
```

## Using as a Go library

```go
import "github.com/philippslang/progressive-api-mcp/pkg/openapimcp"

srv, err := openapimcp.New(config.Config{
    Server: config.ServerConfig{Transport: "stdio"},
    APIs: []config.APIConfig{
        {Name: "petstore", Definition: "./petstore.yaml"},
    },
})
if err != nil {
    log.Fatal(err)
}
if err := srv.Start(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Tool prefix

When running multiple `prograpimcp` instances (one per API), the `tool_prefix` option keeps tool names unique across MCP sessions:

```yaml
# petstore server
server:
  tool_prefix: "store"   # → store_http_get, store_list_api, …

# payments server
server:
  tool_prefix: "pay"     # → pay_http_get, pay_list_api, …
```

The prefix must start with a letter or underscore and contain only letters, digits, and underscores. A trailing underscore is stripped automatically.

