package cmd

import "github.com/spf13/cobra"

func newProvisionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "provision",
		Short: "Provision Azure infrastructure for Drasi",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("provision")
		},
	}
}
