package deployment

import "github.com/lukemurraynz/azd.extensions.drasi/internal/config"

// Action describes what to do with a component.
type Action int

const (
	ActionNoOp Action = iota
	ActionCreate
	ActionDeleteThenApply
)

// ComponentAction pairs a component with its calculated action.
type ComponentAction struct {
	Kind   string
	ID     string
	Hash   string
	Action Action
}

// Diff calculates what actions are needed to reconcile the desired state
// (newHashes) against what is already deployed (existingState).
// existingState maps ComponentHash.StateKey() -> hash string.
func Diff(newHashes []config.ComponentHash, existingState map[string]string) []ComponentAction {
	result := make([]ComponentAction, 0, len(newHashes))
	for _, h := range newHashes {
		existing, ok := existingState[h.StateKey()]
		var action Action
		switch {
		case !ok || existing == "":
			action = ActionCreate
		case existing == h.Hash:
			action = ActionNoOp
		default:
			action = ActionDeleteThenApply
		}
		result = append(result, ComponentAction{
			Kind:   h.Kind,
			ID:     h.ID,
			Hash:   h.Hash,
			Action: action,
		})
	}
	return result
}
