package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/philippslang/progressive-api-mcp/pkg/registry"
)

// SearchResult is one matching endpoint returned by the search_api tool.
type SearchResult struct {
	API    string         `json:"api"`
	Method string         `json:"method"`
	Path   string         `json:"path"`
	Schema map[string]any `json:"schema,omitempty"`
}

// RegisterSearchTools registers the search_api MCP tool.
// prefix is prepended to the tool name; empty means no prefix.
// allowedTools restricts whether the tool is registered; nil means it is always registered.
func RegisterSearchTools(s *server.MCPServer, reg *registry.Registry, prefix string, allowedTools map[string]bool) {
	if allowedTools != nil && !allowedTools["search_api"] {
		return
	}
	s.AddTool(mcp.NewTool(applyPrefix(prefix, "search_api"),
		mcp.WithDescription("Search endpoints across APIs by substring match on path or operation description"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Substring to match (case-insensitive) against path, summary, or description")),
		mcp.WithString("api", mcp.Description("Optional: restrict search to a single registered API by name")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := strings.TrimSpace(req.GetString("query", ""))
		apiFilter := req.GetString("api", "")

		if query == "" {
			r, _ := toolErrorResult("INVALID_INPUT", "query must not be empty", nil, nil)
			return r, nil
		}

		var entries []registry.APIEntry
		if apiFilter != "" {
			entry, ok := reg.Lookup(apiFilter)
			if !ok {
				r, _ := toolErrorResult("API_NOT_FOUND",
					fmt.Sprintf("API %q is not registered", apiFilter), nil, nil)
				return r, nil
			}
			entries = []registry.APIEntry{entry}
		} else {
			for _, name := range reg.ListNames() {
				if e, ok := reg.Lookup(name); ok {
					entries = append(entries, e)
				}
			}
		}

		needle := strings.ToLower(query)
		results := []SearchResult{}

		for _, entry := range entries {
			model, errs := entry.Document().BuildV3Model()
			if len(errs) > 0 || model.Model.Paths == nil || model.Model.Paths.PathItems == nil {
				continue
			}
			allowedPaths := entry.AllowList.Paths["search_api"]
			for path, item := range model.Model.Paths.PathItems.FromOldest() {
				if !IsPathPermitted(path, allowedPaths) {
					continue
				}
				for _, mo := range methodOps(item) {
					if mo.op == nil {
						continue
					}
					if !matches(needle, path, mo.op) {
						continue
					}
					results = append(results, SearchResult{
						API:    entry.Name,
						Method: mo.method,
						Path:   path,
						Schema: requestBodySchema(mo.op),
					})
				}
			}
		}

		data, _ := json.Marshal(results)
		return mcp.NewToolResultText(string(data)), nil
	})
}

// methodOp pairs an HTTP method name with its operation pointer from a PathItem.
type methodOp struct {
	method string
	op     *v3high.Operation
}

// methodOps returns the (method, operation) pairs for a PathItem in a deterministic order.
func methodOps(item *v3high.PathItem) []methodOp {
	return []methodOp{
		{"GET", item.Get},
		{"POST", item.Post},
		{"PUT", item.Put},
		{"PATCH", item.Patch},
		{"DELETE", item.Delete},
	}
}

// matches reports whether the lowercased needle appears in the path, operation summary, or description.
func matches(lowerNeedle, path string, op *v3high.Operation) bool {
	if strings.Contains(strings.ToLower(path), lowerNeedle) {
		return true
	}
	if op.Summary != "" && strings.Contains(strings.ToLower(op.Summary), lowerNeedle) {
		return true
	}
	if op.Description != "" && strings.Contains(strings.ToLower(op.Description), lowerNeedle) {
		return true
	}
	return false
}

// requestBodySchema extracts the JSON request-body schema for an operation, or nil when absent.
func requestBodySchema(op *v3high.Operation) map[string]any {
	if op.RequestBody == nil || op.RequestBody.Content == nil {
		return nil
	}
	out := map[string]any{}
	if op.RequestBody.Required != nil {
		out["required"] = *op.RequestBody.Required
	}
	for contentType, mediaType := range op.RequestBody.Content.FromOldest() {
		if strings.Contains(contentType, "json") && mediaType.Schema != nil {
			if schema := mediaType.Schema.Schema(); schema != nil {
				out["schema"] = schemaToMap(schema)
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
