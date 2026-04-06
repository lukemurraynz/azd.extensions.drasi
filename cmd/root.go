package cmd

import "github.com/spf13/cobra"

const (
	extensionID           = "azure.drasi"
	metadataSchemaVersion = "1.0"
)

// NewRootCommand builds the cobra command tree for azd drasi.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "azd drasi <command> [options]",
		Short:         "Manage Drasi reactive data pipeline workloads",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().String("output", "table", "Output format: table or json")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable verbose debug logging")
	rootCmd.PersistentFlags().StringP("environment", "e", "", "Name of the azd environment to use")

	rootCmd.AddCommand(newListenCommand())
	rootCmd.AddCommand(newMetadataCommand())
	rootCmd.AddCommand(newValidateCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newProvisionCommand())
	rootCmd.AddCommand(newDeployCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newDiagnoseCommand())
	rootCmd.AddCommand(newTeardownCommand())
	rootCmd.AddCommand(newUpgradeCommand())

	return rootCmd
}
