package drasi

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteComponent_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind string
		id   string
	}{
		{name: "successful delete returns nil", kind: "query", id: "alerts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: "deleted"}}})

			err := client.DeleteComponent(context.Background(), tt.kind, tt.id)

			require.NoError(t, err)
			assert.True(t, true)
		})
	}
}

func TestDeleteComponent_NotFound_NonFatal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		id     string
		stderr string
	}{
		{name: "not found is non fatal", kind: "query", id: "missing", stderr: "component not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: 1}}})

			err := client.DeleteComponent(context.Background(), tt.kind, tt.id)

			require.NoError(t, err)
			assert.True(t, true)
		})
	}
}

func TestDeleteComponent_OtherError_Propagates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind string
		id   string
		err  error
	}{
		{name: "other errors propagate", kind: "query", id: "alerts", err: errors.New("permission denied")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{err: tt.err}}})

			err := client.DeleteComponent(context.Background(), tt.kind, tt.id)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
