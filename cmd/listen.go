package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	"github.com/spf13/cobra"
)

var drasiCheckFunc = func(ctx context.Context) error {
	client := drasi.NewClient()
	return client.CheckVersion(ctx)
}

var preDeployValidateFunc = validation.Validate

var (
	postProvisionTimeout  = 60 * time.Second
	postProvisionInterval = 5 * time.Second
)

func newListenCommand() *cobra.Command {
	return azdext.NewListenCommand(func(host *azdext.ExtensionHost) {
		host.
			WithProjectEventHandler("postprovision", handlePostProvision).
			WithProjectEventHandler("predeploy", handlePreDeploy).
			WithProjectEventHandler("predown", handlePreDown)
	})
}

func handlePostProvision(ctx context.Context, args *azdext.ProjectEventArgs) error {
	projectName := ""
	if args != nil && args.Project != nil {
		projectName = args.Project.Name
	}
	slog.InfoContext(ctx, "drasi: post-provision hook fired — checking Drasi API health", slog.String("project", projectName))

	if err := waitForDrasiReady(ctx, postProvisionTimeout, postProvisionInterval); err != nil {
		slog.WarnContext(ctx, "drasi: API not ready after post-provision timeout — proceeding anyway",
			slog.String("project", projectName),
			slog.Any("error", err),
		)
		return nil
	}

	slog.InfoContext(ctx, "drasi: API is ready after post-provision", slog.String("project", projectName))
	return nil
}

func handlePreDeploy(ctx context.Context, args *azdext.ProjectEventArgs) error {
	projectName := ""
	if args != nil && args.Project != nil {
		projectName = args.Project.Name
	}
	slog.InfoContext(ctx, "drasi: pre-deploy hook fired — validating manifest", slog.String("project", projectName))

	result, err := preDeployValidateFunc("drasi", "drasi.yaml", "")
	if err != nil {
		slog.ErrorContext(ctx, "drasi: pre-deploy validation failed to load manifest",
			slog.String("project", projectName),
			slog.Any("error", err),
		)
		return fmt.Errorf("pre-deploy validation: %w", err)
	}

	if result.HasErrors() {
		errorCount := countErrors(result)
		slog.ErrorContext(ctx, "drasi: pre-deploy validation found errors — blocking deploy",
			slog.String("project", projectName),
			slog.Int("error_count", errorCount),
		)
		return fmt.Errorf("pre-deploy validation failed with %d error(s)", errorCount)
	}

	if len(result.Issues) > 0 {
		slog.WarnContext(ctx, "drasi: pre-deploy validation found warnings — proceeding",
			slog.String("project", projectName),
			slog.Int("warning_count", len(result.Issues)),
		)
	}

	slog.InfoContext(ctx, "drasi: pre-deploy validation passed", slog.String("project", projectName))
	return nil
}

func waitForDrasiReady(ctx context.Context, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := drasiCheckFunc(ctx); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}

	return fmt.Errorf("drasi API not ready after %v", timeout)
}

// handlePreDown runs before `azd down` tears down the Azure infrastructure.
//
// NOTE: This handler uninstalls the Drasi runtime from the AKS cluster before
// the AKS resource itself is deleted. Without this step, `azd down --purge`
// would delete the AKS cluster while Drasi namespaces and their PersistentVolumes
// still exist, which can leave Azure Disk resources orphaned (detached disks that
// continue to accrue costs). Running `drasi uninstall` first ensures a clean
// teardown of Drasi-managed Kubernetes resources before AKS is removed.
func handlePreDown(ctx context.Context, args *azdext.ProjectEventArgs) error {
	projectName := ""
	if args != nil && args.Project != nil {
		projectName = args.Project.Name
	}
	slog.InfoContext(ctx, "drasi: pre-down hook fired — uninstalling Drasi runtime", slog.String("project", projectName))

	// `drasi uninstall` removes the drasi-system namespace and all Drasi-managed
	// Kubernetes resources. Errors are logged to stderr but do not block azd down —
	// the infrastructure teardown should proceed even if Drasi uninstall fails
	// (e.g. cluster is already unreachable).
	if err := runDrasiCommand(ctx, "uninstall", "--yes"); err != nil {
		slog.WarnContext(ctx, "drasi uninstall failed during predown — proceeding with infrastructure teardown",
			slog.String("project", projectName),
			slog.Any("error", err),
		)
	}

	// Clear runtime state from azd environment so a subsequent `azd up` starts
	// from a clean slate. Best-effort: failures are logged but do not block teardown.
	clearPreDownState(ctx, args, projectName)

	return nil
}

// clearPreDownState removes DRASI_PROVISIONED and AZURE_AKS_CONTEXT from azd
// environment state after Drasi uninstall. This prevents stale markers from
// short-circuiting a future provision cycle.
func clearPreDownState(ctx context.Context, args *azdext.ProjectEventArgs, projectName string) {
	azdClient, err := azdext.NewAzdClient()
	if err != nil {
		slog.WarnContext(ctx, "drasi: could not create azd client to clear state — skipping",
			slog.String("project", projectName),
			slog.Any("error", err),
		)
		return
	}
	defer azdClient.Close()

	azdCtx := azdext.WithAccessToken(ctx)
	envName, err := resolveEnvironmentNameFromEventArgs(azdCtx, azdClient, args)
	if err != nil || envName == "" {
		slog.WarnContext(ctx, "drasi: could not resolve environment name for state cleanup — skipping",
			slog.String("project", projectName),
			slog.Any("error", err),
		)
		return
	}

	adapter := &azdEnvServiceAdapter{client: azdClient}
	for _, key := range []string{"DRASI_PROVISIONED", "AZURE_AKS_CONTEXT"} {
		if setErr := adapter.SetValue(azdCtx, envName, key, ""); setErr != nil {
			slog.WarnContext(ctx, "drasi: failed to clear env key — skipping",
				slog.String("project", projectName),
				slog.String("key", key),
				slog.Any("error", setErr),
			)
		}
	}
	slog.InfoContext(ctx, "drasi: cleared runtime state from azd environment", slog.String("project", projectName), slog.String("environment", envName))
}

// resolveEnvironmentNameFromEventArgs resolves the environment name from event
// args metadata or falls back to GetCurrent. Lifecycle event handlers do not
// have access to cobra flags, so this uses a simpler resolution path.
func resolveEnvironmentNameFromEventArgs(ctx context.Context, azdClient *azdext.AzdClient, args *azdext.ProjectEventArgs) (string, error) {
	// ProjectEventArgs does not carry an explicit environment name today.
	// Fall back to the azd host's current environment.
	resp, err := azdClient.Environment().GetCurrent(ctx, &azdext.EmptyRequest{})
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Environment == nil {
		return "", fmt.Errorf("current azd environment is not set")
	}
	return resp.Environment.Name, nil
}
