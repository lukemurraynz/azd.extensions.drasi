package cmd

import "github.com/spf13/cobra"

func newTeardownCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "teardown",
		Short: "Tear down Drasi components and infrastructure",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("teardown")
		},
	}
}
