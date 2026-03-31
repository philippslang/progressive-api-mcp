# prograpimcp Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-31

## Active Technologies
- Go 1.22+ (inherits from `001-openapi-mcp-server`) + No new dependencies (002-mcp-tool-prefix)
- Go 1.22+ + mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator, spf13/cobra + viper — no new dependencies (003-api-allowlist)
- N/A (in-process, stateless config load) (003-api-allowlist)
- Go 1.22+ + mark3labs/mcp-go v0.46.0 (StreamableHTTPServer), standard library `net/http` (004-health-endpoint)
- Go 1.22+ + Standard library `net/http`, `fmt`, `os`, `time` — no new dependencies (005-http-request-logging)
- Go 1.22+ + `github.com/mark3labs/mcp-go/client`, `github.com/mark3labs/mcp-go/client/transport`, `github.com/mark3labs/mcp-go/mcp` — all already in go.mod; no new dependencies (006-mcp-test-client)
- Go 1.22+ + `cobra` (CLI), `net/http` stdlib (HTTP server), `encoding/json` stdlib (JSON responses), `sync` stdlib (RWMutex for in-memory store) (008-mock-api-servers)
- In-memory only (`sync.RWMutex`-protected maps); no disk persistence (008-mock-api-servers)
- In-memory only (no persistence) (009-ignore-api-headers)
- Go 1.22+ + mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator — no new dependencies (010-body-only-response)
- N/A (in-memory mock server, no persistence) (012-fix-patch-schema)
- N/A (in-memory mock server) (012-fix-patch-schema)

- Go 1.22+ (001-openapi-mcp-server)

## Project Structure

```text
pkg/openapimcp/       # PRIMARY LIBRARY ENTRY POINT — New(), Start(), Stop()
pkg/config/           # Config struct, LoadFile(); no cobra/viper dependency
pkg/loader/           # OpenAPI file parsing and structure validation
pkg/registry/         # In-memory API registry — exported, usable standalone
pkg/validator/        # Request validation wrapper — exported, usable standalone
pkg/tools/            # MCP tool handlers (http, explore, schema)
pkg/httpclient/       # Outbound HTTP call executor
cmd/prograpimcp/      # CLI binary — thin cobra+viper wrapper around pkg/openapimcp
tests/integration/    # End-to-end tests (httptest.NewServer + library embedding test)
tests/unit/           # Per-package unit tests
tests/contract/       # MCP tool and library API contract shape tests
tests/testdata/       # OpenAPI fixture files (petstore.yaml, bookstore.yaml)
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
- `pkg/` packages MUST NOT import `cobra` or `viper` — those belong in `cmd/` only
- All exported types and functions in stable packages MUST have doc comments

## Recent Changes
- 013-api-skip-validation: Added Go 1.22+ + mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator, spf13/cobra + viper — no new dependencies
- 012-fix-patch-schema: Added Go 1.22+ + mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator — no new dependencies
- 012-fix-patch-schema: Added Go 1.22+ + mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator — no new dependencies


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
