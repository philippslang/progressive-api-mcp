// Package openapimcp is the primary entry point for the OpenAPI MCP server library.
package openapimcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/prograpimcp/prograpimcp/pkg/httpclient"
	"github.com/prograpimcp/prograpimcp/pkg/registry"
	"github.com/prograpimcp/prograpimcp/pkg/tools"
)

// Server is the MCP server instance. Create with New(); start with Start().
type Server struct {
	cfg      config.Config
	registry *registry.Registry
	mcpSrv   *server.MCPServer
	client   *httpclient.Client
	stopFn   func()
}

// New creates a new Server from the given Config.
// It validates the config but does NOT load API definitions yet.
func New(cfg config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &Server{cfg: cfg}, nil
}

// Start loads all API definitions, registers MCP tools, and starts the transport.
// Blocks until ctx is cancelled or Stop() is called.
// Returns an error if any API definition fails to load or the transport fails to bind.
func (s *Server) Start(ctx context.Context) error {
	s.registry = registry.New()
	var loadErrors []string
	for _, apiCfg := range s.cfg.APIs {
		if err := s.registry.Load(apiCfg); err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("  - %s: %v", apiCfg.Name, err))
		}
	}
	if len(loadErrors) > 0 {
		return fmt.Errorf("failed to load API definitions:\n%s",
			strings.Join(loadErrors, "\n"))
	}

	s.client = httpclient.New(30 * time.Second)

	s.mcpSrv = server.NewMCPServer("OpenAPI MCP Server", "1.0.0",
		server.WithToolCapabilities(true),
	)

	effectivePrefix := strings.TrimRight(s.cfg.Server.ToolPrefix, "_")
	allowedTools := computeAllowedTools(s.cfg.APIs)
	if effectivePrefix != "" {
		fmt.Fprintf(os.Stderr, "[prograpimcp] tool prefix: %s\n", effectivePrefix)
	} else {
		fmt.Fprintf(os.Stderr, "[prograpimcp] tool prefix: none\n")
	}

	tools.RegisterHTTPTools(s.mcpSrv, s.registry, s.client, effectivePrefix, allowedTools)
	tools.RegisterExploreTools(s.mcpSrv, s.registry, effectivePrefix, allowedTools)
	tools.RegisterSchemaTools(s.mcpSrv, s.registry, effectivePrefix, allowedTools)

	transport := s.cfg.Server.Transport
	if transport == "" {
		transport = "http"
	}

	ctx2, cancel := context.WithCancel(ctx)
	s.stopFn = cancel
	defer cancel()

	switch transport {
	case "stdio":
		stdioSrv := server.NewStdioServer(s.mcpSrv)
		return stdioSrv.Listen(ctx2, os.Stdin, os.Stdout)
	case "http":
		host := s.cfg.Server.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := s.cfg.Server.Port
		if port == 0 {
			port = 8080
		}
		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Fprintf(os.Stderr, "[prograpimcp] MCP endpoint:    http://%s/mcp\n", addr)
		fmt.Fprintf(os.Stderr, "[prograpimcp] health endpoint: http://%s/health\n", addr)
		httpSrv := server.NewStreamableHTTPServer(s.mcpSrv)

		mux := http.NewServeMux()
		mux.Handle("/mcp", httpSrv)
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"status":"ok"}`)
		})
		httpServer := &http.Server{Addr: addr, Handler: mux}

		// Start the server in a goroutine so we can watch ctx.
		errCh := make(chan error, 1)
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
			} else {
				errCh <- nil
			}
		}()

		select {
		case <-ctx2.Done():
			shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutCancel()
			return httpServer.Shutdown(shutCtx)
		case err := <-errCh:
			return err
		}
	default:
		return fmt.Errorf("unsupported transport %q", transport)
	}
}

// computeAllowedTools returns a union map of allowed tool names across all API configs.
// Returns nil when all APIs have empty AllowList.Tools (allow-all semantics).
func computeAllowedTools(apis []config.APIConfig) map[string]bool {
	anyRestricted := false
	for _, api := range apis {
		if len(api.AllowList.Tools) > 0 {
			anyRestricted = true
			break
		}
	}
	if !anyRestricted {
		return nil
	}
	union := make(map[string]bool)
	for _, api := range apis {
		for _, name := range api.AllowList.Tools {
			union[name] = true
		}
	}
	return union
}

// Stop signals the server to shut down gracefully.
func (s *Server) Stop() error {
	if s.stopFn != nil {
		s.stopFn()
	}
	return nil
}

// APIs returns the names of all successfully loaded API definitions.
// Only valid after Start() has returned without error.
func (s *Server) APIs() []string {
	if s.registry == nil {
		return nil
	}
	return s.registry.ListNames()
}
