package validation

import (
	"fmt"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
)

// ValidateQueryLanguages ensures each query declares a non-empty query body.
// A query is considered valid if either Spec.Query (populated from YAML) or
// QueryLanguage (populated directly in tests) is non-empty.
func ValidateQueryLanguages(resolved *config.ResolvedManifest, result *ValidationResult) {
	for _, query := range resolved.Queries {
		if query.Spec.Query == "" && query.QueryLanguage == "" {
			result.Add(ValidationIssue{
				Level:       LevelError,
				File:        query.FilePath,
				Line:        query.Line,
				Code:        output.ERR_MISSING_QUERY_LANGUAGE,
				Message:     fmt.Sprintf("ContinuousQuery %q is missing required field 'query'", query.ID),
				Remediation: "Add a `query:` field under `spec:` with a valid Cypher query.",
			})
		}
	}
}
