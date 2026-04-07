package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStatusClient is a test double for statusDrasiClient.
// When componentsByKind is set, ListComponents returns per-kind results.
// Otherwise it falls back to the flat components slice for backward compatibility.
type fakeStatusClient struct {
	checkVersionErr  error
	components       []drasi.ComponentSummary
	componentsByKind map[string][]drasi.ComponentSummary
	listErr          error
	lastKind         string
	lastContext      string
	calledKinds      []string
}

func (f *fakeStatusClient) CheckVersion(_ context.Context) error {
	return f.checkVersionErr
}

func (f *fakeStatusClient) ListComponents(_ context.Context, kind string) ([]drasi.ComponentSummary, error) {
	f.lastKind = kind
	f.lastContext = ""
	f.calledKinds = append(f.calledKinds, kind)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.componentsByKind != nil {
		res := f.componentsByKind[kind]
		return append([]drasi.ComponentSummary(nil), res...), nil
	}
	return append([]drasi.ComponentSummary(nil), f.components...), nil
}

func (f *fakeStatusClient) ListComponentsInContext(_ context.Context, kind, kubeContext string) ([]drasi.ComponentSummary, error) {
	f.lastKind = kind
	f.lastContext = kubeContext
	f.calledKinds = append(f.calledKinds, kind)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.componentsByKind != nil {
		res := f.componentsByKind[kind]
		return append([]drasi.ComponentSummary(nil), res...), nil
	}
	return append([]drasi.ComponentSummary(nil), f.components...), nil
}

// --- Existing tests updated for the new all-kinds default behavior ---

// TestStatusCommand_TableSuccess_DefaultKind verifies that when no --kind flag is
// provided, the command queries all 4 component kinds and renders section headers.
func TestStatusCommand_TableSuccess_DefaultKind(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		componentsByKind: map[string][]drasi.ComponentSummary{
			"source":     {{ID: "s1", Kind: "source", Status: "Online"}},
			"query":      {{ID: "q1", Kind: "query", Status: "Running"}},
			"middleware": {},
			"reaction":   {{ID: "r1", Kind: "reaction", Status: "Online"}},
		},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"status"})

	err := root.Execute()
	require.NoError(t, err)

	// All 4 wire kinds should be queried.
	assert.Equal(t, []string{"source", "query", "middleware", "reaction"}, client.calledKinds)

	out := stdout.String()
	assert.Contains(t, out, "Sources:")
	assert.Contains(t, out, "s1")
	assert.Contains(t, out, "Queries:")
	assert.Contains(t, out, "q1")
	assert.Contains(t, out, "Middleware:")
	assert.Contains(t, out, "No middleware components found.")
	assert.Contains(t, out, "Reactions:")
	assert.Contains(t, out, "r1")
	assert.Empty(t, stderr.String())
}

func TestStatusCommand_JSONSuccess_ContinuousQueryKindAlias(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		components: []drasi.ComponentSummary{{ID: "q1", Kind: "query", Status: "Online"}},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--output", "json", "status", "--kind", "continuousquery"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "query", client.lastKind)
	assert.Contains(t, stdout.String(), `"status": "ok"`)
	assert.Contains(t, stdout.String(), `"components"`)
	assert.Empty(t, stderr.String())
}

func TestStatusCommand_ListFailure_ReturnsError(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{listErr: errors.New("ERR_DRASI_CLI_ERROR: list failed")}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status", "--kind", "source"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DRASI_CLI_ERROR")
}

// --- New tests for all-kinds and single-kind modes ---

// TestStatusAllKinds verifies table output for all kinds with mixed populated and empty.
func TestStatusAllKinds(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		componentsByKind: map[string][]drasi.ComponentSummary{
			"source": {
				{ID: "src-1", Kind: "source", Status: "Online"},
				{ID: "src-2", Kind: "source", Status: "Offline"},
			},
			"query":      {{ID: "qry-1", Kind: "query", Status: "Running"}},
			"middleware": {},
			"reaction":   {{ID: "rx-1", Kind: "reaction", Status: "Online"}},
		},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status"})

	err := root.Execute()
	require.NoError(t, err)

	out := stdout.String()

	// Section headers present.
	assert.Contains(t, out, "Sources:")
	assert.Contains(t, out, "Queries:")
	assert.Contains(t, out, "Middleware:")
	assert.Contains(t, out, "Reactions:")

	// Component IDs present.
	assert.Contains(t, out, "src-1")
	assert.Contains(t, out, "src-2")
	assert.Contains(t, out, "qry-1")
	assert.Contains(t, out, "rx-1")

	// Empty kind has placeholder.
	assert.Contains(t, out, "No middleware components found.")

	// All 4 wire kinds were queried in order.
	assert.Equal(t, []string{"source", "query", "middleware", "reaction"}, client.calledKinds)
}

