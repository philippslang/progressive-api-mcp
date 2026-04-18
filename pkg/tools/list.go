package tools

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/philippslang/progressive-api-mcp/pkg/registry"
)

// PathInfo is one entry in the exploration tool's result list.
type PathInfo struct {
	Path        string   `json:"path"`
	Methods     []string `json:"methods"`
	Description string   `json:"description"`
}

// RegisterListTools registers the list_api MCP tool.
// prefix is prepended to the tool name; empty means no prefix.
// allowedTools restricts whether the tool is registered; nil means it is always registered.
func RegisterListTools(s *server.MCPServer, reg *registry.Registry, prefix string, allowedTools map[string]bool) {
	if allowedTools != nil && !allowedTools["list_api"] {
		return
	}
	s.AddTool(mcp.NewTool(applyPrefix(prefix, "list_api"),
		mcp.WithDescription("List available API paths for progressive discovery"),
		mcp.WithString("api", mcp.Description("API identifier (required when multiple APIs loaded)")),
		mcp.WithString("prefix", mcp.Description("Filter: only return paths starting with this string")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		apiName := req.GetString("api", "")
		prefix := req.GetString("prefix", "")

		entry, errResult, err := resolveAPI(reg, apiName)
		if err != nil || errResult != nil {
			return errResult, err
		}

		model, errs := entry.Document().BuildV3Model()
		if len(errs) > 0 {
			r, _ := toolErrorResult("INTERNAL_ERROR", "build model: "+errs[0].Error(), nil, nil)
			return r, nil
		}

		var paths []PathInfo
		if model.Model.Paths != nil && model.Model.Paths.PathItems != nil {
			for p, item := range model.Model.Paths.PathItems.FromOldest() {
				if prefix != "" && !hasPrefix(p, prefix) {
					continue
				}
				var methods []string
				if item.Get != nil {
					methods = append(methods, "GET")
				}
				if item.Post != nil {
					methods = append(methods, "POST")
				}
				if item.Put != nil {
					methods = append(methods, "PUT")
				}
				if item.Patch != nil {
					methods = append(methods, "PATCH")
				}
				if item.Delete != nil {
					methods = append(methods, "DELETE")
				}
				desc := ""
				if item.Get != nil && item.Get.Summary != "" {
					desc = item.Get.Summary
				} else if item.Post != nil && item.Post.Summary != "" {
					desc = item.Post.Summary
				}
				paths = append(paths, PathInfo{Path: p, Methods: methods, Description: desc})
			}
		}

		// Apply path allow-list filter (before prefix filter so both stack).
		if allowedPaths := entry.AllowList.Paths["list_api"]; len(allowedPaths) > 0 {
			filtered := paths[:0]
			for _, pi := range paths {
				if IsPathPermitted(pi.Path, allowedPaths) {
					filtered = append(filtered, pi)
				}
			}
			paths = filtered
		}

		sort.Slice(paths, func(i, j int) bool { return paths[i].Path < paths[j].Path })
		data, _ := json.Marshal(paths)
		return mcp.NewToolResultText(string(data)), nil
	})
}

func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return s[:len(prefix)] == prefix
}
