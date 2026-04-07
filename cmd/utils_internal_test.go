package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/validation"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUtilityTestCommand(t *testing.T) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	return cmd, stdout, stderr
}

func TestCommandError_Error(t *testing.T) {
	t.Parallel()

	err := &commandError{message: "validation failed", exitCode: 1}
	require.NotNil(t, err)

	assert.Equal(t, "validation failed", err.Error())
}

func TestCountErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result *validation.ValidationResult
		want   int
	}{
		{
			name:   "empty result",
			result: &validation.ValidationResult{},
			want:   0,
		},
		{
			name: "warnings do not count",
			result: &validation.ValidationResult{Issues: []validation.ValidationIssue{
				{Level: validation.LevelWarning, Code: "WARN001", Message: "warning"},
				{Level: validation.LevelWarning, Code: "WARN002", Message: "warning"},
			}},
			want: 0,
		},
		{
			name: "only level errors count",
			result: &validation.ValidationResult{Issues: []validation.ValidationIssue{
				{Level: validation.LevelError, Code: "ERR001", Message: "error 1"},
				{Level: validation.LevelWarning, Code: "WARN001", Message: "warning"},
				{Level: validation.LevelError, Code: "ERR002", Message: "error 2"},
			}},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.NotNil(t, tt.result)

			assert.Equal(t, tt.want, countErrors(tt.result))
		})
	}
}

func TestIsCommandError_Classification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "command error",
			err:  &commandError{message: "failed", exitCode: 1},
			want: true,
		},
		{
			name: "wrapped command error",
			err:  errors.Join(errors.New("outer"), &commandError{message: "failed", exitCode: 1}),
			want: true,
		},
		{
			name: "plain error",
			err:  errors.New("plain"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, IsCommandError(tt.err))
		})
	}
}

func TestRenderValidation_Output(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		result       *validation.ValidationResult
		format       output.OutputFormat
		expectStdout []string
		expectStderr []string
		forbidStdout []string
		forbidStderr []string
		validateJSON func(t *testing.T, data []byte)
	}{
		{
			name:         "json ok payload",
			result:       &validation.ValidationResult{},
			format:       output.FormatJSON,
			expectStdout: []string{"\"status\": \"ok\"", "\"issues\": null"},
			validateJSON: func(t *testing.T, data []byte) {
				t.Helper()

				var payload struct {
					Status string                       `json:"status"`
					Issues []validation.ValidationIssue `json:"issues"`
				}
				require.NoError(t, json.Unmarshal(data, &payload))
				assert.Equal(t, "ok", payload.Status)
				assert.Nil(t, payload.Issues)
			},
		},
		{
			name: "json error payload",
			result: &validation.ValidationResult{Issues: []validation.ValidationIssue{
				{Level: validation.LevelError, File: "drasi.yaml", Line: 7, Code: "ERR001", Message: "broken"},
			}},
			format:       output.FormatJSON,
			expectStdout: []string{"\"status\": \"error\"", "\"code\": \"ERR001\""},
			validateJSON: func(t *testing.T, data []byte) {
				t.Helper()

				var payload struct {
					Status string                       `json:"status"`
					Issues []validation.ValidationIssue `json:"issues"`
				}
				require.NoError(t, json.Unmarshal(data, &payload))
				assert.Equal(t, "error", payload.Status)
				require.Len(t, payload.Issues, 1)
				assert.Equal(t, "ERR001", payload.Issues[0].Code)
			},
		},
		{
			name:         "table success message",
			result:       &validation.ValidationResult{},
			format:       output.FormatTable,
			expectStdout: []string{"Validation passed."},
			forbidStderr: []string{"ERR001"},
			forbidStdout: []string{"\"status\""},
		},
		{
			name: "table issues written to stderr",
			result: &validation.ValidationResult{Issues: []validation.ValidationIssue{
				{Level: validation.LevelError, File: "drasi/drasi.yaml", Line: 12, Code: "ERR001", Message: "bad source"},
			}},
			format:       output.FormatTable,
			expectStderr: []string{"[error] drasi/drasi.yaml:12 ERR001 - bad source"},
			forbidStdout: []string{"Validation passed.", "\"status\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, stdout, stderr := newUtilityTestCommand(t)
			require.NotNil(t, cmd)
			require.NotNil(t, tt.result)

			renderValidationOutput(cmd, tt.result, tt.format)

			for _, expected := range tt.expectStdout {
				assert.Contains(t, stdout.String(), expected)
			}
			for _, expected := range tt.expectStderr {
				assert.Contains(t, stderr.String(), expected)
			}
			for _, forbidden := range tt.forbidStdout {
				assert.NotContains(t, stdout.String(), forbidden)
			}
			for _, forbidden := range tt.forbidStderr {
				assert.NotContains(t, stderr.String(), forbidden)
			}
			if tt.validateJSON != nil {
				tt.validateJSON(t, bytes.TrimSpace(stdout.Bytes()))
			}
		})
	}
}

