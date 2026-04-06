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

type fakeLogsClient struct {
	checkVersionErr error
	detail          *drasi.ComponentDetail
	describeErr     error
	logOutput       string
	logsErr         error
	lastLogsArgs    []string
	lastContext     string
}

func (f *fakeLogsClient) CheckVersion(_ context.Context) error {
	return f.checkVersionErr
}

func (f *fakeLogsClient) DescribeComponent(_ context.Context, _, _ string) (*drasi.ComponentDetail, error) {
	f.lastContext = ""
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	if f.detail == nil {
		return &drasi.ComponentDetail{}, nil
	}
	return f.detail, nil
}

func (f *fakeLogsClient) DescribeComponentInContext(_ context.Context, _, _ string, kubeContext string) (*drasi.ComponentDetail, error) {
	f.lastContext = kubeContext
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	if f.detail == nil {
		return &drasi.ComponentDetail{}, nil
	}
	return f.detail, nil
}

func (f *fakeLogsClient) RunCommandOutput(_ context.Context, args ...string) (string, error) {
	f.lastLogsArgs = append([]string(nil), args...)
	if f.logsErr != nil {
		return "", f.logsErr
	}
	return f.logOutput, nil
}

func TestLogsCommand_TableSuccess_PrintsLogs(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		detail:    &drasi.ComponentDetail{ID: "q1", Kind: "query", Status: "Online"},
		logOutput: "line1\nline2\n",
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "q1", "--kind", "continuousquery", "--tail", "5", "--follow"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Equal(t, []string{"watch", "q1"}, client.lastLogsArgs)
	assert.Equal(t, "line1\nline2\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestLogsCommand_JSONSuccess_EmitsPayload(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		detail:    &drasi.ComponentDetail{ID: "s1", Kind: "source", Status: "Online"},
		logOutput: "source ready\n",
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--output", "json", "logs", "--component", "s1", "--kind", "query"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"status": "ok"`)
	assert.Contains(t, stdout.String(), `"component": "s1"`)
	assert.Contains(t, stdout.String(), `"watch": "source ready"`)
	assert.Empty(t, stderr.String())
}

func TestLogsCommand_NoLogs_PrintsNoLogsMessage(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		detail:    &drasi.ComponentDetail{ID: "r1", Kind: "query", Status: "Online"},
		logOutput: "\n",
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "r1", "--kind", "query"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No watch output found for r1 (query).")
	assert.Empty(t, stderr.String())
}

func TestLogsCommand_LogsCallFailure_ReturnsError(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		detail:  &drasi.ComponentDetail{ID: "x", Kind: "source", Status: "Online"},
		logsErr: errors.New("ERR_DRASI_CLI_ERROR: logs failed"),
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--component", "x", "--kind", "source"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_VALIDATION_FAILED")
}

func TestLogsCommand_DescribeFailure_ReturnsError(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		describeErr: errors.New("ERR_DRASI_CLI_ERROR: describe failed"),
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--component", "x", "--kind", "source"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_VALIDATION_FAILED")
}

func TestLogsCommand_WatchFailure_ReturnsDrasiError(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		detail:  &drasi.ComponentDetail{ID: "q", Kind: "query", Status: "Online"},
		logsErr: errors.New("ERR_DRASI_CLI_ERROR: watch failed"),
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--component", "q", "--kind", "query"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DRASI_CLI_ERROR")
	assert.Equal(t, []string{"watch", "q"}, client.lastLogsArgs)
}
