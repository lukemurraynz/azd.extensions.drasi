package output

import (
	"encoding/json"
	"reflect"
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

func TestFormat_SingleStruct(t *testing.T) {
	t.Parallel()

	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name   string
		data   any
		format OutputFormat
		assert func(t *testing.T, got string)
	}{
		{
			name:   "table renders single struct as header plus one row",
			data:   item{ID: 1, Name: "Widget"},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				for _, want := range []string{"id", "name", "1", "Widget"} {
					if !strings.Contains(got, want) {
						t.Fatalf("expected table to contain %q, got %q", want, got)
					}
				}
				lines := strings.Count(strings.TrimSpace(got), "\n") + 1
				if lines != 2 {
					t.Fatalf("expected 2 lines (header + row), got %d", lines)
				}
			},
		},
		{
			name:   "json renders single struct",
			data:   item{ID: 42, Name: "Gadget"},
			format: FormatJSON,
			assert: func(t *testing.T, got string) {
				t.Helper()
				var decoded map[string]any
				if err := json.Unmarshal([]byte(got), &decoded); err != nil {
					t.Fatalf("expected valid json, got error: %v", err)
				}
				if decoded["name"] != "Gadget" {
					t.Fatalf("expected name=Gadget, got %v", decoded["name"])
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

func TestFormat_Map(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		data   any
		format OutputFormat
		assert func(t *testing.T, got string)
	}{
		{
			name:   "table renders map as key-value pairs",
			data:   map[string]any{"region": "eastus", "sku": "standard"},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "key") || !strings.Contains(got, "value") {
					t.Fatalf("expected key/value headers, got %q", got)
				}
				if !strings.Contains(got, "region") || !strings.Contains(got, "eastus") {
					t.Fatalf("expected map entries in output, got %q", got)
				}
			},
		},
		{
			name:   "table renders empty map as empty string",
			data:   map[string]any{},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if got != "" {
					t.Fatalf("expected empty string for empty map, got %q", got)
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

func TestFormat_Pointer(t *testing.T) {
	t.Parallel()

	type row struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name   string
		data   any
		format OutputFormat
		assert func(t *testing.T, got string)
	}{
		{
			name:   "pointer to struct renders table",
			data:   &row{Name: "Alice"},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "Alice") {
					t.Fatalf("expected pointer deref to show Alice, got %q", got)
				}
			},
		},
		{
			name:   "nil pointer returns empty",
			data:   (*row)(nil),
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if got != "" {
					t.Fatalf("expected empty for nil pointer, got %q", got)
				}
			},
		},
		{
			name:   "pointer to slice renders table",
			data:   &[]row{{Name: "Bob"}},
			format: FormatTable,
			assert: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "Bob") {
					t.Fatalf("expected pointer deref to show Bob, got %q", got)
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

func TestFormat_DefaultKind(t *testing.T) {
	t.Parallel()

	// Non-struct, non-slice, non-map values use Sprint fallback.
	tests := []struct {
		name string
		data any
		want string
	}{
		{name: "integer", data: 42, want: "42"},
		{name: "string", data: "hello", want: "hello"},
		{name: "float", data: 3.14, want: "3.14"},
		{name: "bool", data: true, want: "true"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Format(tc.data, FormatTable)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestFormat_SliceOfNonStructs(t *testing.T) {
	t.Parallel()

	data := []string{"alpha", "beta", "gamma"}
	got := Format(data, FormatTable)
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in output, got %q", want, got)
		}
	}
}

func TestFormat_JSONMarshalError(t *testing.T) {
	t.Parallel()

	// Channels cannot be JSON marshaled.
	ch := make(chan int)
	got := Format(ch, FormatJSON)
	if got != "" {
		t.Fatalf("expected empty string for unmarshalable type, got %q", got)
	}
}

func TestStructColumns(t *testing.T) {
	t.Parallel()

	type sample struct {
		Public   string `json:"pub"`
		private  string //nolint:unused // unexported field used to test structColumns filtering
		Skipped  string `json:"-"`
		NoTag    string
		EmptyTag string `json:",omitempty"`
	}

	headers, selectors := structColumns(reflect.TypeOf(sample{}))

	// Expect: "pub", "NoTag", "EmptyTag" (Private skipped, Skipped "-" excluded)
	expectedHeaders := []string{"pub", "NoTag", "EmptyTag"}
	if len(headers) != len(expectedHeaders) {
		t.Fatalf("expected %d headers, got %d: %v", len(expectedHeaders), len(headers), headers)
	}
	for i, want := range expectedHeaders {
		if headers[i] != want {
			t.Fatalf("header[%d]: expected %q, got %q", i, want, headers[i])
		}
	}

	// Verify selectors work
	val := reflect.ValueOf(sample{Public: "A", NoTag: "B", EmptyTag: "C"})
	if selectors[0](val) != "A" {
		t.Fatalf("selector[0] expected A, got %q", selectors[0](val))
	}
	if selectors[1](val) != "B" {
		t.Fatalf("selector[1] expected B, got %q", selectors[1](val))
	}
}

func TestDereferenceValue(t *testing.T) {
	t.Parallel()

	t.Run("non-pointer", func(t *testing.T) {
		t.Parallel()
		v := reflect.ValueOf(42)
		got := dereferenceValue(v)
		if got.Interface() != 42 {
			t.Fatalf("expected 42, got %v", got.Interface())
		}
	})

	t.Run("pointer to value", func(t *testing.T) {
		t.Parallel()
		x := 99
		v := reflect.ValueOf(&x)
		got := dereferenceValue(v)
		if got.Interface() != 99 {
			t.Fatalf("expected 99, got %v", got.Interface())
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		t.Parallel()
		var p *int
		v := reflect.ValueOf(p)
		got := dereferenceValue(v)
		if got.IsValid() {
			t.Fatalf("expected invalid value for nil pointer, got %v", got)
		}
	})

	t.Run("double pointer", func(t *testing.T) {
		t.Parallel()
		x := 77
		px := &x
		ppx := &px
		v := reflect.ValueOf(ppx)
		got := dereferenceValue(v)
		if got.Interface() != 77 {
			t.Fatalf("expected 77, got %v", got.Interface())
		}
	})

	t.Run("pointer to nil pointer", func(t *testing.T) {
		t.Parallel()
		var inner *int
		outer := &inner
		v := reflect.ValueOf(outer)
		got := dereferenceValue(v)
		if got.IsValid() {
			t.Fatalf("expected invalid value for pointer-to-nil-pointer, got %v", got)
		}
	})
}

func TestFormatSliceAsTable_NilFirstElement(t *testing.T) {
	t.Parallel()

	// Slice of pointers where first element is nil.
	type row struct {
		Name string `json:"name"`
	}
	data := []*row{nil, {Name: "Alice"}}
	got := Format(data, FormatTable)
	// First element is nil, so dereferenceValue returns invalid, formatSliceAsTable returns ""
	if got != "" {
		t.Fatalf("expected empty for nil first element, got %q", got)
	}
}

func TestFormat_StructWithNilPointerField(t *testing.T) {
	t.Parallel()

	type row struct {
		Name *string `json:"name"`
		Age  int     `json:"age"`
	}

	data := []row{{Name: nil, Age: 30}}
	got := Format(data, FormatTable)
	// The nil pointer field should render as empty string, not panic
	if !strings.Contains(got, "30") {
		t.Fatalf("expected age 30 in output, got %q", got)
	}
}

func TestFormat_StructWithNoExportedFields(t *testing.T) {
	t.Parallel()

	type hidden struct {
		private string //nolint:unused // unexported field used to test empty structColumns
	}

	data := hidden{}
	got := Format(data, FormatTable)
	// No exported fields means structColumns returns empty headers, formatStructTable returns ""
	if got != "" {
		t.Fatalf("expected empty for struct with no exported fields, got %q", got)
	}
}
