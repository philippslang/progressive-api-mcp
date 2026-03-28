---
description: "Task list for OpenAPI MCP Server"
---

# Tasks: OpenAPI MCP Server

**Input**: Design documents from `/specs/001-openapi-mcp-server/`
**Prerequisites**: plan.md ✅ | spec.md ✅ | research.md ✅ | data-model.md ✅ | contracts/ ✅

**Tests**: Included — TDD is mandatory per the project constitution (Principle II).
Tests MUST be written and confirmed failing before implementation begins.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US4)
- File paths shown relative to repository root

---

## Phase 1: Setup

**Purpose**: Go module initialization, project skeleton, tooling config

- [x] T001 Initialize Go module `github.com/your-org/prograpimcp` with `go mod init` and create `go.mod`
- [x] T002 Add all direct dependencies to `go.mod`: `mark3labs/mcp-go`, `pb33f/libopenapi`, `pb33f/libopenapi-validator`, `spf13/cobra`, `spf13/viper`, `stretchr/testify`; run `go mod tidy`
- [x] T003 [P] Create directory skeleton: `pkg/openapimcp/`, `pkg/config/`, `pkg/loader/`, `pkg/registry/`, `pkg/validator/`, `pkg/tools/`, `pkg/httpclient/`, `cmd/prograpimcp/`, `tests/integration/`, `tests/unit/`, `tests/contract/`, `tests/testdata/`
- [x] T004 [P] Create `.golangci.yml` with linting rules: `errcheck`, `govet`, `staticcheck`, `gocyclo` (max 10), `godot` for doc comments
- [x] T005 [P] Copy `tests/testdata/petstore.yaml` — OpenAPI 3.1 Petstore fixture (used as primary test definition throughout)
- [x] T006 [P] Copy `tests/testdata/bookstore.yaml` — second minimal OpenAPI 3.1 fixture for multi-API tests
- [x] T007 [P] Create `tests/testdata/malformed.yaml` — invalid OpenAPI 3 file for startup-failure tests
- [x] T008 [P] Create `config.yaml.example` at repository root with documented example configuration

---

## Phase 2: Foundational

**Purpose**: Core infrastructure that ALL user story phases depend on. No story work begins until this phase is complete.

**⚠️ CRITICAL**: Phases 3–6 cannot start until this phase is complete.

### Tests for Foundational Layer ⚠️ Write first — confirm they FAIL before implementing

- [x] T009 [P] Write unit tests for `Config` validation in `tests/unit/config_test.go`: missing `name`, invalid `transport`, port out of range, duplicate API names, missing `definition` path
- [x] T010 [P] Write unit tests for `config.LoadFile` in `tests/unit/config_test.go`: valid YAML, malformed YAML, missing file, override precedence for `host`/`base_path`
- [x] T011 [P] Write unit tests for `registry.Registry` in `tests/unit/registry_test.go`: `Load` success, `Load` with invalid file, `Load` with malformed OpenAPI, `Lookup` found/not-found, `ListNames` order, `Len`
- [x] T012 [P] Write unit tests for `loader` in `tests/unit/loader_test.go`: valid OpenAPI 3.0 file, valid 3.1 file, malformed YAML, structurally invalid OpenAPI, external `$ref` resolution failure

### Implementation: `pkg/config`

- [x] T013 Implement `Config`, `ServerConfig`, `APIConfig` structs with YAML tags in `pkg/config/config.go`
- [x] T014 Implement `Config.Validate()` in `pkg/config/config.go`: check required fields, unique names, valid transport, port range
- [x] T015 Implement `config.LoadFile(path string) (Config, error)` in `pkg/config/config.go`: read YAML, unmarshal, call Validate

### Implementation: `pkg/loader`

- [x] T016 Implement `loader.Load(cfg config.APIConfig) (libopenapi.Document, error)` in `pkg/loader/loader.go`: read file, parse with `libopenapi`, validate OpenAPI 3 structure, return error with file name and reason on failure

### Implementation: `pkg/registry`

