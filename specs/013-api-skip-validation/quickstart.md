# Quickstart: Skip Validation for an API

**Feature**: 013-api-skip-validation

## Use case

You have an upstream API whose payloads do not fully conform to its own OpenAPI schema (e.g., accepts extra fields, uses non-standard types), and the MCP server's validator is rejecting requests before they reach the API. You want to bypass validation for that API only.

## Configuration

Add `skip_validation: true` to the API entry in your config file:

```yaml
server:
  transport: http
  port: 8080

apis:
  - name: lenient-api
    definition: ./lenient-api.yaml
    host: https://api.example.com
    skip_validation: true          # <-- add this

  - name: strict-api
    definition: ./strict-api.yaml
    host: https://strict.example.com
    # skip_validation omitted = validation enabled (default)
```

## Effect

- Tool calls targeting `lenient-api` will skip payload schema validation and forward the request directly to `https://api.example.com`.
- Tool calls targeting `strict-api` continue to be validated before forwarding.
- If the upstream API rejects the payload, that error response is returned to the caller as-is.

## No restart required for config changes

Reload the server (restart the process) after editing the config file — there is no hot-reload mechanism.
