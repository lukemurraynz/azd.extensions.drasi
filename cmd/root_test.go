package cmd

import (
	"testing"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
	// Not parallel: mutates package-level observability vars.

	// Reset package-level vars before test.
	rootTracer = nil
	rootMeter = nil
	shutdownTracer = nil
	shutdownMeter = nil

	rootCmd := NewRootCommand()
	rootCmd.SetArgs([]string{"version"})

	// version is already registered in NewRootCommand and triggers
	// PersistentPreRunE on the root before running its own RunE.
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without APPLICATIONINSIGHTS_CONNECTION_STRING, we get no-op providers,
	// but the package-level vars should still be set (non-nil).
	if rootTracer == nil {
		t.Fatal("expected rootTracer to be set after PersistentPreRunE")
	}
	if rootMeter == nil {
		t.Fatal("expected rootMeter to be set after PersistentPreRunE")
	}

	// Verify types are correct interfaces.
	var _ trace.Tracer = rootTracer
	var _ metric.Meter = rootMeter
}

func TestRootCommand_PersistentPostRunE_NilShutdownSafe(t *testing.T) {
	// Not parallel: mutates package-level shutdown vars.

	// Ensure nil shutdown functions do not panic.
	shutdownTracer = nil
	shutdownMeter = nil

	rootCmd := NewRootCommand()
	// version is already registered; just invoke it.
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
