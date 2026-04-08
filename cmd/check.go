package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	drasi "github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

const (
	minimumAzdVersion      = "1.10.0"
	minimumDrasiVersion    = "0.10.0"
	minimumAzureCLIVersion = "2.60.0"
	minimumKubectlVersion  = "1.28.0"
	installedRequirement   = "installed"
	notFoundVersion        = "not found"
	unknownVersion         = "unknown"
	statusPass             = "pass"
	statusFail             = "fail"
)

var (
	versionPattern = regexp.MustCompile(`(?i)\bv?(\d+\.\d+(?:\.\d+)?)\b`)

	checkLookPath   = exec.LookPath
	checkRunCommand = func(ctx context.Context, path string, args ...string) (string, string, error) {
		cmd := exec.CommandContext(ctx, path, args...) //nolint:gosec // path is from exec.LookPath on known binary names
		var stdoutBuf bytes.Buffer
		var stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		err := cmd.Run()
		return stdoutBuf.String(), stderrBuf.String(), err
	}
)

type prerequisiteCheck struct {
	Tool            string `json:"tool"`
	RequiredVersion string `json:"requiredVersion"`
	FoundVersion    string `json:"foundVersion"`
	Status          string `json:"status"`
	Remediation     string `json:"remediation"`
}

type azVersionResponse struct {
	AzureCLI string `json:"azure-cli"`
}

type kubectlVersionResponse struct {
	ClientVersion kubectlClientVersion `json:"clientVersion"`
}

type kubectlClientVersion struct {
	GitVersion string `json:"gitVersion"`
	Major      string `json:"major"`
	Minor      string `json:"minor"`
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check prerequisites for Drasi extension",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			ctx := cmd.Context()

			progress, err := NewProgressHelper(cmd)
			if err != nil {
				progress = &ProgressHelper{noop: true}
			}
			_ = progress.Start()
			defer func() { _ = progress.Stop() }()

			checks := []prerequisiteCheck{
				runVersionCheck(ctx, progress, "Checking azd...", "azd", minimumAzdVersion, detectAzdVersion, "Install or upgrade Azure Developer CLI to >= 1.10.0 from https://aka.ms/azd."),
				runVersionCheck(ctx, progress, "Checking drasi...", "drasi", minimumDrasiVersion, detectDrasiVersion, "Install or upgrade the drasi CLI to >= 0.10.0 from https://drasi.io/docs/getting-started."),
				runVersionCheck(ctx, progress, "Checking az...", "az", minimumAzureCLIVersion, detectAzureCLIVersion, "Install or upgrade Azure CLI to >= 2.60.0 from https://aka.ms/azcli."),
				runPresenceCheck(ctx, progress, "Checking docker...", "docker", installedRequirement, detectDockerVersion, "Install Docker and ensure the docker CLI is available on PATH."),
				runVersionCheck(ctx, progress, "Checking kubectl...", "kubectl", minimumKubectlVersion, detectKubectlVersion, "Install or upgrade kubectl to >= 1.28 from https://kubernetes.io/docs/tasks/tools/."),
			}

			_ = progress.Stop()

			payload := map[string]any{
				"status": "ok",
				"checks": checks,
			}

			failedChecks := failedPrerequisiteChecks(checks)
			if len(failedChecks) > 0 {
				payload["status"] = "error"
			}

			if format == output.FormatJSON {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(checks, output.FormatTable))
			}

			if len(failedChecks) == 0 {
				return nil
			}

			return writeCommandError(
				cmd,
				output.ERR_VALIDATION_FAILED,
				"prerequisite check failed",
				buildPrerequisiteRemediation(failedChecks),
				format,
				output.ExitCodes[output.ERR_VALIDATION_FAILED],
			)
		},
	}
}

