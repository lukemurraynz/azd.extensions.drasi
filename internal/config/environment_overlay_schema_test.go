package config_test

import (
	"encoding/json"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEnvironmentOverlaySchema_ValidComponentExclusions(t *testing.T) {
	t.Parallel()

	schema := compileEnvironmentOverlaySchema(t)
	overlay := []byte(`
name: dev
components:
  exclude:
    - kind: source
      id: postgres-source
    - kind: ContinuousQuery
      id: order-changes
`)

	var environment config.Environment
	require.NoError(t, yaml.Unmarshal(overlay, &environment))

	err := schema.Validate(structToJSONMap(t, environment))
	assert.NoError(t, err)
}

func TestEnvironmentOverlaySchema_InvalidComponentKindFails(t *testing.T) {
	t.Parallel()

	schema := compileEnvironmentOverlaySchema(t)
	overlay := []byte(`
name: dev
components:
  exclude:
    - kind: widget
      id: postgres-source
`)

	var environment config.Environment
	require.NoError(t, yaml.Unmarshal(overlay, &environment))

	err := schema.Validate(structToJSONMap(t, environment))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind")
}

func TestSchemaValidator_EnvironmentOverlayInvalidComponentKind(t *testing.T) {
	t.Parallel()

	environment := config.Environment{
		Name: "dev",
		Components: config.Components{
			Exclude: []config.ComponentRef{{Kind: "widget", ID: "postgres-source"}},
		},
	}

	result := &validation.ValidationResult{}
	validation.ValidateEnvironmentOverlaySchema(environment, "environments/dev.yaml", result)
	require.True(t, result.HasErrors())
	assert.Contains(t, result.Issues[0].Message, "schema validation failed for environment overlay")
}

func compileEnvironmentOverlaySchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	compiler := jsonschema.NewCompiler()
	data, err := config.SchemaFS.ReadFile("schema/environment-overlay.schema.json")
	require.NoError(t, err)

	var doc any
	require.NoError(t, json.Unmarshal(data, &doc))
	require.NoError(t, compiler.AddResource("schema/environment-overlay.schema.json", doc))

	schema, err := compiler.Compile("schema/environment-overlay.schema.json")
	require.NoError(t, err)
	return schema
}

func structToJSONMap(t *testing.T, value any) map[string]any {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(data, &doc))
	return doc
}
