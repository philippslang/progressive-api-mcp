# Research: Health Endpoint

**Branch**: `004-health-endpoint` | **Date**: 2026-03-28

## Decision 1: Same port vs. separate port for `/health`

**Decision**: Serve `/health` on the same host/port as the MCP endpoint (`/mcp`).

**Rationale**: The spec requires zero configuration. A separate health port would need a new config field and docs. Same-port health checks match every standard monitoring convention (Kubernetes liveness probes default to the service port, ELB health checks target the same port).

**Alternatives considered**:
- Separate configurable port — rejected: adds config surface and breaks "zero configuration" requirement.

---

## Decision 2: Integration approach with mcp-go StreamableHTTPServer

**Decision**: Use `StreamableHTTPServer` as an `http.Handler` (it implements `ServeHTTP`) rather than calling `httpSrv.Start()`.

**Rationale**: `StreamableHTTPServer.Start()` creates its own internal `http.ServeMux` that is not accessible after construction. The library explicitly supports the handler pattern:

> "or the server itself can be used as a http.Handler, which is convenient to integrate with existing http servers" — mcp-go source, streamable_http.go lines 169-174

By owning the `http.ServeMux` and `http.Server` ourselves, we can add `/health` without any special mcp-go API. The shutdown path changes from `httpSrv.Shutdown(ctx)` to `httpServer.Shutdown(ctx)` (standard library).

**Integration sketch** (implementation detail — informational only):
```
custom mux:
  /mcp    → StreamableHTTPServer.ServeHTTP (MCP protocol)
  /health → inline handler returning 200 + status body

http.Server{Addr: addr, Handler: mux}.ListenAndServe()
```

**Alternatives considered**:
- `WithStreamableHTTPServer` option — mcp-go does expose this, but passing a pre-built server with a mux pre-populated requires understanding of internal registration order; the ServeHTTP approach is simpler and documented.
- A second goroutine running a separate `http.ListenAndServe` — rejected: two ports, more config, no benefit.

---

## Decision 3: Response body format

**Decision**: Return `{"status":"ok"}` as a JSON body with `Content-Type: application/json`.

**Rationale**: JSON is already the lingua franca of this project's responses. Most health-check consumers (Kubernetes, ELB, Datadog) accept any 200 body, but a minimal JSON payload is friendlier for human operators who `curl` the endpoint.

**Alternatives considered**:
- Plain text `ok` — simpler, but inconsistent with the rest of the server's responses.
- Richer body with version, uptime, dependency status — rejected as out of scope per the spec (liveness only, no dependency checks).

---

## Decision 4: Where to locate the new code

**Decision**: Modify `pkg/openapimcp/server.go` only. No new files or packages.

**Rationale**: The health handler is a single function literal with no logic. Creating a new file or package for it would violate the CLAUDE.md principle against premature abstraction. The HTTP transport setup already lives in `server.go` and that is the correct place to own the mux.

**Alternatives considered**:
- New `pkg/health/` package — rejected: one handler with no state does not warrant a package.
- New file `pkg/openapimcp/health.go` — rejected: adds a file for a two-line handler.

---

## All NEEDS CLARIFICATION markers resolved

None were raised in the spec. No open questions remain.
