package deployment

import (
	"context"
	"time"
)

const (
	// PerComponentTimeout is the deadline for a single component to reach Online status.
	PerComponentTimeout = 5 * time.Minute

	// TotalDeployTimeout is the max time allowed for a full deploy operation.
	TotalDeployTimeout = 15 * time.Minute
)

// WithPerComponentTimeout wraps ctx with a per-component deadline.
func WithPerComponentTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	panic("not implemented")
}

// WithTotalDeployTimeout wraps ctx with the total deploy deadline.
func WithTotalDeployTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	panic("not implemented")
}
