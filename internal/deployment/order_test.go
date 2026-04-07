package deployment

import (
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
)

func makeAction(kind, id string) ComponentAction {
	return ComponentAction{Kind: kind, ID: id}
}

func TestSortForDeploy_OrderIsSourcesQueriesMiddlewareReactions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		actions []ComponentAction
		want    []ComponentAction
	}{
		{
			name: "mixed actions sorted by deploy dependency order",
			actions: []ComponentAction{
				makeAction("reaction", "r1"),
				makeAction("continuousquery", "q1"),
				makeAction("source", "s1"),
				makeAction("middleware", "m1"),
			},
			want: []ComponentAction{
				makeAction("source", "s1"),
				makeAction("continuousquery", "q1"),
				makeAction("middleware", "m1"),
				makeAction("reaction", "r1"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SortForDeploy(tt.actions, &config.ResolvedManifest{})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSortForDelete_OrderIsReverseOfDeploy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		actions []ComponentAction
		want    []ComponentAction
	}{
		{
			name: "mixed actions sorted in reverse dependency order",
			actions: []ComponentAction{
				makeAction("source", "s1"),
				makeAction("reaction", "r1"),
				makeAction("middleware", "m1"),
				makeAction("continuousquery", "q1"),
			},
			want: []ComponentAction{
				makeAction("reaction", "r1"),
				makeAction("middleware", "m1"),
				makeAction("continuousquery", "q1"),
				makeAction("source", "s1"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SortForDelete(tt.actions, &config.ResolvedManifest{})
			assert.Equal(t, tt.want, got)
		})
	}
}
