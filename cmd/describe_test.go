package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDescribeClient struct {
	checkVersionErr error
	detail          *drasi.ComponentDetail
	describeErr     error
	lastKind        string
	lastID          string
	lastContext     string
}

func (f *fakeDescribeClient) CheckVersion(_ context.Context) error {
	return f.checkVersionErr
}

func (f *fakeDescribeClient) DescribeComponent(_ context.Context, kind, id string) (*drasi.ComponentDetail, error) {
	f.lastKind = kind
	f.lastID = id
	f.lastContext = ""
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	return f.detail, nil
}

func (f *fakeDescribeClient) DescribeComponentInContext(_ context.Context, kind, id, kubeContext string) (*drasi.ComponentDetail, error) {
	f.lastKind = kind
	f.lastID = id
	f.lastContext = kubeContext
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	return f.detail, nil
}

func TestDescribeSuccess(t *testing.T) {
	orig := newDescribeDrasiClient
	t.Cleanup(func() { newDescribeDrasiClient = orig })

	client := &fakeDescribeClient{
		detail: &drasi.ComponentDetail{
			ID:     "my-source",
			Kind:   "source",
			Status: "Online",
		},
	}
	newDescribeDrasiClient = func() describeDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"describe", "--kind", "source", "--component", "my-source"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, "source", client.lastKind)
	assert.Equal(t, "my-source", client.lastID)
	assert.Contains(t, stdout.String(), "my-source")
	assert.Contains(t, stdout.String(), "Online")
	assert.Empty(t, stderr.String())
}

func TestDescribeNotFound(t *testing.T) {
	orig := newDescribeDrasiClient
	t.Cleanup(func() { newDescribeDrasiClient = orig })

	client := &fakeDescribeClient{
		describeErr: &drasi.ComponentNotFoundError{Kind: "source", ID: "missing-src"},
	}
	newDescribeDrasiClient = func() describeDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"describe", "--kind", "source", "--component", "missing-src"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source/missing-src")
	assert.Contains(t, err.Error(), "not found")
}

func TestDescribeJSON(t *testing.T) {
	orig := newDescribeDrasiClient
	t.Cleanup(func() { newDescribeDrasiClient = orig })

	client := &fakeDescribeClient{
		detail: &drasi.ComponentDetail{
			ID:     "q1",
			Kind:   "query",
			Status: "Running",
		},
	}
	newDescribeDrasiClient = func() describeDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	// continuousquery should be mapped to query for the CLI.
	root.SetArgs([]string{"--output", "json", "describe", "--kind", "continuousquery", "--component", "q1"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"status": "ok"`)
	assert.Contains(t, stdout.String(), `"kind": "query"`)
	assert.Contains(t, stdout.String(), `"component": "q1"`)
	assert.Contains(t, stdout.String(), `"detail"`)
	assert.Empty(t, stderr.String())
	// Verify the kind was mapped before calling the client.
	assert.Equal(t, "query", client.lastKind)
}

func TestDescribeMissingFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing both", args: []string{"describe"}},
		{name: "missing component", args: []string{"describe", "--kind", "source"}},
		{name: "missing kind", args: []string{"describe", "--component", "my-src"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := newDescribeDrasiClient
			t.Cleanup(func() { newDescribeDrasiClient = orig })

			// Provide a client that succeeds, so the error must come from validation.
			client := &fakeDescribeClient{
				detail: &drasi.ComponentDetail{ID: "x", Kind: "source", Status: "Online"},
			}
			newDescribeDrasiClient = func() describeDrasiClient { return client }

			root := NewRootCommand()
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})
			root.SetArgs(tc.args)

			err := root.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
		})
	}
}

func TestDescribeCheckVersionFailure(t *testing.T) {
	orig := newDescribeDrasiClient
	t.Cleanup(func() { newDescribeDrasiClient = orig })

	client := &fakeDescribeClient{
		checkVersionErr: errors.New(output.ERR_DRASI_CLI_NOT_FOUND + ": drasi not found"),
	}
	newDescribeDrasiClient = func() describeDrasiClient { return client }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"describe", "--kind", "source", "--component", "x"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_DRASI_CLI_NOT_FOUND)
}
