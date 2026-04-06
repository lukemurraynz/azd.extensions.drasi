package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStatusClient struct {
	checkVersionErr error
	components      []drasi.ComponentSummary
	listErr         error
	lastKind        string
	lastContext     string
}

func (f *fakeStatusClient) CheckVersion(_ context.Context) error {
	return f.checkVersionErr
}

func (f *fakeStatusClient) ListComponents(_ context.Context, kind string) ([]drasi.ComponentSummary, error) {
	f.lastKind = kind
	f.lastContext = ""
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]drasi.ComponentSummary(nil), f.components...), nil
}

func (f *fakeStatusClient) ListComponentsInContext(_ context.Context, kind, kubeContext string) ([]drasi.ComponentSummary, error) {
	f.lastKind = kind
	f.lastContext = kubeContext
	if f.listErr != nil {
		return nil, f.listErr
	}
	return append([]drasi.ComponentSummary(nil), f.components...), nil
}

func TestStatusCommand_TableSuccess_DefaultKind(t *testing.T) {
	orig := newStatusDrasiClient
	t.Cleanup(func() { newStatusDrasiClient = orig })

	client := &fakeStatusClient{
		components: []drasi.ComponentSummary{{ID: "s1", Kind: "source", Status: "Online"}},
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
	assert.Equal(t, "source", client.lastKind)
	assert.Contains(t, stdout.String(), "s1")
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
