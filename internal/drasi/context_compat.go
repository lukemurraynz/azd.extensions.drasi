package drasi

import "strings"

func isUnsupportedContextFlagError(stderr string) bool {
	return strings.Contains(strings.ToLower(stderr), "unknown flag: --context")
}
