package deployment

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployLock_AcquireReleaseCycle(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	// Acquire sets a JSON value.
	err := lock.Acquire(ctx)
	require.NoError(t, err)

	raw := client.store[deployLockKey]
	assert.NotEmpty(t, raw, "lock value must not be empty after Acquire")

	var payload lockPayload
	require.NoError(t, json.Unmarshal([]byte(raw), &payload), "lock value must be valid JSON")
	assert.Greater(t, payload.PID, 0, "PID must be positive")
	assert.NotEmpty(t, payload.Started, "started timestamp must be set")

	// Release clears the lock.
	err = lock.Release(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", client.store[deployLockKey], "lock must be empty after Release")

	// Second acquire succeeds after release.
	err = lock.Acquire(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, client.store[deployLockKey], "second Acquire must succeed after Release")
}

func TestDeployLock_IsStale_FreshLock(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	err := lock.Acquire(ctx)
	require.NoError(t, err)

	stale, err := lock.IsStale(ctx, 30*time.Minute)
	require.NoError(t, err)
	assert.False(t, stale, "freshly acquired lock must not be stale")
}

func TestDeployLock_IsStale_OldLock(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	// Write a lock payload with an old timestamp.
	oldPayload := lockPayload{
		PID:     12345,
		Started: time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339),
	}
	data, err := json.Marshal(oldPayload)
	require.NoError(t, err)
	client.store[deployLockKey] = string(data)

	// Use 1ns maxAge to guarantee the lock is stale.
	stale, err := lock.IsStale(ctx, 1*time.Nanosecond)
	require.NoError(t, err)
	assert.True(t, stale, "lock with old timestamp must be stale")
}

func TestDeployLock_IsStale_LegacyTrueValue(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	// Simulate the legacy bare "true" lock value.
	client.store[deployLockKey] = "true"

	stale, err := lock.IsStale(ctx, 30*time.Minute)
	require.NoError(t, err)
	assert.True(t, stale, "legacy bare \"true\" value must be treated as stale")
}

func TestDeployLock_IsStale_EmptyValue(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	// No lock set — empty value.
	stale, err := lock.IsStale(ctx, 30*time.Minute)
	require.NoError(t, err)
	assert.True(t, stale, "empty lock value must be treated as stale (no active lock)")
}

func TestDeployLock_ForceRelease(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	// Set a lock payload.
	err := lock.Acquire(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, client.store[deployLockKey])

	// ForceRelease clears it regardless.
	err = lock.ForceRelease(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", client.store[deployLockKey], "ForceRelease must clear lock regardless of content")

	// Also works with legacy "true" value.
	client.store[deployLockKey] = "true"
	err = lock.ForceRelease(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", client.store[deployLockKey], "ForceRelease must clear legacy lock value")
}

func TestDeployLock_DeferredRelease(t *testing.T) {
	t.Parallel()

	client := &mockEnvClient{store: make(map[string]string)}
	state := NewStateManager(client, "test-env")
	lock := NewDeployLock(state)
	ctx := context.Background()

	// Simulate a function that acquires the lock and defers release.
	simulatedDeploy := func() error {
		if err := lock.Acquire(ctx); err != nil {
			return err
		}
		defer func() {
			_ = lock.Release(ctx)
		}()

		// Simulate an error mid-deploy.
		return assert.AnError
	}

	err := simulatedDeploy()
	require.Error(t, err)

	// Lock must be released even though the function returned an error.
	assert.Equal(t, "", client.store[deployLockKey],
		"deferred Release must clear lock even on error path")
}
