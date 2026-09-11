package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostHTMLCommandWritesStdout(t *testing.T) {
	t.Parallel()

	inputPath := writeMarkdownFixture(t, "## Hello\n\nWordPress post.\n")
	output := new(bytes.Buffer)
	command := newPostHTMLCommand()
	command.SetOut(output)
	command.SetArgs([]string{inputPath})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if got := output.String(); !strings.Contains(got, "<h2") || !strings.Contains(got, "<p>WordPress post.</p>") {
		t.Fatalf("output = %q", got)
	}
}

func TestPostHTMLCommandWritesFile(t *testing.T) {
	t.Parallel()

	inputPath := writeMarkdownFixture(t, "**content**\n")
	outputPath := filepath.Join(t.TempDir(), "post.html")
	command := newPostHTMLCommand()
	command.SetArgs([]string{inputPath, "--output", outputPath})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(): %v", err)
	}
	if expected := "<strong>content</strong>"; !strings.Contains(string(got), expected) {
		t.Fatalf("output file = %q, expected it to contain %q", got, expected)
	}
}

func TestPostHTMLCommandRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		inputPath  string
		outputPath string
		expected   string
	}{
		{
			name:       "non markdown extension",
			inputPath:  filepath.Join(t.TempDir(), "post.txt"),
			outputPath: "-",
			expected:   "input file must use a .md or .markdown extension",
		},
		{
			name:       "same input and output",
			inputPath:  writeMarkdownFixture(t, "content\n"),
			outputPath: "same",
			expected:   "output path must differ from input path",
		},
	}
	tests[1].outputPath = tests[1].inputPath

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			command := newPostHTMLCommand()
			command.SetArgs([]string{test.inputPath, "--output", test.outputPath})
			err := command.Execute()
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("Execute() error = %v, expected it to contain %q", err, test.expected)
			}
		})
	}
}

func writeMarkdownFixture(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "post.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}

	return path
}
