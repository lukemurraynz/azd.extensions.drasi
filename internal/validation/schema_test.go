package validation_test

import (
	"encoding/json"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMiddlewareSchema_ValidMiddleware(t *testing.T) {
	t.Parallel()

	result := &validation.ValidationResult{}
	validation.ValidateMiddlewareSchema(config.Middleware{
		APIVersion:     "v1",
		Kind:           "Middleware",
		ID:             "audit-middleware",
		MiddlewareKind: "Audit",
		Config: map[string]config.Value{
			"endpoint": {StringValue: "https://example.invalid"},
		},
		FilePath: "middleware/audit.yaml",
		Line:     12,
	}, result)

	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Issues)
}

func TestValidateMiddlewareSchema_MissingKind(t *testing.T) {
	t.Parallel()

	result := &validation.ValidationResult{}
	validation.ValidateMiddlewareSchema(config.Middleware{
		APIVersion:     "v1",
		ID:             "audit-middleware",
		MiddlewareKind: "Audit",
		FilePath:       "middleware/audit.yaml",
		Line:           7,
	}, result)

	require.NotEmpty(t, result.Issues)
	assert.True(t, result.HasErrors())
	assert.Contains(t, result.Issues[0].Message, "schema validation failed for middleware")
}

func TestValidateMiddlewareSchema_MissingName(t *testing.T) {
	t.Parallel()

	schema := compileSchema(t, "schema/middleware.schema.json")
	err := schema.Validate(map[string]any{
		"APIVersion":     "v1",
		"Kind":           "Middleware",
		"MiddlewareKind": "Audit",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID")
}

func TestValidateSourceSchema_ValidSource(t *testing.T) {
	t.Parallel()

	result := &validation.ValidationResult{}
	validation.ValidateSourceSchema(config.Source{
		APIVersion: "v1",
		Kind:       "Source",
		ID:         "orders-source",
		SourceKind: "PostgreSQL",
		Properties: map[string]config.Value{
			"host": {StringValue: "localhost"},
		},
		FilePath: "sources/orders.yaml",
		Line:     3,
	}, result)

	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Issues)
}

func TestValidateQuerySchema_ValidQuery(t *testing.T) {
	t.Parallel()

	result := &validation.ValidationResult{}
	validation.ValidateQuerySchema(config.ContinuousQuery{
		APIVersion:    "v1",
		Kind:          "ContinuousQuery",
		ID:            "orders-query",
		QueryLanguage: "Cypher",
		Spec: config.QuerySpec{
			Query: "MATCH (o:Order) RETURN o",
		},
		FilePath: "queries/orders.yaml",
		Line:     5,
	}, result)

	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Issues)
}

func TestValidateReactionSchema_ValidReaction(t *testing.T) {
	t.Parallel()

	result := &validation.ValidationResult{}
	validation.ValidateReactionSchema(config.Reaction{
		APIVersion:   "v1",
		Kind:         "Reaction",
		ID:           "notify-reaction",
		ReactionKind: "Dapr",
		Config: map[string]config.Value{
			"topic": {StringValue: "orders"},
		},
		FilePath: "reactions/notify.yaml",
		Line:     8,
	}, result)

	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Issues)
}

func TestValidateEnvironmentOverlaySchema_ValidOverlay(t *testing.T) {
	t.Parallel()

	result := &validation.ValidationResult{}
	validation.ValidateEnvironmentOverlaySchema(config.Environment{
		Name: "dev",
		Parameters: map[string]string{
			"logLevel": "debug",
		},
		Components: config.Components{
			Exclude: []config.ComponentRef{{
				Kind: "source",
				ID:   "orders-source",
			}},
		},
	}, "environments/dev.yaml", result)

	assert.False(t, result.HasErrors())
	assert.Empty(t, result.Issues)
}

func compileSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()

	compiler := jsonschema.NewCompiler()
	data, err := config.SchemaFS.ReadFile(name)
	require.NoError(t, err)

	var doc any
	require.NoError(t, json.Unmarshal(data, &doc))
	require.NoError(t, compiler.AddResource(name, doc))

	schema, err := compiler.Compile(name)
	require.NoError(t, err)

	return schema
}
