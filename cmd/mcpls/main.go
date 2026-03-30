// mcpls connects to an MCP server, lists all registered tool names, and calls
// http_get with path /pets as a demonstration.
//
// Usage: mcpls <mcp-endpoint-url>
//
// Example: mcpls http://127.0.0.1:8000/mcp
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: mcpls <mcp-endpoint-url>")
		os.Exit(1)
	}
	url := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.NewStreamableHttpClient(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcpls: %v\n", err)
		os.Exit(1)
	}

	if err := c.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "mcpls: cannot connect to %s: %v\n", url, err)
		os.Exit(1)
	}
	defer c.Close()

	_, err = c.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcpls",
				Version: "1.0.0",
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcpls: MCP initialization failed: %v\n", err)
		os.Exit(1)
	}

	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcpls: failed to list tools: %v\n", err)
		os.Exit(1)
	}

	for _, tool := range result.Tools {
		fmt.Println(tool.Name)
	}

	// Call http_get with path /pets as a demonstration.
	getResult, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_schema",
			Arguments: map[string]any{
				"method": "POST",
				"path":   "/pets",
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcpls: http_get failed: %v\n", err)
		os.Exit(1)
	}

	for _, content := range getResult.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			var pretty any
			if json.Unmarshal([]byte(tc.Text), &pretty) == nil {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				enc.Encode(pretty)
			} else {
				fmt.Println(tc.Text)
			}
		}
	}
}
