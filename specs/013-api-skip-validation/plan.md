# Implementation Plan: API-Level Skip Validation Configuration

**Branch**: `013-api-skip-validation` | **Date**: 2026-03-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/013-api-skip-validation/spec.md`

## Summary

Add an optional `skip_validation` boolean field to the per-API configuration. When `true`, the MCP server bypasses request payload schema validation in `validateAndExecute()` and forwards the payload to the upstream API unchanged. Default `false` preserves existing behavior. The change touches three layers: the config struct, the registry entry, and the tool handler.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator, spf13/cobra + viper — no new dependencies
**Storage**: In-memory only (no persistence)
**Testing**: `go test` + testify; table-driven unit tests, httptest-based integration tests, contract shape tests
**Target Platform**: Linux server (HTTP and stdio transports)
**Project Type**: Library (`pkg/openapimcp`) + CLI thin wrapper (`cmd/prograpimcp`)
**Performance Goals**: No change — the skip path removes a validation call, so it is equal or faster than current
**Constraints**: No new dependencies; backward-compatible YAML config (omitted field = `false`)
**Scale/Scope**: Single-binary MCP server; change is confined to three files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution file (`.specify/memory/constitution.md`) contains only the unfilled template — no project-specific gates are defined. Applying the CLAUDE.md guidelines instead:

| Rule | Status |
|------|--------|
| `pkg/` packages MUST NOT import `cobra` or `viper` | PASS — change is in `pkg/config`, `pkg/registry`, `pkg/tools` only |
| No new dependencies | PASS — boolean field addition, no new imports |
| Exported types/functions MUST have doc comments | PASS — new field will be documented |
| No global state; dependencies via constructors | PASS — field flows through existing constructor chain |
| Standard Go idioms; `gofmt` enforced | PASS |

No violations. No Complexity Tracking entries needed.

## Project Structure

### Documentation (this feature)

```text
specs/013-api-skip-validation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── config-schema.md
└── tasks.md             # Phase 2 output (/speckit.tasks — NOT created here)
```

### Source Code (repository root)

```text
pkg/config/
└── config.go            # Add SkipValidation bool to APIConfig

pkg/registry/
└── registry.go          # Add SkipValidation bool to APIEntry; set from APIConfig on Load()

pkg/tools/
└── http.go              # Skip validator.Validate() call when entry.SkipValidation is true

tests/unit/
└── config_test.go       # Add table cases: skip_validation true/false/absent

tests/integration/
└── http_tools_test.go   # Add end-to-end case: invalid payload passes through when skip=true

tests/contract/
└── mcp_tools_contract_test.go  # No change required (tool signatures unchanged)
```

**Structure Decision**: Single-project layout (existing). No new packages or directories required in source. All changes are additive within existing files.
