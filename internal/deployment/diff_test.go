package deployment

import (
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff_UnchangedHash_NoOp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		newHashes     []config.ComponentHash
		existingState map[string]string
		want          ComponentAction
	}{
		{
			name: "same hash in state is noop",
			newHashes: []config.ComponentHash{{
				Kind: "source",
				ID:   "alerts",
				Hash: "abc123",
			}},
			existingState: map[string]string{
				"DRASI_HASH_SOURCE_alerts": "abc123",
			},
			want: ComponentAction{Kind: "source", ID: "alerts", Hash: "abc123", Action: ActionNoOp},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Diff(tt.newHashes, tt.existingState)
			require.Len(t, got, 1)
			assert.Equal(t, tt.want, got[0])
		})
	}
}

func TestDiff_ChangedHash_DeleteThenApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		newHashes     []config.ComponentHash
		existingState map[string]string
		want          ComponentAction
	}{
		{
			name: "different hash in state requires delete then apply",
			newHashes: []config.ComponentHash{{
				Kind: "continuousquery",
				ID:   "severity-escalation",
				Hash: "new-hash",
			}},
			existingState: map[string]string{
				"DRASI_HASH_CONTINUOUSQUERY_severity_escalation": "old-hash",
			},
			want: ComponentAction{Kind: "continuousquery", ID: "severity-escalation", Hash: "new-hash", Action: ActionDeleteThenApply},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Diff(tt.newHashes, tt.existingState)
			require.Len(t, got, 1)
			assert.Equal(t, tt.want, got[0])
		})
	}
}

func TestDiff_MissingInState_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		newHashes     []config.ComponentHash
		existingState map[string]string
		want          ComponentAction
	}{
		{
			name: "missing key in state creates component",
			newHashes: []config.ComponentHash{{
				Kind: "reaction",
				ID:   "alerts-http",
				Hash: "abc123",
			}},
			existingState: map[string]string{},
			want:          ComponentAction{Kind: "reaction", ID: "alerts-http", Hash: "abc123", Action: ActionCreate},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Diff(tt.newHashes, tt.existingState)
			require.Len(t, got, 1)
			assert.Equal(t, tt.want, got[0])
		})
	}
}

func TestDiff_StateKeyFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hash config.ComponentHash
		want string
	}{
		{
			name: "continuous query state key format",
			hash: config.ComponentHash{Kind: "continuousquery", ID: "my-id", Hash: "abc123"},
			want: "DRASI_HASH_CONTINUOUSQUERY_my_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.hash.StateKey())
		})
	}
}
