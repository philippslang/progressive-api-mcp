package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type errorBody struct {
	Message string `json:"message"`
}

func alwaysError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(errorBody{Message: "upstream error"})
}

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "malformed-mock",
		Short: "Mock server that always returns HTTP 500 (simulates a broken upstream API)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := http.NewServeMux()
			mux.HandleFunc("/", alwaysError)

			ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			srv := &http.Server{Handler: mux}
			go func() { _ = srv.Serve(ln) }()
			fmt.Printf("Listening on http://%s\n", ln.Addr())

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return srv.Shutdown(ctx)
		},
	}

	cmd.Flags().IntVar(&port, "port", 9999, "TCP port to listen on (0 = OS-assigned)")
	return cmd
}
