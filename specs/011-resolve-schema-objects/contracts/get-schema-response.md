# Contract: get_schema Response — Resolved Schema Objects

## Breaking Change Notice

The `request_body` field shape changes. The key `"type"` is renamed to `"schema"` and now contains a fully resolved schema object instead of a shallow `{ "type": "object" }`. Callers relying on `request_body["type"]` must update to `request_body["schema"]`.

---

## `get_schema` Tool Response Contract

### Request Body (changed)

```json
{
  "request_body": {
    "required": true,
    "schema": {
      "type": "object",
      "description": "Optional description from OpenAPI spec",
      "properties": {
        "<property-name>": {
          "type": "<primitive-type>",
          "format": "<format-if-present>",
          "description": "<description-if-present>",
          "required": true
        }
      }
    }
  }
}
```

### Response Schemas (unchanged key name, enriched content)

```json
{
  "responses": {
    "<status-code>": {
      "description": "<response-description>",
      "schema": {
        "type": "object",
        "properties": {
          "<property-name>": {
            "type": "<type>",
            "required": false
          }
        }
      }
    }
  }
}
```

### Array Response Example

```json
{
  "responses": {
    "200": {
      "description": "A list of pets",
      "schema": {
        "type": "array",
        "items": {
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
}
```

### Rules

- `properties` is present only when the schema declares named properties.
- `required` on each property is `true` if that property name appears in the parent schema's `required` list.
- `items` is present only when `type` is `"array"` and the schema declares an item schema.
- Recursion terminates at depth 10; deeply nested schemas return `{ "type": "object" }` as a placeholder.
- `allOf`, `oneOf`, `anyOf` are not expanded; they are omitted from the output in this version.
- Primitive schemas (string, integer, boolean, number) include `type`, `format`, and `description` as available.