- [x] T017 Implement `registry.APIEntry` struct in `pkg/registry/registry.go`: `Name`, `Config`, `BaseURL`, `Validator` fields
- [x] T018 Implement `registry.Registry` with `New()`, `Load()`, `Lookup()`, `ListNames()`, `Len()` in `pkg/registry/registry.go`; `Load()` calls `loader.Load` then builds `APIEntry` with resolved `BaseURL`
- [x] T019 Implement `BaseURL` resolution logic in `pkg/registry/registry.go`: `APIConfig.Host + BasePath` > `APIConfig.Host` > `servers[0].url` from definition > error if none resolvable

**Checkpoint**: `go test ./tests/unit/...` MUST pass fully before proceeding to Phase 3.

---

## Phase 3: User Story 1 — Validated HTTP Calls with Schema Feedback (Priority: P1) 🎯 MVP

**Goal**: Agent invokes GET/POST/PUT/PATCH tools; server validates path + schema before any HTTP call; returns structured errors or actual HTTP response.

**Independent Test**: Load `petstore.yaml`, call `http_post` with missing required field — confirm `ToolError{VALIDATION_FAILED}` with field detail returned WITHOUT any network call. Then call with valid data — confirm HTTP response returned.

### Tests for US1 ⚠️ Write first — confirm they FAIL before implementing

- [x] T020 [P] [US1] Write unit tests for `validator.Validator` in `tests/unit/validator_test.go`: valid GET request, POST missing required field, POST wrong field type, POST additional property, missing required query param, path not found — using Petstore fixture
- [x] T021 [P] [US1] Write contract tests for `http_get`/`http_post`/`http_put`/`http_patch` tool output shapes in `tests/contract/mcp_tools_contract_test.go`: verify `HTTPResult` shape on success, `ToolError` shape on validation failure, `ToolError` shape on path-not-found
- [x] T022 [P] [US1] Write integration tests in `tests/integration/http_tools_test.go`: load Petstore, start `httptest.NewServer` as target, exercise GET/POST/PUT/PATCH validation and execution end-to-end

### Implementation: `pkg/validator`

- [x] T023 Implement `ValidationError` and `Result` structs in `pkg/validator/validator.go`
- [x] T024 Implement `validator.New(doc libopenapi.Document) *Validator` in `pkg/validator/validator.go`
- [x] T025 Implement `Validator.Validate(r *http.Request) Result` in `pkg/validator/validator.go`: build `libopenapi-validator` from document, validate request, map errors to `[]ValidationError` with correct `Type` values

### Implementation: `pkg/httpclient`

- [x] T026 Implement `httpclient.Client` with configurable timeout and `Do(req *http.Request) (*http.Response, error)` in `pkg/httpclient/client.go` using `net/http`

### Implementation: `pkg/tools` — HTTP tools

- [x] T027 Implement shared tool helper in `pkg/tools/http.go`: resolve API from registry (single-API bypass, multi-API require `api` param, return `ToolError{AMBIGUOUS_API}` with hints if omitted)
- [x] T028 [US1] Implement path template matching in `pkg/tools/http.go`: match concrete path (e.g., `/pets/42`) to OpenAPI path template (e.g., `/pets/{id}`) using libopenapi path resolution; return `ToolError{PATH_NOT_FOUND}` with hint list on no match
- [x] T029 [US1] Implement `httpGetHandler` in `pkg/tools/http.go`: resolve API → match path → validate query params/headers via `Validator.Validate` → execute via `httpclient` → return `HTTPResult` or `ToolError`
- [x] T030 [US1] Implement `httpPostHandler` in `pkg/tools/http.go`: same as GET plus validate request body schema → execute POST
- [x] T031 [US1] Implement `httpPutHandler` in `pkg/tools/http.go` (mirrors POST)
- [x] T032 [US1] Implement `httpPatchHandler` in `pkg/tools/http.go` (mirrors POST)
- [x] T033 [US1] Implement `HTTPResult` builder in `pkg/tools/http.go`: extract status code, headers (first value per key), parse JSON body if Content-Type is `application/json` else return raw string

