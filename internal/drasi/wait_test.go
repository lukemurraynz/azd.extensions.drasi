package drasi

import (
	"context"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitOnline_ImmediatelyOnline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		id     string
		timeout int
		stdout string
	}{
		{name: "component already online", kind: "query", id: "alerts", timeout: 5, stdout: "STATUS: Online"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}}
			client := newTestClient(runner)

			err := client.WaitOnline(context.Background(), tt.kind, tt.id, tt.timeout)

			require.NoError(t, err)
			assert.Equal(t, 1, runner.callCount)
		})
	}
}

func TestWaitOnline_OnlineAfterPolls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kind     string
		id       string
		timeout  int
		responses []runnerResponse
	}{
		{
			name:    "component becomes online after polling",
			kind:    "query",
			id:      "alerts",
			timeout: 5,
			responses: []runnerResponse{
				{stdout: "STATUS: Pending"},
				{stdout: "STATUS: Pending"},
				{stdout: "STATUS: Online"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: tt.responses}
			client := newTestClient(runner)

			err := client.WaitOnline(context.Background(), tt.kind, tt.id, tt.timeout)

			require.NoError(t, err)
			assert.Equal(t, 3, runner.callCount)
		})
	}
}

func TestWaitOnline_ExceedsTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kind     string
		id       string
		timeout  int
		responses []runnerResponse
		want     string
	}{
		{
			name:    "timeout returns component timeout error",
			kind:    "query",
			id:      "alerts",
			timeout: 1,
			responses: []runnerResponse{
				{stdout: "STATUS: Pending"},
				{stdout: "STATUS: Pending"},
			},
			want: output.ERR_COMPONENT_TIMEOUT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: tt.responses}
			client := newTestClient(runner)

			err := client.WaitOnline(context.Background(), tt.kind, tt.id, tt.timeout)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
