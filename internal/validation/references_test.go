package validation_test

import (
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReferenceValidator_ValidRefs(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "my-source"}},
		Queries: []config.ContinuousQuery{{
			ID: "my-query", FilePath: "q.yaml", Line: 1,
			Sources: []config.SourceRef{{ID: "my-source"}},
		}},
		Reactions: []config.Reaction{{ID: "my-reaction"}},
	}
	result := &validation.ValidationResult{}
	validation.ValidateReferences(resolved, result)
	assert.False(t, result.HasErrors())
}

func TestReferenceValidator_UnknownSourceRef(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Sources: []config.Source{},
		Queries: []config.ContinuousQuery{{
			ID: "my-query", FilePath: "q.yaml", Line: 5,
			Sources: []config.SourceRef{{ID: "missing-source"}},
		}},
	}
	result := &validation.ValidationResult{}
	validation.ValidateReferences(resolved, result)
	assert.True(t, result.HasErrors())
	assert.Len(t, result.Issues, 1)
	assert.Equal(t, "ERR_MISSING_REFERENCE", result.Issues[0].Code)
	assert.Equal(t, "q.yaml", result.Issues[0].File)
	assert.Equal(t, 5, result.Issues[0].Line)
}

func TestReferenceValidator_UnknownReactionRef(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Sources: []config.Source{{ID: "my-source"}},
		Queries: []config.ContinuousQuery{{
			ID: "my-query", FilePath: "q.yaml", Line: 2,
			Sources:   []config.SourceRef{{ID: "my-source"}},
			Reactions: []string{"missing-reaction"},
		}},
		Reactions: []config.Reaction{},
	}
	result := &validation.ValidationResult{}
	validation.ValidateReferences(resolved, result)
	assert.True(t, result.HasErrors())
	assert.Equal(t, "ERR_MISSING_REFERENCE", result.Issues[0].Code)
}

func TestReferenceValidator_MultipleErrors(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Sources: []config.Source{},
		Queries: []config.ContinuousQuery{
			{ID: "q1", FilePath: "q1.yaml", Line: 1, Sources: []config.SourceRef{{ID: "bad1"}}},
			{ID: "q2", FilePath: "q2.yaml", Line: 1, Sources: []config.SourceRef{{ID: "bad2"}}},
		},
	}
	result := &validation.ValidationResult{}
	validation.ValidateReferences(resolved, result)
	assert.Len(t, result.Issues, 2)
}
