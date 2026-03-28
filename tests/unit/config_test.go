package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid single API",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 8080, Transport: "http"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
		{
			name: "empty APIs list",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 8080, Transport: "http"},
				APIs:   []config.APIConfig{},
			},
			wantErr: true,
			errMsg:  "at least one API",
		},
		{
			name: "invalid transport",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 8080, Transport: "websocket"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "transport",
		},
		{
			name: "port out of range",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 99999, Transport: "http"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "duplicate API names",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 8080, Transport: "http"},
				APIs: []config.APIConfig{
					{Name: "petstore", Definition: "./petstore.yaml"},
					{Name: "petstore", Definition: "./other.yaml"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate",
		},
		{
			name: "missing API name",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 8080, Transport: "http"},
				APIs:   []config.APIConfig{{Name: "", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "missing definition",
			cfg: config.Config{
				Server: config.ServerConfig{Host: "127.0.0.1", Port: 8080, Transport: "http"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: ""}},
			},
			wantErr: true,
			errMsg:  "definition",
		},
		{
			name: "stdio transport valid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
		// ToolPrefix validation
		{
			name: "tool_prefix valid simple",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "myapi"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
		{
			name: "tool_prefix with trailing underscore stripped silently",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "myapi_"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
		{
			name: "tool_prefix empty string is valid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: ""},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
		{
			name: "tool_prefix starts with digit is invalid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "123abc"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "tool_prefix",
		},
		{
			name: "tool_prefix purely numeric is invalid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "123"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "tool_prefix",
		},
		{
			name: "tool_prefix contains hyphen is invalid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "my-api"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "tool_prefix",
		},
		{
			name: "tool_prefix contains space is invalid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "my api"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: true,
			errMsg:  "tool_prefix",
		},
		{
			name: "tool_prefix starts with letter+digits is valid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "v2svc"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
		{
			name: "tool_prefix starts with underscore is valid",
			cfg: config.Config{
				Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "_internal"},
				APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestToolPrefixPrecedence(t *testing.T) {
	// Simulates CLI flag overriding config-file value: after loading config with
	// ToolPrefix "fromfile", the CLI wiring sets it to "fromcli".
	cfg := config.Config{
		Server: config.ServerConfig{Transport: "stdio", ToolPrefix: "fromfile"},
		APIs:   []config.APIConfig{{Name: "petstore", Definition: "./petstore.yaml"}},
	}
	require.NoError(t, cfg.Validate())
	assert.Equal(t, "fromfile", cfg.Server.ToolPrefix)

	// Simulate CLI override (equivalent of viper.GetString("server.tool_prefix") after flag set).
	cfg.Server.ToolPrefix = "fromcli"
	assert.Equal(t, "fromcli", cfg.Server.ToolPrefix)
	require.NoError(t, cfg.Validate())
}

func TestLoadFile(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		_, err := config.LoadFile("/nonexistent/path/config.yaml")
		require.Error(t, err)
	})

	t.Run("valid YAML file", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		content := `
server:
  host: "127.0.0.1"
  port: 8080
  transport: "http"
apis:
  - name: petstore
    definition: "./petstore.yaml"
`
		err := os.WriteFile(cfgPath, []byte(content), 0600)
		require.NoError(t, err)

		cfg, err := config.LoadFile(cfgPath)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "http", cfg.Server.Transport)
		require.Len(t, cfg.APIs, 1)
		assert.Equal(t, "petstore", cfg.APIs[0].Name)
	})

	t.Run("malformed YAML", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		err := os.WriteFile(cfgPath, []byte(":\t invalid yaml {{{"), 0600)
		require.NoError(t, err)
		_, err = config.LoadFile(cfgPath)
		require.Error(t, err)
	})

	t.Run("invalid config", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yaml")
		content := `
server:
  transport: "http"
apis: []
`
		err := os.WriteFile(cfgPath, []byte(content), 0600)
		require.NoError(t, err)
		_, err = config.LoadFile(cfgPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one API")
	})
}
