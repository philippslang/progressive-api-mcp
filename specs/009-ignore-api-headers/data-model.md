# Data Model: Ignore Headers for APIs

## Entities

### `config.APIConfig` (modified)

Existing struct in `pkg/config/config.go`. Gains one new field.

| Field           | Type       | YAML key         | Validation                                   |
|-----------------|------------|------------------|----------------------------------------------|
| `IgnoreHeaders` | `[]string` | `ignore_headers` | Optional. Each entry must be a non-empty string. Duplicates silently collapsed. |

No existing fields are removed or renamed.

---

### `registry.APIEntry` (modified)

Existing struct in `pkg/registry/registry.go`. Gains one new field populated at load time.

| Field              | Type                  | Source                                      |
|--------------------|-----------------------|---------------------------------------------|
| `IgnoreHeaders`    | `map[string]struct{}` | Built from `config.APIConfig.IgnoreHeaders` at `Load` time, keys lowercased. |

Zero value (`nil` map) means no headers are suppressed — identical to current behaviour.

---

### `tools.SchemaResult` (modified)

Existing struct in `pkg/tools/schema.go`. Gains one new optional output field.

| Field              | Type              | JSON key            | Populated when                            |
|--------------------|-------------------|---------------------|-------------------------------------------|
| `HeaderParameters` | `map[string]any`  | `header_parameters` | At least one non-ignored header parameter exists on the operation. Omitted (`omitempty`) when empty. |

Existing fields (`PathParameters`, `QueryParameters`, `RequestBody`, `Responses`) are unchanged.

---

## State Transitions

No persistent state. All data lives in memory for the lifetime of the server process.

- **Config loaded** → `APIConfig.IgnoreHeaders` populated from YAML.
- **Registry.Load called** → `APIEntry.IgnoreHeaders` built as lowercase map.
- **Tool call (get_schema)** → header parameters extracted from OpenAPI model; each filtered against `APIEntry.IgnoreHeaders` before inclusion in `SchemaResult.HeaderParameters`.
- **Tool call (http_*)** → for each name in `APIEntry.IgnoreHeaders`, if not already present in the synthetic validation request, a placeholder header value is injected; validation proceeds; placeholder headers are NOT forwarded to the upstream API.

## Validation Rules

- `IgnoreHeaders` entries with empty strings are invalid; `config.Validate()` MUST return an error.
- Duplicate names in `IgnoreHeaders` (case-insensitive) are collapsed to one entry at load time; no error raised.
- An `IgnoreHeaders` list that refers to a header not present in the API schema is silently accepted.
