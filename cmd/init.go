package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/scaffold"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var templateName string
	var force bool
	var outputDir string
	var envName string

	command := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Drasi project scaffold",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")

			files, err := scaffold.Scaffold(templateName, outputDir, force)
			if err != nil {
				return err
			}

			if outputFormat == string(output.FormatJSON) {
				payload := map[string]any{"status": "ok", "files": files}
				data, _ := json.MarshalIndent(payload, "", "  ")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Initialized Drasi project in %s\n", outputDir)
			for _, f := range files {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  Created %s\n", f)
			}
			return nil
		},
	}

	command.Flags().StringVar(&templateName, "template", "blank", "Template to scaffold. One of: blank, blank-terraform, event-hub-routing, query-subscription, postgresql-source")
	command.Flags().BoolVar(&force, "force", false, "Overwrite existing files")
	command.Flags().StringVar(&outputDir, "output-dir", ".", "Directory to scaffold into")
	// NOTE: --environment is accepted for future use (environment overlay selection).
	command.Flags().StringVar(&envName, "environment", "", "Environment name")

	return command
}
