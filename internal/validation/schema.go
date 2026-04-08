package validation

import (
	"encoding/json"
	"fmt"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
)

var (
	loadSchemaOnce sync.Once
	loadSchemaErr  error
	schemas        map[string]*jsonschema.Schema
)

// ValidateSourceSchema validates a Source against the JSON schema.
func ValidateSourceSchema(source config.Source, result *ValidationResult) {
	validateSchema("schema/source.schema.json", source, source.FilePath, source.Line, "source", result)
}

// ValidateQuerySchema validates a ContinuousQuery against the JSON schema.
func ValidateQuerySchema(query config.ContinuousQuery, result *ValidationResult) {
	validateSchema("schema/continuousquery.schema.json", query, query.FilePath, query.Line, "query", result)
}

// ValidateReactionSchema validates a Reaction against the JSON schema.
func ValidateReactionSchema(reaction config.Reaction, result *ValidationResult) {
	validateSchema("schema/reaction.schema.json", reaction, reaction.FilePath, reaction.Line, "reaction", result)
}

// ValidateMiddlewareSchema validates a Middleware against the JSON schema.
func ValidateMiddlewareSchema(middleware config.Middleware, result *ValidationResult) {
	validateSchema("schema/middleware.schema.json", middleware, middleware.FilePath, middleware.Line, "middleware", result)
}

// ValidateEnvironmentOverlaySchema validates an environment overlay against the JSON schema.
func ValidateEnvironmentOverlaySchema(environment config.Environment, file string, result *ValidationResult) {
	validateSchema("schema/environment-overlay.schema.json", environment, file, 1, "environment overlay", result)
}

// ValidateManifestSchema validates a DrasiManifest against the JSON schema.
func ValidateManifestSchema(manifest config.DrasiManifest, file string, result *ValidationResult) {
	validateSchema("schema/manifest.schema.json", manifestSchemaValue(manifest), file, 1, "manifest", result)
}

func manifestSchemaValue(manifest config.DrasiManifest) map[string]any {
	m := map[string]any{}

	if manifest.APIVersion != "" {
		m["APIVersion"] = manifest.APIVersion
	}
	if manifest.Includes != nil {
		includes := make([]map[string]any, 0, len(manifest.Includes))
		for _, include := range manifest.Includes {
			item := map[string]any{}
			if include.Kind != "" {
				item["Kind"] = include.Kind
			}
			if include.Pattern != "" {
				item["Pattern"] = include.Pattern
			}
			includes = append(includes, item)
		}
		m["Includes"] = includes
	}
	if manifest.Environments != nil {
		m["Environments"] = manifest.Environments
	}
	if manifest.FeatureFlags != nil {
		m["FeatureFlags"] = manifest.FeatureFlags
	}
	if manifest.SecretMappings != nil {
		secretMappings := make([]map[string]any, 0, len(manifest.SecretMappings))
		for _, mapping := range manifest.SecretMappings {
			item := map[string]any{}
			if mapping.VaultName != "" {
				item["vaultName"] = mapping.VaultName
			}
			if mapping.SecretName != "" {
				item["secretName"] = mapping.SecretName
			}
			if mapping.K8sSecret != "" {
				item["k8sSecret"] = mapping.K8sSecret
			}
			if mapping.K8sKey != "" {
				item["k8sKey"] = mapping.K8sKey
			}
			if mapping.Namespace != "" {
				item["namespace"] = mapping.Namespace
			}
			secretMappings = append(secretMappings, item)
		}
		m["SecretMappings"] = secretMappings
	}

	return m
}

func validateSchema(schemaName string, value any, file string, line int, componentKind string, result *ValidationResult) {
	schemaMap, err := compiledSchemas()
	if err != nil {
		result.Add(ValidationIssue{
			Level:       LevelError,
			File:        file,
			Line:        line,
			Code:        output.ERR_VALIDATION_FAILED,
			Message:     fmt.Sprintf("schema initialization failed for %s: %s", componentKind, err),
			Remediation: "Check embedded validation schemas.",
		})
		return
	}

	schema := schemaMap[schemaName]
	if schema == nil {
		result.Add(ValidationIssue{
			Level:       LevelError,
			File:        file,
			Line:        line,
			Code:        output.ERR_VALIDATION_FAILED,
			Message:     fmt.Sprintf("schema %s is not available", schemaName),
			Remediation: "Restore the missing validation schema.",
		})
		return
	}

	if err := schema.Validate(structToMap(value)); err != nil {
		result.Add(ValidationIssue{
			Level:       LevelError,
			File:        file,
			Line:        line,
			Code:        output.ERR_VALIDATION_FAILED,
			Message:     fmt.Sprintf("schema validation failed for %s: %s", componentKind, err),
			Remediation: "Update the file to match the expected schema.",
		})
	}
}

func compiledSchemas() (map[string]*jsonschema.Schema, error) {
	loadSchemaOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		schemas = make(map[string]*jsonschema.Schema, 6)
		for _, name := range []string{
			"schema/source.schema.json",
			"schema/continuousquery.schema.json",
			"schema/reaction.schema.json",
			"schema/middleware.schema.json",
			"schema/environment-overlay.schema.json",
			"schema/manifest.schema.json",
		} {
			data, err := config.SchemaFS.ReadFile(name)
			if err != nil {
				loadSchemaErr = fmt.Errorf("read %s: %w", name, err)
				return
			}
			// AddResource requires a pre-decoded JSON value (any), not an io.Reader.
			var doc any
			if err := json.Unmarshal(data, &doc); err != nil {
				loadSchemaErr = fmt.Errorf("parse %s: %w", name, err)
				return
			}
			if err := compiler.AddResource(name, doc); err != nil {
				loadSchemaErr = fmt.Errorf("add %s: %w", name, err)
				return
			}
		}

		for _, name := range []string{
			"schema/source.schema.json",
			"schema/continuousquery.schema.json",
			"schema/reaction.schema.json",
			"schema/middleware.schema.json",
			"schema/environment-overlay.schema.json",
			"schema/manifest.schema.json",
		} {
			compiled, err := compiler.Compile(name)
			if err != nil {
				loadSchemaErr = fmt.Errorf("compile %s: %w", name, err)
				return
			}
			schemas[name] = compiled
		}
	})

	return schemas, loadSchemaErr
}

func structToMap(value any) map[string]any {
	data, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]any{}
	}
	// Remove runtime-only fields and null values so optional schema properties
	// that are unset do not produce false validation errors.
	delete(m, "FilePath")
	delete(m, "Line")
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
	}
	return m
}
