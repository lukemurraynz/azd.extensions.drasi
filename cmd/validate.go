package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/azure/azd.extensions.drasi/internal/validation"
	"github.com/spf13/cobra"
)

type commandError struct {
	message  string
	exitCode int
}

func (e *commandError) Error() string {
	return e.message
}

func newValidateCommand() *cobra.Command {
	var configPath string
	var strict bool
	var envName string

	command := &cobra.Command{
		Use:   "validate",
		Short: "Validate Drasi configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")
			format := output.FormatTable
			if outputFormat == string(output.FormatJSON) {
				format = output.FormatJSON
			}

			absoluteConfigPath, err := filepath.Abs(configPath)
			if err != nil {
				return writeCommandError(cmd, output.ERR_NO_MANIFEST, fmt.Sprintf("cannot resolve config path %q: %s", configPath, err), "Ensure the --config path is valid.", format, output.ExitCodes[output.ERR_NO_MANIFEST])
			}

			result, err := validation.Validate(filepath.Dir(absoluteConfigPath), filepath.Base(absoluteConfigPath), envName)
			if err != nil {
				return writeCommandError(cmd, output.ERR_NO_MANIFEST, err.Error(), "Ensure the manifest file exists and is valid YAML.", format, output.ExitCodes[output.ERR_NO_MANIFEST])
			}

			issues := append([]validation.ValidationIssue(nil), result.Issues...)
			if strict {
				for i := range issues {
					if issues[i].Level == validation.LevelWarning {
						issues[i].Level = validation.LevelError
					}
				}
			}

			effective := &validation.ValidationResult{Issues: issues}
			renderValidationOutput(cmd, effective, format)

			if effective.HasErrors() {
				return &commandError{message: fmt.Sprintf("validation failed with %d issue(s)", countErrors(effective)), exitCode: output.ExitCodes[output.ERR_VALIDATION_FAILED]}
			}
			return nil
		},
	}

	command.Flags().StringVar(&configPath, "config", filepath.Join("drasi", "drasi.yaml"), "Path to drasi.yaml manifest")
	command.Flags().BoolVar(&strict, "strict", false, "Promote warnings to errors")
	command.Flags().StringVar(&envName, "environment", "", "Environment overlay to validate")

	return command
}

func renderValidationOutput(cmd *cobra.Command, result *validation.ValidationResult, format output.OutputFormat) {
	if format == output.FormatJSON {
		status := "ok"
		if result.HasErrors() {
			status = "error"
		}
		payload := map[string]any{
			"status": status,
			"issues":  result.Issues,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err == nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		}
		return
	}

	if len(result.Issues) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Validation passed.")
		return
	}

	for _, issue := range result.Issues {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[%s] %s:%d %s - %s\n", issue.Level, issue.File, issue.Line, issue.Code, issue.Message)
	}
}

func writeCommandError(cmd *cobra.Command, code, message, remediation string, format output.OutputFormat, exitCode int) error {
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output.FormatError(code, message, remediation, format))
	// Include the error code in the message so callers (and tests) can inspect it via err.Error().
	return &commandError{message: fmt.Sprintf("%s: %s", code, message), exitCode: exitCode}
}

func countErrors(result *validation.ValidationResult) int {
	count := 0
	for _, issue := range result.Issues {
		if issue.Level == validation.LevelError {
			count++
		}
	}
	return count
}

func IsCommandError(err error) bool {
	var target *commandError
	return errors.As(err, &target)
}
