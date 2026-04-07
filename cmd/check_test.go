package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCommand_TableDriven(t *testing.T) {
	// NOTE: not parallel because subtests mutate package-level checkLookPath and checkRunCommand.

	tests := []struct {
		name               string
		lookPath           func(string) (string, error)
		runCommand         func(context.Context, string, ...string) (string, string, error)
		args               []string
		wantErr            bool
		wantErrCode        string
		wantStdoutContains []string
		wantStderrContains []string
	}{
		{
			name:     "all tools present and valid versions",
			lookPath: fakeLookPathSuccess,
			runCommand: func(_ context.Context, path string, args ...string) (string, string, error) {
				return fakeCommandOutput(path, args...), "", nil
			},
			args:    []string{"check"},
			wantErr: false,
			wantStdoutContains: []string{
				"tool",
				"azd",
				"drasi",
				"docker",
				"kubectl",
				statusPass,
			},
		},
		{
			name: "missing drasi on path",
			lookPath: func(file string) (string, error) {
				if file == "drasi" {
					return "", exec.ErrNotFound
				}
				return fakeLookPathSuccess(file)
			},
			runCommand: func(_ context.Context, path string, args ...string) (string, string, error) {
				return fakeCommandOutput(path, args...), "", nil
			},
			args:        []string{"check"},
			wantErr:     true,
			wantErrCode: output.ERR_VALIDATION_FAILED,
			wantStdoutContains: []string{
				"drasi",
				notFoundVersion,
				statusFail,
			},
			wantStderrContains: []string{
				"Install or upgrade the drasi CLI",
			},
		},
		{
			name:     "json output format",
			lookPath: fakeLookPathSuccess,
			runCommand: func(_ context.Context, path string, args ...string) (string, string, error) {
				return fakeCommandOutput(path, args...), "", nil
			},
			args:    []string{"--output", "json", "check"},
			wantErr: false,
			wantStdoutContains: []string{
				`"status": "ok"`,
				`"checks"`,
				`"tool": "azd"`,
				`"tool": "kubectl"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restoreCheckTestHooks(t, tt.lookPath, tt.runCommand)

			root := NewRootCommand()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			root.SetOut(stdout)
			root.SetErr(stderr)
			root.SetArgs(tt.args)

			err := root.Execute()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrCode)
			} else {
				require.NoError(t, err)
			}

			for _, want := range tt.wantStdoutContains {
				assert.Contains(t, stdout.String(), want)
			}

			for _, want := range tt.wantStderrContains {
				assert.Contains(t, stderr.String(), want)
			}
		})
	}
}

func restoreCheckTestHooks(
	t *testing.T,
	lookPath func(string) (string, error),
	runCommand func(context.Context, string, ...string) (string, string, error),
) {
	t.Helper()
	originalLookPath := checkLookPath
	originalRunCommand := checkRunCommand
	checkLookPath = lookPath
	checkRunCommand = runCommand
	t.Cleanup(func() {
		checkLookPath = originalLookPath
		checkRunCommand = originalRunCommand
	})
}

func fakeLookPathSuccess(file string) (string, error) {
	return fmt.Sprintf("/fake/%s", file), nil
}

func fakeCommandOutput(path string, args ...string) string {
	command := strings.TrimPrefix(path, "/fake/")
	switch command {
	case "azd":
		return "azd version 1.10.1"
	case "drasi":
		return "drasi version: 0.10.2"
	case "az":
		return `{"azure-cli":"2.61.0"}`
	case "docker":
		return "Docker version 24.0.7, build deadbeef"
	case "kubectl":
		return `{"clientVersion":{"gitVersion":"v1.28.4","major":"1","minor":"28"}}`
	default:
		panic(fmt.Sprintf("unexpected command: %s %v", path, args))
	}
}
