package cmd

import (
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

func newUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Drasi runtime assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010; remove when upgrade logic is implemented
			outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")
			format := output.FormatTable
			if outputFormat == string(output.FormatJSON) {
				format = output.FormatJSON
			}
			return writeCommandError(
				cmd,
				output.ERR_NOT_IMPLEMENTED,
				"upgrade is planned for a future release \u2014 see https://github.com/azure/azd.extensions.drasi/issues for the roadmap",
				"Check the roadmap for the planned upgrade release.",
				format,
				output.ExitCodes[output.ERR_NOT_IMPLEMENTED],
			)
		},
	}
	cmd.Flags().String("environment", "", "Target azd environment name")
	return cmd
}
