package drasi

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescribeComponent_ParsesStatusAndErrorReason(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		id     string
		stdout string
		want   *ComponentDetail
	}{
		{
			name:   "describe output parsed into detail",
			kind:   "query",
			id:     "alerts",
			stdout: "ID: alerts\nKind: query\nStatus: TerminalError\nErrorReason: parse failure\n",
			want: &ComponentDetail{
				ID:          "alerts",
				Kind:        "query",
				Status:      "TerminalError",
				ErrorReason: "parse failure",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}})

			got, err := client.DescribeComponent(context.Background(), tt.kind, tt.id)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDescribeComponent_NotFound_TypedError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		id     string
		stderr string
	}{
		{name: "missing component returns typed not found error", kind: "query", id: "missing", stderr: "component not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: 1}}})

			got, err := client.DescribeComponent(context.Background(), tt.kind, tt.id)

			require.Error(t, err)
			assert.Nil(t, got)

			var notFoundErr *ComponentNotFoundError
			assert.True(t, errors.As(err, &notFoundErr))
		})
	}
}
