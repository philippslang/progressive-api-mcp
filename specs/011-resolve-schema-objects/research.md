# Research: Fully Resolved Schema Objects in get_schema

## Finding 1: Root cause — `schemaToMap` is shallow

**Decision**: The only change needed is to replace `schemaToMap` in `pkg/tools/schema.go` with a recursive version that walks the `Properties`, `Required`, `Description`, and `Items` fields of `*base.Schema`.

**Rationale**: The current `schemaToMap` extracts only `type` and `format`. The libopenapi high-level `base.Schema` struct already exposes fully resolved fields — `Properties`, `Required`, `Items`, `AllOf`, etc. — so no additional parsing or `$ref` resolution is needed; libopenapi has already done that work when `BuildV3Model()` is called and when `.Schema()` is called on a `*base.SchemaProxy`.

**Alternatives considered**: Post-processing the raw YAML node tree to resolve `$ref`s manually. Rejected — libopenapi already resolves references when `.Schema()` is called on a proxy; using the high-level API is correct and idiomatic.

---

## Finding 2: High-level `base.Schema` field API (libopenapi v0.21.12)

**Decision**: Use these fields from `*base.Schema` in the recursive resolver:

| Field | Type | Usage |
|-------|------|-------|
| `Type` | `[]string` | Take `[0]` if non-empty |
| `Format` | `string` | Include if non-empty |
| `Description` | `string` | Include if non-empty |
| `Properties` | `*orderedmap.Map[string, *base.SchemaProxy]` | Iterate with `.FromOldest()` |
| `Required` | `[]string` | Build a set; mark each property as required/optional |
| `Items` | `*base.DynamicValue[*base.SchemaProxy, bool]` | If `Items.N == 0`, resolve `Items.A.Schema()` as the array item schema |
| `AllOf` | `[]*base.SchemaProxy` | Included as count/presence indicator only (out of scope for deep expansion) |

**Rationale**: These are the fields sufficient to fully describe a schema shape for MCP callers. All are already resolved by libopenapi — no extra passes needed.

---

## Finding 3: Circular reference and depth protection

**Decision**: Limit recursion to a maximum depth of 10 levels using a `depth int` parameter threaded through the recursive call. Do NOT use a visited-pointer map.

**Rationale**: Depth-limiting is simple, reliable, and sufficient for all real-world OpenAPI schemas (typically 3–5 levels deep). A visited-pointer approach is fragile because `SchemaProxy.Schema()` may return newly allocated objects on each call, making pointer equality unreliable for cycle detection. At depth > 10, return `map[string]any{"type": "object"}` as a placeholder, indicating the schema continues deeper than the resolved view.

**Alternatives considered**: Visited-pointer set. Rejected for the reason above.

---

## Finding 4: How `required` is represented in output

**Decision**: Represent `required` at the property level — each property entry includes `"required": true/false` — rather than as a separate top-level array. Build a set from `schema.Required []string` and check membership for each property.

**Rationale**: Callers using an AI agent benefit most from property-level required flags; they do not need to cross-reference a separate list. This matches the assumption documented in the spec.

---

## Finding 5: Request body schema extraction path

**Decision**: In the request body extraction block, change `bodySchema["type"] = schemaToMap(schema)` to `bodySchema["schema"] = schemaToMap(schema)`. The current code puts the resolved schema under the key `"type"` which is confusing and wrong — it nests the entire schema map under the key `"type"`. The correct key is `"schema"`.

**Rationale**: The current output `{ "required": true, "type": { "type": "object" } }` shows `"type"` holding a nested object, which is a bug in the key name. Fixing this to `"schema"` produces `{ "required": true, "schema": { "type": "object", "properties": { ... } } }` which is semantically correct.

**Impact**: This is a breaking change to the `request_body` shape in `get_schema` output. It is justified because the current output is incorrect/misleading.

---

## Finding 6: No new dependencies required

**Decision**: Use only existing libopenapi high-level API (`base.Schema`, `base.SchemaProxy`, `base.DynamicValue`, `orderedmap`). All are already imported transitively.

**Rationale**: Everything needed is already available through the existing import of `"github.com/pb33f/libopenapi/datamodel/high/base"`.
