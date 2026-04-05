package deployment

import "github.com/azure/azd.extensions.drasi/internal/config"

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
	panic("not implemented")
}