### Implementation: `pkg/openapimcp` — server wiring (partial, for US1)

- [x] T034 [US1] Implement `openapimcp.Server` struct in `pkg/openapimcp/server.go` with `New(cfg config.Config) (*Server, error)` — validates config only, does not load files yet
- [x] T035 [US1] Implement `Server.Start(ctx context.Context) error` in `pkg/openapimcp/server.go`: load all APIs via registry, register MCP tools with `mcp-go`, start configured transport (HTTP or stdio)
- [x] T036 [US1] Implement `Server.Stop() error` and `Server.APIs() []string` in `pkg/openapimcp/server.go`
- [x] T037 [US1] Register `http_get`, `http_post`, `http_put`, `http_patch` tools with `mcp-go` server in `pkg/openapimcp/server.go` using handlers from `pkg/tools/http.go`

### Implementation: `cmd/prograpimcp` — CLI (partial, for US1 smoke test)

- [x] T038 [US1] Implement cobra root command with `--config`, `--host`, `--port`, `--transport` flags in `cmd/prograpimcp/main.go`
- [x] T039 [US1] Implement viper binding in `cmd/prograpimcp/main.go`: bind cobra flags to viper keys, set env prefix `PROGRAPIMCP_`, set defaults, load config file via `config.LoadFile`, construct `config.Config`, call `openapimcp.New().Start(ctx)`

**Checkpoint**: `go test ./tests/...` MUST pass. User Story 1 is fully functional and independently testable.

---

## Phase 4: User Story 2 — Progressive API Path Discovery (Priority: P2)

**Goal**: Agent calls `explore_api` with optional prefix filter; gets back all matching paths with methods and descriptions.

**Independent Test**: Load Petstore (20 paths), call `explore_api` with no filter → all 20 paths returned with methods and summary. Call with prefix `/pet` → only `/pet` prefixed paths returned.

### Tests for US2 ⚠️ Write first — confirm they FAIL before implementing

- [x] T040 [P] [US2] Write unit tests for `explore_api` handler in `tests/unit/explore_test.go`: no filter returns all paths, prefix filter returns subset, prefix matches nothing returns empty array, multi-API with `api` param, multi-API without `api` param returns `AMBIGUOUS_API`
- [x] T041 [P] [US2] Write integration tests for `explore_api` in `tests/integration/explore_schema_test.go`: load Petstore, call explore with/without prefix, verify `PathInfo` structure and sort order

### Implementation: `pkg/tools` — explore tool

- [x] T042 [US2] Implement `PathInfo` struct (`Path`, `Methods`, `Description`) in `pkg/tools/explore.go`
- [x] T043 [US2] Implement `exploreAPIHandler` in `pkg/tools/explore.go`: look up API entry, iterate OpenAPI paths, optionally filter by prefix, collect `PathInfo` per path, sort lexicographically, return array
- [x] T044 [US2] Register `explore_api` tool with `mcp-go` server in `pkg/openapimcp/server.go`

**Checkpoint**: `go test ./tests/...` MUST pass. User Story 2 independently testable alongside US1.

---

## Phase 5: User Story 3 — Schema Retrieval for a Specific Endpoint (Priority: P2)

**Goal**: Agent calls `get_schema` with a path and method; receives full parameter and body schemas for that endpoint.

**Independent Test**: Load Petstore, call `get_schema` for `POST /pets` → response includes required fields, types, request body schema, response schemas. Call for non-existent path → `ToolError{PATH_NOT_FOUND}`.

### Tests for US3 ⚠️ Write first — confirm they FAIL before implementing

- [x] T045 [P] [US3] Write unit tests for `get_schema` handler in `tests/unit/schema_test.go`: known path+method returns `SchemaResult`, concrete path resolves to template, unknown path returns `ToolError{PATH_NOT_FOUND}`, unknown method on known path returns `ToolError{PATH_NOT_FOUND}`
- [x] T046 [P] [US3] Write integration tests for `get_schema` in `tests/integration/explore_schema_test.go`: load Petstore, retrieve schema for `GET /pets/{id}` and `POST /pets`, validate shape matches `SchemaResult` contract

