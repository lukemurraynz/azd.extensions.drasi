package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadManifest reads a manifest and expands includes relative to dir.
func LoadManifest(dir, manifestFile string) (
	manifest DrasiManifest,
	sources []Source,
	queries []ContinuousQuery,
	reactions []Reaction,
	middlewares []Middleware,
	err error,
) {
	manifestPath := filepath.Join(dir, manifestFile)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return manifest, nil, nil, nil, nil, fmt.Errorf("reading manifest %s: %w", manifestPath, err)
	}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, nil, nil, nil, nil, fmt.Errorf("parsing manifest %s: %w", manifestPath, err)
	}

	for _, include := range manifest.Includes {
		matches, err := expandIncludePattern(dir, include.Pattern)
		if err != nil {
			return manifest, nil, nil, nil, nil, fmt.Errorf("resolving include %q: %w", include.Pattern, err)
		}

		for _, match := range matches {
			rel, err := filepath.Rel(dir, match)
			if err != nil {
				return manifest, nil, nil, nil, nil, fmt.Errorf("rel path for %s: %w", match, err)
			}

			switch include.Kind {
			case "sources":
				source, err := loadSource(match, rel)
				if err != nil {
					return manifest, nil, nil, nil, nil, err
				}
				sources = append(sources, source)
			case "queries":
				query, err := loadQuery(match, rel)
				if err != nil {
					return manifest, nil, nil, nil, nil, err
				}
				queries = append(queries, query)
			case "reactions":
				reaction, err := loadReaction(match, rel)
				if err != nil {
					return manifest, nil, nil, nil, nil, err
				}
				reactions = append(reactions, reaction)
			case "middlewares":
				middleware, err := loadMiddleware(match, rel)
				if err != nil {
					return manifest, nil, nil, nil, nil, err
				}
				middlewares = append(middlewares, middleware)
			}
		}
	}

	return manifest, sources, queries, reactions, middlewares, nil
}

func loadSource(path, rel string) (Source, error) {
	node, err := decodeNode(path)
	if err != nil {
		return Source{}, err
	}
	var source Source
	if err := node.Decode(&source); err != nil {
		return Source{}, fmt.Errorf("decoding source %s: %w", path, err)
	}
	source.FilePath = filepath.ToSlash(rel)
	source.Line = node.Line
	return source, nil
}

func loadQuery(path, rel string) (ContinuousQuery, error) {
	node, err := decodeNode(path)
	if err != nil {
		return ContinuousQuery{}, err
	}
	var query ContinuousQuery
	if err := node.Decode(&query); err != nil {
		return ContinuousQuery{}, fmt.Errorf("decoding query %s: %w", path, err)
	}
	query.FilePath = filepath.ToSlash(rel)
	query.Line = node.Line
	return query, nil
}

func loadReaction(path, rel string) (Reaction, error) {
	node, err := decodeNode(path)
	if err != nil {
		return Reaction{}, err
	}
	var reaction Reaction
	if err := node.Decode(&reaction); err != nil {
		return Reaction{}, fmt.Errorf("decoding reaction %s: %w", path, err)
	}
	reaction.FilePath = filepath.ToSlash(rel)
	reaction.Line = node.Line
	return reaction, nil
}

func loadMiddleware(path, rel string) (Middleware, error) {
	node, err := decodeNode(path)
	if err != nil {
		return Middleware{}, err
	}
	var middleware Middleware
	if err := node.Decode(&middleware); err != nil {
		return Middleware{}, fmt.Errorf("decoding middleware %s: %w", path, err)
	}
	middleware.FilePath = filepath.ToSlash(rel)
	middleware.Line = node.Line
	return middleware, nil
}

func decodeNode(path string) (*yaml.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if document.Kind == yaml.DocumentNode && len(document.Content) > 0 {
		return document.Content[0], nil
	}
	return &document, nil
}

func expandIncludePattern(dir, pattern string) ([]string, error) {
	normalizedPattern := filepath.ToSlash(pattern)
	if !strings.Contains(normalizedPattern, "**") {
		matches, err := filepath.Glob(filepath.Join(dir, filepath.FromSlash(pattern)))
		if err != nil {
			return nil, err
		}
		sort.Strings(matches)
		return matches, nil
	}

	root := includeRoot(dir, normalizedPattern)
	matches := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if matchDoubleStarPattern(normalizedPattern, filepath.ToSlash(rel)) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

func includeRoot(dir, pattern string) string {
	idx := strings.Index(pattern, "**")
	if idx < 0 {
		return dir
	}
	prefix := strings.TrimSuffix(pattern[:idx], "/")
	if prefix == "" {
		return dir
	}
	return filepath.Join(dir, filepath.FromSlash(prefix))
}

func matchDoubleStarPattern(pattern, value string) bool {
	patternParts := splitPattern(pattern)
	valueParts := splitPattern(value)
	return matchPatternParts(patternParts, valueParts)
}

func splitPattern(value string) []string {
	parts := strings.Split(value, "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}

func matchPatternParts(patternParts, valueParts []string) bool {
	if len(patternParts) == 0 {
		return len(valueParts) == 0
	}

	if patternParts[0] == "**" {
		if matchPatternParts(patternParts[1:], valueParts) {
			return true
		}
		for i := range valueParts {
			if matchPatternParts(patternParts[1:], valueParts[i+1:]) {
				return true
			}
		}
		return false
	}

	if len(valueParts) == 0 {
		return false
	}

	matched, err := filepath.Match(patternParts[0], valueParts[0])
	if err != nil || !matched {
		return false
	}
	return matchPatternParts(patternParts[1:], valueParts[1:])
}
