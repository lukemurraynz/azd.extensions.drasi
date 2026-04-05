package scaffold

import "embed"

// templateFS contains all scaffold template files.
// NOTE: The "all:" prefix is required to include files and directories whose
// names begin with "." (e.g. .vscode/launch.json). Without it, Go's embed
// directive silently skips dot-prefixed entries.
//
//go:embed all:templates
var templateFS embed.FS
