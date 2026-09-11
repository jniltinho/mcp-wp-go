// mcp-wp-go is a WordPress REST API MCP server and content utility CLI.
package main

import (
	"context"
	"fmt"
	"os"

	"mcp-wp-go/cmd"
)

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "mcp-wp-go:", err)
		os.Exit(1)
	}
}