### Implementation: `pkg/tools` — schema tool

- [x] T047 [US3] Implement `SchemaResult` struct in `pkg/tools/schema.go`: `Path`, `Method`, `PathParameters`, `QueryParameters`, `RequestBody`, `Responses` (all as `map[string]any` JSON schema representations)
- [x] T048 [US3] Implement `getSchemaHandler` in `pkg/tools/schema.go`: resolve API, resolve path template (concrete or template input), extract and serialize parameter definitions and body/response schemas from libopenapi document model
- [x] T049 [US3] Register `get_schema` tool with `mcp-go` server in `pkg/openapimcp/server.go`

**Checkpoint**: `go test ./tests/...` MUST pass. User Stories 1, 2, and 3 all independently testable.

---

## Phase 6: User Story 4 — Operator Configures Multiple APIs at Startup (Priority: P3)

**Goal**: Server loads two distinct APIs from config; agent can target each by `name`; config host/base_path override applies correctly; malformed definition causes startup failure.

**Independent Test**: Start server with `petstore` + `bookstore` configs. Verify `explore_api(api="petstore")` returns only Petstore paths. Verify `explore_api(api="bookstore")` returns only Bookstore paths. Verify `explore_api()` without `api` returns `AMBIGUOUS_API` with both names in hints.

### Tests for US4 ⚠️ Write first — confirm they FAIL before implementing

- [x] T050 [P] [US4] Write integration tests for multi-API in `tests/integration/http_tools_test.go`: load Petstore + Bookstore, verify API isolation, ambiguity error, host override applied to outbound URL
- [x] T051 [P] [US4] Write integration test for malformed definition in `tests/integration/http_tools_test.go`: verify `openapimcp.New(cfg).Start(ctx)` returns descriptive error and does NOT start
- [x] T052 [P] [US4] Write unit tests for `BaseURL` resolution in `tests/unit/registry_test.go`: host+basePath, host only, basePath only (appended to servers URL), neither (uses servers[0].url), empty servers array returns error

### Implementation (all changes in existing packages — no new files needed)

- [x] T053 [US4] Verify multi-API ambiguity logic in `pkg/tools/http.go` helper covers all 6 tools (should already work from T027; add targeted tests to confirm)
- [x] T054 [US4] Implement startup-abort error formatting in `pkg/openapimcp/server.go`: when any `registry.Load` fails, collect all failures, return single error listing each failed API name and its reason
- [ ] T055 [US4] Write integration test for CLI env var and flag override: `PROGRAPIMCP_SERVER_PORT=9999` overrides config file port; `--host` flag overrides env var — in `tests/integration/http_tools_test.go`

**Checkpoint**: `go test ./tests/...` MUST pass. All four user stories independently testable.

---

## Phase 7: Library Embedding & Polish

**Purpose**: Validate the library is importable and functional without the CLI; cross-cutting quality improvements.

- [ ] T056 [P] Write library embedding test in `tests/integration/embedding_test.go`: construct `config.Config` programmatically (no YAML file, no CLI), call `openapimcp.New(cfg).Start(ctx)` in a goroutine, call `explore_api` via the MCP client, verify response — proves library is independently usable
- [ ] T057 [P] Write benchmark `BenchmarkValidatedHTTPCall` in `tests/integration/http_tools_test.go`: measures end-to-end path+schema validation + HTTP round-trip against `httptest.NewServer`; baseline must be ≤ 2s (SC-005)
- [ ] T058 [P] Write benchmark `BenchmarkServerStartup` in `tests/integration/http_tools_test.go`: measures time to load 10 API definitions; baseline must be ≤ 3s (SC-006)
- [ ] T059 Add `golangci-lint run` to CI check; confirm zero lint warnings across all packages
- [ ] T060 [P] Verify all exported symbols in `pkg/config`, `pkg/openapimcp`, `pkg/validator`, `pkg/registry` have godoc comments; add missing ones
- [ ] T061 [P] Add `pkg/httpclient` timeout configuration: ensure outbound HTTP calls have a sensible default timeout (30s); verify it is configurable via `ServerConfig` if needed
- [ ] T062 Run `quickstart.md` validation: build binary, create config pointing to `tests/testdata/petstore.yaml`, start server in stdio mode, run a manual `explore_api` call, confirm response matches expected format
- [ ] T063 [P] Update `config.yaml.example` with all fields documented and annotated

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — BLOCKS all user story phases
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on Phase 2 (can run in parallel with Phase 3 once Phase 2 done)
- **Phase 5 (US3)**: Depends on Phase 2 (can run in parallel with Phases 3–4)
- **Phase 6 (US4)**: Depends on Phases 3–5 (multi-API test exercises all tools)
- **Phase 7 (Polish)**: Depends on Phase 6

