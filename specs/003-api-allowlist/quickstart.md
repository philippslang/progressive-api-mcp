# Quickstart: API Tool & Path Allow List

**Feature**: 003-api-allowlist

## What this does

Adds an optional `allow_list` block to each API entry in `config.yaml`. It controls:
1. **Which MCP tools** are available for that API (tool-level restriction)
2. **Which OpenAPI paths** each tool may access (path-level restriction)

Both levels are optional — omitting either gives full access.

## Minimal example: read-only API

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    allow_list:
      tools:
        - explore_api
        - get_schema
        - http_get
```

Result: `http_post`, `http_put`, `http_patch` are not registered in the MCP session. An agent cannot mutate the API.

## Path restriction example

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    allow_list:
      tools:
        - explore_api
        - http_get
        - http_post
      paths:
        http_post:
          - /pets          # only POST to /pets is allowed; /owners is blocked
```

Result: `http_post /pets` works; `http_post /owners` returns `PATH_NOT_PERMITTED`.

## Multi-API example

```yaml
apis:
  - name: internal-admin
    definition: ./admin.yaml
    host: https://admin.internal
    allow_list:
      tools:
        - explore_api
        - get_schema     # agents can explore and read schemas but not call anything

  - name: public-api
    definition: ./public.yaml
    host: https://api.example.com
    # No allow_list — all tools and paths open
```

## Default: no allow_list = allow all

```yaml
apis:
  - name: petstore
    definition: ./petstore.yaml
    host: https://api.example.com
    # No allow_list — identical to pre-feature behaviour
```

## Path template vs. concrete path

`allowed_paths` uses OpenAPI path templates, not concrete URLs:

```yaml
paths:
  http_get:
    - /pets          # matches GET /pets exactly
    - /pets/{id}     # matches GET /pets/42, /pets/fluffy, etc.
```

`/pets/42` is **not** a valid allow-list entry — use `/pets/{id}` instead.

## Error responses

| Situation | Error code |
|-----------|-----------|
| Path not in OpenAPI spec | `PATH_NOT_FOUND` (existing) |
| Path in spec but not in allow list | `PATH_NOT_PERMITTED` (new) |
| Tool restricted on this API (multi-API) | `TOOL_NOT_PERMITTED` (new) |
