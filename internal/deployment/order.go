package deployment

import "github.com/azure/azd.extensions.drasi/internal/config"

// SortForDeploy returns actions ordered: sources → queries → middleware → reactions.
func SortForDeploy(actions []ComponentAction, manifest *config.ResolvedManifest) []ComponentAction {
	panic("not implemented")
}

// SortForDelete returns actions in reverse order: reactions → middleware → queries → sources.
func SortForDelete(actions []ComponentAction, manifest *config.ResolvedManifest) []ComponentAction {
	panic("not implemented")
}
