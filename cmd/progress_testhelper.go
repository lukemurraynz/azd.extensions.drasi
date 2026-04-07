package cmd

import "github.com/spf13/cobra"

// NewTestableProgressCommand returns a hidden subcommand that exercises
// ProgressHelper. It is exported for use in black-box tests (cmd_test package).
// This command is not registered in NewRootCommand and has no user-facing surface.
func NewTestableProgressCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "_test-progress",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := NewProgressHelper(cmd)
			if err != nil {
				return err
			}

			if err := p.Start(); err != nil {
				return err
			}
			p.Message("testing")
			return p.Stop()
		},
	}
}
