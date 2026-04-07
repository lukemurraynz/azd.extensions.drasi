package cmd

import (
	"context"
	"fmt"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

type upgradeDrasiClient interface {
	CheckVersion(ctx context.Context) error
	GetVersion(ctx context.Context) (string, error)
}

var newUpgradeDrasiClient = func() upgradeDrasiClient {
	return drasi.NewClient()
}

var resolveUpgradeKubeContext = resolvedKubeContextForCommand
var switchUpgradeKubectlContext = switchKubectlContext
var runUpgradeDrasiCommand = runDrasiCommand

func newUpgradeCommand() *cobra.Command {
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Drasi runtime assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			envLabel, _ := cmd.Root().PersistentFlags().GetString("environment")

			confirmed, err := ConfirmDestructive("This will upgrade the Drasi runtime on the active cluster. Continue?", force)
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_FORCE_REQUIRED,
					err.Error(),
					"Re-run with --force for non-interactive environments.",
					format,
					output.ExitCodes[output.ERR_FORCE_REQUIRED],
				)
			}
			if !confirmed {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Upgrade aborted by user.")
				return nil
			}

			// Resolve kube context from --environment root flag (same pattern as status/diagnose).
			progress, progressErr := NewProgressHelper(cmd)
			if progressErr != nil {
				progress = &ProgressHelper{noop: true}
			}
			_ = progress.Start()
			defer func() { _ = progress.Stop() }()

			progress.Message("Resolving cluster context...")

			kubeContext, err := resolveUpgradeKubeContext(cmd.Context(), cmd, "")
			if err != nil {
				code := errorCodeFromError(err, output.ERR_AKS_CONTEXT_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(),
					"Set the target azd environment and ensure AZURE_AKS_CONTEXT is present.",
					format, output.ExitCodes[code])
			}

			if kubeContext != "" {
				if err := switchUpgradeKubectlContext(cmd.Context(), kubeContext); err != nil {
					return writeCommandError(cmd, output.ERR_AKS_CONTEXT_NOT_FOUND,
						fmt.Sprintf("switching kubectl context to %s: %s", kubeContext, err),
						"Ensure the AKS context exists in your kubeconfig.",
						format, output.ExitCodes[output.ERR_AKS_CONTEXT_NOT_FOUND])
				}
				if err := runUpgradeDrasiCommand(cmd.Context(), "env", "kube"); err != nil {
					return writeCommandError(cmd, output.ERR_DRASI_CLI_ERROR,
						fmt.Sprintf("registering drasi environment: %s", err),
						"Ensure the drasi CLI is installed and the cluster is reachable.",
						format, output.ExitCodes[output.ERR_DRASI_CLI_ERROR])
				}
			}

			client := newUpgradeDrasiClient()
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

			if dryRun {
				_ = progress.Stop()

				currentVersion, vErr := client.GetVersion(cmd.Context())
				versionStr := "unknown"
				if vErr == nil {
					versionStr = currentVersion
				}

				if format == output.FormatJSON {
					payload := map[string]any{"status": "dry-run", "currentVersion": versionStr, "environment": envLabel}
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
					return nil
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current Drasi runtime version: %s\n", versionStr)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Dry-run: upgrade would reinstall the Drasi runtime using the installed CLI version.")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run without --dry-run and with --force to apply.")
				return nil
			}

			progress.Message("Upgrading Drasi runtime...")

			if err := runUpgradeDrasiCommand(cmd.Context(), "upgrade"); err != nil {
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

			_ = progress.Stop()

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
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what the upgrade would do without applying changes")
	return cmd
}
