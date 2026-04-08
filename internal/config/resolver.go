package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ResolveManifest applies the requested environment overlay and sorts components deterministically.
func ResolveManifest(
	manifest DrasiManifest,
	sources []Source,
	queries []ContinuousQuery,
	reactions []Reaction,
	middlewares []Middleware,
	dir string,
	envName string,
) (ResolvedManifest, []OverlayWarning, error) {
	resolved := ResolvedManifest{
		Sources:        append([]Source(nil), sources...),
		Queries:        append([]ContinuousQuery(nil), queries...),
		Reactions:      append([]Reaction(nil), reactions...),
		Middlewares:    append([]Middleware(nil), middlewares...),
		FeatureFlags:   manifest.FeatureFlags,
		SecretMappings: manifest.SecretMappings,
		ManifestDir:    dir,
	}

	warnings := make([]OverlayWarning, 0)
	if envName != "" {
		envRelPath, ok := manifest.Environments[envName]
		if ok {
			envPath := filepath.Join(dir, envRelPath)
			data, err := os.ReadFile(envPath)
			if err != nil {
				return resolved, nil, fmt.Errorf("reading environment file %s: %w", envPath, err)
			}

			var environment Environment
			if err := yaml.Unmarshal(data, &environment); err != nil {
				return resolved, nil, fmt.Errorf("parsing environment file %s: %w", envPath, err)
			}
			resolved.Environment = environment

			for key := range environment.Parameters {
				// NOTE: The manifest model has no formal parameter declaration block yet.
				// Until that exists, treat snake_case overlay keys as likely undeclared.
				if strings.ContainsRune(key, '_') {
					warnings = append(warnings, OverlayWarning{
						Code:    WarningUndeclaredOverlayParameter,
						Message: fmt.Sprintf("overlay parameter %q in environment %q has no base declaration", key, envName),
					})
				}
			}

			warnings = append(warnings, validateExcludedComponents(environment.Components.Exclude, sources, queries, reactions, middlewares, envName)...)
		}
		// NOTE: An environment with no overlay entry in the manifest is valid — it uses the
		// base manifest as-is with no parameter overrides. This allows `azd drasi deploy`
		// to succeed on default environments (e.g. "dev") that require no customisation.
	}

	sort.Slice(resolved.Sources, func(i, j int) bool { return resolved.Sources[i].ID < resolved.Sources[j].ID })
	sort.Slice(resolved.Queries, func(i, j int) bool { return resolved.Queries[i].ID < resolved.Queries[j].ID })
	sort.Slice(resolved.Reactions, func(i, j int) bool { return resolved.Reactions[i].ID < resolved.Reactions[j].ID })
	sort.Slice(resolved.Middlewares, func(i, j int) bool { return resolved.Middlewares[i].ID < resolved.Middlewares[j].ID })

	if len(resolved.Environment.Components.Exclude) > 0 {
		excludeSet := make(map[string]bool, len(resolved.Environment.Components.Exclude))
		for _, ref := range resolved.Environment.Components.Exclude {
			key := componentKey(ref.Kind, ref.ID)
			excludeSet[key] = true
		}

		resolved.Sources = filterSources(resolved.Sources, excludeSet)
		resolved.Queries = filterQueries(resolved.Queries, excludeSet)
		resolved.Reactions = filterReactions(resolved.Reactions, excludeSet)
		resolved.Middlewares = filterMiddlewares(resolved.Middlewares, excludeSet)
	}

	return resolved, warnings, nil
}

func validateExcludedComponents(
	excluded []ComponentRef,
	sources []Source,
	queries []ContinuousQuery,
	reactions []Reaction,
	middlewares []Middleware,
	envName string,
) []OverlayWarning {
	if len(excluded) == 0 {
		return nil
	}

	knownKinds := map[string]bool{
		"source":          true,
		"continuousquery": true,
		"reaction":        true,
		"middleware":      true,
	}

	existing := make(map[string]bool, len(sources)+len(queries)+len(reactions)+len(middlewares))
	for _, source := range sources {
		existing[componentKey("source", source.ID)] = true
	}
	for _, query := range queries {
		existing[componentKey("continuousquery", query.ID)] = true
	}
	for _, reaction := range reactions {
		existing[componentKey("reaction", reaction.ID)] = true
	}
	for _, middleware := range middlewares {
		existing[componentKey("middleware", middleware.ID)] = true
	}

	warnings := make([]OverlayWarning, 0)
	for _, ref := range excluded {
		kind := strings.ToLower(ref.Kind)
		if !knownKinds[kind] {
			warnings = append(warnings, OverlayWarning{
				Code:    WarningInvalidComponentExclusion,
				Message: fmt.Sprintf("overlay component exclusion kind %q in environment %q is invalid; valid kinds are source, continuousquery, reaction, middleware", ref.Kind, envName),
			})
			continue
		}

		key := componentKey(kind, ref.ID)
		if !existing[key] {
			warnings = append(warnings, OverlayWarning{
				Code:    WarningMissingComponentExclusion,
				Message: fmt.Sprintf("overlay component exclusion %q for kind %q in environment %q does not match any component in the base manifest", ref.ID, ref.Kind, envName),
			})
		}
	}

	return warnings
}

func componentKey(kind, id string) string {
	return strings.ToLower(kind) + "/" + id
}

func filterSources(sources []Source, excludeSet map[string]bool) []Source {
	result := make([]Source, 0, len(sources))
	for _, source := range sources {
		if !excludeSet[componentKey("source", source.ID)] {
			result = append(result, source)
		}
	}
	return result
}

func filterQueries(queries []ContinuousQuery, excludeSet map[string]bool) []ContinuousQuery {
	result := make([]ContinuousQuery, 0, len(queries))
	for _, query := range queries {
		if !excludeSet[componentKey("continuousquery", query.ID)] {
			result = append(result, query)
		}
	}
	return result
}

func filterReactions(reactions []Reaction, excludeSet map[string]bool) []Reaction {
	result := make([]Reaction, 0, len(reactions))
	for _, reaction := range reactions {
		if !excludeSet[componentKey("reaction", reaction.ID)] {
			result = append(result, reaction)
		}
	}
	return result
}

func filterMiddlewares(middlewares []Middleware, excludeSet map[string]bool) []Middleware {
	result := make([]Middleware, 0, len(middlewares))
	for _, middleware := range middlewares {
		if !excludeSet[componentKey("middleware", middleware.ID)] {
			result = append(result, middleware)
		}
	}
	return result
}
