# Research: API-Level Skip Validation Configuration

**Branch**: `013-api-skip-validation` | **Date**: 2026-03-31

## Decision 1: Where to add the skip flag

**Decision**: Add `SkipValidation bool` to both `APIConfig` (config layer) and `APIEntry` (registry layer).

**Rationale**: The config struct (`pkg/config/config.go`) is the source of truth for operator-supplied settings. `APIEntry` (`pkg/registry/registry.go`) is the runtime representation of a loaded API and already holds per-API state (validator, doc, etc.). Storing the flag on `APIEntry` avoids passing the full config down to the tool handler layer — the tool handler only interacts with registry entries, not raw config.

**Alternatives considered**:
- Store only on `APIConfig`, look up from config in tool handler — rejected because tool handlers receive `APIEntry` from the registry, not config; adding a config lookup would couple the tool layer to the config package.
- Global server-level skip flag — rejected per spec: must be per-API.

---

## Decision 2: Where to enforce the skip in the call path

**Decision**: Add a single conditional guard at line 253 of `pkg/tools/http.go` inside `validateAndExecute()`, just before the `entry.Validator.Validate(req)` call.

**Rationale**: `validateAndExecute()` is the single choke point for all four HTTP tool methods (`GET`, `POST`, `PUT`, `PATCH`). A single check here covers all methods with no duplication. The existing code structure is:

```
buildRequest → matchPath → checkAllowList → buildHTTPRequest → [VALIDATE] → executeHTTP
```

Skipping validation means branching at `[VALIDATE]` based on `entry.SkipValidation`.

**Alternatives considered**:
- Skip at the validator level — rejected because `validator.Validate()` has no access to per-API config; it operates on a raw `*http.Request`.
- Skip in `Registry.Load()` by not creating the validator — rejected because `explore_api` and `get_schema` tools may still use the document for schema exposure; the validator is separate from the document.

---

## Decision 3: Default value behavior

**Decision**: Omitting `skip_validation` from YAML config defaults to `false` (validation enabled).

**Rationale**: Go zero-value for `bool` is `false`. No explicit default-setting code is needed. Existing configs that omit the field continue to work identically.

---

## Decision 4: No changes to MCP tool signatures or schema exposure

**Decision**: The tool input schemas exposed to the LLM via MCP are generated from the OpenAPI document and remain unchanged.

**Rationale**: `skip_validation` controls whether the MCP server validates before forwarding — it does not change what the tool accepts at the MCP protocol level. The OpenAPI schema is still used to describe the tool's parameters; only the runtime enforcement of that schema is skipped.

---

## Unknowns resolved

No NEEDS CLARIFICATION items remain. All decisions above are derived from reading existing source code and the feature spec.
