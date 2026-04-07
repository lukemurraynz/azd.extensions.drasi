package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/deployment"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newTeardownCommand() *cobra.Command {
	var configPath string
	var force bool
	var includeInfrastructure bool
	var envName string

	cmd := &cobra.Command{
		Use:   "teardown",
		Short: "Tear down Drasi components and infrastructure",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)

			prompt := "This will remove all deployed Drasi components. Continue?"
			if includeInfrastructure {
				prompt = "This will delete all Drasi components AND the Azure resource group. This is irreversible. Continue?"
			}
			confirmed, err := ConfirmDestructive(prompt, force)
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
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Teardown aborted by user.")
				return nil
			}

			ctx := azdext.WithAccessToken(cmd.Context())

			progress, progressErr := NewProgressHelper(cmd)
			if progressErr != nil {
				progress = &ProgressHelper{noop: true}
			}
			_ = progress.Start()
			defer func() { _ = progress.Stop() }()

			progress.Message("Resolving environment...")

			azdClient, err := azdext.NewAzdClient()
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_NO_AUTH,
					fmt.Sprintf("creating azd client: %s", err),
					"Run `azd auth login` and ensure AZD_SERVER is set.",
					format,
					output.ExitCodes[output.ERR_NO_AUTH],
				)
			}
			defer azdClient.Close()

			resolvedEnv, err := resolveEnvironmentName(ctx, cmd, azdClient, envName)
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_NO_AUTH,
					fmt.Sprintf("resolving environment: %s", err),
					"Run `azd auth login` and set/select an azd environment.",
					format,
					output.ExitCodes[output.ERR_NO_AUTH],
				)
			}

			manifestPath, err := filepath.Abs(configPath)
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_NO_MANIFEST,
					fmt.Sprintf("cannot resolve config path %q: %s", configPath, err),
					"Ensure the --config path is valid.",
					format,
					output.ExitCodes[output.ERR_NO_MANIFEST],
				)
			}
			manifestDir := filepath.Dir(manifestPath)
			manifestFile := filepath.Base(manifestPath)

			manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(manifestDir, manifestFile)
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_NO_MANIFEST,
					err.Error(),
					"Ensure drasi.yaml and component files are present and valid.",
					format,
					output.ExitCodes[output.ERR_NO_MANIFEST],
				)
			}
			resolved, _, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, manifestDir, resolvedEnv)
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_VALIDATION_FAILED,
					err.Error(),
					"Fix environment overlay references and retry.",
					format,
					output.ExitCodes[output.ERR_VALIDATION_FAILED],
				)
			}

			drasiClient := drasi.NewClient()
			if err := drasiClient.CheckVersion(ctx); err != nil {
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

			state := deployment.NewStateManagerFromClient(&azdEnvServiceAdapter{client: azdClient}, resolvedEnv)
			engine := deployment.NewEngine(state, drasiClient)

			progress.Message("Tearing down components...")

			if err := engine.Teardown(ctx, &resolved, deployment.DeployOptions{Environment: resolvedEnv}); err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Check Drasi runtime health and retry.",
					format,
					output.ExitCodes[code],
				)
			}

			if includeInfrastructure {
				progress.Message("Deleting Azure infrastructure...")

				rgName, err := getEnvValue(ctx, azdClient, resolvedEnv, "AZURE_RESOURCE_GROUP")
				if err != nil || rgName == "" {
					return writeCommandError(
						cmd,
						output.ERR_NO_AUTH,
						"could not resolve AZURE_RESOURCE_GROUP for infrastructure teardown",
						"Ensure azd environment values include AZURE_RESOURCE_GROUP and retry.",
						format,
						output.ExitCodes[output.ERR_NO_AUTH],
					)
				}
				if err := runAzGroupDelete(ctx, rgName); err != nil {
					return writeCommandError(
						cmd,
						output.ERR_DRASI_CLI_ERROR,
						fmt.Sprintf("deleting resource group %s: %s", rgName, err),
						"Remove resource locks and retry infrastructure teardown.",
						format,
						output.ExitCodes[output.ERR_DRASI_CLI_ERROR],
					)
				}
			}

			payload := map[string]any{"status": "ok", "environment": resolvedEnv, "infrastructure": includeInfrastructure}
			_ = progress.Stop()

			if format == output.FormatJSON {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Teardown completed for environment %s\n", resolvedEnv)
			if includeInfrastructure {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Infrastructure deletion triggered.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", filepath.Join("drasi", "drasi.yaml"), "Path to drasi.yaml manifest")
	cmd.Flags().BoolVar(&force, "force", false, "Confirm destructive teardown")
	cmd.Flags().BoolVar(&includeInfrastructure, "infrastructure", false, "Delete provisioned Azure infrastructure after component teardown")
	cmd.Flags().StringVar(&envName, "environment", "", "Target azd environment name")

	return cmd
}

func runAzGroupDelete(ctx context.Context, resourceGroup string) error {
	azPath, err := exec.LookPath("az")
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, azPath, "group", "delete", "--name", resourceGroup, "--yes", "--no-wait")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
