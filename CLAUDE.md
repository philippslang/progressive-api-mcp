# prograpimcp Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-28

## Active Technologies

- Go 1.22+ (001-openapi-mcp-server)

## Project Structure

```text
cmd/prograpimcp/main.go       # Binary entry point
internal/config/              # Config struct, YAML loading, viper/cobra binding
internal/loader/              # OpenAPI file parsing and structure validation
internal/registry/            # In-memory API registry
internal/validator/           # Request validation wrapper (libopenapi-validator)
internal/tools/               # MCP tool handlers (http, explore, schema)
internal/httpclient/          # Outbound HTTP call executor
internal/server/              # MCP server init and transport selection
tests/integration/            # End-to-end tests using httptest.NewServer
tests/unit/                   # Per-package unit tests
tests/contract/               # MCP tool contract shape tests
tests/testdata/               # OpenAPI fixture files (petstore.yaml, etc.)
```

## Commands

```bash
go build -o prograpimcp ./cmd/prograpimcp   # Build binary
go test ./...                               # Run all tests
go test -bench=. ./...                      # Run benchmarks
golangci-lint run                           # Lint
```

## Code Style

- Standard Go idioms; `gofmt` enforced
- Table-driven tests (`[]struct{ name, input, want }`)
- Errors wrapped with `fmt.Errorf("context: %w", err)`
- No global state; dependencies passed via constructors

## Recent Changes

- 001-openapi-mcp-server: Added Go 1.22+

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
