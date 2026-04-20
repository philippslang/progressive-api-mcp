# Quickstart: search_api Tool

**Feature**: 014-search-api-tool

## Use case

You have two APIs loaded — `petstore` and `bookstore` — and you want to find every endpoint that deals with "order" without navigating each API's paths via `explore_api`.

## Example call — all APIs

```json
{
  "tool": "search_api",
  "arguments": { "query": "order" }
}
```

Returns every endpoint in every registered API whose path or description contains "order":

```json
[
  { "api": "petstore",   "method": "GET",  "path": "/orders/{id}" },
  { "api": "petstore",   "method": "POST", "path": "/orders", "schema": { "...": "..." } },
  { "api": "bookstore",  "method": "GET",  "path": "/orders/{orderId}/status" }
]
```

## Example call — single API filter

```json
{
  "tool": "search_api",
  "arguments": { "query": "pet", "api": "petstore" }
}
```

Returns only petstore endpoints matching "pet".

## Example — empty query

```json
{ "tool": "search_api", "arguments": { "query": "  " } }
```

→ `{"code": "INVALID_INPUT", "message": "query must not be empty"}`

## Example — unknown API

```json
{ "tool": "search_api", "arguments": { "query": "pet", "api": "nosuch" } }
```

→ `{"code": "API_NOT_FOUND", "message": "API \"nosuch\" is not registered"}`

## Typical workflow

1. `search_api` with a keyword → pick the most relevant `{api, method, path}` tuple
2. `get_schema` with that `(api, method, path)` → inspect full details
3. `http_<method>` with that `(api, path, body)` → call the endpoint

## Allow-list restriction

If your config restricts tools:

```yaml
apis:
  - name: petstore
    allow_list:
      tools: [search_api, get_schema]
      paths:
        search_api: ["/pets", "/pets/{id}"]
```

`search_api` is registered (it's in `tools`), and only `/pets` and `/pets/{id}` for petstore appear in results — other paths are filtered.