### Within Each Phase

- Tests (marked ⚠️) MUST be written and confirmed FAILING before implementation begins
- Entity/struct tasks before handler tasks
- Handler tasks before server registration tasks
- `pkg/` package tasks before `cmd/` tasks

### User Story Dependencies

- **US1 (P1)**: No dependency on other stories — implement first
- **US2 (P2)**: No dependency on US1 implementation (both use registry + loader from Phase 2)
- **US3 (P2)**: No dependency on US1/US2 implementation
- **US4 (P3)**: Depends on US1–US3 being complete (its tests exercise all tools)

---

## Parallel Opportunities

### Phase 1 — all setup tasks [P] can run in parallel after T001+T002

```
T003, T004, T005, T006, T007, T008 — all in parallel
```

### Phase 2 — tests in parallel, then implementation sequentially

```
T009, T010, T011, T012 — write all unit tests in parallel
T013 → T014 → T015   (config: structs → validate → load)
T016                  (loader, depends on config structs)
T017 → T018 → T019   (registry: entry → registry → BaseURL logic)
```

### Phase 3 — tests in parallel before implementation

```
T020, T021, T022 — write all US1 tests in parallel
T023 → T024 → T025   (validator: structs → New → Validate)
T026                  (httpclient, independent)
T027 → T028 → T029–T032 → T033  (tools: shared helper → path match → handlers → result builder)
T034 → T035 → T036 → T037  (server: New → Start → Stop/APIs → register tools)
T038 → T039  (CLI: flags → viper binding)
```

### Phases 4 and 5 — can run in parallel once Phase 2 complete

```
Phase 4 tasks (T040–T044) and Phase 5 tasks (T045–T049) can proceed in parallel
if staffed — they touch different files (explore.go vs schema.go)
```

---

## Implementation Strategy

### MVP (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks everything)
3. Complete Phase 3: US1 — validated HTTP tools + CLI binary
4. **STOP and VALIDATE**: `go test ./tests/...` all pass; run binary against Petstore
5. Ship/demo MVP

### Incremental Delivery

1. Setup + Foundational → all packages compilable
2. US1 → validated HTTP calls → **MVP**
3. US2 → `explore_api` discovery
4. US3 → `get_schema` schema retrieval
5. US4 → multi-API + malformed definition handling
6. Polish → benchmarks, library embedding test, lint clean

### Parallel Team Strategy

Once Phase 2 is complete:
- Developer A: US1 (validator + HTTP tools + server wiring + CLI)
- Developer B: US2 + US3 (explore and schema tools — independent files)

---

## Notes

- `[P]` tasks = different files, no blocking dependencies
- `[USn]` label maps each task to its user story for traceability
- Tests marked ⚠️ MUST fail before implementation — Red-Green-Refactor strictly enforced
- Each user story phase has an independent test checkpoint before moving on
- `pkg/config`, `pkg/openapimcp`, `pkg/validator`, `pkg/registry` are stable public API — doc comments required on all exported symbols before Phase 7 ends
- `pkg/tools/`, `pkg/loader/`, `pkg/httpclient/` are exported but internal-implementation — doc comments recommended but not gated
