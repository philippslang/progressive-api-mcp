# Research: Ignore Headers for APIs

## Finding 1: How headers appear in the current codebase

**Decision**: Header parameters (`in: "header"` in OpenAPI) are currently NOT extracted or displayed by either `get_schema` or `explore_api`. The `get_schema` parameter loop (`switch param.In`) only handles `path` and `query`; header params are silently skipped. The `explore_api` tool only lists paths and HTTP methods — no parameters at all.

**Rationale**: To make "ignore headers" meaningful for schema display, `get_schema` must first be extended to output header parameters, then apply the filter. For `explore_api`, headers are already invisible so no filter is needed there.

**Alternatives considered**: Leave schema display unchanged and only fix validation. Rejected because the spec explicitly calls out `get_schema` as a place where ignored headers should not appear — making the feature incomplete if we skip it.

---

## Finding 2: How to suppress required-header validation

**Decision**: In `validateAndExecute` (`pkg/tools/http.go`), inject a placeholder value (e.g. `"*"`) into the synthetic validation request for every header listed in `ignore_headers` that the caller did not explicitly supply. This makes `libopenapi-validator` see the header as present, so it does not raise a required-parameter error.

**Rationale**: The `libopenapi-validator` operates on an `http.Request` — it cannot be told to skip specific parameters. Injecting a placeholder into the synthetic request (used only for schema validation, not forwarded) is the minimal-invasive approach: it requires no changes to `pkg/validator`, introduces no new dependency, and is fully reversible.

**Alternatives considered**:
- Post-validation error filtering: filter out errors whose `Field` matches an ignored header. Rejected because `extractField` in the validator is heuristic and may miss edge cases; injecting a placeholder is more robust.
- Rebuild a stripped OpenAPI document without the ignored headers. Rejected as expensive and complex — requires rebuilding the `libopenapi` document and validator per request or per config load.

---

## Finding 3: Where to store `IgnoreHeaders` at runtime

**Decision**: Add `IgnoreHeaders []string` to both `config.APIConfig` (YAML source) and `registry.APIEntry` (runtime). Copy the value in `registry.Load` alongside `AllowList`. Build a case-insensitive lookup set in tools at call time (or store a pre-built `map[string]struct{}` on `APIEntry`).

**Rationale**: This follows the exact pattern already used for `AllowList` — config struct carries the raw value, registry entry carries the runtime copy. A pre-built lowercase set on `APIEntry` avoids repeated allocations per tool call.

**Alternatives considered**: Store `IgnoreHeaders` only on `config.APIConfig` and look it up via `entry.Config.IgnoreHeaders`. Rejected to maintain parity with `AllowList` and to allow future callers to use the entry directly.

---

## Finding 4: Case-insensitive matching strategy

**Decision**: Normalize to lowercase at load time (when copying from config to `APIEntry`). Store as `map[string]struct{}` keyed on `strings.ToLower(name)`. Lookups use `strings.ToLower(name)` before checking the map.

**Rationale**: HTTP header names are case-insensitive per RFC 7230. Normalising once at load time is cheaper than repeated `strings.EqualFold` comparisons per parameter in every tool call.

**Alternatives considered**: `strings.EqualFold` loop at each lookup. Rejected for performance on APIs with many headers; also more code per lookup site.

---

## Finding 5: No new dependencies required

**Decision**: Implement entirely with the existing stdlib (`strings`) and current packages (`pkg/config`, `pkg/registry`, `pkg/tools`, `pkg/validator`). No new modules needed.

**Rationale**: All required primitives (string normalisation, map lookup, request mutation) are in the standard library. `libopenapi-validator` is already present and its behaviour is exploited, not replaced.
