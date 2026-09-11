// mcp-wp-go is a stdio MCP server for a single WordPress REST API site.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-wp-go/internal/config"
	"mcp-wp-go/internal/server"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("mcp-wp-go: ")

	envFile := flag.String("env-file", "", "optional path to a KEY=VALUE environment file")
	flag.Parse()
	if *envFile != "" {
		if err := config.LoadEnvFile(*envFile); err != nil {
			log.Fatalf("load env file: %v", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration: %v", err)
	}
	if err := server.New(cfg).Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, "mcp-wp-go: server stopped:", err)
	}
}
