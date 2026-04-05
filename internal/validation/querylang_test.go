package validation_test

import (
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestQueryLangValidator_CypherPasses(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "q1", QueryLanguage: "Cypher", FilePath: "q1.yaml", Line: 1},
		},
	}
	result := &validation.ValidationResult{}
	validation.ValidateQueryLanguages(resolved, result)
	assert.False(t, result.HasErrors())
}

func TestQueryLangValidator_GQLPasses(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "q1", QueryLanguage: "GQL", FilePath: "q1.yaml", Line: 1},
		},
	}
	result := &validation.ValidationResult{}
	validation.ValidateQueryLanguages(resolved, result)
	assert.False(t, result.HasErrors())
}

func TestQueryLangValidator_MissingLanguage(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "q1", QueryLanguage: "", FilePath: "q1.yaml", Line: 3},
		},
	}
	result := &validation.ValidationResult{}
	validation.ValidateQueryLanguages(resolved, result)
	assert.True(t, result.HasErrors())
	assert.Equal(t, "ERR_MISSING_QUERY_LANGUAGE", result.Issues[0].Code)
	assert.Equal(t, "q1.yaml", result.Issues[0].File)
	assert.Equal(t, 3, result.Issues[0].Line)
}

func TestQueryLangValidator_AccumulatesAllErrors(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "q1", QueryLanguage: "", FilePath: "q1.yaml", Line: 1},
			{ID: "q2", QueryLanguage: "", FilePath: "q2.yaml", Line: 1},
		},
	}
	result := &validation.ValidationResult{}
	validation.ValidateQueryLanguages(resolved, result)
	assert.Len(t, result.Issues, 2)
}
