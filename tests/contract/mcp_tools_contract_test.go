package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/philippslang/progressive-api-mcp/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPResultShape verifies the HTTPResult JSON shape.
func TestHTTPResultShape(t *testing.T) {
	raw := `{"status_code":200,"headers":{"Content-Type":"application/json"},"body":{"id":1,"name":"Fido"}}`
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)

	assert.Contains(t, result, "status_code", "HTTPResult must have status_code")
	assert.Contains(t, result, "headers", "HTTPResult must have headers")
	assert.Contains(t, result, "body", "HTTPResult must have body")
	assert.Equal(t, float64(200), result["status_code"])
}

// TestToolErrorShape verifies the ToolError JSON shape.
func TestToolErrorShape(t *testing.T) {
	raw := `{"code":"VALIDATION_FAILED","message":"validation error","details":[{"type":"MISSING_REQUIRED_FIELD","field":"name","message":"Field name is required"}],"hints":[]}`
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)

	assert.Contains(t, result, "code", "ToolError must have code")
	assert.Contains(t, result, "message", "ToolError must have message")
	assert.Equal(t, "VALIDATION_FAILED", result["code"])
}

// TestToolErrorPathNotFound verifies PATH_NOT_FOUND ToolError shape.
func TestToolErrorPathNotFound(t *testing.T) {
	raw := `{"code":"PATH_NOT_FOUND","message":"path not found","hints":["/pets","/pets/{id}"]}`
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)

	assert.Equal(t, "PATH_NOT_FOUND", result["code"])
	hints, ok := result["hints"].([]any)
	assert.True(t, ok, "hints must be array")
	assert.Greater(t, len(hints), 0, "hints must have entries for PATH_NOT_FOUND")
}

// TestPathInfoShape verifies the PathInfo JSON shape used by explore_api.
func TestPathInfoShape(t *testing.T) {
	raw := `[{"path":"/pets","methods":["GET","POST"],"description":"List or create pets"}]`
	var result []map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)

	require.Len(t, result, 1)
	assert.Contains(t, result[0], "path")
	assert.Contains(t, result[0], "methods")
	assert.Contains(t, result[0], "description")
}

// TestAPIAllowListShape verifies the APIAllowList config struct shape.
func TestAPIAllowListShape(t *testing.T) {
	al := config.APIAllowList{
		Tools: []string{"http_get"},
		Paths: map[string][]string{"http_get": {"/pets"}},
	}
	assert.Equal(t, []string{"http_get"}, al.Tools)
	assert.Equal(t, []string{"/pets"}, al.Paths["http_get"])
}

// TestToolErrorPathNotPermitted verifies PATH_NOT_PERMITTED ToolError shape.
func TestToolErrorPathNotPermitted(t *testing.T) {
	raw := `{"code":"PATH_NOT_PERMITTED","message":"path \"/owners\" is not in the allow list for http_get on API \"petstore\""}`
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)
	assert.Equal(t, "PATH_NOT_PERMITTED", result["code"])
	assert.Contains(t, result["message"], "/owners")
}

// TestToolErrorToolNotPermitted verifies TOOL_NOT_PERMITTED ToolError shape.
func TestToolErrorToolNotPermitted(t *testing.T) {
	raw := `{"code":"TOOL_NOT_PERMITTED","message":"tool \"http_post\" is not permitted for API \"petstore\""}`
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)
	assert.Equal(t, "TOOL_NOT_PERMITTED", result["code"])
}

// TestSchemaResultShape verifies the SchemaResult JSON shape used by get_schema.
func TestSchemaResultShape(t *testing.T) {
	raw := `{"path":"/pets/{id}","method":"GET","path_parameters":{"id":{"type":"integer","required":true}}}`
	var result map[string]any
	err := json.Unmarshal([]byte(raw), &result)
	require.NoError(t, err)

	assert.Contains(t, result, "path")
	assert.Contains(t, result, "method")
	assert.Equal(t, "/pets/{id}", result["path"])
	assert.Equal(t, "GET", result["method"])
}
