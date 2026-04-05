package output

import (
	"encoding/json"
	fmtpkg "fmt"
	"time"
)

// AuditEvent records a structured log entry for a mutating command.
type AuditEvent struct {
	Operation     string    `json:"operation"`
	Environment   string    `json:"environment"`
	CorrelationID string    `json:"correlationId"`
	Target        string    `json:"target"`
	Result        string    `json:"result"`
	StartedAtUtc  time.Time `json:"startedAtUtc"`
	EndedAtUtc    time.Time `json:"endedAtUtc"`
}

// FormatAuditEvent renders an AuditEvent in the requested output format.
func FormatAuditEvent(event AuditEvent, fmt OutputFormat) string {
	switch fmt {
	case FormatJSON:
		b, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			return ""
		}
		return string(b)
	default:
		duration := event.EndedAtUtc.Sub(event.StartedAtUtc).Round(time.Millisecond)
		return fmtpkg.Sprintf("[%s] %s on %s — %s (%s)",
			event.StartedAtUtc.UTC().Format(time.RFC3339),
			event.Operation,
			event.Environment,
			event.Result,
			duration,
		)
	}
}
