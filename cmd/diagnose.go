package cmd

import "github.com/spf13/cobra"

func newDiagnoseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Run Drasi diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("diagnose")
		},
	}
}
