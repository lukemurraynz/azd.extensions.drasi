package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestFormatAuditEvent(t *testing.T) {
	t.Parallel()

	event := AuditEvent{
		Operation:     "deploy",
		Environment:   "dev",
		CorrelationID: "corr-123",
		Target:        "query/orders",
		Result:        "success",
		StartedAtUtc:  time.Date(2026, time.April, 5, 12, 0, 0, 0, time.UTC),
		EndedAtUtc:    time.Date(2026, time.April, 5, 12, 0, 2, 0, time.UTC),
	}

	t.Run("table mode renders human readable line", func(t *testing.T) {
		t.Parallel()

		got := FormatAuditEvent(event, FormatTable)
		for _, want := range []string{"deploy", "dev", "success", "2026-04-05T12:00:00Z"} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected table audit output to contain %q, got %q", want, got)
			}
		}
	})

	t.Run("json mode renders valid json with all fields", func(t *testing.T) {
		t.Parallel()

		got := FormatAuditEvent(event, FormatJSON)

		var decoded map[string]any
		if err := json.Unmarshal([]byte(got), &decoded); err != nil {
			t.Fatalf("expected valid json output, got error: %v; output=%q", err, got)
		}

		for _, field := range []string{"operation", "environment", "correlationId", "target", "result", "startedAtUtc", "endedAtUtc"} {
			if _, ok := decoded[field]; !ok {
				t.Fatalf("expected field %q in audit json payload: %#v", field, decoded)
			}
		}

		started, _ := decoded["startedAtUtc"].(string)
		ended, _ := decoded["endedAtUtc"].(string)
		if _, err := time.Parse(time.RFC3339, started); err != nil {
			t.Fatalf("expected startedAtUtc to be RFC3339, got %q (%v)", started, err)
		}
		if _, err := time.Parse(time.RFC3339, ended); err != nil {
			t.Fatalf("expected endedAtUtc to be RFC3339, got %q (%v)", ended, err)
		}
	})
}
