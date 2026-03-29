package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/philippslang/progressive-api-mcp/pkg/config"
	"github.com/philippslang/progressive-api-mcp/pkg/openapimcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogging(t *testing.T) {
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

	// Redirect stderr so we can capture log output.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	// Wait for the server to be ready.
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	require.Eventually(t, func() bool {
		resp, e := http.Get(healthURL) //nolint:noctx
		if e != nil {
			return false
		}
		resp.Body.Close()
		return true
	}, 5*time.Second, 20*time.Millisecond, "server did not become ready")

	// Send a second request to generate a clean log line after startup noise.
	resp, err := http.Get(healthURL) //nolint:noctx
	require.NoError(t, err)
	resp.Body.Close()

	// Shut down and restore stderr before reading captured output.
	cancel()
	select {
	case e := <-errCh:
		assert.NoError(t, e)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down in time")
	}

	w.Close()
	os.Stderr = origStderr
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()

	// At least one log line for GET /health should be present.
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var healthLogLines []string
	for _, line := range lines {
		if strings.Contains(line, "GET /health") && strings.Contains(line, "200") {
			healthLogLines = append(healthLogLines, line)
		}
	}
	require.NotEmpty(t, healthLogLines, "expected at least one log line for GET /health 200; got output:\n%s", output)

	// Verify the log line contains method, path, status code, and duration.
	line := healthLogLines[len(healthLogLines)-1]
	assert.Contains(t, line, "[prograpimcp]")
	assert.Contains(t, line, "GET")
	assert.Contains(t, line, "/health")
	assert.Contains(t, line, "200")
	assert.Contains(t, line, "ms")
}
