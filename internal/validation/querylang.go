package validation

import (
	"fmt"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/azure/azd.extensions.drasi/internal/output"
)

// ValidateQueryLanguages ensures each query declares a query language.
func ValidateQueryLanguages(resolved *config.ResolvedManifest, result *ValidationResult) {
	for _, query := range resolved.Queries {
		if query.QueryLanguage == "" {
			result.Add(ValidationIssue{
				Level:       LevelError,
				File:        query.FilePath,
				Line:        query.Line,
				Code:        output.ERR_MISSING_QUERY_LANGUAGE,
				Message:     fmt.Sprintf("ContinuousQuery %q is missing required field 'queryLanguage'", query.ID),
				Remediation: "Add `queryLanguage: Cypher` or `queryLanguage: GQL`.",
			})
		}
	}
}
