package cmd

import "github.com/spf13/cobra"

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a Drasi project scaffold",
		RunE: func(cmd *cobra.Command, args []string) error {
			// intentional stub — FR-010
			return notImplemented("init")
		},
	}
}
