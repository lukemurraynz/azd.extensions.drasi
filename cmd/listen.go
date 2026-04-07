package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

var drasiCheckFunc = func(ctx context.Context) error {
	client := drasi.NewClient()
	return client.CheckVersion(ctx)
}

var preDeployValidateFunc = func(dir, file, env string) (*validation.ValidationResult, error) {
	return validation.Validate(dir, file, env)
}

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

	result, err := preDeployValidateFunc(filepath.Join("drasi"), "drasi.yaml", "")
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

	return nil
}
