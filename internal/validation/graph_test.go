package validation_test

import (
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestGraphValidator_NoCycle(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "q1", FilePath: "q1.yaml", Line: 1, Reactions: []string{"r1"}},
			{ID: "q2", FilePath: "q2.yaml", Line: 1, Reactions: []string{"r2"}},
		},
		Reactions: []config.Reaction{{ID: "r1"}, {ID: "r2"}},
	}
	result := &validation.ValidationResult{}
	validation.ValidateDependencyGraph(resolved, result)
	assert.False(t, result.HasErrors())
}

func TestGraphValidator_DirectCycle(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "query-a", FilePath: "qa.yaml", Line: 1, Reactions: []string{"query-b"}},
			{ID: "query-b", FilePath: "qb.yaml", Line: 1, Reactions: []string{"query-a"}},
		},
		Reactions: []config.Reaction{},
	}
	result := &validation.ValidationResult{}
	validation.ValidateDependencyGraph(resolved, result)
	assert.True(t, result.HasErrors())
	assert.Equal(t, "ERR_CIRCULAR_DEPENDENCY", result.Issues[0].Code)
}

func TestGraphValidator_DisconnectedGraph(t *testing.T) {
	t.Parallel()
	resolved := &config.ResolvedManifest{
		Queries: []config.ContinuousQuery{
			{ID: "q1", FilePath: "q1.yaml", Line: 1},
			{ID: "q2", FilePath: "q2.yaml", Line: 1},
		},
	}
	result := &validation.ValidationResult{}
	validation.ValidateDependencyGraph(resolved, result)
	assert.False(t, result.HasErrors())
}
