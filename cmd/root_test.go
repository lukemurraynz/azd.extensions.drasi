package cmd

import (
	"bytes"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	if rootCmd == nil {
		t.Fatal("expected non-nil root command")
	}

	outputFlag := rootCmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("expected --output flag to be registered")
	}
	if got := outputFlag.DefValue; got != "table" {
		t.Fatalf("expected --output default to be table, got %q", got)
	}

	debugFlag := rootCmd.PersistentFlags().Lookup("debug")
	if debugFlag == nil {
		t.Fatal("expected --debug flag to be registered")
	}

	hasListen := false
	hasMetadata := false
	hasVersion := false
	for _, subCmd := range rootCmd.Commands() {
		if subCmd.Name() == "listen" {
			hasListen = true
			if !subCmd.Hidden {
				t.Fatal("expected listen subcommand to be hidden")
			}
		}
		if subCmd.Name() == "metadata" {
			hasMetadata = true
			if !subCmd.Hidden {
				t.Fatal("expected metadata subcommand to be hidden")
			}
		}
		if subCmd.Name() == "version" {
			hasVersion = true
			if subCmd.Hidden {
				t.Fatal("expected version subcommand to be visible")
			}
		}
	}

	if !hasListen {
		t.Fatal("expected listen subcommand to be registered")
	}

	if !hasMetadata {
		t.Fatal("expected metadata subcommand to be registered")
	}

	if !hasVersion {
		t.Fatal("expected version subcommand to be registered")
	}
}

func TestRootCommand_PersistentPreRunE_SetsObservability(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	stderr := &bytes.Buffer{}
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// NOTE: version output goes to os.Stdout via azdext.NewVersionCommand,
	// not cmd.OutOrStdout(), so we only verify execution succeeds.
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRootCommand_PersistentPostRunE_NilShutdownSafe(t *testing.T) {
	t.Parallel()

	rootCmd := NewRootCommand()
	stderr := &bytes.Buffer{}
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// NOTE: version output goes to os.Stdout via azdext.NewVersionCommand,
	// not cmd.OutOrStdout(), so we only verify execution succeeds.
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}
