# Implementation Plan: Fully Resolved Schema Objects in get_schema

**Branch**: `011-resolve-schema-objects` | **Date**: 2026-03-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-resolve-schema-objects/spec.md`

## Summary

Replace the shallow `schemaToMap` helper in `pkg/tools/schema.go` with a recursive version that fully expands `Properties`, `Required`, `Items`, `Description`, and `Format` fields from the libopenapi high-level `base.Schema` struct. Also fix the wrong key name `"type"` → `"schema"` for the request body schema entry. All changes are confined to `pkg/tools/schema.go`. No new dependencies.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: mark3labs/mcp-go v0.46.0, pb33f/libopenapi + libopenapi-validator — no new dependencies
**Storage**: In-memory only (no persistence)
**Testing**: `go test ./...` — existing suite covers contract shape
**Target Platform**: Linux server
**Project Type**: CLI binary + embeddable library
**Performance Goals**: No measurable overhead — depth-limited recursion over in-memory schema objects
**Constraints**: No new dependencies; max recursion depth 10
**Scale/Scope**: One function rewrite in one file

## Constitution Check

Constitution file is a blank template with no ratified principles. No gates to evaluate.

## Project Structure

### Documentation (this feature)

```text
specs/011-resolve-schema-objects/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── get-schema-response.md  # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
pkg/tools/
└── schema.go            # Replace schemaToMap with recursive schemaToMapDepth;
                         # fix request_body key "type" -> "schema"

tests/contract/
└── mcp_tools_contract_test.go  # Update: request_body["schema"] instead of request_body["type"]
```

## Key Implementation Detail

The new recursive function signature:

```
schemaToMap(schema *base.Schema) map[string]any
  -> calls schemaToMapDepth(schema, 0)

schemaToMapDepth(schema *base.Schema, depth int) map[string]any:
  - Guard: if schema == nil or depth > 10, return {}
  - Extract: type, format, description
  - If Properties non-nil: iterate FromOldest(), build required-set from schema.Required,
    recurse into each property's .Schema() at depth+1
  - If Items non-nil and Items.N==0: recurse into Items.A.Schema() at depth+1
  - Return assembled map
```

The request body extraction changes from:
  bodySchema["type"] = schemaToMap(schema)
to:
  bodySchema["schema"] = schemaToMap(schema)

## Complexity Tracking

No constitution violations. No complexity justification required.
