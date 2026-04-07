package main

import (
	"log/slog"
	"os"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/lukemurraynz/azd.extensions.drasi/cmd"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	cmd.SetVersion(version)
	azdext.Run(cmd.NewRootCommand())
}
