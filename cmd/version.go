package cmd

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newVersionCommand(outputFormat *string) *cobra.Command {
	return azdext.NewVersionCommand(extensionID, extensionVersion, outputFormat)
}
