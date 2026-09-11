package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"mcp-wp-go/internal/posthtml"
)

func newPostHTMLCommand() *cobra.Command {
	var outputPath string

	command := &cobra.Command{
		Use:     "post-html FILE.md",
		Aliases: []string{"markdown", "md"},
		Short:   "Convert a Markdown file to WordPress post HTML",
		Args:    cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return convertPostHTML(command, args[0], outputPath)
		},
		ValidArgsFunction: func(
			_ *cobra.Command,
			_ []string,
			_ string,
		) ([]string, cobra.ShellCompDirective) {
			return []string{"md", "markdown"}, cobra.ShellCompDirectiveFilterFileExt
		},
	}
	command.Flags().StringVarP(
		&outputPath,
		"output",
		"o",
		"-",
		"output file, or - for stdout",
	)

	return command
}

func convertPostHTML(command *cobra.Command, inputPath, outputPath string) error {
	extension := strings.ToLower(filepath.Ext(inputPath))
	if extension != ".md" && extension != ".markdown" {
		return fmt.Errorf("input file must use a .md or .markdown extension: %q", inputPath)
	}
	if outputPath != "-" && samePath(inputPath, outputPath) {
		return fmt.Errorf("output path must differ from input path: %q", outputPath)
	}

	markdownSource, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read markdown file %q: %w", inputPath, err)
	}

	postHTML := posthtml.FromMarkdown(markdownSource)
	if outputPath == "-" {
		if _, err := command.OutOrStdout().Write(postHTML); err != nil {
			return fmt.Errorf("write html to stdout: %w", err)
		}
		return nil
	}

	if err := os.WriteFile(outputPath, postHTML, 0o644); err != nil {
		return fmt.Errorf("write html file %q: %w", outputPath, err)
	}

	return nil
}

func samePath(first, second string) bool {
	firstPath, firstErr := filepath.Abs(first)
	secondPath, secondErr := filepath.Abs(second)
	if firstErr != nil || secondErr != nil {
		return filepath.Clean(first) == filepath.Clean(second)
	}

	return firstPath == secondPath
}
