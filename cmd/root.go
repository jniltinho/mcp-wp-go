// Package cmd defines the mcp-wp-go command tree.
package cmd

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"mcp-wp-go/internal/config"
	"mcp-wp-go/internal/server"
)

type serverRunner func(context.Context, string) error

// Execute runs the mcp-wp-go command tree.
func Execute(ctx context.Context) error {
	return NewRootCommand().ExecuteContext(ctx)
}

// NewRootCommand builds a fresh root command.
func NewRootCommand() *cobra.Command {
	return newRootCommand(runServer)
}

func newRootCommand(run serverRunner) *cobra.Command {
	var envFile string

	command := &cobra.Command{
		Use:           "mcp-wp-go",
		Short:         "Manage WordPress through MCP and prepare post content",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return run(command.Context(), envFile)
		},
	}
	command.PersistentFlags().StringVar(
		&envFile,
		"env-file",
		"",
		"optional path to a KEY=VALUE environment file",
	)
	command.AddCommand(newPostHTMLCommand())

	return command
}

func runServer(ctx context.Context, envFile string) error {
	if envFile != "" {
		if err := config.LoadEnvFile(envFile); err != nil {
			return fmt.Errorf("load env file: %w", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	if err := server.New(cfg).Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}

	return nil
}
