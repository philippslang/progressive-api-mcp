package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/prograpimcp/prograpimcp/pkg/openapimcp"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var cfgFile string

	cmd := &cobra.Command{
		Use:   "prograpimcp",
		Short: "OpenAPI MCP Server — expose OpenAPI definitions as MCP tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cfgFile)
			if err != nil {
				return err
			}

			// CLI flags override config file values.
			if cmd.Flags().Changed("host") {
				cfg.Server.Host = viper.GetString("server.host")
			}
			if cmd.Flags().Changed("port") {
				cfg.Server.Port = viper.GetInt("server.port")
			}
			if cmd.Flags().Changed("transport") {
				cfg.Server.Transport = viper.GetString("server.transport")
			}
			if cmd.Flags().Changed("tool-prefix") {
				cfg.Server.ToolPrefix = viper.GetString("server.tool_prefix")
			}

			srv, err := openapimcp.New(cfg)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
			}()

			fmt.Fprintf(os.Stderr, "[prograpimcp] starting server (transport: %s)\n", cfg.Server.Transport)
			return srv.Start(ctx)
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file path")
	cmd.Flags().String("host", "", "MCP server host (overrides config)")
	cmd.Flags().Int("port", 0, "MCP server port (overrides config)")
	cmd.Flags().String("transport", "", "MCP transport: http or stdio (overrides config)")
	cmd.Flags().String("tool-prefix", "", "MCP tool name prefix (overrides config)")

	viper.SetEnvPrefix("PROGRAPIMCP")
	viper.AutomaticEnv()
	if err := viper.BindPFlag("server.host", cmd.Flags().Lookup("host")); err != nil {
		fmt.Fprintf(os.Stderr, "warn: bind host flag: %v\n", err)
	}
	if err := viper.BindPFlag("server.port", cmd.Flags().Lookup("port")); err != nil {
		fmt.Fprintf(os.Stderr, "warn: bind port flag: %v\n", err)
	}
	if err := viper.BindPFlag("server.transport", cmd.Flags().Lookup("transport")); err != nil {
		fmt.Fprintf(os.Stderr, "warn: bind transport flag: %v\n", err)
	}
	if err := viper.BindPFlag("server.tool_prefix", cmd.Flags().Lookup("tool-prefix")); err != nil {
		fmt.Fprintf(os.Stderr, "warn: bind tool-prefix flag: %v\n", err)
	}

	return cmd
}

func loadConfig(cfgFile string) (config.Config, error) {
	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config.Config{}, fmt.Errorf("config file %q not found; create one or use --config", cfgFile)
		}
		return config.Config{}, err
	}
	return cfg, nil
}
