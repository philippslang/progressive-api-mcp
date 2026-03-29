# Implementation Plan: Ignore Headers for APIs

**Branch**: `009-ignore-api-headers` | **Date**: 2026-03-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/009-ignore-api-headers/spec.md`

## Summary

Add a per-API `ignore_headers` config option. Declared headers are excluded from `get_schema` header parameter output and are treated as satisfied (not required) during request validation in all HTTP tools. No new dependencies; changes are confined to `pkg/config`, `pkg/registry`, and `pkg/tools`.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator, spf13/cobra + viper — no new dependencies
**Storage**: In-memory only (no persistence)
**Testing**: `go test ./...` — table-driven unit tests, integration tests via `httptest`
**Target Platform**: Linux server (also runs on any Go-supported OS)
**Project Type**: CLI binary + embeddable library
**Performance Goals**: No measurable overhead — O(n) header lookup where n = number of ignored headers per API (typically < 10)
**Constraints**: No new dependencies; `pkg/` packages must not import `cobra` or `viper`
**Scale/Scope**: Per-API, per-header-name suppression; affects all 4 HTTP tools uniformly

## Constitution Check

Constitution file is a blank template with no ratified principles. No gates to evaluate.

## Project Structure

### Documentation (this feature)

```text
specs/009-ignore-api-headers/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── config-schema.md # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
pkg/config/
└── config.go            # Add IgnoreHeaders []string to APIConfig; validate no empty entries

pkg/registry/
└── registry.go          # Add IgnoreHeaders map[string]struct{} to APIEntry; build at Load time

pkg/tools/
├── http.go              # validateAndExecute: inject placeholder headers before validation
└── schema.go            # get_schema: extract header params into HeaderParameters, filter ignored

tests/unit/
├── config_test.go       # New cases: valid ignore_headers, empty string entry rejected, duplicates ok
└── registry_test.go     # New cases: IgnoreHeaders map built correctly (case-insensitive)

tests/integration/
└── ignore_headers_test.go  # End-to-end: get_schema hides ignored header, http_get skips validation

tests/contract/
└── mcp_tools_contract_test.go  # Add: header_parameters field in get_schema response shape
```

## Complexity Tracking

No constitution violations. No complexity justification required.
