# Contract: mcpls CLI

**Version**: 1.0 | **Date**: 2026-03-28

## Synopsis

```
mcpls <mcp-endpoint-url>
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<mcp-endpoint-url>` | Yes | Full HTTP URL of the MCP server endpoint (e.g. `http://127.0.0.1:8080/mcp`) |

## Output

### stdout — success

One tool name per line, in the order returned by the server. No headers, no separators, no trailing whitespace.

```
http_get
http_post
http_put
http_patch
explore_api
get_schema
```

If the server has no tools registered, stdout is empty.

### stderr — errors

A single human-readable error line prefixed with `mcpls: `. Examples:

```
mcpls: cannot connect to http://bad-host/mcp: dial tcp: lookup bad-host: no such host
mcpls: request timed out after 10s
mcpls: MCP initialization failed: ...
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success — tool list printed (or empty) |
| 1 | Any error — connection failure, timeout, protocol error, wrong usage |

## Usage Errors

If the wrong number of arguments is supplied, the CLI prints usage to stderr and exits 1:

```
usage: mcpls <mcp-endpoint-url>
```

## Timeout

Default: 10 seconds for the entire operation (connect + initialize + list tools). Not configurable in v1.

## Stability

The output format (one name per line, stdout/stderr split, exit codes) is a stable contract. Tool names themselves are determined by the server.
