package main

import (
	"log/slog"
	"os"

	"github.com/azure/azd.extensions.drasi/cmd"
	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	cmd.SetVersion(version)
	azdext.Run(cmd.NewRootCommand())
}
