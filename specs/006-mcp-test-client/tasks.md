# Tasks: MCP Test Client CLI

**Input**: Design documents from `/specs/006-mcp-test-client/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, contracts/ ✓

**Organization**: Tasks grouped by user story. Single story (P1) covers the entire feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to ([US1])

---

## Phase 1: Setup (Shared Infrastructure)

*(No setup tasks — project exists, no new dependencies, no new packages)*

---

## Phase 2: Foundational (Blocking Prerequisites)

*(No foundational tasks — new binary is fully self-contained)*

---

## Phase 3: User Story 1 - List All MCP Tools (Priority: P1) 🎯 MVP

**Goal**: `mcpls <url>` connects to the MCP server, initializes the session, and prints one tool name per line to stdout. Errors go to stderr with exit code 1.

**Independent Test**: With `prograpimcp` running against petstore.yaml, run `go run ./cmd/mcpls http://127.0.0.1:8080/mcp` and verify the output contains the expected tool names (e.g. `http_get`, `explore_api`) one per line.

### Implementation for User Story 1

- [x] T001 [US1] Create `cmd/mcpls/main.go`: parse `os.Args[1]` as the MCP endpoint URL (print usage + exit 1 if missing); create a Streamable HTTP client using `client.NewStreamableHttpClient`; call `Start`, `Initialize`, `ListTools`; print each tool name to stdout one per line; print errors to stderr and exit 1 on any failure; use a 10-second context timeout for the full operation

**Checkpoint**: After T001, `go run ./cmd/mcpls http://127.0.0.1:8080/mcp` lists all tools and `go build ./cmd/mcpls` produces a working binary.

---

## Phase 4: Polish & Cross-Cutting Concerns

- [x] T002 Run `go build -o /dev/null ./cmd/mcpls` and `go test ./...` to verify the binary compiles and all existing tests remain green

---

## Dependencies & Execution Order

- T001 has no dependencies — start immediately
- T002 must run after T001

---

## Implementation Strategy

1. T001 — Write `cmd/mcpls/main.go`
2. T002 — Build + full test suite green

**Total**: 2 tasks. Feature complete after T002.

---

## Notes

- Import paths needed: `github.com/mark3labs/mcp-go/client`, `github.com/mark3labs/mcp-go/mcp`
- No Cobra — `os.Args[1]` is sufficient per research decision 3
- `mcp.InitializeRequest` requires `Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION` and `Params.ClientInfo`
- Tool names are in `result.Tools[i].Name`
- Always call `defer c.Close()` after `Start` succeeds
