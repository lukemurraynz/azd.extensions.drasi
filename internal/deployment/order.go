package deployment

import (
	"sort"

	"github.com/azure/azd.extensions.drasi/internal/config"
)

var deployPriority = map[string]int{
	"source":          0,
	"continuousquery": 1,
	"middleware":      2,
	"reaction":        3,
}

var deletePriority = map[string]int{
	"reaction":        0,
	"middleware":      1,
	"continuousquery": 2,
	"source":          3,
}

// SortForDeploy returns actions ordered: sources → queries → middleware → reactions.
func SortForDeploy(actions []ComponentAction, manifest *config.ResolvedManifest) []ComponentAction {
	result := make([]ComponentAction, len(actions))
	copy(result, actions)
	sort.SliceStable(result, func(i, j int) bool {
		return deployPriority[result[i].Kind] < deployPriority[result[j].Kind]
	})
	return result
}

// SortForDelete returns actions in reverse order: reactions → middleware → queries → sources.
func SortForDelete(actions []ComponentAction, manifest *config.ResolvedManifest) []ComponentAction {
	result := make([]ComponentAction, len(actions))
	copy(result, actions)
	sort.SliceStable(result, func(i, j int) bool {
		return deletePriority[result[i].Kind] < deletePriority[result[j].Kind]
	})
	return result
}
