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

// EnvServiceClient is the exported equivalent of envServiceClient, used by
// integration-test stubs that live in external packages.
type EnvServiceClient interface {
	GetValue(ctx context.Context, envName, key string) (string, error)
	SetValue(ctx context.Context, envName, key, value string) error
}

// exportedEnvStateClient adapts an EnvServiceClient into the unexported envStateClient
// interface so external test stubs can be passed to NewStateManager via NewStateManagerFromClient.
type exportedEnvStateClient struct {
	svc EnvServiceClient
}

func (e *exportedEnvStateClient) environment() envServiceClient { return e.svc }

// StateManager reads and writes component hashes via the azd environment gRPC service.
type StateManager struct {
	client  envStateClient
	envName string
}

// NewStateManager creates a StateManager.
func NewStateManager(client envStateClient, envName string) *StateManager {
	return &StateManager{client: client, envName: envName}
}

// NewStateManagerFromClient creates a StateManager from an exported EnvServiceClient.
// This constructor exists for integration tests in external packages that need to
// provide a stub without access to the unexported envStateClient interface.
func NewStateManagerFromClient(svc EnvServiceClient, envName string) *StateManager {
	return &StateManager{client: &exportedEnvStateClient{svc: svc}, envName: envName}
}

// ReadHash returns the stored hash for a component key, or "" if absent.
func (s *StateManager) ReadHash(ctx context.Context, stateKey string) (string, error) {
	return s.client.environment().GetValue(ctx, s.envName, stateKey)
}

// WriteHash persists a hash for a component key.
func (s *StateManager) WriteHash(ctx context.Context, stateKey, hash string) error {
	return s.client.environment().SetValue(ctx, s.envName, stateKey, hash)
}
