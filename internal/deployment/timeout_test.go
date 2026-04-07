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
		name     string
		override time.Duration
		want     time.Duration
	}{
		{name: "deadline is set in future using default", override: 0, want: PerComponentTimeout},
		{name: "deadline uses override", override: 2 * time.Second, want: 2 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := WithPerComponentTimeout(context.Background(), tt.override)
			defer cancel()

			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			remaining := time.Until(deadline)
			assert.True(t, remaining > 0)
			assert.LessOrEqual(t, remaining, tt.want)
			assert.Greater(t, remaining, tt.want-time.Second)
		})
	}
}

func TestWithTotalDeployTimeout_DeadlineSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		override time.Duration
		want     time.Duration
	}{
		{name: "total deadline is set using default", override: 0, want: TotalDeployTimeout},
		{name: "total deadline uses override", override: 3 * time.Second, want: 3 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := WithTotalDeployTimeout(context.Background(), tt.override)
			defer cancel()

			deadline, ok := ctx.Deadline()
			require.True(t, ok)
			remaining := time.Until(deadline)
			assert.True(t, remaining > 0)
			assert.LessOrEqual(t, remaining, tt.want)
			assert.Greater(t, remaining, tt.want-time.Second)
		})
	}
}
