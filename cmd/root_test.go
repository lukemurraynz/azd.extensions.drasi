package cmd

import "testing"

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
	}

	if !hasListen {
		t.Fatal("expected listen subcommand to be registered")
	}

	if !hasMetadata {
		t.Fatal("expected metadata subcommand to be registered")
	}
}
