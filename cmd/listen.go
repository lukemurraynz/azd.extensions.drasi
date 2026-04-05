package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newListenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "listen",
		Short: "Subscribe to azd lifecycle events (invoked by azd host)",
		RunE:  runListen,
	}
}

func runListen(cmd *cobra.Command, _ []string) error {
	ctx := azdext.WithAccessToken(cmd.Context())

	azdClient, err := azdext.NewAzdClient()
	if err != nil {
		return fmt.Errorf("creating azd client: %w", err)
	}
	defer azdClient.Close()

	eventManager := azdext.NewEventManager("azd-drasi", azdClient, log.New(os.Stderr, "", 0))
	defer func() {
		_ = eventManager.Close()
	}()

	if err := eventManager.AddProjectEventHandler(ctx, "postprovision", handlePostProvision); err != nil {
		return fmt.Errorf("subscribing to postprovision: %w", err)
	}
	if err := eventManager.AddProjectEventHandler(ctx, "predeploy", handlePreDeploy); err != nil {
		return fmt.Errorf("subscribing to predeploy: %w", err)
	}

	return eventManager.Receive(ctx)
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
