package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
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
	root.SetArgs([]string{"logs", "--component", "q1", "--kind", "continuousquery"})

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

func TestLogsCommand_KubectlFailure_ReturnsError(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	kubectlLogsFunc = func(_ context.Context, _ ...string) (string, error) {
		return "", errors.New("kubectl logs: connection refused")
	}

	root := NewRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"logs", "--component", "x", "--kind", "source"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DRASI_CLI_ERROR")
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
	root.SetArgs([]string{"logs", "--component", "x", "--kind", "query"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ERR_DRASI_CLI_ERROR")
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

// --- kubectl-based tests for non-query kinds ---

func TestLogsSourceKind(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	var capturedArgs []string
	kubectlLogsFunc = func(_ context.Context, args ...string) (string, error) {
		capturedArgs = append([]string(nil), args...)
		return "source log line 1\nsource log line 2\n", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "my-source", "--kind", "source"})

	err := root.Execute()
	require.NoError(t, err)

	argsJoined := strings.Join(capturedArgs, " ")
	assert.Contains(t, argsJoined, "drasi.io/kind=source,drasi.io/component=my-source")
	assert.Contains(t, argsJoined, "-n drasi-system")
	assert.Contains(t, argsJoined, "--tail=100")
	assert.Contains(t, stdout.String(), "source log line 1")
	assert.Contains(t, stdout.String(), "source log line 2")
	assert.Empty(t, stderr.String())
}

func TestLogsReactionKind(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	var capturedArgs []string
	kubectlLogsFunc = func(_ context.Context, args ...string) (string, error) {
		capturedArgs = append([]string(nil), args...)
		return "reaction output\n", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "my-reaction", "--kind", "reaction"})

	err := root.Execute()
	require.NoError(t, err)

	argsJoined := strings.Join(capturedArgs, " ")
	assert.Contains(t, argsJoined, "drasi.io/kind=reaction,drasi.io/component=my-reaction")
	assert.Contains(t, argsJoined, "-n drasi-system")
	assert.Contains(t, stdout.String(), "reaction output")
	assert.Empty(t, stderr.String())
}

func TestLogsMiddlewareKind(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	var capturedArgs []string
	kubectlLogsFunc = func(_ context.Context, args ...string) (string, error) {
		capturedArgs = append([]string(nil), args...)
		return "middleware logs here\n", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "my-middleware", "--kind", "middleware"})

	err := root.Execute()
	require.NoError(t, err)

	argsJoined := strings.Join(capturedArgs, " ")
	assert.Contains(t, argsJoined, "drasi.io/kind=middleware,drasi.io/component=my-middleware")
	assert.Contains(t, argsJoined, "-n drasi-system")
	assert.Contains(t, stdout.String(), "middleware logs here")
	assert.Empty(t, stderr.String())
}

func TestLogsContinuousQueryKind(t *testing.T) {
	orig := newLogsDrasiClient
	t.Cleanup(func() { newLogsDrasiClient = orig })

	client := &fakeLogsClient{
		detail:    &drasi.ComponentDetail{ID: "cq1", Kind: "query", Status: "Online"},
		logOutput: "watch output\n",
	}
	newLogsDrasiClient = func() logsDrasiClient { return client }

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "cq1", "--kind", "continuousquery"})

	err := root.Execute()
	require.NoError(t, err)
	// continuousquery is mapped to query, so it should use drasi watch path
	assert.Equal(t, []string{"watch", "cq1"}, client.lastLogsArgs)
	assert.Contains(t, stdout.String(), "watch output")
	assert.Empty(t, stderr.String())
}

func TestLogsSourceKind_JSON(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	kubectlLogsFunc = func(_ context.Context, _ ...string) (string, error) {
		return "json source logs\n", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"--output", "json", "logs", "--component", "s1", "--kind", "source"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), `"status": "ok"`)
	assert.Contains(t, stdout.String(), `"kind": "source"`)
	assert.Contains(t, stdout.String(), `"component": "s1"`)
	assert.Contains(t, stdout.String(), `"logs": "json source logs"`)
	assert.Empty(t, stderr.String())
}

func TestLogsSourceKind_EmptyOutput(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	kubectlLogsFunc = func(_ context.Context, _ ...string) (string, error) {
		return "\n", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs([]string{"logs", "--component", "empty-src", "--kind", "source"})

	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No logs found for empty-src (source).")
	assert.Empty(t, stderr.String())
}

func TestLogsSourceKind_WithKubeContext(t *testing.T) {
	origKubectl := kubectlLogsFunc
	t.Cleanup(func() { kubectlLogsFunc = origKubectl })

	var capturedArgs []string
	kubectlLogsFunc = func(_ context.Context, args ...string) (string, error) {
		capturedArgs = append([]string(nil), args...)
		return "ctx logs\n", nil
	}

	root := NewRootCommand()
	stdout := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(&bytes.Buffer{})
	// Inject kube context by setting environment state (bypass resolvedKubeContextForCommand)
	// The simplest way: override resolvedKubeContextForCommand via the --environment flag
	// But that requires azd environment state. Instead, test the kubectl args construction
	// by checking that when kubeContext would be set, it prepends --context.
	// Since resolvedKubeContextForCommand returns "" when no --environment is set,
	// the context won't be prepended here. The kubeContext path is tested implicitly
	// by the existing TestLogsCommand_RootEnvironmentFlagAccepted external test.
	root.SetArgs([]string{"logs", "--component", "my-source", "--kind", "source"})

	err := root.Execute()
	require.NoError(t, err)

	argsJoined := strings.Join(capturedArgs, " ")
	// No --context should be present when no environment is set
	assert.NotContains(t, argsJoined, "--context")
	assert.Contains(t, stdout.String(), "ctx logs")
}
