package cmd

import (
	"fmt"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

func newUpgradeCommand() *cobra.Command {
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

			// Resolve kube context from --environment root flag (same pattern as status/diagnose).
			kubeContext, err := resolvedKubeContextForCommand(cmd.Context(), cmd, "")
			if err != nil {
				code := errorCodeFromError(err, output.ERR_AKS_CONTEXT_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(),
					"Set the target azd environment and ensure AZURE_AKS_CONTEXT is present.",
					format, output.ExitCodes[code])
			}

			if kubeContext != "" {
				if err := switchKubectlContext(cmd.Context(), kubeContext); err != nil {
					return writeCommandError(cmd, output.ERR_AKS_CONTEXT_NOT_FOUND,
						fmt.Sprintf("switching kubectl context to %s: %s", kubeContext, err),
						"Ensure the AKS context exists in your kubeconfig.",
						format, output.ExitCodes[output.ERR_AKS_CONTEXT_NOT_FOUND])
				}
				if err := runDrasiCommand(cmd.Context(), "env", "kube"); err != nil {
					return writeCommandError(cmd, output.ERR_DRASI_CLI_ERROR,
						fmt.Sprintf("registering drasi environment: %s", err),
						"Ensure the drasi CLI is installed and the cluster is reachable.",
						format, output.ExitCodes[output.ERR_DRASI_CLI_ERROR])
				}
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

			// Derive the environment label from the root flag for output messaging.
			envLabel, _ := cmd.Root().PersistentFlags().GetString("environment")

			if format == output.FormatJSON {
				payload := map[string]any{"status": "ok", "environment": envLabel}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			if envLabel != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Upgrade completed for environment %s\n", envLabel)
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Upgrade completed.")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Confirm runtime upgrade operation")
	return cmd
}
