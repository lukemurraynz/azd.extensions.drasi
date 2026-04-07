package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// RecordCommandExecution records command execution metrics using the provided meter.
// If meter is nil, all recording is skipped silently.
// Metrics emitted:
//   - drasi.command.invocations (counter): tags command + status (success|error)
//   - drasi.command.duration_ms (histogram): tag command
//   - drasi.command.errors (counter): tags command + error_code (only on error)
//
// All telemetry is opt-in: metrics are only exported when APPLICATIONINSIGHTS_CONNECTION_STRING is set.
// Command argument values are never recorded — only the command name (privacy).
func RecordCommandExecution(ctx context.Context, meter metric.Meter, command string, duration time.Duration, err error) {
	if meter == nil {
		return
	}

	status := "success"
	errorCode := ""
	if err != nil {
		status = "error"
		errorCode = errorCodeFromErr(err)
	}

	// drasi.command.invocations counter
	invocations, counterErr := meter.Int64Counter("drasi.command.invocations",
		metric.WithDescription("Total number of Drasi command invocations."),
	)
	if counterErr == nil {
		invocations.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("command", command),
				attribute.String("status", status),
			),
		)
	}

	// drasi.command.duration_ms histogram
	durationHist, histErr := meter.Int64Histogram("drasi.command.duration_ms",
		metric.WithDescription("Duration of Drasi command execution in milliseconds."),
		metric.WithUnit("ms"),
	)
	if histErr == nil {
		durationHist.Record(ctx, duration.Milliseconds(),
			metric.WithAttributes(
				attribute.String("command", command),
			),
		)
	}

	// drasi.command.errors counter — only recorded on error.
	if err != nil {
		errors, errCounterErr := meter.Int64Counter("drasi.command.errors",
			metric.WithDescription("Total number of Drasi command errors."),
		)
		if errCounterErr == nil {
			errors.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("command", command),
					attribute.String("error_code", errorCode),
				),
			)
		}
	}
}

// errorCodeFromErr extracts a string error code from an error.
// Returns "UNKNOWN" if the error does not implement a Code() method.
func errorCodeFromErr(err error) string {
	if err == nil {
		return ""
	}
	type coder interface {
		Code() string
	}
	if c, ok := err.(coder); ok {
		return c.Code()
	}
	return "UNKNOWN"
}