func TestWriteCommandError_ReturnsCommandError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		format         output.OutputFormat
		code           string
		message        string
		remediation    string
		exitCode       int
		expectStderr   []string
		expectErrValue string
	}{
		{
			name:        "table output",
			format:      output.FormatTable,
			code:        output.ERR_VALIDATION_FAILED,
			message:     "validation failed",
			remediation: "Fix the manifest.",
			exitCode:    output.ExitCodes[output.ERR_VALIDATION_FAILED],
			expectStderr: []string{
				"[ERR_VALIDATION_FAILED] validation failed",
				"Remediation: Fix the manifest.",
			},
			expectErrValue: "ERR_VALIDATION_FAILED: validation failed",
		},
		{
			name:        "json output",
			format:      output.FormatJSON,
			code:        output.ERR_NO_MANIFEST,
			message:     "missing manifest",
			remediation: "Create drasi.yaml.",
			exitCode:    output.ExitCodes[output.ERR_NO_MANIFEST],
			expectStderr: []string{
				"\"status\": \"error\"",
				"\"code\": \"ERR_NO_MANIFEST\"",
				"\"message\": \"missing manifest\"",
			},
			expectErrValue: "ERR_NO_MANIFEST: missing manifest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, _, stderr := newUtilityTestCommand(t)
			require.NotNil(t, cmd)

			err := writeCommandError(cmd, tt.code, tt.message, tt.remediation, tt.format, tt.exitCode)
			require.Error(t, err)

			cmdErr, ok := err.(*commandError)
			require.True(t, ok)
			assert.Equal(t, tt.expectErrValue, cmdErr.Error())
			assert.Equal(t, tt.exitCode, cmdErr.exitCode)

			for _, expected := range tt.expectStderr {
				assert.Contains(t, stderr.String(), expected)
			}
		})
	}
}

func TestDigitsOnly_Utility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "mixed characters", value: "v1.28.4+k3s1", want: "128431"},
		{name: "whitespace and punctuation", value: " 2.60.0 ", want: "2600"},
		{name: "no digits", value: "minor+", want: ""},
		{name: "digits only", value: "12345", want: "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, digitsOnly(tt.value))
		})
	}
}

func TestFailedPrerequisiteChecks_Filtering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		checks []prerequisiteCheck
		want   []prerequisiteCheck
	}{
		{
			name:   "empty input",
			checks: nil,
			want:   []prerequisiteCheck{},
		},
		{
			name: "no failures",
			checks: []prerequisiteCheck{
				{Tool: "azd", Status: statusPass},
				{Tool: "docker", Status: statusPass},
			},
			want: []prerequisiteCheck{},
		},
		{
			name: "only failures returned in order",
			checks: []prerequisiteCheck{
				{Tool: "azd", Status: statusPass},
				{Tool: "drasi", Status: statusFail, Remediation: "install drasi"},
				{Tool: "kubectl", Status: statusFail, Remediation: "install kubectl"},
			},
			want: []prerequisiteCheck{
				{Tool: "drasi", Status: statusFail, Remediation: "install drasi"},
				{Tool: "kubectl", Status: statusFail, Remediation: "install kubectl"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			failed := failedPrerequisiteChecks(tt.checks)
			assert.Equal(t, tt.want, failed)
		})
	}
}

