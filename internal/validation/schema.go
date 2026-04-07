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
		schemas = make(map[string]*jsonschema.Schema, 5)
		for _, name := range []string{
			"schema/source.schema.json",
			"schema/continuousquery.schema.json",
			"schema/reaction.schema.json",
			"schema/middleware.schema.json",
			"schema/environment-overlay.schema.json",
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
