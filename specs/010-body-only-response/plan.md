# Implementation Plan: Body-Only HTTP Responses

**Branch**: `010-body-only-response` | **Date**: 2026-03-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/010-body-only-response/spec.md`

## Summary

Add a per-API `response_body_only` boolean config flag. When `true`, all HTTP tool calls for that API return only the response body value instead of the full `{ status_code, headers, body }` envelope. The change is confined to one field in `pkg/config/config.go` and two lines of conditional logic in `executeHTTP` in `pkg/tools/http.go`. No new dependencies.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator — no new dependencies
**Storage**: In-memory only (no persistence)
**Testing**: `go test ./...` — existing suite; no new test files required for this change
**Target Platform**: Linux server
**Project Type**: CLI binary + embeddable library
**Performance Goals**: No measurable overhead — one boolean check per HTTP tool response
**Constraints**: No new dependencies; `pkg/` packages must not import `cobra` or `viper`
**Scale/Scope**: One boolean field; affects response shape only

## Constitution Check

Constitution file is a blank template with no ratified principles. No gates to evaluate.

## Project Structure

### Documentation (this feature)

```text
specs/010-body-only-response/
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
└── config.go            # Add ResponseBodyOnly bool to APIConfig

pkg/tools/
└── http.go              # executeHTTP: branch on entry.Config.ResponseBodyOnly
```

## Complexity Tracking

No constitution violations. No complexity justification required.