func runVersionCheck(
	ctx context.Context,
	progress *ProgressHelper,
	message string,
	tool string,
	requiredVersion string,
	versionDetector func(context.Context, string) (string, error),
	remediation string,
) prerequisiteCheck {
	progress.Message(message)

	path, err := checkLookPath(tool)
	if err != nil {
		return prerequisiteCheck{
			Tool:            tool,
			RequiredVersion: requiredVersion,
			FoundVersion:    notFoundVersion,
			Status:          statusFail,
			Remediation:     remediation,
		}
	}

	foundVersion, err := versionDetector(ctx, path)
	if err != nil {
		return prerequisiteCheck{
			Tool:            tool,
			RequiredVersion: requiredVersion,
			FoundVersion:    fallbackFoundVersion(foundVersion),
			Status:          statusFail,
			Remediation:     fmt.Sprintf("%s Details: %s", remediation, err.Error()),
		}
	}

	ok, err := isVersionAtLeast(foundVersion, requiredVersion)
	if err != nil {
		return prerequisiteCheck{
			Tool:            tool,
			RequiredVersion: requiredVersion,
			FoundVersion:    fallbackFoundVersion(foundVersion),
			Status:          statusFail,
			Remediation:     fmt.Sprintf("%s Details: %s", remediation, err.Error()),
		}
	}
	if !ok {
		return prerequisiteCheck{
			Tool:            tool,
			RequiredVersion: requiredVersion,
			FoundVersion:    foundVersion,
			Status:          statusFail,
			Remediation:     remediation,
		}
	}

	return prerequisiteCheck{
		Tool:            tool,
		RequiredVersion: requiredVersion,
		FoundVersion:    foundVersion,
		Status:          statusPass,
		Remediation:     "",
	}
}

func runPresenceCheck(
	ctx context.Context,
	progress *ProgressHelper,
	message string,
	tool string,
	requiredVersion string,
	versionDetector func(context.Context, string) (string, error),
	remediation string,
) prerequisiteCheck {
	progress.Message(message)

	path, err := checkLookPath(tool)
	if err != nil {
		return prerequisiteCheck{
			Tool:            tool,
			RequiredVersion: requiredVersion,
			FoundVersion:    notFoundVersion,
			Status:          statusFail,
			Remediation:     remediation,
		}
	}

	foundVersion, err := versionDetector(ctx, path)
	if err != nil {
		return prerequisiteCheck{
			Tool:            tool,
			RequiredVersion: requiredVersion,
			FoundVersion:    fallbackFoundVersion(foundVersion),
			Status:          statusFail,
			Remediation:     fmt.Sprintf("%s Details: %s", remediation, err.Error()),
		}
	}

	return prerequisiteCheck{
		Tool:            tool,
		RequiredVersion: requiredVersion,
		FoundVersion:    foundVersion,
		Status:          statusPass,
		Remediation:     "",
	}
}

func detectAzdVersion(ctx context.Context, path string) (string, error) {
	stdout, stderr, err := checkRunCommand(ctx, path, "version")
	if err != nil {
		parsedVersion, parseErr := extractFirstVersion(strings.TrimSpace(stdout))
		if parseErr != nil {
			parsedVersion = ""
		}
		return parsedVersion, formatCommandError("azd version", stderr, err)
	}

	version, parseErr := extractFirstVersion(strings.TrimSpace(stdout))
	if parseErr != nil {
		return "", parseErr
	}

	return version, nil
}

func detectDrasiVersion(ctx context.Context, path string) (string, error) {
	stdout, stderr, err := checkRunCommand(ctx, path, "version")
	if err != nil {
		return drasi.ParseSemverFromVersionOutput(stdout), formatCommandError("drasi version", stderr, err)
	}

	version := drasi.ParseSemverFromVersionOutput(stdout)
	if strings.TrimSpace(version) == "" {
		return "", errors.New("cannot parse drasi version output")
	}

	return normalizeVersion(version), nil
}

