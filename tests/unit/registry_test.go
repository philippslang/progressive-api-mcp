package unit_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/prograpimcp/prograpimcp/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../testdata", name)
}

func TestRegistry(t *testing.T) {
	t.Run("load valid API", func(t *testing.T) {
		reg := registry.New()
		err := reg.Load(config.APIConfig{
			Name:       "petstore",
			Definition: testdataPath("petstore.yaml"),
			Host:       "http://localhost:8080",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, reg.Len())
	})

	t.Run("lookup found", func(t *testing.T) {
		reg := registry.New()
		_ = reg.Load(config.APIConfig{Name: "petstore", Definition: testdataPath("petstore.yaml"), Host: "http://localhost:8080"})
		entry, ok := reg.Lookup("petstore")
		require.True(t, ok)
		assert.Equal(t, "petstore", entry.Name)
		assert.NotNil(t, entry.Validator)
		assert.Equal(t, "http://localhost:8080", entry.BaseURL)
	})

	t.Run("lookup not found", func(t *testing.T) {
		reg := registry.New()
		_, ok := reg.Lookup("nonexistent")
		assert.False(t, ok)
	})

	t.Run("load invalid file", func(t *testing.T) {
		reg := registry.New()
		err := reg.Load(config.APIConfig{Name: "bad", Definition: "/nonexistent/file.yaml"})
		require.Error(t, err)
	})

	t.Run("load malformed openapi", func(t *testing.T) {
		reg := registry.New()
		err := reg.Load(config.APIConfig{Name: "bad", Definition: testdataPath("malformed.yaml")})
		require.Error(t, err)
	})

	t.Run("list names", func(t *testing.T) {
		reg := registry.New()
		_ = reg.Load(config.APIConfig{Name: "petstore", Definition: testdataPath("petstore.yaml"), Host: "http://localhost:8080"})
		_ = reg.Load(config.APIConfig{Name: "bookstore", Definition: testdataPath("bookstore.yaml"), Host: "http://localhost:9090"})
		names := reg.ListNames()
		assert.Equal(t, []string{"petstore", "bookstore"}, names)
	})
}

func TestBaseURLResolution(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.APIConfig
		wantURL string
	}{
		{
			name:    "host and base_path",
			cfg:     config.APIConfig{Name: "api", Definition: testdataPath("petstore.yaml"), Host: "https://api.example.com", BasePath: "/v2"},
			wantURL: "https://api.example.com/v2",
		},
		{
			name:    "host only",
			cfg:     config.APIConfig{Name: "api", Definition: testdataPath("petstore.yaml"), Host: "https://api.example.com"},
			wantURL: "https://api.example.com",
		},
		{
			name:    "neither host nor base_path uses servers[0].url",
			cfg:     config.APIConfig{Name: "api", Definition: testdataPath("petstore.yaml")},
			wantURL: "http://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New()
			err := reg.Load(tt.cfg)
			require.NoError(t, err)
			entry, ok := reg.Lookup("api")
			require.True(t, ok)
			assert.Equal(t, tt.wantURL, entry.BaseURL)
		})
	}
}
