# Data Model: Fully Resolved Schema Objects in get_schema

## Entities

### `tools.SchemaResult` (unchanged struct, changed content)

No fields are added or removed. The content of `RequestBody` and `Responses` fields changes:

| Field | Before | After |
|-------|--------|-------|
| `RequestBody["type"]` | Shallow map `{ "type": "object" }` | Removed — was a wrong key name |
| `RequestBody["schema"]` | Absent | Fully resolved schema map (new key name) |
| `Responses[code]["schema"]` | Shallow map `{ "type": "object" }` | Fully resolved schema map |

---

### Resolved Schema Map (runtime value, not a Go struct)

The `map[string]any` produced by `schemaToMap` gains new keys:

| Key | Type | Condition |
|-----|------|-----------|
| `type` | `string` | When `schema.Type` is non-empty |
| `format` | `string` | When `schema.Format` is non-empty |
| `description` | `string` | When `schema.Description` is non-empty |
| `properties` | `map[string]any` | When `schema.Properties` is non-nil and non-empty; each value is itself a resolved schema map with an added `required: bool` key |
| `items` | `map[string]any` | When schema type is array and `schema.Items.N == 0` |
| `required` | `bool` | On each property entry — `true` if the property name appears in the parent schema's `Required` list |

Depth limit: recursion terminates at depth 10. At that depth, returns `{ "type": "object" }` as a placeholder.

---

## Output Shape Change — Request Body

**Before**:
```json
{
  "request_body": {
    "required": true,
    "type": {
      "type": "object"
    }
  }
}
```

**After**:
```json
{
  "request_body": {
    "required": true,
    "schema": {
      "type": "object",
      "description": "A new pet",
      "properties": {
        "name": { "type": "string", "description": "The name of the pet", "required": true },
        "age":  { "type": "integer", "required": false },
        "tag":  { "type": "string", "required": false }
      }
    }
  }
}
```

## Output Shape Change — Responses

**Before**:
```json
{
  "responses": {
    "201": {
      "description": "Pet created",
      "schema": { "type": "object" }
    }
  }
}
```

**After**:
```json
{
  "responses": {
    "201": {
      "description": "Pet created",
      "schema": {
        "type": "object",
        "properties": {
          "id":   { "type": "integer", "required": true },
          "name": { "type": "string",  "required": true },
          "tag":  { "type": "string",  "required": false }
        }
      }
    }
  }
}
```
