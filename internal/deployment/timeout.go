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
// If override is zero, uses the default PerComponentTimeout.
func WithPerComponentTimeout(ctx context.Context, override time.Duration) (context.Context, context.CancelFunc) {
	d := PerComponentTimeout
	if override > 0 {
		d = override
	}
	return context.WithTimeout(ctx, d)
}

// WithTotalDeployTimeout wraps ctx with the total deploy deadline.
// If override is zero, uses the default TotalDeployTimeout.
func WithTotalDeployTimeout(ctx context.Context, override time.Duration) (context.Context, context.CancelFunc) {
	d := TotalDeployTimeout
	if override > 0 {
		d = override
	}
	return context.WithTimeout(ctx, d)
}
