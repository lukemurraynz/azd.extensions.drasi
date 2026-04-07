package observability_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/observability"
	"github.com/stretchr/testify/assert"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
)

type codedError struct {
	message string
	code    string
}

func (e codedError) Error() string {
	return e.message
}

func (e codedError) Code() string {
	return e.code
}

func TestRecordCommandExecution_SuccessDoesNotPanic(t *testing.T) {
	t.Parallel()

	meter := metricnoop.NewMeterProvider().Meter("test")
	assert.NotPanics(t, func() {
		observability.RecordCommandExecution(context.Background(), meter, "deploy", 500*time.Millisecond, nil)
	})
}

func TestRecordCommandExecution_ErrorDoesNotPanic(t *testing.T) {
	t.Parallel()

	meter := metricnoop.NewMeterProvider().Meter("test")
	assert.NotPanics(t, func() {
		observability.RecordCommandExecution(context.Background(), meter, "deploy", 250*time.Millisecond, errors.New("some error"))
	})
}

func TestRecordCommandExecution_NilMeterDoesNotPanic(t *testing.T) {
	t.Parallel()

	// nil meter = no APPLICATIONINSIGHTS_CONNECTION_STRING — must be a silent no-op.
	assert.NotPanics(t, func() {
		observability.RecordCommandExecution(context.Background(), nil, "deploy", 100*time.Millisecond, nil)
	})
}

func TestRecordCommandExecution_CodedErrorDoesNotPanic(t *testing.T) {
	t.Parallel()

	meter := metricnoop.NewMeterProvider().Meter("test")
	assert.NotPanics(t, func() {
		observability.RecordCommandExecution(
			context.Background(),
			meter,
			"deploy",
			125*time.Millisecond,
			codedError{message: "coded failure", code: "E_TEST"},
		)
	})
}
