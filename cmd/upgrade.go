package cmd

import "github.com/spf13/cobra"

func newUpgradeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Drasi runtime assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("upgrade")
		},
	}
}
