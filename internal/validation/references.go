package validation

import (
	"fmt"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
)

// ValidateReferences checks query references against declared components.
func ValidateReferences(resolved *config.ResolvedManifest, result *ValidationResult) {
	sourceIDs := make(map[string]struct{}, len(resolved.Sources))
	for _, source := range resolved.Sources {
		sourceIDs[source.ID] = struct{}{}
	}

	reactionIDs := make(map[string]struct{}, len(resolved.Reactions))
	for _, reaction := range resolved.Reactions {
		reactionIDs[reaction.ID] = struct{}{}
	}

	queryIDs := make(map[string]struct{}, len(resolved.Queries))
	for _, query := range resolved.Queries {
		queryIDs[query.ID] = struct{}{}
	}

	for _, query := range resolved.Queries {
		for _, sourceRef := range query.Sources {
			if _, ok := sourceIDs[sourceRef.ID]; !ok {
				result.Add(ValidationIssue{
					Level:       LevelError,
					File:        query.FilePath,
					Line:        query.Line,
					Code:        output.ERR_MISSING_REFERENCE,
					Message:     fmt.Sprintf("query %q references unknown source %q", query.ID, sourceRef.ID),
					Remediation: fmt.Sprintf("Declare a source with id %q or update the query reference.", sourceRef.ID),
				})
			}
		}

		for _, reactionRef := range query.Reactions {
			_, isReaction := reactionIDs[reactionRef]
			_, isQuery := queryIDs[reactionRef]
			if !isReaction && !isQuery {
				result.Add(ValidationIssue{
					Level:       LevelError,
					File:        query.FilePath,
					Line:        query.Line,
					Code:        output.ERR_MISSING_REFERENCE,
					Message:     fmt.Sprintf("query %q references unknown reaction %q", query.ID, reactionRef),
					Remediation: fmt.Sprintf("Declare a reaction with id %q or update the query reference.", reactionRef),
				})
			}
		}
	}
}
