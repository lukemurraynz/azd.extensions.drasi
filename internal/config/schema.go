package config

import "embed"

//go:embed schema/*.json
var SchemaFS embed.FS
