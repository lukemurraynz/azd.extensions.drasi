package drasi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListComponents_ParsesTabularOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		kind    string
		stdout  string
		expected []ComponentSummary
	}{
		{
			name:   "tabular output parsed into summaries",
			kind:   "query",
			stdout: "ID KIND STATUS\nalerts query Online\norders query Pending\n",
			expected: []ComponentSummary{
				{ID: "alerts", Kind: "query", Status: "Online"},
				{ID: "orders", Kind: "query", Status: "Pending"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}})

			got, err := client.ListComponents(context.Background(), tt.kind)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestListComponents_EmptyList_ReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		stdout string
	}{
		{name: "empty output returns empty slice", kind: "query", stdout: "ID KIND STATUS\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}})

			got, err := client.ListComponents(context.Background(), tt.kind)

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Empty(t, got)
		})
	}
}

func TestListComponents_ParseError_Surfaced(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		stdout string
	}{
		{name: "malformed output returns parse error", kind: "query", stdout: "not a drasi table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(&mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}})

			got, err := client.ListComponents(context.Background(), tt.kind)

			require.Error(t, err)
			assert.Nil(t, got)
		})
	}
}
