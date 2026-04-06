package output

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestErrorConstantsAreNonEmpty(t *testing.T) {
	t.Parallel()

	allCodes := []string{
		ERR_NO_AUTH,
		ERR_DRASI_CLI_NOT_FOUND,
		ERR_DRASI_CLI_VERSION,
		ERR_DRASI_CLI_ERROR,
		ERR_COMPONENT_TIMEOUT,
		ERR_TOTAL_TIMEOUT,
		ERR_VALIDATION_FAILED,
		ERR_MISSING_REFERENCE,
		ERR_CIRCULAR_DEPENDENCY,
		ERR_MISSING_QUERY_LANGUAGE,
		ERR_KV_AUTH_FAILED,
		ERR_AKS_CONTEXT_NOT_FOUND,
		ERR_FORCE_REQUIRED,
		ERR_NO_MANIFEST,
		ERR_DEPLOY_IN_PROGRESS,
		ERR_DAPR_NOT_READY,
	}

	if len(allCodes) != 16 {
		t.Fatalf("expected 16 error codes, got %d", len(allCodes))
	}

	for _, code := range allCodes {
		if strings.TrimSpace(code) == "" {
			t.Fatal("expected all ERR_* constants to be non-empty")
		}
	}
}

func TestFormatError_JSONMode(t *testing.T) {
	t.Parallel()

	got := FormatError(ERR_VALIDATION_FAILED, "validation failed", "fix the config", FormatJSON)

	var response ErrorResponse
	if err := json.Unmarshal([]byte(got), &response); err != nil {
		t.Fatalf("expected valid json error output, got error: %v; output=%q", err, got)
	}

	if response.Code != ERR_VALIDATION_FAILED {
		t.Fatalf("expected code %q, got %q", ERR_VALIDATION_FAILED, response.Code)
	}
	if response.Message != "validation failed" {
		t.Fatalf("expected message to round-trip, got %q", response.Message)
	}
	if response.Remediation != "fix the config" {
		t.Fatalf("expected remediation to round-trip, got %q", response.Remediation)
	}
	if response.Status != "error" {
		t.Fatalf("expected status error, got %q", response.Status)
	}
}

func TestFormatError_TableMode(t *testing.T) {
	t.Parallel()

	got := FormatError(ERR_FORCE_REQUIRED, "force is required", "rerun with --force", FormatTable)

	for _, want := range []string{ERR_FORCE_REQUIRED, "force is required", "rerun with --force"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected human-readable output to contain %q, got %q", want, got)
		}
	}
}

func TestExitCodesCoverAllErrorConstants(t *testing.T) {
	t.Parallel()

	allCodes := []string{
		ERR_NO_AUTH,
		ERR_DRASI_CLI_NOT_FOUND,
		ERR_DRASI_CLI_VERSION,
		ERR_DRASI_CLI_ERROR,
		ERR_COMPONENT_TIMEOUT,
		ERR_TOTAL_TIMEOUT,
		ERR_VALIDATION_FAILED,
		ERR_MISSING_REFERENCE,
		ERR_CIRCULAR_DEPENDENCY,
		ERR_MISSING_QUERY_LANGUAGE,
		ERR_KV_AUTH_FAILED,
		ERR_AKS_CONTEXT_NOT_FOUND,
		ERR_FORCE_REQUIRED,
		ERR_NO_MANIFEST,
		ERR_DEPLOY_IN_PROGRESS,
		ERR_DAPR_NOT_READY,
	}

	for _, code := range allCodes {
		exitCode, ok := ExitCodes[code]
		if !ok {
			t.Fatalf("expected exit code mapping for %q", code)
		}
		if exitCode <= 0 {
			t.Fatalf("expected positive exit code for %q, got %d", code, exitCode)
		}
	}
}
