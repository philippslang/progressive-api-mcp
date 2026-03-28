# Research: MCP Test Client CLI

**Branch**: `006-mcp-test-client` | **Date**: 2026-03-28

## Decision 1: MCP client API to use

**Decision**: Use `client.NewStreamableHttpClient(url)` (convenience factory) to create a connected HTTP MCP client, then call `client.Initialize()` and `client.ListTools()`.

**Key API** (from mcp-go v0.46.0):
```
client.NewStreamableHttpClient(baseURL string, opts ...transport.StreamableHTTPCOption) (*Client, error)
client.Start(ctx) error
client.Initialize(ctx, mcp.InitializeRequest{...}) (*mcp.InitializeResult, error)
client.ListTools(ctx, mcp.ListToolsRequest{}) (*mcp.ListToolsResult, error)
client.Close() error
```

**Import paths**:
- `github.com/mark3labs/mcp-go/client`
- `github.com/mark3labs/mcp-go/mcp`

**Rationale**: The convenience factory handles transport construction in one call. All three imports are already resolved in go.mod — no `go get` needed.

**Alternatives considered**:
- Manual `transport.NewStreamableHTTP` + `client.NewClient` — more flexible but no benefit for this use case.

---

## Decision 2: Timeout handling

**Decision**: Use `context.WithTimeout(context.Background(), 10*time.Second)` for the entire operation (connect + init + list).

**Rationale**: Matches the spec's 10-second default. A single context covers all operations cleanly; no per-call timeout needed.

**Alternatives considered**:
- Per-call timeouts — more granular but overkill for a diagnostic tool.
- `--timeout` flag — out of scope for v1.

---

## Decision 3: CLI argument parsing

**Decision**: Use `os.Args` directly — check `len(os.Args) != 2`, print usage to stderr and exit 1 if wrong. No Cobra, no `flag` package.

**Rationale**: The CLI has exactly one positional argument. Cobra is available but adds ~3ms startup time and 5MB to binary size for zero user-facing benefit. The spec says "only depends on mcp."

**Alternatives considered**:
- Cobra — available in go.mod but unnecessary for one argument.
- `flag` package — fine but still unnecessary; `os.Args[1]` is clear and direct.

---

## Decision 4: Output format

**Decision**: Print tool names to stdout, one per line (`fmt.Println(tool.Name)`). All errors go to stderr (`fmt.Fprintf(os.Stderr, ...)`). Exit 0 on success, exit 1 on any error.

**Rationale**: Matches FR-003/FR-004/FR-005 exactly. One name per line is the unix convention for machine-parseable list output.

**Alternatives considered**:
- Printing name + description — rejected; spec says "lists tool names," and descriptions would break `wc -l` counting.
- JSON output flag — out of scope for v1.

---

## All NEEDS CLARIFICATION markers resolved

None were raised in the spec. No open questions remain.
