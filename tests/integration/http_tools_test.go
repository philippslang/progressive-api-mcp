package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/prograpimcp/prograpimcp/pkg/httpclient"
	"github.com/prograpimcp/prograpimcp/pkg/openapimcp"
	"github.com/prograpimcp/prograpimcp/pkg/registry"
	"github.com/prograpimcp/prograpimcp/pkg/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../testdata", name)
}

// makeTestClient builds an MCPServer with the given registry, wraps it in an
// InProcessTransport, initializes a client, and returns the client.
func makeTestClient(t *testing.T, reg *registry.Registry, httpClient *httpclient.Client) *client.Client {
	t.Helper()
	return makeTestClientWithPrefix(t, "", reg, httpClient)
}

// makeTestClientWithPrefix builds an MCPServer with the given registry and tool prefix.
func makeTestClientWithPrefix(t *testing.T, prefix string, reg *registry.Registry, httpClient *httpclient.Client) *client.Client {
	t.Helper()
	srv := mcpserver.NewMCPServer("test", "1.0.0", mcpserver.WithToolCapabilities(true))
	tools.RegisterHTTPTools(srv, reg, httpClient, prefix)
	tools.RegisterExploreTools(srv, reg, prefix)
	tools.RegisterSchemaTools(srv, reg, prefix)

	tr := transport.NewInProcessTransport(srv)
	c := client.NewClient(tr)
	ctx := context.Background()
	require.NoError(t, c.Start(ctx))
	_, err := c.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// callToolRaw invokes a named tool and returns the raw result and error.
func callToolRaw(t *testing.T, c *client.Client, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args
	return c.CallTool(context.Background(), req)
}

// callTool is a helper to invoke a named tool and return the result text.
func callTool(t *testing.T, c *client.Client, toolName string, args map[string]any) string {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args
	result, err := c.CallTool(context.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, result.Content)
	first := result.Content[0]
	tc, ok := first.(mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", first)
	return tc.Text
}

// TestOpenAPIMCPServerLoad tests that New accepts valid config and that invalid config is rejected.
func TestOpenAPIMCPServerLoad(t *testing.T) {
	t.Run("valid config creates server", func(t *testing.T) {
		cfg := config.Config{
			Server: config.ServerConfig{Transport: "http", Host: "127.0.0.1", Port: 8080},
			APIs: []config.APIConfig{
				{Name: "petstore", Definition: testdataPath("petstore.yaml"), Host: "http://localhost:8080"},
			},
		}
		srv, err := openapimcp.New(cfg)
		require.NoError(t, err)
		assert.NotNil(t, srv)
	})

	t.Run("invalid config returns error", func(t *testing.T) {
		cfg := config.Config{
			Server: config.ServerConfig{Transport: "invalid"},
			APIs:   []config.APIConfig{{Name: "x", Definition: "./x.yaml"}},
		}
		_, err := openapimcp.New(cfg)
		require.Error(t, err)
	})
}

// TestHTTPToolsEndToEnd exercises all HTTP tools end-to-end via in-process transport.
func TestHTTPToolsEndToEnd(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "GET" && r.URL.Path == "/pets":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 1, "name": "Fido", "species": "dog"},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Fido", "species": "dog"})
		}
	}))
	defer target.Close()

	reg := registry.New()
	require.NoError(t, reg.Load(config.APIConfig{
		Name:       "petstore",
		Definition: testdataPath("petstore.yaml"),
		Host:       target.URL,
	}))

	httpClient := httpclient.New(10 * time.Second)
	c := makeTestClient(t, reg, httpClient)

	t.Run("http_get valid path returns HTTPResult", func(t *testing.T) {
		text := callTool(t, c, "http_get", map[string]any{"path": "/pets"})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, float64(200), result["status_code"])
		assert.Contains(t, result, "body")
		assert.Contains(t, result, "headers")
	})

	t.Run("http_get path not found returns PATH_NOT_FOUND", func(t *testing.T) {
		text := callTool(t, c, "http_get", map[string]any{"path": "/nonexistent"})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "PATH_NOT_FOUND", result["code"])
		_, hasHints := result["hints"]
		assert.True(t, hasHints)
	})

	t.Run("http_get with path param", func(t *testing.T) {
		text := callTool(t, c, "http_get", map[string]any{"path": "/pets/42"})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, float64(200), result["status_code"])
	})

	t.Run("http_post missing required field returns VALIDATION_FAILED", func(t *testing.T) {
		text := callTool(t, c, "http_post", map[string]any{
			"path": "/pets",
			"body": map[string]any{"species": "dog"}, // missing 'name'
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "VALIDATION_FAILED", result["code"])
		details, ok := result["details"].([]any)
		require.True(t, ok, "details should be an array")
		require.NotEmpty(t, details)
	})

	t.Run("http_post valid body executes and returns HTTPResult", func(t *testing.T) {
		text := callTool(t, c, "http_post", map[string]any{
			"path": "/pets",
			"body": map[string]any{"name": "Fido", "species": "dog"},
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, float64(200), result["status_code"])
	})

	t.Run("http_put valid body", func(t *testing.T) {
		text := callTool(t, c, "http_put", map[string]any{
			"path": "/pets/1",
			"body": map[string]any{"name": "Rex", "species": "cat"},
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, float64(200), result["status_code"])
	})

	t.Run("http_patch valid body", func(t *testing.T) {
		text := callTool(t, c, "http_patch", map[string]any{
			"path": "/pets/1",
			"body": map[string]any{"name": "Rex"},
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, float64(200), result["status_code"])
	})
}

// TestExploreAndSchemaTools exercises the explore_api and get_schema tools.
func TestExploreAndSchemaTools(t *testing.T) {
	reg := registry.New()
	require.NoError(t, reg.Load(config.APIConfig{
		Name:       "petstore",
		Definition: testdataPath("petstore.yaml"),
		Host:       "http://localhost:8080",
	}))

	httpClient := httpclient.New(10 * time.Second)
	c := makeTestClient(t, reg, httpClient)

	t.Run("explore_api returns all paths sorted", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.Greater(t, len(paths), 0)
		// Check sorted.
		for i := 1; i < len(paths); i++ {
			assert.LessOrEqual(t, paths[i-1]["path"].(string), paths[i]["path"].(string))
		}
	})

	t.Run("explore_api with /pets prefix filters", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{"prefix": "/pets"})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		for _, p := range paths {
			path := p["path"].(string)
			assert.True(t, len(path) >= 5 && path[:5] == "/pets", "path should start with /pets, got %s", path)
		}
	})

	t.Run("explore_api with /owners prefix filters", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{"prefix": "/owners"})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.NotEmpty(t, paths)
		for _, p := range paths {
			path := p["path"].(string)
			assert.True(t, len(path) >= 7 && path[:7] == "/owners", "expected /owners prefix, got %s", path)
		}
	})

	t.Run("explore_api prefix matches nothing returns empty array", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{"prefix": "/nonexistent"})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.Empty(t, paths)
	})

	t.Run("get_schema for GET /pets/{id}", func(t *testing.T) {
		text := callTool(t, c, "get_schema", map[string]any{
			"path":   "/pets/{id}",
			"method": "GET",
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "/pets/{id}", result["path"])
		assert.Equal(t, "GET", result["method"])
	})

	t.Run("get_schema concrete path resolves to template", func(t *testing.T) {
		text := callTool(t, c, "get_schema", map[string]any{
			"path":   "/pets/42",
			"method": "GET",
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "/pets/{id}", result["path"])
	})

	t.Run("get_schema unknown path returns PATH_NOT_FOUND", func(t *testing.T) {
		text := callTool(t, c, "get_schema", map[string]any{
			"path":   "/nonexistent",
			"method": "GET",
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "PATH_NOT_FOUND", result["code"])
	})

	t.Run("get_schema unknown method returns PATH_NOT_FOUND", func(t *testing.T) {
		text := callTool(t, c, "get_schema", map[string]any{
			"path":   "/pets",
			"method": "DELETE",
		})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "PATH_NOT_FOUND", result["code"])
	})
}

// TestMultiAPIAmbiguity verifies multi-API ambiguity handling.
func TestMultiAPIAmbiguity(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer target.Close()

	reg := registry.New()
	require.NoError(t, reg.Load(config.APIConfig{
		Name: "petstore", Definition: testdataPath("petstore.yaml"), Host: target.URL,
	}))
	require.NoError(t, reg.Load(config.APIConfig{
		Name: "bookstore", Definition: testdataPath("bookstore.yaml"), Host: target.URL,
	}))

	httpClient := httpclient.New(10 * time.Second)
	c := makeTestClient(t, reg, httpClient)

	t.Run("omitting api with multiple APIs returns AMBIGUOUS_API", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{})
		var result map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &result))
		assert.Equal(t, "AMBIGUOUS_API", result["code"])
		hints, ok := result["hints"].([]any)
		require.True(t, ok)
		assert.Contains(t, hints, "petstore")
		assert.Contains(t, hints, "bookstore")
	})

	t.Run("specifying api=petstore targets petstore paths", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{"api": "petstore"})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.NotEmpty(t, paths)
		for _, p := range paths {
			path := p["path"].(string)
			assert.True(t, len(path) >= 5 && (path[:5] == "/pets" || path[:7] == "/owners"),
				"expected petstore paths, got %s", path)
		}
	})

	t.Run("specifying api=bookstore targets bookstore paths", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{"api": "bookstore"})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.NotEmpty(t, paths)
		for _, p := range paths {
			path := p["path"].(string)
			assert.True(t, len(path) >= 6 && path[:6] == "/books",
				"expected bookstore paths, got %s", path)
		}
	})
}

