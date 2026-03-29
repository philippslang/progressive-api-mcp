# Data Model: Body-Only HTTP Responses

## Entities

### `config.APIConfig` (modified)

Gains one new field.

| Field              | Type   | YAML key              | Validation                        |
|--------------------|--------|-----------------------|-----------------------------------|
| `ResponseBodyOnly` | `bool` | `response_body_only`  | Optional. Default: `false`. No additional validation needed. |

No existing fields are removed or renamed.

---

### `tools.HTTPResult` (unchanged)

The `HTTPResult` struct is not modified. In body-only mode the struct is simply not used — `bodyVal` is marshalled directly instead of populating an `HTTPResult`. The struct remains the shape returned in the default (non-body-only) mode.

---

## Behaviour

- **`response_body_only: false` (default)**: `executeHTTP` returns `json.Marshal(HTTPResult{StatusCode, Headers, Body})` — current behaviour, unchanged.
- **`response_body_only: true`**: `executeHTTP` returns `json.Marshal(bodyVal)` where `bodyVal` is the same value that would have been placed in `HTTPResult.Body` — a parsed JSON value or a raw string.

## Validation Rules

- No validation is needed beyond YAML bool parsing. An absent field defaults to `false`.
