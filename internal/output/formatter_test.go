package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	t.Parallel()

	type sampleRow struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}

	tests := []struct {
		name   string
		data   any
		format OutputFormat
		assert func(t *testing.T, got string)
	}{
		{
			name:   "table renders slice of structs",
			data:   []sampleRow{{Name: "Alice", Role: "Admin"}, {Name: "Bob", Role: "Reader"}},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if got == "" {
					t.Fatal("expected non-empty table output")
				}
				for _, want := range []string{"name", "role", "Alice", "Admin", "Bob", "Reader"} {
					if !strings.Contains(got, want) {
						t.Fatalf("expected table output to contain %q, got %q", want, got)
					}
				}
				if lines := strings.Count(strings.TrimSpace(got), "\n") + 1; lines < 3 {
					t.Fatalf("expected at least 3 lines in table output, got %d (%q)", lines, got)
				}
			},
		},
		{
			name:   "json renders valid json",
			data:   []sampleRow{{Name: "Alice", Role: "Admin"}},
			format: FormatJSON,
			assert: func(t *testing.T, got string) {
				t.Helper()
				var decoded []map[string]string
				if err := json.Unmarshal([]byte(got), &decoded); err != nil {
					t.Fatalf("expected valid json output, got error: %v; output=%q", err, got)
				}
				if len(decoded) != 1 || decoded[0]["name"] != "Alice" || decoded[0]["role"] != "Admin" {
					t.Fatalf("unexpected decoded json payload: %#v", decoded)
				}
			},
		},
		{
			name:   "nil data handled gracefully",
			data:   nil,
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if got != "" && got != "{}" {
					t.Fatalf("expected empty string or {} for nil input, got %q", got)
				}
			},
		},
		{
			name:   "empty slice handled gracefully in table mode",
			data:   []sampleRow{},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if strings.TrimSpace(got) != "" {
					t.Fatalf("expected empty table output for empty slice, got %q", got)
				}
			},
		},
		{
			name:   "empty slice handled gracefully in json mode",
			data:   []sampleRow{},
			format: FormatJSON,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if strings.TrimSpace(got) != "[]" {
					t.Fatalf("expected [] for empty slice json output, got %q", got)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Format(tc.data, tc.format)
			tc.assert(t, got)
		})
	}
}
