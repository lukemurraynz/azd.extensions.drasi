package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
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

// kubectlLogsFunc shells out to kubectl for non-query kind log retrieval.
// Replaceable in tests to avoid requiring a real kubectl binary.
var kubectlLogsFunc = func(ctx context.Context, args ...string) (string, error) {
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput() //nolint:gosec // kubectl CLI with caller-controlled arguments
	if err != nil {
		return "", fmt.Errorf("kubectl %s: %w\noutput: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func newLogsCommand() *cobra.Command {
	var componentID string
	var kind string

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
			// Non-query kinds: retrieve logs via kubectl label selector.
			if kind != "query" {
				selector := fmt.Sprintf("drasi.io/kind=%s,drasi.io/component=%s", kind, componentID)
				kubectlArgs := []string{"logs", "-l", selector, "-n", "drasi-system", "--tail=100"}
				if kubeContext != "" {
					kubectlArgs = append([]string{"--context", kubeContext}, kubectlArgs...)
				}
				logOutput, kubectlErr := kubectlLogsFunc(cmd.Context(), kubectlArgs...)
				if kubectlErr != nil {
					code := errorCodeFromError(kubectlErr, output.ERR_DRASI_CLI_ERROR)
					return writeCommandError(cmd, code, kubectlErr.Error(),
						"Ensure kubectl is configured and the component is running.", format, output.ExitCodes[code])
				}

				if format == output.FormatJSON {
					payload := map[string]any{"status": "ok", "kind": kind, "component": componentID, "logs": strings.TrimSpace(logOutput)}
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
					return nil
				}
				if strings.TrimSpace(logOutput) == "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No logs found for %s (%s).\n", componentID, kind)
					return nil
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(logOutput, "\n"))
				return nil
			}

			// Query kind: existing drasi watch path.
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

	return cmd
}
