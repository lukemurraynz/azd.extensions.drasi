package validation

import (
	"fmt"
	"path/filepath"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
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

	ValidateManifestSchema(manifest, manifestFile, result)

	if envName != "" {
		envFile := filepath.ToSlash(filepath.Join("environments", envName+".yaml"))
		if rel, ok := manifest.Environments[envName]; ok {
			envFile = filepath.ToSlash(rel)
		}
		ValidateEnvironmentOverlaySchema(resolved.Environment, envFile, result)
	}

	for _, warning := range warnings {
		code := warning.Code
		remediation := "Remove undeclared overlay parameters or align them with manifest expectations."
		switch warning.Code {
		case config.WarningInvalidComponentExclusion:
			remediation = "Use component kinds source, continuousquery, reaction, or middleware in environment overlay exclusions."
		case config.WarningMissingComponentExclusion:
			remediation = "Update the environment overlay to exclude only components that exist in the base manifest."
		case "":
			code = config.WarningUndeclaredOverlayParameter
		}

		result.Add(ValidationIssue{
			Level:       LevelWarning,
			File:        filepath.ToSlash(manifestFile),
			Line:        1,
			Code:        code,
			Message:     warning.Message,
			Remediation: remediation,
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
