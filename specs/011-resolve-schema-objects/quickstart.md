# Quickstart: Fully Resolved Schema Objects in get_schema

## What changed

`get_schema` now returns fully expanded object schemas. Before, any object type showed only `{ "type": "object" }`. After, all properties, types, required flags, descriptions, and nested objects are included.

## Before / After — POST /pets

**Before**:
```json
{
  "method": "POST",
  "path": "/pets",
  "request_body": {
    "required": true,
    "type": { "type": "object" }
  },
  "responses": {
    "201": { "description": "Pet created", "schema": { "type": "object" } }
  }
}
```

**After**:
```json
{
  "method": "POST",
  "path": "/pets",
  "request_body": {
    "required": true,
    "schema": {
      "type": "object",
      "properties": {
        "name": { "type": "string", "required": true },
        "age":  { "type": "integer", "required": false },
        "tag":  { "type": "string", "required": false }
      }
    }
  },
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

## Breaking change

The `request_body` key is renamed from `"type"` to `"schema"`:
- **Before**: `request_body.type` held the schema map
- **After**: `request_body.schema` holds the schema map

## Scope

- Applies to all endpoints in all APIs.
- No config change needed.
- `explore_api` is unchanged.
