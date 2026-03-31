# Data Model: API-Level Skip Validation Configuration

**Branch**: `013-api-skip-validation` | **Date**: 2026-03-31

## Changed Entities

### APIConfig (`pkg/config/config.go`)

Adds one optional boolean field. All existing fields unchanged.

| Field | Type | YAML Tag | Default | Description |
|-------|------|----------|---------|-------------|
| Name | string | `name` | — | Unique API identifier |
| Definition | string | `definition` | — | Path to OpenAPI spec file |
| Host | string | `host` | — | Override base URL |
| BasePath | string | `base_path` | — | Override base path |
| AllowList | APIAllowList | `allow_list` | — | Tool/path restrictions |
| IgnoreHeaders | []string | `ignore_headers` | — | Headers to suppress |
| ResponseBodyOnly | bool | `response_body_only` | false | Omit response envelope |
| **SkipValidation** | **bool** | **`skip_validation`** | **false** | **Bypass payload validation** |

**Validation rules**: No new validation required. `SkipValidation` is always valid as a bool; zero value (`false`) is the correct default.

---

### APIEntry (`pkg/registry/registry.go`)

Adds one boolean field mirroring the config field. Set during `Registry.Load()` from the corresponding `APIConfig`.

| Field | Type | Set From | Description |
|-------|------|----------|-------------|
| Name | string | APIConfig.Name | Lookup key |
| Doc | libopenapi.Document | parsed spec | OpenAPI document |
| Validator | *validator.Validator | created on load | Schema validator |
| Config | APIConfig (or subset) | APIConfig | Per-API settings |
| **SkipValidation** | **bool** | **APIConfig.SkipValidation** | **Bypass validation at call time** |

---

## Call Flow Change

Current flow in `validateAndExecute()` (`pkg/tools/http.go`):

```
resolveAPI → matchPath → checkAllowList → buildHTTPRequest → Validator.Validate(req) → executeHTTP
                                                              ↑
                                                     always executed today
```

New flow:

```
resolveAPI → matchPath → checkAllowList → buildHTTPRequest → [if !entry.SkipValidation] Validator.Validate(req) → executeHTTP
                                                              ↑
                                                     skipped when SkipValidation=true
```

The `executeHTTP()` call is unchanged. Error handling for validation failures is only reached when validation runs.

---

## State Transitions

None. `SkipValidation` is set once at server startup during config load and is immutable at runtime (no hot-reload mechanism exists).

---

## Backward Compatibility

All existing YAML configs that omit `skip_validation` continue to work identically. The Go zero value for `bool` is `false`, so no migration or default-setting code is needed.
