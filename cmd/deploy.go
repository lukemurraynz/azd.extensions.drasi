package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/azure/azd.extensions.drasi/internal/deployment"
	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	var configPath string
	var dryRun bool
	var envName string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy Drasi components",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			ctx := azdext.WithAccessToken(cmd.Context())

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

			absoluteConfigPath, err := filepath.Abs(configPath)
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

			manifestDir := filepath.Dir(absoluteConfigPath)
			manifestFile := filepath.Base(absoluteConfigPath)

			validationResult, err := validation.Validate(manifestDir, manifestFile, resolvedEnv)
			if err != nil {
				code := errorCodeFromError(err, output.ERR_NO_MANIFEST)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Fix validation issues before deploying.",
					format,
					output.ExitCodes[code],
				)
			}
			if validationResult.HasErrors() {
				renderValidationOutput(cmd, validationResult, format)
				return &commandError{message: fmt.Sprintf("%s: validation failed", output.ERR_VALIDATION_FAILED), exitCode: output.ExitCodes[output.ERR_VALIDATION_FAILED]}
			}

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
			deployLock, err := state.ReadHash(ctx, "DRASI_DEPLOY_IN_PROGRESS")
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_DEPLOY_IN_PROGRESS,
					fmt.Sprintf("reading deploy lock: %s", err),
					"Retry after the current deploy lock check completes.",
					format,
					output.ExitCodes[output.ERR_DEPLOY_IN_PROGRESS],
				)
			}
			if deployLock == "true" {
				return writeCommandError(
					cmd,
					output.ERR_DEPLOY_IN_PROGRESS,
					"a deploy is already in progress for this environment",
					"Wait for the current deploy to finish, then retry.",
					format,
					output.ExitCodes[output.ERR_DEPLOY_IN_PROGRESS],
				)
			}
			if err := state.WriteHash(ctx, "DRASI_DEPLOY_IN_PROGRESS", "true"); err != nil {
				return writeCommandError(
					cmd,
					output.ERR_DEPLOY_IN_PROGRESS,
					fmt.Sprintf("writing deploy lock: %s", err),
					"Ensure azd environment state is writable and retry.",
					format,
					output.ExitCodes[output.ERR_DEPLOY_IN_PROGRESS],
				)
			}
			defer func() {
				_ = state.WriteHash(ctx, "DRASI_DEPLOY_IN_PROGRESS", "")
			}()

			engine := deployment.NewEngine(state, drasiClient)

			if err := engine.Deploy(ctx, &resolved, deployment.DeployOptions{DryRun: dryRun, Environment: resolvedEnv}); err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Check Drasi runtime health with `azd drasi diagnose`, then retry.",
					format,
					output.ExitCodes[code],
				)
			}

			if format == output.FormatJSON {
				payload := map[string]any{"status": "ok", "environment": resolvedEnv, "dryRun": dryRun}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			if dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Dry-run succeeded for environment %s\n", resolvedEnv)
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deploy succeeded for environment %s\n", resolvedEnv)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", filepath.Join("drasi", "drasi.yaml"), "Path to drasi.yaml manifest")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Compute changes without applying resources")
	cmd.Flags().StringVar(&envName, "environment", "", "Target azd environment name")

	return cmd
}
