package cmd

import "github.com/spf13/cobra"

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Drasi component status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("status")
		},
	}
}
