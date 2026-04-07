package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const deployLockKey = "DRASI_DEPLOY_IN_PROGRESS"

// lockPayload is the JSON structure stored in the deploy lock state key.
type lockPayload struct {
	PID     int    `json:"pid"`
	Started string `json:"started"`
}

// DeployLock provides crash-safe deployment locking via azd environment state.
// It stores a JSON timestamp payload instead of a bare "true" so stale locks
// from crashed processes can be detected and force-released.
type DeployLock struct {
	state *StateManager
}

// NewDeployLock creates a DeployLock backed by the given StateManager.
func NewDeployLock(state *StateManager) *DeployLock {
	return &DeployLock{state: state}
}

// Acquire writes the deploy lock with a JSON payload containing the current PID and timestamp.
func (l *DeployLock) Acquire(ctx context.Context) error {
	payload := lockPayload{
		PID:     os.Getpid(),
		Started: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling deploy lock payload: %w", err)
	}
	return l.state.WriteHash(ctx, deployLockKey, string(data))
}

// Release clears the deploy lock by writing an empty string.
func (l *DeployLock) Release(ctx context.Context) error {
	return l.state.WriteHash(ctx, deployLockKey, "")
}

// IsStale checks whether the current lock value represents a stale or absent lock.
// Returns true (stale) when:
//   - the lock value is empty (no active lock)
//   - the lock value is the legacy bare "true" (backwards compat — cannot determine age)
//   - the lock payload timestamp is older than maxAge
//
// Returns false when the lock holds a valid JSON payload with a fresh timestamp.
func (l *DeployLock) IsStale(ctx context.Context, maxAge time.Duration) (bool, error) {
	val, err := l.state.ReadHash(ctx, deployLockKey)
	if err != nil {
		return false, fmt.Errorf("reading deploy lock: %w", err)
	}

	if val == "" {
		return true, nil
	}

	if val == "true" {
		// Legacy bare "true" value — treat as stale because we cannot determine age.
		return true, nil
	}

	var payload lockPayload
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		// Unrecognised format — treat as stale for safety.
		return true, nil
	}

	started, err := time.Parse(time.RFC3339, payload.Started)
	if err != nil {
		// Malformed timestamp — treat as stale.
		return true, nil
	}

	return time.Since(started) > maxAge, nil
}

// ForceRelease unconditionally clears the deploy lock regardless of its current content.
func (l *DeployLock) ForceRelease(ctx context.Context) error {
	return l.state.WriteHash(ctx, deployLockKey, "")
}
