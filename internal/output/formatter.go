package output

import (
	"bytes"
	"encoding/json"
	fmtpkg "fmt"
	"reflect"
	"strings"
	"text/tabwriter"
)

// OutputFormat controls how command output is rendered.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// Format renders data in the requested format.
// data may be a struct, slice of structs, or map[string]any.
// Returns empty string for nil input.
func Format(data any, fmt OutputFormat) string {
	if data == nil {
		return ""
	}

	if fmt == FormatJSON {
		payload, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return ""
		}
		return string(payload)
	}

	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}

	switch value.Kind() { //nolint:exhaustive // Only slice/array/struct/map get table formatting; all other kinds use Sprint.
	case reflect.Slice, reflect.Array:
		return formatSliceAsTable(value)
	case reflect.Struct:
		return formatStructTable(value)
	case reflect.Map:
		return formatMapTable(value)
	default:
		return fmtpkg.Sprint(data)
	}
}

func formatSliceAsTable(value reflect.Value) string {
	if value.Len() == 0 {
		return ""
	}

	first := dereferenceValue(value.Index(0))
	if !first.IsValid() {
		return ""
	}

	if first.Kind() != reflect.Struct {
		rows := make([]string, 0, value.Len())
		for i := range value.Len() {
			rows = append(rows, fmtpkg.Sprint(value.Index(i).Interface()))
		}
		return strings.Join(rows, "\n")
	}

	headers, selectors := structColumns(first.Type())
	if len(headers) == 0 {
		return ""
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmtpkg.Fprintln(tw, strings.Join(headers, "\t"))
	for i := range value.Len() {
		rowValue := dereferenceValue(value.Index(i))
		cells := make([]string, 0, len(selectors))
		for _, selector := range selectors {
			cells = append(cells, selector(rowValue))
		}
		_, _ = fmtpkg.Fprintln(tw, strings.Join(cells, "\t"))
	}
	_ = tw.Flush()
	return buf.String()
}

func formatStructTable(value reflect.Value) string {
	headers, selectors := structColumns(value.Type())
	if len(headers) == 0 {
		return ""
	}

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmtpkg.Fprintln(tw, strings.Join(headers, "\t"))
	row := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		row = append(row, selector(value))
	}
	_, _ = fmtpkg.Fprintln(tw, strings.Join(row, "\t"))
	_ = tw.Flush()
	return buf.String()
}

func formatMapTable(value reflect.Value) string {
	if value.Len() == 0 {
		return ""
	}

	iter := value.MapRange()
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmtpkg.Fprintln(tw, "key\tvalue")
	for iter.Next() {
		_, _ = fmtpkg.Fprintf(tw, "%v\t%v\n", iter.Key().Interface(), iter.Value().Interface())
	}
	_ = tw.Flush()
	return buf.String()
}

func structColumns(typ reflect.Type) ([]string, []func(reflect.Value) string) {
	headers := make([]string, 0, typ.NumField())
	selectors := make([]func(reflect.Value) string, 0, typ.NumField())

	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		header := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			name := strings.Split(tag, ",")[0]
			if name == "-" {
				continue
			}
			if name != "" {
				header = name
			}
		}

		index := i
		headers = append(headers, header)
		selectors = append(selectors, func(v reflect.Value) string {
			fieldValue := dereferenceValue(v.Field(index))
			if !fieldValue.IsValid() {
				return ""
			}
			return fmtpkg.Sprint(fieldValue.Interface())
		})
	}

	return headers, selectors
}

func dereferenceValue(value reflect.Value) reflect.Value {
	for value.IsValid() && value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}
