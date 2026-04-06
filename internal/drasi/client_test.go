package drasi

import (
	"context"
	"errors"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runnerResponse struct {
	stdout   string
	stderr   string
	exitCode int
	err      error
}

type mockRunner struct {
	responses []runnerResponse
	callCount int
	args      [][]string
}

func (m *mockRunner) Run(ctx context.Context, args ...string) (string, string, int, error) {
	m.callCount++
	m.args = append(m.args, append([]string(nil), args...))

	if len(m.responses) == 0 {
		return "", "", 0, nil
	}

	index := m.callCount - 1
	if index >= len(m.responses) {
		index = len(m.responses) - 1
	}

	resp := m.responses[index]
	return resp.stdout, resp.stderr, resp.exitCode, resp.err
}

func newTestClient(r commandRunner) *Client {
	return &Client{runner: r}
}

func TestClient_CheckVersion_PassesAtMinimumVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
	}{
		{name: "minimum version accepted", version: "0.10.0"},
		{name: "prefixed version accepted", version: "v0.10.0"},
		{name: "labelled version accepted", version: "Drasi CLI version: 0.10.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stdout: tt.version}}}
			client := newTestClient(runner)

			err := client.CheckVersion(context.Background())

			require.NoError(t, err)
			assert.Equal(t, 1, runner.callCount)
		})
	}
}

func TestClient_CheckVersion_FailsBelowMinimum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "below minimum version rejected", version: "0.9.2", want: output.ERR_DRASI_CLI_VERSION},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stdout: tt.version}}}
			client := newTestClient(runner)

			err := client.CheckVersion(context.Background())

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestClient_CheckVersion_BinaryNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "missing binary returns not found error", err: errors.New("executable file not found in %PATH%"), want: output.ERR_DRASI_CLI_NOT_FOUND},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{err: tt.err}}}
			client := newTestClient(runner)

			err := client.CheckVersion(context.Background())

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestClient_RunCommand_NonZeroExit_MapsToCliError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		stderr   string
		exitCode int
		want     string
	}{
		{name: "non-zero exit maps to CLI error", args: []string{"list", "query"}, stderr: "drasi command failed", exitCode: 1, want: output.ERR_DRASI_CLI_ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: tt.exitCode}}}
			client := newTestClient(runner)

			err := client.RunCommand(context.Background(), tt.args...)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestClient_RunCommand_NoAutomaticRetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		stderr   string
		exitCode int
	}{
		{name: "non-zero exit is attempted once", args: []string{"apply", "-f", "manifest.yaml"}, stderr: "boom", exitCode: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: tt.exitCode}}}
			client := newTestClient(runner)

			err := client.RunCommand(context.Background(), tt.args...)

			require.Error(t, err)
			assert.Equal(t, 1, runner.callCount)
		})
	}
}

func TestClient_RunCommandOutput_ReturnsStdout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		stdout     string
		wantOutput string
	}{
		{name: "returns stdout on success", args: []string{"list", "source", "--output", "json"}, stdout: `[{"id":"s1"}]`, wantOutput: `[{"id":"s1"}]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stdout: tt.stdout}}}
			client := newTestClient(runner)

			out, err := client.RunCommandOutput(context.Background(), tt.args...)

			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, out)
		})
	}
}

func TestClient_RunCommandOutput_NonZeroExit_MapsToCliError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		stderr   string
		exitCode int
		want     string
	}{
		{name: "non-zero exit maps to CLI error", args: []string{"list", "source"}, stderr: "drasi command failed", exitCode: 1, want: output.ERR_DRASI_CLI_ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: tt.exitCode}}}
			client := newTestClient(runner)

			_, err := client.RunCommandOutput(context.Background(), tt.args...)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestClient_RunCommandOutput_NoAutomaticRetry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		stderr   string
		exitCode int
	}{
		{name: "non-zero exit is attempted once", args: []string{"list", "source"}, stderr: "boom", exitCode: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := &mockRunner{responses: []runnerResponse{{stderr: tt.stderr, exitCode: tt.exitCode}}}
			client := newTestClient(runner)

			_, err := client.RunCommandOutput(context.Background(), tt.args...)

			require.Error(t, err)
			assert.Equal(t, 1, runner.callCount)
		})
	}
}
