package cmd

import (
	"context"
	"fmt"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

type statusDrasiClient interface {
	CheckVersion(ctx context.Context) error
	ListComponents(ctx context.Context, kind string) ([]drasi.ComponentSummary, error)
	ListComponentsInContext(ctx context.Context, kind, kubeContext string) ([]drasi.ComponentSummary, error)
}

var newStatusDrasiClient = func() statusDrasiClient {
	return drasi.NewClient()
}

func newStatusCommand() *cobra.Command {
	var kind string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Drasi component status",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			selectedKind := kind
			kubeContext, err := resolvedKubeContextForCommand(cmd.Context(), cmd, "")
			if err != nil {
				code := errorCodeFromError(err, output.ERR_AKS_CONTEXT_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(), "Set the target azd environment and ensure AZURE_AKS_CONTEXT is present.", format, output.ExitCodes[code])
			}

			if selectedKind == "" {
				selectedKind = "source"
			}
			if selectedKind == "continuousquery" {
				selectedKind = "query"
			}

			client := newStatusDrasiClient()
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

			var resources []drasi.ComponentSummary
			if kubeContext == "" {
				resources, err = client.ListComponents(cmd.Context(), selectedKind)
			} else {
				resources, err = client.ListComponentsInContext(cmd.Context(), selectedKind, kubeContext)
			}
			if err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Check cluster connectivity and Drasi runtime health, then retry.",
					format,
					output.ExitCodes[code],
				)
			}

			if format == output.FormatJSON {
				payload := map[string]any{
					"status":      "ok",
					"kind":        selectedKind,
					"components":  resources,
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			if len(resources) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No %s components found.\n", selectedKind)
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(resources, output.FormatTable))
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "", "Component kind to query (source, continuousquery, middleware, reaction)")

	return cmd
}
