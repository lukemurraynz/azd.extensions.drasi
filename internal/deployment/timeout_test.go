package deployment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerComponentTimeout_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want time.Duration
	}{
		{name: "constant equals five minutes", want: 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, PerComponentTimeout)
		})
	}
}

func TestTotalDeployTimeout_Value(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want time.Duration
	}{
		{name: "constant equals fifteen minutes", want: 15 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, TotalDeployTimeout)
		})
	}
}

func TestWithPerComponentTimeout_DeadlineSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "deadline is set in future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := WithPerComponentTimeout(context.Background())
			defer cancel()

			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			assert.True(t, deadline.After(time.Now()))
		})
	}
}

func TestWithTotalDeployTimeout_DeadlineSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "total deadline is set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := WithTotalDeployTimeout(context.Background())
			defer cancel()

			_, ok := ctx.Deadline()
			require.True(t, ok)
		})
	}
}
