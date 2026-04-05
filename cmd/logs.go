package cmd

import "github.com/spf13/cobra"

func newLogsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Show Drasi runtime logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("logs")
		},
	}
}
