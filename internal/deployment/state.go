package deployment

import "context"

// envStateClient is the consumer-side interface for azd environment state access.
// azdext.AzdClient is a concrete struct (not an interface), so we define this
// local interface for testability.
type envStateClient interface {
	environment() envServiceClient
}

// envServiceClient covers the gRPC methods we actually use.
type envServiceClient interface {
	GetValue(ctx context.Context, envName, key string) (string, error)
	SetValue(ctx context.Context, envName, key, value string) error
}

// StateManager reads and writes component hashes via the azd environment gRPC service.
type StateManager struct {
	client  envStateClient
	envName string
}

// NewStateManager creates a StateManager.
func NewStateManager(client envStateClient, envName string) *StateManager {
	return &StateManager{client: client, envName: envName}
}

// ReadHash returns the stored hash for a component key, or "" if absent.
func (s *StateManager) ReadHash(ctx context.Context, stateKey string) (string, error) {
	return s.client.environment().GetValue(ctx, s.envName, stateKey)
}

// WriteHash persists a hash for a component key.
func (s *StateManager) WriteHash(ctx context.Context, stateKey, hash string) error {
	return s.client.environment().SetValue(ctx, s.envName, stateKey, hash)
}
