package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFallbackFoundVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "empty returns unknown", version: "", want: unknownVersion},
		{name: "whitespace returns unknown", version: " \t\n ", want: unknownVersion},
		{name: "trims surrounding whitespace", version: " 1.2.3 ", want: "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, fallbackFoundVersion(tt.version))
		})
	}
}

func TestFormatCommandError(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("boom")

	tests := []struct {
		name           string
		stderr         string
		wantContains   []string
		wantNotContain string
	}{
		{
			name:           "without stderr",
			stderr:         "   ",
			wantContains:   []string{"kubectl", "boom"},
			wantNotContain: "stderr:",
		},
		{
			name:         "with trimmed stderr",
			stderr:       "  context missing  ",
			wantContains: []string{"kubectl", "boom", "stderr: context missing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := formatCommandError("kubectl", tt.stderr, baseErr)
			require.Error(t, err)
			assert.ErrorIs(t, err, baseErr)
			for _, want := range tt.wantContains {
				assert.Contains(t, err.Error(), want)
			}
			if tt.wantNotContain != "" {
				assert.NotContains(t, err.Error(), tt.wantNotContain)
			}
		})
	}
}

func TestDigitsOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "mixed characters", value: "v1.28.4+k3s1", want: "128431"},
		{name: "no digits", value: "minor+", want: ""},
		{name: "already digits", value: "12345", want: "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, digitsOnly(tt.value), fmt.Sprintf("value=%q", tt.value))
		})
	}
}
