package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/observability"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	extensionID           = "azure.drasi"
	metadataSchemaVersion = "1.0"
)

var extensionVersion = "dev"

type rootState struct {
	tracer           trace.Tracer
	meter            metric.Meter
	shutdownTracer   func(context.Context) error
	shutdownMeter    func(context.Context) error
	commandStartTime time.Time
}

func SetVersion(version string) {
	extensionVersion = version
}

func NewRootCommand() *cobra.Command {
	var outputFormat string
	state := &rootState{}

	rootCmd := &cobra.Command{
		Use:           "azd drasi <command> [options]",
		Short:         "Manage Drasi reactive data pipeline workloads",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			state.tracer = nil
			state.meter = nil
			state.shutdownTracer = nil
			state.shutdownMeter = nil
			state.commandStartTime = time.Now()

			ctx := cmd.Context()

			t, tShutdown, err := observability.NewTracer(ctx)
			if err != nil {
				slog.WarnContext(ctx, "tracer init failed, continuing with no-op", slog.Any("error", err))
			} else {
				state.tracer = t
				state.shutdownTracer = tShutdown
			}

			m, mShutdown, err := observability.NewMeter(ctx)
			if err != nil {
				slog.WarnContext(ctx, "meter init failed, continuing with no-op", slog.Any("error", err))
			} else {
				state.meter = m
				state.shutdownMeter = mShutdown
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Record command success metrics before shutting down the meter.
			// PersistentPostRunE only runs when RunE succeeds (Cobra behavior).
			if state.meter != nil {
				elapsed := time.Since(state.commandStartTime)
				observability.RecordCommandExecution(ctx, state.meter, cmd.Name(), elapsed, nil)
			}

			if state.shutdownTracer != nil {
				if err := state.shutdownTracer(ctx); err != nil {
					slog.WarnContext(ctx, "tracer shutdown error", slog.Any("error", err))
				}
			}

			if state.shutdownMeter != nil {
				if err := state.shutdownMeter(ctx); err != nil {
					slog.WarnContext(ctx, "meter shutdown error", slog.Any("error", err))
				}
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table or json")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable verbose debug logging")
	rootCmd.PersistentFlags().StringP("environment", "e", "", "Name of the azd environment to use")

	rootCmd.AddCommand(newListenCommand())
	rootCmd.AddCommand(newMetadataCommand())
	rootCmd.AddCommand(newVersionCommand(&outputFormat))
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newProvisionCommand())
	rootCmd.AddCommand(newDeployCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newCheckCommand())
	rootCmd.AddCommand(newDiagnoseCommand())
	rootCmd.AddCommand(newTeardownCommand())
	rootCmd.AddCommand(newUpgradeCommand())
	rootCmd.AddCommand(newDescribeCommand())

	return rootCmd
}