func detectAzureCLIVersion(ctx context.Context, path string) (string, error) {
	stdout, stderr, err := checkRunCommand(ctx, path, "version", "--output", "json")
	if err != nil {
		return "", formatCommandError("az version --output json", stderr, err)
	}

	var response azVersionResponse
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		return "", fmt.Errorf("parsing az version output: %w", err)
	}
	if strings.TrimSpace(response.AzureCLI) == "" {
		return "", errors.New("az version output missing azure-cli field")
	}

	return normalizeVersion(response.AzureCLI), nil
}

func detectDockerVersion(ctx context.Context, path string) (string, error) {
	stdout, stderr, err := checkRunCommand(ctx, path, "--version")
	if err != nil {
		parsedVersion, parseErr := extractFirstVersion(strings.TrimSpace(stdout))
		if parseErr != nil {
			parsedVersion = ""
		}
		return parsedVersion, formatCommandError("docker --version", stderr, err)
	}

	version, parseErr := extractFirstVersion(strings.TrimSpace(stdout))
	if parseErr != nil {
		return unknownVersion, nil
	}

	return version, nil
}

func detectKubectlVersion(ctx context.Context, path string) (string, error) {
	stdout, stderr, err := checkRunCommand(ctx, path, "version", "--client", "--output", "json")
	if err != nil {
		return "", formatCommandError("kubectl version --client --output json", stderr, err)
	}

	var response kubectlVersionResponse
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		return "", fmt.Errorf("parsing kubectl version output: %w", err)
	}

	if strings.TrimSpace(response.ClientVersion.GitVersion) != "" {
		version, err := extractFirstVersion(response.ClientVersion.GitVersion)
		if err != nil {
			return "", err
		}
		return version, nil
	}

	major := digitsOnly(response.ClientVersion.Major)
	minor := digitsOnly(response.ClientVersion.Minor)
	if major == "" || minor == "" {
		return "", errors.New("kubectl version output missing clientVersion")
	}

	return normalizeVersion(fmt.Sprintf("%s.%s", major, minor)), nil
}

func failedPrerequisiteChecks(checks []prerequisiteCheck) []prerequisiteCheck {
	failed := make([]prerequisiteCheck, 0)
	for _, check := range checks {
		if check.Status == statusFail {
			failed = append(failed, check)
		}
	}
	return failed
}

func buildPrerequisiteRemediation(checks []prerequisiteCheck) string {
	lines := make([]string, 0, len(checks))
	for _, check := range checks {
		lines = append(lines, fmt.Sprintf("- %s: %s", check.Tool, check.Remediation))
	}
	return strings.Join(lines, "\n")
}

func fallbackFoundVersion(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return unknownVersion
	}
	return trimmed
}

func formatCommandError(command string, stderr string, err error) error {
	trimmedStderr := strings.TrimSpace(stderr)
	if trimmedStderr == "" {
		return fmt.Errorf("%s: %w", command, err)
	}
	return fmt.Errorf("%s: %w (stderr: %s)", command, err, trimmedStderr)
}

func extractFirstVersion(text string) (string, error) {
	match := versionPattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return "", fmt.Errorf("cannot parse version from %q", text)
	}
	return normalizeVersion(match[1]), nil
}

func normalizeVersion(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "v")
	trimmed = strings.TrimPrefix(trimmed, "V")

	parts := strings.Split(trimmed, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}

	return strings.Join(parts[:3], ".")
}

func digitsOnly(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if r < '0' || r > '9' {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func isVersionAtLeast(foundVersion string, minimumVersion string) (bool, error) {
	found, err := semver.NewVersion(normalizeVersion(foundVersion))
	if err != nil {
		return false, fmt.Errorf("parsing found version %q: %w", foundVersion, err)
	}

	minimum, err := semver.NewVersion(normalizeVersion(minimumVersion))
	if err != nil {
		return false, fmt.Errorf("parsing minimum version %q: %w", minimumVersion, err)
	}

	return !found.LessThan(minimum), nil
}
