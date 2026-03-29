// Package tools implements the MCP tool handlers for the OpenAPI MCP server.
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/philippslang/progressive-api-mcp/pkg/httpclient"
	"github.com/philippslang/progressive-api-mcp/pkg/registry"
)

// ToolError is the envelope returned by tools on failure.
type ToolError struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []ValidationDetail `json:"details,omitempty"`
	Hints   []string           `json:"hints,omitempty"`
}

// ValidationDetail describes a single field-level violation.
type ValidationDetail struct {
	Type    string `json:"type"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// HTTPResult is returned on a successful HTTP call.
type HTTPResult struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       any               `json:"body"`
}

func toolErrorResult(code, message string, details []ValidationDetail, hints []string) (*mcp.CallToolResult, error) {
	te := ToolError{Code: code, Message: message, Details: details, Hints: hints}
	data, _ := json.Marshal(te)
	return mcp.NewToolResultText(string(data)), nil
}

func resolveAPI(reg *registry.Registry, apiName string) (registry.APIEntry, *mcp.CallToolResult, error) {
	if apiName == "" {
		if reg.Len() == 1 {
			names := reg.ListNames()
			entry, _ := reg.Lookup(names[0])
			return entry, nil, nil
		}
		if reg.Len() > 1 {
			r, _ := toolErrorResult("AMBIGUOUS_API",
				"multiple APIs are loaded; specify the 'api' parameter",
				nil, reg.ListNames())
			return registry.APIEntry{}, r, nil
		}
		r, _ := toolErrorResult("INTERNAL_ERROR", "no APIs loaded", nil, nil)
		return registry.APIEntry{}, r, nil
	}
	entry, ok := reg.Lookup(apiName)
	if !ok {
		r, _ := toolErrorResult("AMBIGUOUS_API",
			fmt.Sprintf("API %q not found", apiName),
			nil, reg.ListNames())
		return registry.APIEntry{}, r, nil
	}
	return entry, nil, nil
}

// matchPath tries to find the concrete path against OpenAPI path templates.
// Returns the matched template and all available paths.
func matchPath(entry registry.APIEntry, concretePath string) (string, []string) {
	model, errs := entry.Document().BuildV3Model()
	if len(errs) > 0 {
		return "", nil
	}
	paths := model.Model.Paths
	if paths == nil || paths.PathItems == nil {
		return "", nil
	}

	var allPaths []string
	for template, _ := range paths.PathItems.FromOldest() {
		allPaths = append(allPaths, template)
		if template == concretePath || matchesTemplate(template, concretePath) {
			return template, nil
		}
	}
	return "", allPaths
}

// matchesTemplate checks if concretePath matches an OpenAPI path template.
// e.g., /pets/42 matches /pets/{id}.
func matchesTemplate(template, concrete string) bool {
	tParts := strings.Split(strings.TrimPrefix(template, "/"), "/")
	cParts := strings.Split(strings.TrimPrefix(concrete, "/"), "/")
	if len(tParts) != len(cParts) {
		return false
	}
	for i, tp := range tParts {
		if strings.HasPrefix(tp, "{") && strings.HasSuffix(tp, "}") {
			continue
		}
		if tp != cParts[i] {
			return false
		}
	}
	return true
}

func executeHTTP(ctx context.Context, client *httpclient.Client, entry registry.APIEntry,
	method, path string, queryParams map[string]string, headers map[string]string, body any) (*mcp.CallToolResult, error) {

	fullURL := strings.TrimRight(entry.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	if len(queryParams) > 0 {
		q := url.Values{}
		for k, v := range queryParams {
			q.Set(k, v)
		}
		fullURL += "?" + q.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		var bodyBytes []byte
		switch b := body.(type) {
		case string:
			bodyBytes = []byte(b)
		default:
			var err error
			bodyBytes, err = json.Marshal(b)
			if err != nil {
				r, _ := toolErrorResult("INTERNAL_ERROR", "marshal request body: "+err.Error(), nil, nil)
				return r, nil
			}
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		r, _ := toolErrorResult("INTERNAL_ERROR", "build request: "+err.Error(), nil, nil)
		return r, nil
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		r, _ := toolErrorResult("HTTP_ERROR", "execute request: "+err.Error(), nil, nil)
		return r, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	respHeaders := make(map[string]string)
	for k, vs := range resp.Header {
		if len(vs) > 0 {
			respHeaders[k] = vs[0]
		}
	}

	var bodyVal any = string(respBody)
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		var jsonVal any
		if err := json.Unmarshal(respBody, &jsonVal); err == nil {
			bodyVal = jsonVal
		}
	}

	result := HTTPResult{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       bodyVal,
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

func validateAndExecute(ctx context.Context, reg *registry.Registry, client *httpclient.Client,
	method, apiName, path string, queryParams map[string]string, headers map[string]string, body any) (*mcp.CallToolResult, error) {

	entry, errResult, err := resolveAPI(reg, apiName)
	if err != nil || errResult != nil {
		return errResult, err
	}

	// Match path against OpenAPI templates.
	matched, allPaths := matchPath(entry, path)
	if matched == "" {
		hints := allPaths
		if len(hints) > 10 {
			hints = hints[:10]
		}
		r, _ := toolErrorResult("PATH_NOT_FOUND",
			fmt.Sprintf("path %q does not exist in API %q", path, entry.Name),
			nil, hints)
		return r, nil
	}

	// Check path allow list.
	if !IsPathPermitted(matched, entry.AllowList.Paths[method2tool(method)]) {
		return pathNotPermittedResult(matched, method2tool(method), entry.Name)
	}

	// Build a synthetic request for validation.
	var bodyReader io.Reader
	if body != nil {
		switch b := body.(type) {
		case string:
			bodyReader = strings.NewReader(b)
		default:
			bs, _ := json.Marshal(b)
			bodyReader = bytes.NewReader(bs)
		}
	}
	req, _ := http.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range queryParams {
		q := req.URL.Query()
		q.Set(k, v)
		req.URL.RawQuery = q.Encode()
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Inject placeholder values for ignored headers that the caller did not supply.
	// This prevents the validator from raising required-parameter errors for headers
	// that the caller is not expected to provide. The placeholder is never forwarded.
	for name := range entry.IgnoreHeaders {
		if req.Header.Get(name) == "" {
			req.Header.Set(name, "*")
		}
	}

	result := entry.Validator.Validate(req)
	if !result.Valid {
		details := make([]ValidationDetail, 0, len(result.Errors))
		for _, e := range result.Errors {
			details = append(details, ValidationDetail{Type: e.Type, Field: e.Field, Message: e.Message})
		}
		r, _ := toolErrorResult("VALIDATION_FAILED",
			fmt.Sprintf("request validation failed for %s %s", method, path),
			details, nil)
		return r, nil
	}

	return executeHTTP(ctx, client, entry, method, path, queryParams, headers, body)
}

// applyPrefix prepends prefix+"_" to name if prefix is non-empty (trailing "_" stripped first).
// Returns name unchanged when prefix is empty.
func applyPrefix(prefix, name string) string {
	p := strings.TrimRight(prefix, "_")
	if p == "" {
		return name
	}
	return p + "_" + name
}

// method2tool maps an HTTP method string to its MCP tool base name.
func method2tool(method string) string {
	switch method {
	case "GET":
		return "http_get"
	case "POST":
		return "http_post"
	case "PUT":
		return "http_put"
	case "PATCH":
		return "http_patch"
	default:
		return strings.ToLower("http_" + method)
	}
}

// IsPathPermitted returns true when allowed is nil/empty (allow-all) or template is in the list.
func IsPathPermitted(template string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, p := range allowed {
		if p == template {
			return true
		}
	}
	return false
}

// toolNotPermittedResult returns a PATH_NOT_PERMITTED ToolError result.
func toolNotPermittedResult(toolName, apiName string) (*mcp.CallToolResult, error) {
	return toolErrorResult("TOOL_NOT_PERMITTED",
		fmt.Sprintf("tool %q is not permitted for API %q", toolName, apiName),
		nil, nil)
}

// pathNotPermittedResult returns a PATH_NOT_PERMITTED ToolError result.
func pathNotPermittedResult(path, toolName, apiName string) (*mcp.CallToolResult, error) {
	return toolErrorResult("PATH_NOT_PERMITTED",
		fmt.Sprintf("path %q is not in the allow list for %s on API %q", path, toolName, apiName),
		nil, nil)
}

// RegisterHTTPTools registers the http_get, http_post, http_put, http_patch MCP tools.
// prefix is prepended to each tool name; empty means no prefix.
// allowedTools restricts which tools are registered; nil means all four are registered.
func RegisterHTTPTools(s *server.MCPServer, reg *registry.Registry, client *httpclient.Client, prefix string, allowedTools map[string]bool) {
	// http_get
	if allowedTools == nil || allowedTools["http_get"] {
		s.AddTool(mcp.NewTool(applyPrefix(prefix, "http_get"),
			mcp.WithDescription("Execute a validated HTTP GET request against a loaded API"),
			mcp.WithString("api", mcp.Description("API identifier (required when multiple APIs are loaded)")),
			mcp.WithString("path", mcp.Required(), mcp.Description("Concrete path with parameters substituted (e.g. /pets/42)")),
			mcp.WithObject("query_params", mcp.Description("Key-value pairs appended as query string parameters")),
			mcp.WithObject("headers", mcp.Description("Key-value pairs added as request headers")),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			apiName := req.GetString("api", "")
			path := req.GetString("path", "")
			queryParams := toStringMap(req.GetArguments()["query_params"])
			headers := toStringMap(req.GetArguments()["headers"])
			return validateAndExecute(ctx, reg, client, "GET", apiName, path, queryParams, headers, nil)
		})
	}

	// http_post
	if allowedTools == nil || allowedTools["http_post"] {
		s.AddTool(mcp.NewTool(applyPrefix(prefix, "http_post"),
			mcp.WithDescription("Execute a validated HTTP POST request against a loaded API"),
			mcp.WithString("api", mcp.Description("API identifier (required when multiple APIs are loaded)")),
			mcp.WithString("path", mcp.Required(), mcp.Description("Concrete path with parameters substituted")),
			mcp.WithObject("query_params", mcp.Description("Key-value pairs appended as query string parameters")),
			mcp.WithObject("headers", mcp.Description("Key-value pairs added as request headers")),
			mcp.WithObject("body", mcp.Description("Request body (must conform to the endpoint schema)")),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			apiName := req.GetString("api", "")
			path := req.GetString("path", "")
			queryParams := toStringMap(req.GetArguments()["query_params"])
			headers := toStringMap(req.GetArguments()["headers"])
			body := req.GetArguments()["body"]
			return validateAndExecute(ctx, reg, client, "POST", apiName, path, queryParams, headers, body)
		})
	}

	// http_put
	if allowedTools == nil || allowedTools["http_put"] {
		s.AddTool(mcp.NewTool(applyPrefix(prefix, "http_put"),
			mcp.WithDescription("Execute a validated HTTP PUT request against a loaded API"),
			mcp.WithString("api", mcp.Description("API identifier (required when multiple APIs are loaded)")),
			mcp.WithString("path", mcp.Required(), mcp.Description("Concrete path with parameters substituted")),
			mcp.WithObject("query_params", mcp.Description("Key-value pairs appended as query string parameters")),
			mcp.WithObject("headers", mcp.Description("Key-value pairs added as request headers")),
			mcp.WithObject("body", mcp.Description("Request body")),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			apiName := req.GetString("api", "")
			path := req.GetString("path", "")
			queryParams := toStringMap(req.GetArguments()["query_params"])
			headers := toStringMap(req.GetArguments()["headers"])
			body := req.GetArguments()["body"]
			return validateAndExecute(ctx, reg, client, "PUT", apiName, path, queryParams, headers, body)
		})
	}

	// http_patch
	if allowedTools == nil || allowedTools["http_patch"] {
		s.AddTool(mcp.NewTool(applyPrefix(prefix, "http_patch"),
			mcp.WithDescription("Execute a validated HTTP PATCH request against a loaded API"),
			mcp.WithString("api", mcp.Description("API identifier (required when multiple APIs are loaded)")),
			mcp.WithString("path", mcp.Required(), mcp.Description("Concrete path with parameters substituted")),
			mcp.WithObject("query_params", mcp.Description("Key-value pairs appended as query string parameters")),
			mcp.WithObject("headers", mcp.Description("Key-value pairs added as request headers")),
			mcp.WithObject("body", mcp.Description("Request body (partial update)")),
		), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			apiName := req.GetString("api", "")
			path := req.GetString("path", "")
			queryParams := toStringMap(req.GetArguments()["query_params"])
			headers := toStringMap(req.GetArguments()["headers"])
			body := req.GetArguments()["body"]
			return validateAndExecute(ctx, reg, client, "PATCH", apiName, path, queryParams, headers, body)
		})
	}
}

func toStringMap(v any) map[string]string {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, val := range m {
		result[k] = fmt.Sprint(val)
	}
	return result
}
