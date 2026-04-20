# Research: Rename `explore_api` to `list_api`

**Feature**: 015-rename-explore-to-list | **Date**: 2026-04-18

No `NEEDS CLARIFICATION` markers remained after the spec phase, so research here is short and exists mainly to lock in decisions that affect scope.

## Decision 1: Hard rename, no alias

- **Decision**: Remove the name `explore_api` outright. Do not keep it registered as an alias, and do not accept it in `allow_list.tools` or `allow_list.paths`.
- **Rationale**:
  - The user said "change the name," not "add an alias." An alias would widen scope and obscure the user's intent.
  - Config validation in `pkg/config/config.go` uses a whitelist (`knownToolNames`) that today contains `explore_api`. Replacing (rather than adding) that entry means operators with stale configs see a clear validation error at startup instead of silently losing the tool.
  - The hint string (`knownToolNamesHint`) is user-visible in the error — it will now list `list_api` and not `explore_api`, guiding operators directly to the fix.
- **Alternatives considered**:
  - *Dual-name alias*: accept both `explore_api` and `list_api`. Rejected — adds hidden state, defers the cleanup forever, and contradicts the user's wording.
  - *Keep old name, add `list_api` as additional alias*: even worse — double-registration would bloat `tools/list` output and complicate prefix handling.

## Decision 2: Rename only live code, tests, and user-facing docs

- **Decision**: Update the following buckets:
  - Live code: `pkg/tools/explore.go` (file rename + identifier renames), `pkg/config/config.go`, `pkg/openapimcp/server.go`.
  - Tests: `tests/integration/http_tools_test.go`, `tests/contract/mcp_tools_contract_test.go`, `tests/unit/config_test.go`.
  - User-facing docs: `README.md`, `config.yaml.example`.
  - Leave historical spec artefacts under `specs/001-…` through `specs/014-…` unchanged.
- **Rationale**:
  - Spec artefacts are frozen records of what was designed and shipped at a point in time. Rewriting history inside those files would obscure the project's evolution and is explicitly excluded by the spec's Assumptions section (SC-004 applies to user-facing docs, not archival spec files).
  - Live code and tests must change in lockstep with the rename — otherwise the build breaks or tests fail.
  - `README.md` and `config.yaml.example` are the two files a new user reads first; leaving them on the old name would directly contradict SC-004.
- **Alternatives considered**:
  - *Global search-and-replace across the entire repo, including `specs/`*: rejected — pollutes historical records and produces a noisy diff that buries the real change.

## Decision 3: Rename the Go file `explore.go` → `list.go`

- **Decision**: Rename `pkg/tools/explore.go` to `pkg/tools/list.go` and rename `RegisterExploreTools` → `RegisterListTools` inside it. Keep `PathInfo` (the result-element type) unchanged — the name is generic and already reused in contract tests.
- **Rationale**:
  - The file is the implementation for the renamed tool. Keeping the old filename while the identifiers and tool name change would be a long-lived source of confusion for anyone navigating the package.
  - Renaming `PathInfo` would cause churn with zero user-visible benefit; the type name does not refer to the tool's name.
- **Alternatives considered**:
  - *Keep filename `explore.go`*: rejected — filename and identifier naming should match.
  - *Rename `PathInfo` → `ListEntry`*: rejected — gratuitous; `PathInfo` describes what the type is, not which tool returns it.

## Decision 4: Preserve the tool's handler signature and behaviour

- **Decision**: The exported Go function signature (`RegisterListTools(s *server.MCPServer, reg *registry.Registry, prefix string, allowedTools map[string]bool)`) matches the pre-rename `RegisterExploreTools` signature byte-for-byte apart from the function name. Tool description string, parameter schemas (`api`, `prefix`), response JSON shape, error envelope, and allow-list lookup key (`"list_api"`) change only in the renamed key — the rest stays.
- **Rationale**: FR-003 mandates no behavioural change.
- **Alternatives considered**: none — this is not a judgement call.
