# Quickstart: `list_api` Tool

**Feature**: 015-rename-explore-to-list

This is the same tool previously known as `explore_api`. Only the name has changed.

## Discovering endpoints — single API

```json
{ "tool": "list_api", "arguments": {} }
```

Returns every path in the single registered API, sorted by path:

```json
[
  { "path": "/pets",      "methods": ["GET", "POST"], "description": "List all pets" },
  { "path": "/pets/{id}", "methods": ["GET", "PUT", "PATCH", "DELETE"], "description": "Get a pet by ID" }
]
```

## Discovering endpoints — multiple APIs

With two APIs registered (`petstore`, `bookstore`):

```json
{ "tool": "list_api", "arguments": { "api": "petstore" } }
```

Returns only petstore paths. Omitting `api` when multiple APIs are registered returns an `AMBIGUOUS_API` error whose `hints` array lists the registered API names.

## Prefix filter

```json
{ "tool": "list_api", "arguments": { "api": "petstore", "prefix": "/pets" } }
```

Returns only paths starting with `/pets`.

## Tool prefix

With `server.tool_prefix: "store"` configured, the tool is advertised as `store_list_api`:

```json
{ "tool": "store_list_api", "arguments": { "api": "petstore" } }
```

## Allow-list restriction

```yaml
apis:
  - name: petstore
    allow_list:
      tools: [list_api, get_schema]
      paths:
        list_api: ["/pets", "/pets/{id}"]
```

`list_api` is registered (listed in `tools`). Only `/pets` and `/pets/{id}` appear in the output — other paths are filtered.

## Migration from `explore_api`

- Replace every call-site's tool name from `explore_api` → `list_api`.
- Replace every `allow_list.tools: [..., explore_api, ...]` entry with `list_api`.
- Replace every `allow_list.paths: { explore_api: [...] }` key with `list_api`.
- Configs that still reference `explore_api` are rejected at startup with a validation error naming `list_api` as a valid alternative.
