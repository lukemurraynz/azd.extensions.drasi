package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/drasi"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
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

			// Key Vault check
			vaultName := os.Getenv("AZURE_KEYVAULT_NAME")
			if strings.TrimSpace(vaultName) == "" {
				checks = append(checks, diagnosticCheck{
					Name:        "key-vault-auth",
					Status:      "skipped",
					Detail:      "AZURE_KEYVAULT_NAME not set; skipping Key Vault connectivity check",
					Remediation: "Set AZURE_KEYVAULT_NAME in the azd environment to enable this check.",
				})
			} else {
				kvStatus, kvDetail, kvErr := azKeyVaultCheck(ctx, vaultName)
				if kvErr != nil {
					kvStatus = "failed"
					kvDetail = kvErr.Error()
				}
				remediation := ""
				if kvStatus == "failed" {
					remediation = "Ensure the Key Vault exists and the managed identity has 'Key Vault Secrets User' role."
				}
				checks = append(checks, diagnosticCheck{Name: "key-vault-auth", Status: kvStatus, Detail: kvDetail, Remediation: remediation})
			}

			// Log Analytics check
			wsName := os.Getenv("AZURE_LOG_ANALYTICS_WORKSPACE_NAME")
			rgName := os.Getenv("AZURE_RESOURCE_GROUP")
			if strings.TrimSpace(wsName) == "" || strings.TrimSpace(rgName) == "" {
				checks = append(checks, diagnosticCheck{
					Name:        "log-analytics",
					Status:      "skipped",
					Detail:      "AZURE_LOG_ANALYTICS_WORKSPACE_NAME or AZURE_RESOURCE_GROUP not set; skipping Log Analytics check",
					Remediation: "Set AZURE_LOG_ANALYTICS_WORKSPACE_NAME and AZURE_RESOURCE_GROUP to enable this check.",
				})
			} else {
				laStatus, laDetail, laErr := azLogAnalyticsCheck(ctx, rgName, wsName)
				if laErr != nil {
					laStatus = "failed"
					laDetail = laErr.Error()
				}
				remediation := ""
				if laStatus == "failed" {
					remediation = "Ensure the Log Analytics workspace exists and is accessible from the current subscription."
				}
				checks = append(checks, diagnosticCheck{Name: "log-analytics", Status: laStatus, Detail: laDetail, Remediation: remediation})
			}

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

// azKeyVaultCheck shells out to az CLI to verify Key Vault accessibility.
var azKeyVaultCheck = func(ctx context.Context, vaultName string) (string, string, error) {
	out, err := exec.CommandContext(ctx, "az", "keyvault", "show", "--name", vaultName).CombinedOutput() //nolint:gosec // az CLI with validated vault name
	if err != nil {
		return "failed", strings.TrimSpace(string(out)), err
	}
	return "ok", fmt.Sprintf("Key Vault %s is accessible", vaultName), nil
}

// azLogAnalyticsCheck shells out to az CLI to verify Log Analytics workspace accessibility.
var azLogAnalyticsCheck = func(ctx context.Context, resourceGroup, workspaceName string) (string, string, error) {
	out, err := exec.CommandContext(ctx, "az", "monitor", "log-analytics", "workspace", "show", //nolint:gosec // az CLI with validated parameters
		"--resource-group", resourceGroup,
		"--workspace-name", workspaceName).CombinedOutput()
	if err != nil {
		return "failed", strings.TrimSpace(string(out)), err
	}
	return "ok", fmt.Sprintf("Log Analytics workspace %s is accessible", workspaceName), nil
}
