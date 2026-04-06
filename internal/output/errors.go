package output

import (
	"encoding/json"
	fmtpkg "fmt"
)

const (
	ERR_NO_AUTH                = "ERR_NO_AUTH"
	ERR_DRASI_CLI_NOT_FOUND    = "ERR_DRASI_CLI_NOT_FOUND"
	ERR_DRASI_CLI_VERSION      = "ERR_DRASI_CLI_VERSION"
	ERR_DRASI_CLI_ERROR        = "ERR_DRASI_CLI_ERROR"
	ERR_COMPONENT_TIMEOUT      = "ERR_COMPONENT_TIMEOUT"
	ERR_TOTAL_TIMEOUT          = "ERR_TOTAL_TIMEOUT"
	ERR_VALIDATION_FAILED      = "ERR_VALIDATION_FAILED"
	ERR_MISSING_REFERENCE      = "ERR_MISSING_REFERENCE"
	ERR_CIRCULAR_DEPENDENCY    = "ERR_CIRCULAR_DEPENDENCY"
	ERR_MISSING_QUERY_LANGUAGE = "ERR_MISSING_QUERY_LANGUAGE"
	ERR_KV_AUTH_FAILED         = "ERR_KV_AUTH_FAILED"
	ERR_AKS_CONTEXT_NOT_FOUND  = "ERR_AKS_CONTEXT_NOT_FOUND"
	ERR_FORCE_REQUIRED         = "ERR_FORCE_REQUIRED"
	ERR_NO_MANIFEST            = "ERR_NO_MANIFEST"
	ERR_DEPLOY_IN_PROGRESS     = "ERR_DEPLOY_IN_PROGRESS"
	ERR_DAPR_NOT_READY         = "ERR_DAPR_NOT_READY"
)

type ErrorResponse struct {
	Status      string `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	Remediation string `json:"remediation"`
	Detail      any    `json:"detail,omitempty"`
}

func FormatError(code, msg, remediation string, fmt OutputFormat) string {
	response := ErrorResponse{
		Status:      "error",
		Code:        code,
		Message:     msg,
		Remediation: remediation,
		Detail:      map[string]any{},
	}

	if fmt == FormatJSON {
		payload, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return ""
		}
		return string(payload)
	}

	return fmtpkg.Sprintf("[%s] %s\nRemediation: %s", code, msg, remediation)
}

var ExitCodes = map[string]int{
	ERR_NO_AUTH:                2,
	ERR_DRASI_CLI_NOT_FOUND:    2,
	ERR_DRASI_CLI_VERSION:      2,
	ERR_DRASI_CLI_ERROR:        1,
	ERR_COMPONENT_TIMEOUT:      1,
	ERR_TOTAL_TIMEOUT:          1,
	ERR_VALIDATION_FAILED:      1,
	ERR_MISSING_REFERENCE:      1,
	ERR_CIRCULAR_DEPENDENCY:    1,
	ERR_MISSING_QUERY_LANGUAGE: 1,
	ERR_KV_AUTH_FAILED:         2,
	ERR_AKS_CONTEXT_NOT_FOUND:  2,
	ERR_FORCE_REQUIRED:         2,
	ERR_NO_MANIFEST:            2,
	ERR_DEPLOY_IN_PROGRESS:     2,
	ERR_DAPR_NOT_READY:         2,
}
