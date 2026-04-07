package cmd

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/azure/azd.extensions.drasi/internal/drasi"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

type diagnosticCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Detail      string `json:"detail"`
	Remediation string `json:"remediation"`
}

type diagnoseDrasiClient interface {
	CheckVersion(ctx context.Context) error
	ListComponents(ctx context.Context, kind string) ([]drasi.ComponentSummary, error)
	ListComponentsInContext(ctx context.Context, kind, kubeContext string) ([]drasi.ComponentSummary, error)
}

var newDiagnoseDrasiClient = func() diagnoseDrasiClient {
	return drasi.NewClient()
}

var kubectlClientVersionCheck = func(ctx context.Context, kubeContext string) error {
	args := []string{"version", "--client"}
	if strings.TrimSpace(kubeContext) != "" {
		args = append([]string{"--context", kubeContext}, args...)
	}
	_, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	return err
}

var kubectlOnPathCheck = func() error {
	_, err := exec.LookPath("kubectl")
	return err
}

func newDiagnoseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Run Drasi diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			format := outputFormatFromCommand(cmd)
			ctx := cmd.Context()
			kubeContext, err := resolvedKubeContextForCommand(ctx, cmd, "")
			if err != nil {
				code := errorCodeFromError(err, output.ERR_AKS_CONTEXT_NOT_FOUND)
				return writeCommandError(cmd, code, err.Error(), "Set the target azd environment and ensure AZURE_AKS_CONTEXT is present.", format, output.ExitCodes[code])
			}

			checks := []diagnosticCheck{}

			if err := kubectlClientVersionCheck(ctx, kubeContext); err != nil {
				return writeCommandError(
					cmd,
					output.ERR_AKS_CONTEXT_NOT_FOUND,
					fmt.Sprintf("checking AKS kubectl connectivity: %s", err),
					"Ensure kubectl is configured with the target AKS context.",
					format,
					output.ExitCodes[output.ERR_AKS_CONTEXT_NOT_FOUND],
				)
			}
			checks = append(checks, diagnosticCheck{
				Name:        "aks-connectivity",
				Status:      "ok",
				Detail:      "kubectl client is available for the active context",
				Remediation: "",
			})

			drasiClient := newDiagnoseDrasiClient()
			if err := drasiClient.CheckVersion(ctx); err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_NOT_FOUND)
				return writeCommandError(
					cmd,
					code,
					err.Error(),
					"Install or upgrade the drasi CLI and retry.",
					format,
					output.ExitCodes[code],
				)
			}
			checks = append(checks, diagnosticCheck{
				Name:        "drasi-cli",
				Status:      "ok",
				Detail:      "drasi CLI is available and meets minimum version",
				Remediation: "",
			})

			if err := kubectlOnPathCheck(); err != nil {
				return writeCommandError(
					cmd,
					output.ERR_DRASI_CLI_ERROR,
					"kubectl not found on PATH",
					"Install kubectl and ensure it is available on PATH.",
					format,
					output.ExitCodes[output.ERR_DRASI_CLI_ERROR],
				)
			}
			checks = append(checks, diagnosticCheck{
				Name:        "kubectl",
				Status:      "ok",
				Detail:      "kubectl is available",
				Remediation: "",
			})

			daprReady, daprDetail, daprErr := isDaprReady(ctx, kubeContext)
			if daprErr != nil {
				return writeCommandError(
					cmd,
					output.ERR_DAPR_NOT_READY,
					fmt.Sprintf("checking Dapr runtime: %s", daprErr),
					"Ensure kubeconfig points to the target cluster and retry.",
					format,
					output.ExitCodes[output.ERR_DAPR_NOT_READY],
				)
			}
			if !daprReady {
				return writeCommandError(
					cmd,
					output.ERR_DAPR_NOT_READY,
					daprDetail,
					"Install or restart Dapr in the target AKS cluster.",
					format,
					output.ExitCodes[output.ERR_DAPR_NOT_READY],
				)
			}
			checks = append(checks, diagnosticCheck{
				Name:        "dapr-runtime",
				Status:      "ok",
				Detail:      daprDetail,
				Remediation: "",
			})

			if kubeContext == "" {
				_, err = drasiClient.ListComponents(ctx, "source")
			} else {
				_, err = drasiClient.ListComponentsInContext(ctx, "source", kubeContext)
			}
			if err != nil {
				code := errorCodeFromError(err, output.ERR_DRASI_CLI_ERROR)
				return writeCommandError(
					cmd,
					code,
					fmt.Sprintf("checking Drasi API health: %s", err),
					"Verify Drasi runtime is installed and reachable from kubectl context.",
					format,
					output.ExitCodes[code],
				)
			}
			checks = append(checks, diagnosticCheck{
				Name:        "drasi-api",
				Status:      "ok",
				Detail:      "Drasi API responded to list operation",
				Remediation: "",
			})

			checks = append(checks, diagnosticCheck{
				Name:        "key-vault-auth",
				Status:      "skipped",
				Detail:      "Key Vault auth requires a deployed secret reference; validated at provision/deploy time",
				Remediation: "Deploy a component with a Key Vault secret reference to validate auth end-to-end.",
			})

			checks = append(checks, diagnosticCheck{
				Name:        "log-analytics",
				Status:      "skipped",
				Detail:      "Log Analytics workspace wiring is validated during provision and runtime observability checks",
				Remediation: "Run `azd drasi provision` to validate Log Analytics workspace integration.",
			})

			payload := map[string]any{"status": "ok", "checks": checks}
			if format == output.FormatJSON {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(payload, output.FormatJSON))
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), output.Format(checks, output.FormatTable))
			return nil
		},
	}

	return cmd
}

var isDaprReady = func(ctx context.Context, kubeContext string) (bool, string, error) {
	args := []string{"get", "pods", "-n", "dapr-system", "-l", "app=dapr-operator", "--no-headers"}
	if strings.TrimSpace(kubeContext) != "" {
		args = append([]string{"--context", kubeContext}, args...)
	}
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	if err != nil {
		return false, "", err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return false, "no Dapr operator pod found in dapr-system namespace", nil
	}
	return true, "Dapr operator pod is present", nil
}