// TestToolPrefix verifies that tools are registered with a prefix when ToolPrefix is set.
func TestToolPrefix(t *testing.T) {
	reg := registry.New()
	require.NoError(t, reg.Load(config.APIConfig{
		Name:       "petstore",
		Definition: testdataPath("petstore.yaml"),
		Host:       "http://localhost:8080",
	}))
	httpClient := httpclient.New(10 * time.Second)
	c := makeTestClientWithPrefix(t, "test", reg, httpClient)

	t.Run("test_explore_api returns path list", func(t *testing.T) {
		text := callTool(t, c, "test_explore_api", map[string]any{})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.Greater(t, len(paths), 0)
	})

	t.Run("unprefixed explore_api returns tool-not-found error", func(t *testing.T) {
		_, err := callToolRaw(t, c, "explore_api", map[string]any{})
		assert.Error(t, err)
	})
}

// TestNoPrefixDefaultBehavior verifies that empty ToolPrefix leaves original tool names intact.
func TestNoPrefixDefaultBehavior(t *testing.T) {
	reg := registry.New()
	require.NoError(t, reg.Load(config.APIConfig{
		Name:       "petstore",
		Definition: testdataPath("petstore.yaml"),
		Host:       "http://localhost:8080",
	}))
	httpClient := httpclient.New(10 * time.Second)
	c := makeTestClientWithPrefix(t, "", reg, httpClient)

	t.Run("explore_api works with empty prefix", func(t *testing.T) {
		text := callTool(t, c, "explore_api", map[string]any{})
		var paths []map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &paths))
		assert.Greater(t, len(paths), 0)
	})

	t.Run("http_get works with empty prefix", func(t *testing.T) {
		text := callTool(t, c, "http_get", map[string]any{"path": "/pets"})
		assert.NotEmpty(t, text)
	})
}

