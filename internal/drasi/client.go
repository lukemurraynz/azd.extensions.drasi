package drasi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/azure/azd.extensions.drasi/internal/output"
)

// minimumVersion is the lowest drasi CLI version this extension supports (FR-046).
const minimumVersion = "0.10.0"

// commandRunner abstracts drasi CLI execution for tests.
type commandRunner interface {
	Run(ctx context.Context, args ...string) (stdout, stderr string, exitCode int, err error)
}

// realRunner implements commandRunner using the drasi binary on PATH.
type realRunner struct{}

func (r *realRunner) Run(ctx context.Context, args ...string) (string, string, int, error) {
	path, err := exec.LookPath("drasi")
	if err != nil {
		return "", "", -1, err
	}

	cmd := exec.CommandContext(ctx, path, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
			runErr = nil // exitCode carries the signal; clear the error
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode, runErr
}

// Client wraps the drasi CLI binary.
type Client struct {
	runner commandRunner
}

// NewClient creates a Client using the drasi binary on PATH.
func NewClient() *Client {
	return &Client{runner: &realRunner{}}
}

// CheckVersion verifies the drasi binary meets the minimum version requirement.
func (c *Client) CheckVersion(ctx context.Context) error {
	stdout, _, _, err := c.runner.Run(ctx, "version")
	if err != nil {
		return fmt.Errorf("%s: %w", output.ERR_DRASI_CLI_NOT_FOUND, err)
	}

	raw := parseSemverFromVersionOutput(stdout)
	got, parseErr := semver.NewVersion(raw)
	if parseErr != nil {
		return fmt.Errorf("%s: cannot parse version %q: %w", output.ERR_DRASI_CLI_VERSION, raw, parseErr)
	}

	min, _ := semver.NewVersion(minimumVersion)
	if got.LessThan(min) {
		return fmt.Errorf("%s: drasi CLI version %s is below minimum required %s", output.ERR_DRASI_CLI_VERSION, got, min)
	}

	return nil
}

func parseSemverFromVersionOutput(stdout string) string {
	raw := strings.TrimSpace(stdout)
	if raw == "" {
		return raw
	}

	line := raw
	if lines := strings.Split(raw, "\n"); len(lines) > 0 {
		line = strings.TrimSpace(lines[0])
	}

	if idx := strings.IndexByte(line, ':'); idx >= 0 {
		line = strings.TrimSpace(line[idx+1:])
	}

	if strings.HasPrefix(line, "v") || strings.HasPrefix(line, "V") {
		line = strings.TrimSpace(line[1:])
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return strings.TrimSpace(stdout)
	}

	return line
}

// RunCommand executes a drasi CLI subcommand and returns an error on non-zero exit.
// No automatic retry is performed (FR-046).
func (c *Client) RunCommand(ctx context.Context, args ...string) error {
	_, stderr, exitCode, err := c.runner.Run(ctx, args...)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("%s: drasi %s: %s", output.ERR_DRASI_CLI_ERROR, strings.Join(args, " "), strings.TrimSpace(stderr))
	}
	return nil
}

// RunCommandOutput executes a drasi CLI subcommand and returns stdout on success.
// No automatic retry is performed (FR-046).
func (c *Client) RunCommandOutput(ctx context.Context, args ...string) (string, error) {
	stdout, stderr, exitCode, err := c.runner.Run(ctx, args...)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", fmt.Errorf("%s: drasi %s: %s", output.ERR_DRASI_CLI_ERROR, strings.Join(args, " "), strings.TrimSpace(stderr))
	}
	return stdout, nil
}
