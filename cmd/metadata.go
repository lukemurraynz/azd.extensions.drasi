package cmd

import (
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/spf13/cobra"
)

func newMetadataCommand() *cobra.Command {
	return azdext.NewMetadataCommand(metadataSchemaVersion, extensionID, NewRootCommand)
}