// TestTrailingUnderscoreStripped verifies that a trailing underscore in ToolPrefix is stripped,
// so tools are registered as "myapi_http_get" not "myapi__http_get".
func TestTrailingUnderscoreStripped(t *testing.T) {
	reg := registry.New()
	require.NoError(t, reg.Load(config.APIConfig{
		Name:       "petstore",
		Definition: testdataPath("petstore.yaml"),
		Host:       "http://localhost:8080",
	}))
	httpClient := httpclient.New(10 * time.Second)
	c := makeTestClientWithPrefix(t, "myapi_", reg, httpClient)

	t.Run("myapi_http_get works (single underscore)", func(t *testing.T) {
		text := callTool(t, c, "myapi_http_get", map[string]any{"path": "/pets"})
		assert.NotEmpty(t, text)
	})

	t.Run("myapi__http_get does not exist (double underscore)", func(t *testing.T) {
		_, err := callToolRaw(t, c, "myapi__http_get", map[string]any{"path": "/pets"})
		assert.Error(t, err)
	})
}

// TestMalformedDefinitionFailsStartup verifies startup abort on bad definition.
func TestMalformedDefinitionFailsStartup(t *testing.T) {
	cfg := config.Config{
		Server: config.ServerConfig{Transport: "stdio"},
		APIs: []config.APIConfig{
			{Name: "bad", Definition: testdataPath("malformed.yaml")},
		},
	}
	srv, err := openapimcp.New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load API definitions")
}
