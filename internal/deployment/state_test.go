package deployment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnvClient struct {
	store map[string]string
}

func (m *mockEnvClient) environment() envServiceClient { return &mockEnvService{store: m.store} }

type mockEnvService struct {
	store map[string]string
}

func (m *mockEnvService) GetValue(_ context.Context, _, key string) (string, error) {
	return m.store[key], nil
}

func (m *mockEnvService) SetValue(_ context.Context, _, key, value string) error {
	m.store[key] = value
	return nil
}

func TestStateManager_ReadHash_MissingKey_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stateKey string
	}{
		{name: "missing key returns empty string", stateKey: "DRASI_HASH_SOURCE_missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockEnvClient{store: make(map[string]string)}
			manager := NewStateManager(client, "test-env")

			got, err := manager.ReadHash(context.Background(), tt.stateKey)
			require.NoError(t, err)
			assert.Equal(t, "", got)
		})
	}
}

func TestStateManager_WriteHash_PersistsValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stateKey string
		hash     string
	}{
		{name: "write then read persists value", stateKey: "DRASI_HASH_SOURCE_alerts", hash: "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockEnvClient{store: make(map[string]string)}
			manager := NewStateManager(client, "test-env")

			require.NoError(t, manager.WriteHash(context.Background(), tt.stateKey, tt.hash))
			got, err := manager.ReadHash(context.Background(), tt.stateKey)
			require.NoError(t, err)
			assert.Equal(t, tt.hash, got)
		})
	}
}

func TestStateManager_RoundTrip_PreservesHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stateKey string
		hash     string
	}{
		{name: "round trip preserves exact hash", stateKey: "DRASI_HASH_REACTION_alerts-http", hash: "sha256:deadbeef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockEnvClient{store: make(map[string]string)}
			manager := NewStateManager(client, "test-env")

			require.NoError(t, manager.WriteHash(context.Background(), tt.stateKey, tt.hash))
			got, err := manager.ReadHash(context.Background(), tt.stateKey)
			require.NoError(t, err)
			assert.Equal(t, tt.hash, got)
		})
	}
}
