# Quickstart: OpenAPI MCP Server

**Branch**: `001-openapi-mcp-server` | **Date**: 2026-03-28

## Prerequisites

- Go 1.22 or later
- An OpenAPI 3.x definition file (YAML or JSON)
- Access to the target API host

## 1. Build

```bash
go build -o prograpimcp ./cmd/prograpimcp
```

## 2. Create a Configuration File

Create `config.yaml` in your working directory:

```yaml
server:
  host: "127.0.0.1"
  port: 8080
  transport: "http"   # "http" (Streamable HTTP) or "stdio"

apis:
  - name: petstore
    definition: "./petstore.yaml"
    host: "https://api.example.com"
    base_path: "/v2"
```

**Minimal single-API config (no host override)**:

```yaml
server:
  transport: "stdio"

apis:
  - name: myapi
    definition: "./myapi.yaml"
```

When `host` is omitted, the base URL is taken from `servers[0].url` in the OpenAPI definition.

## 3. Run the Server

**HTTP transport (default)**:
```bash
./prograpimcp --config config.yaml
```

**Override host/port via flags**:
```bash
./prograpimcp --config config.yaml --host 0.0.0.0 --port 9090
```

**Override via environment variables**:
```bash
PROGRAPIMCP_SERVER_HOST=0.0.0.0 PROGRAPIMCP_SERVER_PORT=9090 ./prograpimcp
```

**Stdio transport (Claude Desktop / local agent)**:
```bash
./prograpimcp --config config.yaml --transport stdio
```

## 4. Configuration Precedence

Settings are resolved in this order (highest wins):

1. CLI flags (`--host`, `--port`, `--transport`)
2. Environment variables (`PROGRAPIMCP_SERVER_HOST`, `PROGRAPIMCP_SERVER_PORT`,
   `PROGRAPIMCP_SERVER_TRANSPORT`)
3. Config file values
4. Defaults (`127.0.0.1:8080`, transport `http`)

## 5. Verify Startup

On successful startup the server prints:

```
[prograpimcp] Loaded API: petstore (42 paths) from ./petstore.yaml
[prograpimcp] MCP server listening on http://127.0.0.1:8080
```

If a definition file fails validation, the server prints the error and exits with code 1:

```
[prograpimcp] ERROR: failed to load API 'petstore': invalid OpenAPI 3.1 document:
  - paths./pets.get.responses is required
```

## 6. Connect an Agent

Point your MCP-compatible agent at `http://127.0.0.1:8080` (or your configured host/port).
The server exposes these tools:

| Tool          | Purpose |
|---------------|---------|
| `explore_api` | List available paths (optionally filtered by prefix) |
| `get_schema`  | Get the full input/output schema for one endpoint |
| `http_get`    | Validated GET request |
| `http_post`   | Validated POST request |
| `http_put`    | Validated PUT request |
| `http_patch`  | Validated PATCH request |

## 7. Typical Agent Discovery Flow

```
1. explore_api(api="petstore")
   → Returns all paths with methods and descriptions

2. get_schema(api="petstore", path="/pets", method="POST")
   → Returns required fields, types, constraints

3. http_post(api="petstore", path="/pets", body={"name": "Fido", "species": "dog"})
   → Validates body → executes POST → returns response
```

If the agent makes an invalid call:

```
http_post(api="petstore", path="/pets", body={"species": "dog"})
→ ToolError: VALIDATION_FAILED
  details: [{type: "MISSING_REQUIRED_FIELD", field: "name", message: "..."}]
```

The agent can correct and retry immediately without human intervention.

## 8. Claude Desktop Integration (stdio mode)

Add to `~/.config/claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "myapi": {
      "command": "/path/to/prograpimcp",
      "args": ["--config", "/path/to/config.yaml", "--transport", "stdio"]
    }
  }
}
```

## 9. Use as a Go Library

Other Go programs can embed the MCP server without the CLI. Add the dependency:

```bash
go get github.com/your-org/prograpimcp
```

**Minimal embedding example**:

```go
import (
    "context"
    "github.com/your-org/prograpimcp/pkg/config"
    "github.com/your-org/prograpimcp/pkg/openapimcp"
)

cfg := config.Config{
    Server: config.ServerConfig{Host: "0.0.0.0", Port: 9090, Transport: "http"},
    APIs: []config.APIConfig{
        {Name: "myapi", Definition: "./myapi.yaml", Host: "https://api.example.com"},
    },
}
srv, err := openapimcp.New(cfg)
if err != nil { /* handle */ }
if err := srv.Start(context.Background()); err != nil { /* handle */ }
```

**Use only the validator** (no MCP server):

```go
import (
    "github.com/your-org/prograpimcp/pkg/registry"
    "github.com/your-org/prograpimcp/pkg/validator"
)

reg := registry.New()
reg.Load(config.APIConfig{Name: "myapi", Definition: "./myapi.yaml"})
entry, _ := reg.Lookup("myapi")
result := entry.Validator.Validate(myHttpRequest)
if !result.Valid {
    for _, e := range result.Errors {
        fmt.Printf("[%s] %s: %s\n", e.Type, e.Field, e.Message)
    }
}
```

See [contracts/library-api.md](contracts/library-api.md) for the full stable API reference.

## Validation

Run the full test suite to verify a working installation:

```bash
go test ./...
```

Integration tests start an `httptest.Server` as the target API; no live network access
is required to run the tests.
