package cmd

import (
	"fmt"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

func newUpgradeCommand() *cobra.Command {
	var envName string
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Drasi runtime assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			if !force {
				return writeCommandError(
					cmd,
					output.ERR_FORCE_REQUIRED,
					"upgrade requires --force",
					"Re-run with --force to confirm runtime upgrade operations.",
					format,
					output.ExitCodes[output.ERR_FORCE_REQUIRED],
				)
			}

			client := drasi.NewClient()
			if err := client.CheckVersion(cmd.Context()); err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_NOT_FOUND)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Install or upgrade the drasi CLI and retry.",
					format,
					output.ExitCodes[code],
				)
			}

			if err := runDrasiCommand(cmd.Context(), "upgrade"); err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Check cluster reachability and drasi runtime status, then retry.",
					format,
					output.ExitCodes[code],
				)
			}

			if format == output.FormatJSON {
				payload := map[string]any{"status": "ok", "environment": envName}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			if envName != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Upgrade completed for environment %s\n", envName)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Upgrade completed.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&envName, "environment", "", "Target azd environment name")
	cmd.Flags().BoolVar(&force, "force", false, "Confirm runtime upgrade operation")
	return cmd
}
