package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/prograpimcp/prograpimcp/pkg/openapimcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	require.NoError(t, ln.Close())
	return port
}

func TestHealthEndpoint(t *testing.T) {
	port := freePort(t)

	cfg := config.Config{
		Server: config.ServerConfig{
			Host:      "127.0.0.1",
			Port:      port,
			Transport: "http",
		},
		APIs: []config.APIConfig{
			{Name: "petstore", Definition: testdataPath("petstore.yaml")},
		},
	}

	srv, err := openapimcp.New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for the server to be ready.
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	var resp *http.Response
	require.Eventually(t, func() bool {
		r, e := http.Get(healthURL) //nolint:noctx
		if e != nil {
			return false
		}
		resp = r
		return true
	}, 5*time.Second, 20*time.Millisecond, "server did not become ready")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])

	// Shut down the server.
	cancel()
	select {
	case e := <-errCh:
		assert.NoError(t, e)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