func TestBuildPrerequisiteRemediation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		checks []prerequisiteCheck
		want   string
	}{
		{
			name:   "empty input",
			checks: nil,
			want:   "",
		},
		{
			name: "single remediation",
			checks: []prerequisiteCheck{
				{Tool: "drasi", Status: statusFail, Remediation: "Install drasi CLI."},
			},
			want: "- drasi: Install drasi CLI.",
		},
		{
			name: "multiple remediations",
			checks: []prerequisiteCheck{
				{Tool: "azd", Status: statusFail, Remediation: "Install azd."},
				{Tool: "kubectl", Status: statusFail, Remediation: "Install kubectl."},
			},
			want: "- azd: Install azd.\n- kubectl: Install kubectl.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, buildPrerequisiteRemediation(tt.checks))
		})
	}
}

func TestFallbackFoundVersion_Utility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "empty returns unknown", version: "", want: unknownVersion},
		{name: "whitespace returns unknown", version: " \t\n ", want: unknownVersion},
		{name: "trims value", version: " 1.2.3 ", want: "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, fallbackFoundVersion(tt.version))
		})
	}
}

func TestFormatCommandError_Utility(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("boom")

	tests := []struct {
		name           string
		stderr         string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "without stderr",
			stderr:         "",
			wantContains:   []string{"kubectl version", "boom"},
			wantNotContain: []string{"stderr:"},
		},
		{
			name:           "whitespace stderr ignored",
			stderr:         "  \n\t ",
			wantContains:   []string{"kubectl version", "boom"},
			wantNotContain: []string{"stderr:"},
		},
		{
			name:           "stderr included when present",
			stderr:         "  context deadline exceeded  ",
			wantContains:   []string{"kubectl version", "boom", "stderr: context deadline exceeded"},
			wantNotContain: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := formatCommandError("kubectl version", tt.stderr, baseErr)
			require.Error(t, err)
			assert.ErrorIs(t, err, baseErr)
			for _, expected := range tt.wantContains {
				assert.Contains(t, err.Error(), expected)
			}
			for _, unexpected := range tt.wantNotContain {
				assert.NotContains(t, err.Error(), unexpected)
			}
		})
	}
}

func TestOutputFormatFromCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  output.OutputFormat
	}{
		{name: "json flag", value: "json", want: output.FormatJSON},
		{name: "table flag", value: "table", want: output.FormatTable},
		{name: "unexpected flag falls back to table", value: "yaml", want: output.FormatTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := NewRootCommand()
			cmd, _, err := root.Find([]string{"status"})
			require.NoError(t, err)
			require.NotNil(t, cmd)
			require.NoError(t, cmd.Root().PersistentFlags().Set("output", tt.value))

			assert.Equal(t, tt.want, outputFormatFromCommand(cmd))
		})
	}
}

func TestWireKindForDrasiCLI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind string
		want string
	}{
		{name: "continuous query becomes query", kind: "continuousquery", want: "query"},
		{name: "source unchanged", kind: "source", want: "source"},
		{name: "middleware unchanged", kind: "middleware", want: "middleware"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, wireKindForDrasiCLI(tt.kind))
		})
	}
}

func TestNonNilResources(t *testing.T) {
	t.Parallel()

	t.Run("nil slice becomes empty slice", func(t *testing.T) {
		t.Parallel()

		got := nonNilResources(nil)
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("non nil slice is preserved", func(t *testing.T) {
		t.Parallel()

		resources := []drasi.ComponentSummary{{ID: "q1", Kind: "query", Status: "Running"}}
		got := nonNilResources(resources)
		require.Len(t, got, 1)
		assert.Equal(t, resources, got)
		assert.Same(t, &resources[0], &got[0])
	})
}

func TestSetVersion_StateMutation(t *testing.T) {
	originalVersion := extensionVersion
	t.Cleanup(func() {
		extensionVersion = originalVersion
	})

	SetVersion("9.9.9-test")

	assert.Equal(t, "9.9.9-test", extensionVersion)
}
