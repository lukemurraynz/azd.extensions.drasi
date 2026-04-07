package config_test

import (
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestSchemaValidator_ValidSource(t *testing.T) {
	t.Parallel()
	src := config.Source{
		APIVersion: "v1",
		Kind:       "Source",
		ID:         "my-source",
		SourceKind: "PostgreSQL",
		FilePath:   "sources/my-source.yaml",
		Line:       1,
	}
	result := &validation.ValidationResult{}
	validation.ValidateSourceSchema(src, result)
	assert.False(t, result.HasErrors())
}

func TestSchemaValidator_SourceMissingKind(t *testing.T) {
	t.Parallel()
	src := config.Source{
		APIVersion: "v1",
		ID:         "my-source",
		SourceKind: "PostgreSQL",
		FilePath:   "sources/my-source.yaml",
		Line:       1,
	}
	result := &validation.ValidationResult{}
	validation.ValidateSourceSchema(src, result)
	assert.True(t, result.HasErrors())
}

func TestSchemaValidator_ValidQuery(t *testing.T) {
	t.Parallel()
	q := config.ContinuousQuery{
		APIVersion:    "v1",
		Kind:          "ContinuousQuery",
		ID:            "my-query",
		QueryLanguage: "Cypher",
		Spec: config.QuerySpec{
			Query: "MATCH (n) RETURN n",
		},
		FilePath: "queries/my-query.yaml",
		Line:     1,
	}
	result := &validation.ValidationResult{}
	validation.ValidateQuerySchema(q, result)
	assert.False(t, result.HasErrors())
}

func TestSchemaValidator_ValidReaction(t *testing.T) {
	t.Parallel()
	r := config.Reaction{
		APIVersion:   "v1",
		Kind:         "Reaction",
		ID:           "my-reaction",
		ReactionKind: "Dapr",
		FilePath:     "reactions/my-reaction.yaml",
		Line:         1,
	}
	result := &validation.ValidationResult{}
	validation.ValidateReactionSchema(r, result)
	assert.False(t, result.HasErrors())
}
