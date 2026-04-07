package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
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

// allComponentKinds is the ordered list of kinds queried when no --kind flag is provided.
var allComponentKinds = []string{"source", "continuousquery", "middleware", "reaction"}

// kindDisplayNames maps kind identifiers to human-friendly section headers.
var kindDisplayNames = map[string]string{
	"source":          "Sources",
	"continuousquery": "Queries",
	"middleware":      "Middleware",
	"reaction":        "Reactions",
}

// kindJSONKeys maps kind identifiers to JSON object keys for the all-kinds payload.
var kindJSONKeys = map[string]string{
	"source":          "sources",
	"continuousquery": "queries",
	"middleware":      "middleware",
	"reaction":        "reactions",
}

// wireKindForDrasiCLI maps the canonical kind to the value the drasi CLI expects.
// The drasi CLI uses "query" instead of "continuousquery".
func wireKindForDrasiCLI(kind string) string {
	if kind == "continuousquery" {
		return "query"
	}
	return kind
}

// nonNilResources ensures a nil slice is returned as an empty slice so JSON
// serialization produces [] instead of null.
func nonNilResources(r []drasi.ComponentSummary) []drasi.ComponentSummary {
	if r == nil {
		return []drasi.ComponentSummary{}
	}
	return r
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

			// When no --kind flag is provided, query all component kinds.
			if selectedKind == "" {
				return statusAllKinds(cmd, client, kubeContext, format)
			}

			// Single-kind mode: keep existing behavior unchanged.
			wireKind := wireKindForDrasiCLI(selectedKind)

			var resources []drasi.ComponentSummary
			if kubeContext == "" {
				resources, err = client.ListComponents(cmd.Context(), wireKind)
			} else {
				resources, err = client.ListComponentsInContext(cmd.Context(), wireKind, kubeContext)
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
					"status":     "ok",
					"kind":       wireKind,
					"components": resources,
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			if len(resources) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No %s components found.\n", wireKind)
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(resources, output.FormatTable))
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "", "Component kind to query (source, continuousquery, middleware, reaction)")

	return cmd
}

// statusAllKinds queries and displays all 4 component kinds.
func statusAllKinds(cmd *cobra.Command, client statusDrasiClient, kubeContext string, format output.OutputFormat) error {
	type kindResult struct {
		kind      string
		resources []drasi.ComponentSummary
	}

	results := make([]kindResult, 0, len(allComponentKinds))
	for _, k := range allComponentKinds {
		wireKind := wireKindForDrasiCLI(k)

		var (
			res []drasi.ComponentSummary
			err error
		)
		if kubeContext == "" {
			res, err = client.ListComponents(cmd.Context(), wireKind)
		} else {
			res, err = client.ListComponentsInContext(cmd.Context(), wireKind, kubeContext)
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
		results = append(results, kindResult{kind: k, resources: res})
	}

	if format == output.FormatJSON {
		payload := map[string]any{
			"status": "ok",
		}
		for _, r := range results {
			payload[kindJSONKeys[r.kind]] = nonNilResources(r.resources)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
		return nil
	}

	// Table output: section header per kind.
	var sb strings.Builder
	for i, r := range results {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(kindDisplayNames[r.kind])
		sb.WriteString(":\n")
		if len(r.resources) == 0 {
			fmt.Fprintf(&sb, "  No %s components found.\n", r.kind)
		} else {
			sb.WriteString(output.Format(r.resources, output.FormatTable))
		}
	}
	_, _ = fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}
