package drasi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyFile_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		stdout string
	}{
		{name: "apply succeeds and captures stdout", path: "manifests/query.yaml", stdout: "applied successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}}
			client := newTestClient(runner)

			err := client.ApplyFile(context.Background(), tt.path)

			require.NoError(t, err)
			assert.Equal(t, 1, runner.callCount)
		})
	}
}

func TestApplyFile_FailedApply_PropagatesStderr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		stderr string
	}{
		{name: "apply failure surfaces stderr", path: "manifests/query.yaml", stderr: "failed to apply manifest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: 1}}}
			client := newTestClient(runner)

			err := client.ApplyFile(context.Background(), tt.path)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.stderr)
		})
	}
}

func TestApplyFile_ContextCancellation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: "cancelled context returns error", path: "manifests/query.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			client := newTestClient(&mockRunner{})

			err := client.ApplyFile(ctx, tt.path)

			require.Error(t, err)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}
