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
) (ResolvedManifest, []string, error) {
	resolved := ResolvedManifest{
		Sources:      append([]Source(nil), sources...),
		Queries:      append([]ContinuousQuery(nil), queries...),
		Reactions:    append([]Reaction(nil), reactions...),
		Middlewares:  append([]Middleware(nil), middlewares...),
		FeatureFlags: manifest.FeatureFlags,
	}

	warnings := make([]string, 0)
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
					warnings = append(warnings, fmt.Sprintf("overlay parameter %q in environment %q has no base declaration", key, envName))
				}
			}
		}
		// NOTE: An environment with no overlay entry in the manifest is valid — it uses the
		// base manifest as-is with no parameter overrides. This allows `azd drasi deploy`
		// to succeed on default environments (e.g. "dev") that require no customisation.
	}

	sort.Slice(resolved.Sources, func(i, j int) bool { return resolved.Sources[i].ID < resolved.Sources[j].ID })
	sort.Slice(resolved.Queries, func(i, j int) bool { return resolved.Queries[i].ID < resolved.Queries[j].ID })
	sort.Slice(resolved.Reactions, func(i, j int) bool { return resolved.Reactions[i].ID < resolved.Reactions[j].ID })
	sort.Slice(resolved.Middlewares, func(i, j int) bool { return resolved.Middlewares[i].ID < resolved.Middlewares[j].ID })

	return resolved, warnings, nil
}
