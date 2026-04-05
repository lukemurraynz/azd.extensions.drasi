package validation

import (
	"fmt"
	"path/filepath"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/azure/azd.extensions.drasi/internal/output"
)

// Validate runs the full configuration validation pipeline.
func Validate(dir, manifestFile, envName string) (*ValidationResult, error) {
	manifest, sources, queries, reactions, middlewares, err := config.LoadManifest(dir, manifestFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", output.ERR_NO_MANIFEST, err)
	}

	resolved, warnings, err := config.ResolveManifest(manifest, sources, queries, reactions, middlewares, dir, envName)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{}
	for _, warning := range warnings {
		result.Add(ValidationIssue{
			Level:       LevelWarning,
			File:        filepath.ToSlash(manifestFile),
			Line:        1,
			Code:        "WARN_UNDECLARED_PARAMETER",
			Message:     warning,
			Remediation: "Remove undeclared overlay parameters or align them with manifest expectations.",
		})
	}

	for _, source := range resolved.Sources {
		ValidateSourceSchema(source, result)
	}
	for _, query := range resolved.Queries {
		ValidateQuerySchema(query, result)
	}
	for _, reaction := range resolved.Reactions {
		ValidateReactionSchema(reaction, result)
	}
	for _, middleware := range resolved.Middlewares {
		ValidateMiddlewareSchema(middleware, result)
	}

	ValidateReferences(&resolved, result)
	ValidateDependencyGraph(&resolved, result)
	ValidateQueryLanguages(&resolved, result)

	return result, nil
}
