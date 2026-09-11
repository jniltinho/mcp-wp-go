package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRootCommandKeepsDefaultServerBehavior(t *testing.T) {
	t.Parallel()

	var receivedEnvFile string
	root := newRootCommand(func(_ context.Context, envFile string) error {
		receivedEnvFile = envFile
		return nil
	})
	root.SetArgs([]string{"--env-file", "wordpress.env"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if receivedEnvFile != "wordpress.env" {
		t.Fatalf("env file = %q", receivedEnvFile)
	}
}

func TestRootCommandHelpListsPostHTML(t *testing.T) {
	t.Parallel()

	output := new(bytes.Buffer)
	root := newRootCommand(func(context.Context, string) error {
		t.Fatal("server runner called while displaying help")
		return nil
	})
	root.SetOut(output)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if got := output.String(); !strings.Contains(got, "post-html") {
		t.Fatalf("help output = %q", got)
	}
}
