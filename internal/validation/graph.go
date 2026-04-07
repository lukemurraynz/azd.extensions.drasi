package validation

import (
	"fmt"
	"strings"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
)

// ValidateDependencyGraph detects cycles between queries that reference other queries as reactions.
func ValidateDependencyGraph(resolved *config.ResolvedManifest, result *ValidationResult) {
	queryIDs := make(map[string]struct{}, len(resolved.Queries))
	queryByID := make(map[string]config.ContinuousQuery, len(resolved.Queries))
	edges := make(map[string][]string, len(resolved.Queries))

	for _, query := range resolved.Queries {
		queryIDs[query.ID] = struct{}{}
		queryByID[query.ID] = query
	}

	for _, query := range resolved.Queries {
		for _, reactionRef := range query.Reactions {
			if _, ok := queryIDs[reactionRef]; ok {
				edges[query.ID] = append(edges[query.ID], reactionRef)
			}
		}
	}

	const (
		stateUnvisited = iota
		stateVisiting
		stateDone
	)
	state := make(map[string]int, len(resolved.Queries))
	cycle := make([]string, 0)
	found := false

	var dfs func(string, []string)
	dfs = func(node string, path []string) {
		if found {
			return
		}

		state[node] = stateVisiting
		path = append(path, node)

		for _, next := range edges[node] {
			switch state[next] {
			case stateVisiting:
				start := 0
				for i, item := range path {
					if item == next {
						start = i
						break
					}
				}
				cycle = append(append([]string(nil), path[start:]...), next)
				found = true
				return
			case stateUnvisited:
				dfs(next, path)
			}
			if found {
				return
			}
		}

		state[node] = stateDone
	}

	for _, query := range resolved.Queries {
		if state[query.ID] == stateUnvisited {
			dfs(query.ID, nil)
		}
		if found {
			break
		}
	}

	if !found {
		return
	}

	query := queryByID[cycle[0]]
	result.Add(ValidationIssue{
		Level:       LevelError,
		File:        query.FilePath,
		Line:        query.Line,
		Code:        output.ERR_CIRCULAR_DEPENDENCY,
		Message:     fmt.Sprintf("circular dependency detected: %s", strings.Join(cycle, " -> ")),
		Remediation: "Remove the query-to-query reaction cycle.",
	})
}
