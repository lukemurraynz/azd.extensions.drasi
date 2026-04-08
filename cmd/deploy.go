package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/google/uuid"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/deployment"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	"github.com/spf13/cobra"
)

// envVarRefPattern matches $(VARNAME) placeholders in YAML files.
var envVarRefPattern = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*)\)`)

// scanEnvVarRefs reads component YAML files from the manifest directory and
// returns the unique set of $(VARNAME) variable names referenced across all files.
func scanEnvVarRefs(resolved *config.ResolvedManifest) []string {
	seen := make(map[string]struct{})
	var paths []string

	for _, s := range resolved.Sources {
		if s.FilePath != "" {
			paths = append(paths, s.FilePath)
		}
	}
	for _, q := range resolved.Queries {
		if q.FilePath != "" {
			paths = append(paths, q.FilePath)
		}
	}
	for _, r := range resolved.Reactions {
		if r.FilePath != "" {
			paths = append(paths, r.FilePath)
		}
	}
	for _, m := range resolved.Middlewares {
		if m.FilePath != "" {
			paths = append(paths, m.FilePath)
		}
	}

	for _, relPath := range paths {
		absPath := filepath.Join(resolved.ManifestDir, filepath.FromSlash(relPath))
		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		for _, match := range envVarRefPattern.FindAllSubmatch(data, -1) {
			if len(match) > 1 {
				seen[string(match[1])] = struct{}{}
			}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	return names
}

func newDeployCommand() *cobra.Command {
	var configPath string
	var dryRun bool
	var noRollback bool
	var timeoutStr string

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy Drasi components",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			ctx := azdext.WithAccessToken(cmd.Context())
			var err error

			startedAt := time.Now().UTC()
			correlationID := uuid.New().String()

			var totalTimeout time.Duration
			if timeoutStr != "" {
				totalTimeout, err = time.ParseDuration(timeoutStr)
				if err != nil {
					return writeCommandError(
						cmd,
						output.ERR_VALIDATION_FAILED,
						fmt.Sprintf("invalid --timeout value %q: %s", timeoutStr, err),
						"Use Go duration format: 30m, 1h, 2h30m.",
						format,
						output.ExitCodes[output.ERR_VALIDATION_FAILED],
					)
				}
			}

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

			envFlag, _ := cmd.Root().PersistentFlags().GetString("environment")
			resolvedEnv, err := resolveEnvironmentName(ctx, cmd, azdClient, envFlag)
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

			progress.Message("Validating configuration...")

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
			lock := deployment.NewDeployLock(state)

			stale, err := lock.IsStale(ctx, 30*time.Minute)
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
			if !stale {
				return writeCommandError(
					cmd,
					output.ERR_DEPLOY_IN_PROGRESS,
					"a deploy is already in progress for this environment",
					"Wait for the current deploy to finish, then retry. If the previous deploy crashed, the lock will expire after 30 minutes.",
					format,
					output.ExitCodes[output.ERR_DEPLOY_IN_PROGRESS],
				)
			}

			// Check if a stale lock was left behind by a crashed deploy and force-release it.
			if currentVal, readErr := state.ReadHash(ctx, "DRASI_DEPLOY_IN_PROGRESS"); readErr == nil && currentVal != "" {
				slog.WarnContext(ctx, "stale deploy lock detected, force-releasing", "value", currentVal)
				if releaseErr := lock.ForceRelease(ctx); releaseErr != nil {
					slog.ErrorContext(ctx, "failed to force-release stale lock", "error", releaseErr)
				}
			}

			if err := lock.Acquire(ctx); err != nil {
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
				if releaseErr := lock.Release(ctx); releaseErr != nil {
					slog.ErrorContext(ctx, "failed to release deploy lock", "error", releaseErr)
				}
			}()

			engine := deployment.NewEngine(state, drasiClient)

			// Scan component YAML files for $(VARNAME) references and resolve
			// each from azd environment state so the deploy engine can substitute
			// them before passing YAML to `drasi apply`.
			envVars := make(map[string]string)
			for _, varName := range scanEnvVarRefs(&resolved) {
				val, err := getEnvValue(ctx, azdClient, resolvedEnv, varName)
				if err != nil {
					slog.WarnContext(ctx, "could not resolve env var for YAML substitution",
						slog.String("var", varName), slog.Any("error", err))
					continue
				}
				if val != "" {
					envVars[varName] = val
				}
			}

			progress.Message("Deploying components...")

			if err := engine.Deploy(ctx, &resolved, deployment.DeployOptions{
				DryRun:       dryRun,
				Environment:  resolvedEnv,
				NoRollback:   noRollback,
				TotalTimeout: totalTimeout,
				EnvVars:      envVars,
			}); err != nil {
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

			slog.InfoContext(ctx, "deploy completed",
				slog.String("environment", resolvedEnv),
				slog.String("correlation_id", correlationID),
				slog.Bool("dry_run", dryRun),
				slog.Int("component_count", len(resolved.Sources)+len(resolved.Queries)+len(resolved.Reactions)+len(resolved.Middlewares)),
			)

			endedAt := time.Now().UTC()
			deployResult := "success"
			if dryRun {
				deployResult = "dry-run"
			}

			// Emit structured audit event to stderr (matching provision command pattern).
			auditEvent := output.AuditEvent{
				Operation:     "deploy",
				Environment:   resolvedEnv,
				CorrelationID: correlationID,
				Target:        "drasi-components",
				Result:        deployResult,
				StartedAtUtc:  startedAt,
				EndedAtUtc:    endedAt,
			}
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output.FormatAuditEvent(auditEvent, format))

			if format == output.FormatJSON {
				_ = progress.Stop()
				payload := map[string]any{"status": "ok", "environment": resolvedEnv, "dryRun": dryRun}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			_ = progress.Stop()

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
	cmd.Flags().BoolVar(&noRollback, "no-rollback", false, "Skip rollback on deploy failure")
	cmd.Flags().StringVar(&timeoutStr, "timeout", "", "Total deploy timeout (e.g. 30m, 1h). Default: 15m")

	return cmd
}
