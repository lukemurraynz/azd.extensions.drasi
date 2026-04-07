package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

type describeDrasiClient interface {
	CheckVersion(ctx context.Context) error
	DescribeComponent(ctx context.Context, kind, id string) (*drasi.ComponentDetail, error)
	DescribeComponentInContext(ctx context.Context, kind, id, kubeContext string) (*drasi.ComponentDetail, error)
}

var newDescribeDrasiClient = func() describeDrasiClient {
	return drasi.NewClient()
}

func newDescribeCommand() *cobra.Command {
	var kind string
	var componentID string

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details for a single Drasi component",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			ctx := cmd.Context()

			if strings.TrimSpace(kind) == "" || strings.TrimSpace(componentID) == "" {
				return writeCommandError(cmd, output.ERR_VALIDATION_FAILED,
					"describe requires both --kind and --component",
					"Provide --kind <kind> and --component <id>.",
					format, output.ExitCodes[output.ERR_VALIDATION_FAILED])
			}

			kubeContext, err := resolvedKubeContextForCommand(ctx, cmd, "")
			if err != nil {
				code := errorCodeFromError(err, output.ERR_AKS_CONTEXT_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(),
					"Set the target azd environment and ensure AZURE_AKS_CONTEXT is present.",
					format, output.ExitCodes[code])
			}

			// Map continuousquery → query for the drasi CLI (consistent with status/logs).
			selectedKind := kind
			if selectedKind == "continuousquery" {
				selectedKind = "query"
			}

			client := newDescribeDrasiClient()
			if err := client.CheckVersion(ctx); err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(),
					"Install or upgrade the drasi CLI and retry.",
					format, output.ExitCodes[code])
			}

			var detail *drasi.ComponentDetail
			if kubeContext == "" {
				detail, err = client.DescribeComponent(ctx, selectedKind, componentID)
			} else {
				detail, err = client.DescribeComponentInContext(ctx, selectedKind, componentID, kubeContext)
			}
			if err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(cmd, code,
					fmt.Sprintf("component %s/%s not found or not accessible: %s", selectedKind, componentID, err),
					"Verify component kind and ID, check cluster connectivity, then retry.",
					format, output.ExitCodes[code])
			}

			if format == output.FormatJSON {
				payload := map[string]any{
					"status":    "ok",
					"kind":      selectedKind,
					"component": componentID,
					"detail":    detail,
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(detail, output.FormatTable))
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "", "Component kind (source, continuousquery, middleware, reaction)")
	cmd.Flags().StringVar(&componentID, "component", "", "Component ID to describe")

	return cmd
}