// TestStatusSingleKind verifies that --kind filters to a single kind with no section header.
func TestStatusSingleKind(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		components: []drasi.ComponentSummary{
			{ID: "rx-1", Kind: "reaction", Status: "Online"},
		},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status", "--kind", "reaction"})

	err := root.Execute()
	require.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "rx-1")

	// Single-kind mode does NOT produce section headers.
	assert.NotContains(t, out, "Sources:")
	assert.NotContains(t, out, "Queries:")
	assert.NotContains(t, out, "Reactions:")

	// Only the specified kind was queried.
	assert.Equal(t, []string{"reaction"}, client.calledKinds)
	assert.Equal(t, "reaction", client.lastKind)
}

// TestStatusAllKindsJSON verifies JSON output when no --kind flag is provided.
func TestStatusAllKindsJSON(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		componentsByKind: map[string][]drasi.ComponentSummary{
			"source":     {{ID: "s1", Kind: "source", Status: "Online"}},
			"query":      {{ID: "q1", Kind: "query", Status: "Running"}},
			"middleware": {},
			"reaction":   {},
		},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "status"})

	err := root.Execute()
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &payload))

	assert.Equal(t, "ok", payload["status"])

	// All 4 keys present.
	for _, key := range []string{"sources", "queries", "middleware", "reactions"} {
		val, ok := payload[key]
		require.True(t, ok, "missing JSON key %q", key)
		// Must be an array (not null).
		arr, isArr := val.([]any)
		require.True(t, isArr, "expected array for %q, got %T", key, val)
		_ = arr
	}

	// sources and queries have items.
	sources := payload["sources"].([]any)
	assert.Len(t, sources, 1)
	queries := payload["queries"].([]any)
	assert.Len(t, queries, 1)

	// middleware and reactions are empty arrays (not null).
	middleware := payload["middleware"].([]any)
	assert.Empty(t, middleware)
	reactions := payload["reactions"].([]any)
	assert.Empty(t, reactions)

	// Single-kind JSON keys should NOT be present.
	_, hasKind := payload["kind"]
	assert.False(t, hasKind, "all-kinds JSON should not have 'kind' key")
	_, hasComponents := payload["components"]
	assert.False(t, hasComponents, "all-kinds JSON should not have 'components' key")
}

// TestStatusEmpty verifies that when all kinds return empty, the table shows
// placeholder messages for every kind.
func TestStatusEmpty(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		componentsByKind: map[string][]drasi.ComponentSummary{
			"source":     {},
			"query":      {},
			"middleware": {},
			"reaction":   {},
		},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"status"})

	err := root.Execute()
	require.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "No source components found.")
	assert.Contains(t, out, "No continuousquery components found.")
	assert.Contains(t, out, "No middleware components found.")
	assert.Contains(t, out, "No reaction components found.")
}

// TestStatusSingleKindJSON verifies JSON output when --kind is specified (backward compat).
func TestStatusSingleKindJSON(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		components: []drasi.ComponentSummary{
			{ID: "s1", Kind: "source", Status: "Online"},
		},
	}
	newStatusDrasiClient = func() statusDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--output", "json", "status", "--kind", "source"})

	err := root.Execute()
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &payload))

	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, "source", payload["kind"])

	_, hasComponents := payload["components"]
	assert.True(t, hasComponents, "single-kind JSON must have 'components' key")

	// All-kinds keys should NOT be present.
	for _, key := range []string{"sources", "queries", "middleware", "reactions"} {
		_, exists := payload[key]
		assert.False(t, exists, "single-kind JSON should not have %q key", key)
	}
}
