package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

type logsDrasiClient interface {
	CheckVersion(ctx context.Context) error
	DescribeComponent(ctx context.Context, kind, id string) (*drasi.ComponentDetail, error)
	DescribeComponentInContext(ctx context.Context, kind, id, kubeContext string) (*drasi.ComponentDetail, error)
	RunCommandOutput(ctx context.Context, args ...string) (string, error)
}

var newLogsDrasiClient = func() logsDrasiClient {
	return drasi.NewClient()
}

func newLogsCommand() *cobra.Command {
	var componentID string
	var kind string
	var tail int
	var follow bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Watch Drasi query results",
		RunE: func(cmd *cobra.Command, _ []string) error {
			format := outputFormatFromCommand(cmd)
			kubeContext, err := resolvedKubeContextForCommand(cmd.Context(), cmd, "")
			if err != nil {
				code := errorCodeFromError(err, output.ERR_AKS_CONTEXT_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(), "Set the target azd environment and ensure AZURE_AKS_CONTEXT is present.", format, output.ExitCodes[code])
			}
			if kind == "continuousquery" {
				kind = "query"
			}

			if strings.TrimSpace(componentID) == "" || strings.TrimSpace(kind) == "" {
				return writeCommandError(
					cmd,
					output.ERR_VALIDATION_FAILED,
					"logs requires both --component and --kind",
					"Provide --component <query-id> and --kind continuousquery/query for watch mode.",
					format,
					output.ExitCodes[output.ERR_VALIDATION_FAILED],
				)
			}
			if kind != "query" {
				return writeCommandError(
					cmd,
					output.ERR_VALIDATION_FAILED,
					"logs only supports continuousquery/query components in this Drasi CLI version",
					"Use --kind continuousquery (or query) and pass a query component id.",
					format,
					output.ExitCodes[output.ERR_VALIDATION_FAILED],
				)
			}

			client := newLogsDrasiClient()
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

			var detail *drasi.ComponentDetail
			if kubeContext == "" {
				detail, err = client.DescribeComponent(cmd.Context(), kind, componentID)
			} else {
				detail, err = client.DescribeComponentInContext(cmd.Context(), kind, componentID, kubeContext)
			}
			if err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Verify component identifiers and cluster connectivity, then retry.",
					format,
					output.ExitCodes[code],
				)
			}

			logArgs := []string{"watch", componentID}
			if kubeContext != "" {
				logArgs = append([]string{"--context", kubeContext}, logArgs...)
			}
			_ = tail
			if follow {
				// drasi watch streams by default; --follow remains a compatibility alias.
			}

			logOutput, err := client.RunCommandOutput(cmd.Context(), logArgs...)
			if err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Verify drasi watch command support and cluster connectivity, then retry.",
					format,
					output.ExitCodes[code],
				)
			}

			payload := map[string]any{
				"status":    "ok",
				"component": componentID,
				"kind":      kind,
				"follow":    follow,
				"detail":    detail,
				"watch":     strings.TrimSpace(logOutput),
			}

			if format == output.FormatJSON {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			if strings.TrimSpace(logOutput) == "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No watch output found for %s (%s).\n", componentID, kind)
				return nil
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(logOutput, "\n"))

			return nil
		},
	}

	cmd.Flags().StringVar(&componentID, "component", "", "Filter logs by component ID")
	cmd.Flags().StringVar(&kind, "kind", "", "Filter logs by component kind (source, continuousquery, middleware, reaction)")
	cmd.Flags().IntVar(&tail, "tail", 0, "Number of recent log lines to show (0 = all)")
	cmd.Flags().BoolVar(&follow, "follow", false, "Stream log output (compatibility alias)")

	return cmd
}
