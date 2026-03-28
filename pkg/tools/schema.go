package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/prograpimcp/prograpimcp/pkg/registry"
)

// SchemaResult is the full schema for one endpoint.
type SchemaResult struct {
	Path            string         `json:"path"`
	Method          string         `json:"method"`
	PathParameters  map[string]any `json:"path_parameters,omitempty"`
	QueryParameters map[string]any `json:"query_parameters,omitempty"`
	RequestBody     map[string]any `json:"request_body,omitempty"`
	Responses       map[string]any `json:"responses,omitempty"`
}

// RegisterSchemaTools registers the get_schema MCP tool.
func RegisterSchemaTools(s *server.MCPServer, reg *registry.Registry) {
	s.AddTool(mcp.NewTool("get_schema",
		mcp.WithDescription("Return the full schema for one endpoint"),
		mcp.WithString("api", mcp.Description("API identifier (required when multiple APIs loaded)")),
		mcp.WithString("path", mcp.Required(), mcp.Description("OpenAPI path template or concrete path (e.g. /pets/{id} or /pets/42)")),
		mcp.WithString("method", mcp.Required(), mcp.Description("HTTP method: GET, POST, PUT, PATCH")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		apiName := req.GetString("api", "")
		path := req.GetString("path", "")
		method := strings.ToUpper(req.GetString("method", ""))

		entry, errResult, err := resolveAPI(reg, apiName)
		if err != nil || errResult != nil {
			return errResult, err
		}

		// Resolve path (may be concrete or template).
		matched, _ := matchPath(entry, path)
		if matched == "" {
			// Try as a template directly.
			model, errs := entry.Document().BuildV3Model()
			if len(errs) == 0 && model.Model.Paths != nil && model.Model.Paths.PathItems != nil {
				for pi := range model.Model.Paths.PathItems.FromOldest() {
					if pi == path {
						matched = path
						break
					}
				}
			}
		}
		if matched == "" {
			r, _ := toolErrorResult("PATH_NOT_FOUND",
				fmt.Sprintf("path %q not found in API %q", path, entry.Name),
				nil, nil)
			return r, nil
		}

		model, errs := entry.Document().BuildV3Model()
		if len(errs) > 0 {
			r, _ := toolErrorResult("INTERNAL_ERROR", "build model: "+errs[0].Error(), nil, nil)
			return r, nil
		}

		pathItem, ok := model.Model.Paths.PathItems.Get(matched)
		if !ok {
			r, _ := toolErrorResult("PATH_NOT_FOUND", fmt.Sprintf("path %q not found", matched), nil, nil)
			return r, nil
		}

		var op *v3high.Operation
		switch method {
		case "GET":
			op = pathItem.Get
		case "POST":
			op = pathItem.Post
		case "PUT":
			op = pathItem.Put
		case "PATCH":
			op = pathItem.Patch
		case "DELETE":
			op = pathItem.Delete
		default:
			r, _ := toolErrorResult("PATH_NOT_FOUND", fmt.Sprintf("method %s not supported", method), nil, nil)
			return r, nil
		}

		if op == nil {
			r, _ := toolErrorResult("PATH_NOT_FOUND", fmt.Sprintf("%s %s is not defined", method, matched), nil, nil)
			return r, nil
		}

		result := SchemaResult{
			Path:   matched,
			Method: method,
		}

		// Extract parameters.
		if len(op.Parameters) > 0 {
			pathParams := make(map[string]any)
			queryParams := make(map[string]any)
			for _, param := range op.Parameters {
				paramInfo := map[string]any{
					"description": param.Description,
				}
				if param.Required != nil {
					paramInfo["required"] = *param.Required
				} else {
					paramInfo["required"] = false
				}
				if param.Schema != nil {
					schema := param.Schema.Schema()
					if schema != nil {
						if len(schema.Type) > 0 {
							paramInfo["type"] = schema.Type[0]
						}
						if schema.Format != "" {
							paramInfo["format"] = schema.Format
						}
					}
				}
				switch param.In {
				case "path":
					pathParams[param.Name] = paramInfo
				case "query":
					queryParams[param.Name] = paramInfo
				}
			}
			if len(pathParams) > 0 {
				result.PathParameters = pathParams
			}
			if len(queryParams) > 0 {
				result.QueryParameters = queryParams
			}
		}

		// Extract request body.
		if op.RequestBody != nil && op.RequestBody.Content != nil {
			bodySchema := map[string]any{}
			if op.RequestBody.Required != nil {
				bodySchema["required"] = *op.RequestBody.Required
			}
			for contentType, mediaType := range op.RequestBody.Content.FromOldest() {
				if strings.Contains(contentType, "json") && mediaType.Schema != nil {
					schema := mediaType.Schema.Schema()
					if schema != nil {
						bodySchema["type"] = schemaToMap(schema)
					}
				}
			}
			result.RequestBody = bodySchema
		}

		// Extract responses.
		if op.Responses != nil && op.Responses.Codes != nil {
			responses := make(map[string]any)
			for code, resp := range op.Responses.Codes.FromOldest() {
				respInfo := map[string]any{"description": resp.Description}
				if resp.Content != nil {
					for contentType, mediaType := range resp.Content.FromOldest() {
						if strings.Contains(contentType, "json") && mediaType.Schema != nil {
							schema := mediaType.Schema.Schema()
							if schema != nil {
								respInfo["schema"] = schemaToMap(schema)
							}
						}
					}
				}
				responses[code] = respInfo
			}
			result.Responses = responses
		}

		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	})
}

// schemaToMap converts a libopenapi Schema to a map[string]any for JSON output.
func schemaToMap(schema *base.Schema) map[string]any {
	m := map[string]any{}
	if len(schema.Type) > 0 {
		m["type"] = schema.Type[0]
	}
	if schema.Format != "" {
		m["format"] = schema.Format
	}
	return m
}
