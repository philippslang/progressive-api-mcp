# Research: Body-Only HTTP Responses

## Finding 1: Where the response shape is currently determined

**Decision**: The full response object (`status_code`, `headers`, `body`) is assembled in `executeHTTP` in `pkg/tools/http.go` and returned as a JSON-encoded `HTTPResult`. To support body-only mode, `executeHTTP` needs to know whether the caller API has `response_body_only` set, and if so, skip the `HTTPResult` envelope and return the body value directly.

**Rationale**: `executeHTTP` is the single place where the MCP response is built for all four HTTP tools. Adding the body-only branch here requires no changes to the tool registration code or the validator path — only to the result construction.

**Alternatives considered**: Filtering in the tool handler closures (http_get, http_post, etc.) after `validateAndExecute` returns. Rejected because it would duplicate the same logic four times and requires exposing the body-only flag to all four closures independently.

---

## Finding 2: How to pass the body-only flag to `executeHTTP`

**Decision**: Add `ResponseBodyOnly bool` to `config.APIConfig` (yaml: `response_body_only`). Copy to `registry.APIEntry` at load time. Pass `entry.IgnoreHeaders` is already passed implicitly via `entry` — do the same for `ResponseBodyOnly`: read it from `entry.Config.ResponseBodyOnly` inside `executeHTTP` (which already receives `entry`).

**Rationale**: `executeHTTP` already receives the `registry.APIEntry`, which contains `entry.Config`. Reading `entry.Config.ResponseBodyOnly` inside `executeHTTP` requires zero additional function parameters. This follows the same pattern already established for `AllowList` and `IgnoreHeaders`.

**Alternatives considered**: Add a separate `ResponseBodyOnly bool` field directly on `APIEntry` (promoted out of Config). Marginally cleaner but unnecessary — `Config` is already on the entry and other booleans follow the same pattern.

---

## Finding 3: What "body only" means for JSON vs non-JSON responses

**Decision**: The `body` field in `HTTPResult` is already typed `any` and is either a parsed JSON value or a raw string. When body-only mode is active, marshal and return that same `any` value directly — same serialisation logic, just without the `HTTPResult` wrapper.

**Rationale**: No new parsing is needed. The existing `bodyVal` local variable in `executeHTTP` already holds the right value in the right form. A single `json.Marshal(bodyVal)` replaces `json.Marshal(result)`.

**Alternatives considered**: Return a different type for body-only mode. Rejected — using the same `bodyVal` keeps the output format identical to what `body` would have contained in the full response, ensuring predictable caller behaviour.

---

## Finding 4: No new dependencies required

**Decision**: Implement using existing stdlib and packages only. No new modules.

**Rationale**: The change is two lines of conditional logic in `executeHTTP` plus one new boolean field in the config struct.
