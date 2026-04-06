package cmd

import (
	"context"
	"log/slog"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
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
	slog.InfoContext(ctx, "drasi: post-provision hook fired", slog.String("project", projectName))
	return nil
}

func handlePreDeploy(ctx context.Context, args *azdext.ProjectEventArgs) error {
	projectName := ""
	if args != nil && args.Project != nil {
		projectName = args.Project.Name
	}
	slog.InfoContext(ctx, "drasi: pre-deploy hook fired", slog.String("project", projectName))
	return nil
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
