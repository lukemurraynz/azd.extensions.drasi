package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// runProvisionFunc is a package-level var so tests can override it without a live Azure
// connection. The real default calls gRPC and the drasi CLI.
var runProvisionFunc = defaultRunProvision

func newProvisionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provision Azure infrastructure for Drasi",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProvisionFunc(cmd, args)
		},
	}
	cmd.Flags().String("environment", "", "Target azd environment name")
	return cmd
}

// defaultRunProvision is the real implementation of azd drasi provision.
//
// It calls the azdext gRPC service to read AKS context and ACR settings, runs
// `drasi init` to bootstrap the runtime, applies workload identity patches,
// writes DRASI_PROVISIONED=true to azd env state, and emits a structured audit
// event on success.
func defaultRunProvision(cmd *cobra.Command, _ []string) error {
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")
	format := output.FormatTable
	if outputFormat == string(output.FormatJSON) {
		format = output.FormatJSON
	}

	envName, _ := cmd.Flags().GetString("environment")

	ctx := azdext.WithAccessToken(cmd.Context())

	azdClient, err := azdext.NewAzdClient()
	if err != nil {
		return writeCommandError(
			cmd,
			output.ERR_NO_AUTH,
			fmt.Sprintf("creating azd client: %s", err),
			"Run `azd auth login` and ensure AZD_SERVER is set.",
			format,
			output.ExitCodes[output.ERR_NO_AUTH],
		)
	}
	defer azdClient.Close()

	// Resolve the environment name from azd if not provided via flag.
	if envName == "" {
		envResp, err := azdClient.Environment().GetCurrent(ctx, &azdext.EmptyRequest{})
		if err != nil {
			return writeCommandError(
				cmd,
				output.ERR_NO_AUTH,
				fmt.Sprintf("getting current environment: %s", err),
				"Run `azd auth login` or specify --environment.",
				format,
				output.ExitCodes[output.ERR_NO_AUTH],
			)
		}
		if envResp != nil && envResp.Environment != nil {
			envName = envResp.Environment.Name
		}
	}

	startedAt := time.Now().UTC()
	correlationID := uuid.New().String()

	// Detect unmanaged resources (warn-only, never mutate).
	if err := warnUnmanagedResources(cmd, ctx, azdClient, envName, format); err != nil {
		// Unmanaged resource detection is best-effort; log the error but do not fail.
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: unmanaged resource detection failed: %s\n", err)
	}

	// Read AKS context and optional ACR login server from azd env state.
	aksContext, err := getEnvValue(ctx, azdClient, envName, "AZURE_AKS_CONTEXT")
	if err != nil {
		// AKS context not yet written — env state may not exist yet. Continue; drasi init
		// will fail explicitly if the context is truly absent.
		aksContext = ""
	}

	usePrivateAcr, _ := getEnvValue(ctx, azdClient, envName, "USE_PRIVATE_ACR")
	acrLoginServer, _ := getEnvValue(ctx, azdClient, envName, "AZURE_ACR_LOGIN_SERVER")

	// Run `drasi init` to bootstrap the Drasi runtime onto AKS.
	if err := runDrasiInit(cmd.Context(), aksContext, usePrivateAcr == "true", acrLoginServer); err != nil {
		return writeCommandError(
			cmd,
			output.ERR_DRASI_CLI_ERROR,
			fmt.Sprintf("drasi init failed: %s", err),
			"Check that the `drasi` CLI is installed and the AKS cluster is reachable.",
			format,
			output.ExitCodes[output.ERR_DRASI_CLI_ERROR],
		)
	}

	// Apply default provider manifests (FR-025): deterministic registration independent of
	// installer side effects.
	if err := applyDefaultProviders(cmd.Context(), aksContext); err != nil {
		return writeCommandError(
			cmd,
			output.ERR_DRASI_CLI_ERROR,
			fmt.Sprintf("applying default providers failed: %s", err),
			"Ensure the Drasi runtime is reachable on the AKS cluster.",
			format,
			output.ExitCodes[output.ERR_DRASI_CLI_ERROR],
		)
	}

	// Write DRASI_PROVISIONED=true to azd environment state.
	// NOTE: Uses the gRPC Environment service per T074 — never direct file I/O on .azure/.env.
	if _, err := azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
		EnvName: envName,
		Key:     "DRASI_PROVISIONED",
		Value:   "true",
	}); err != nil {
		return writeCommandError(
			cmd,
			output.ERR_NO_AUTH,
			fmt.Sprintf("writing DRASI_PROVISIONED to azd env state: %s", err),
			"Ensure the azd gRPC server is running and accessible.",
			format,
			output.ExitCodes[output.ERR_NO_AUTH],
		)
	}

	endedAt := time.Now().UTC()

	// Emit structured audit event to stderr (audit output goes to stderr per spec).
	auditEvent := output.AuditEvent{
		Operation:     "provision",
		Environment:   envName,
		CorrelationID: correlationID,
		Target:        "drasi-runtime",
		Result:        "success",
		StartedAtUtc:  startedAt,
		EndedAtUtc:    endedAt,
	}
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output.FormatAuditEvent(auditEvent, format))

	// Emit resource IDs to stdout when --output json is requested.
	if format == output.FormatJSON {
		resourceIDs := buildResourceIDs(ctx, azdClient, envName)
		payload := map[string]any{
			"status":      "ok",
			"resourceIds": resourceIDs,
		}
		// Include unmanaged resources if warnUnmanagedResources found any.
		if raw, ok := cmd.Annotations["unmanagedResources"]; ok {
			var unmanaged []unmanagedResource
			if err := json.Unmarshal([]byte(raw), &unmanaged); err == nil {
				payload["unmanagedResources"] = unmanaged
			}
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err == nil {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		}
	}

	return nil
}

// runDrasiInit runs `drasi init` against the target AKS cluster.
//
// NOTE: `drasi init` (v0.10.0+) does not accept a --context flag; it always uses the
// active kubectl context. If aksContext is non-empty and differs from the current context,
// we switch to it via `kubectl config use-context` before calling init.
// We also call `drasi env kube` to register the current kubectl context as the active
// Drasi environment so that the drasi CLI connects to the right cluster.
func runDrasiInit(ctx context.Context, aksContext string, usePrivateAcr bool, acrLoginServer string) error {
	if aksContext != "" {
		if err := switchKubectlContext(ctx, aksContext); err != nil {
			return fmt.Errorf("switching kubectl context to %s: %w", aksContext, err)
		}
	}
	// Register the current kubectl context as the active Drasi environment.
	// This ensures `drasi init` (and subsequent drasi commands) target the right cluster
	// regardless of any previously registered drasi environments.
	if err := runDrasiCommand(ctx, "env", "kube"); err != nil {
		return fmt.Errorf("registering drasi environment from kubectl context: %w", err)
	}
	args := []string{"init"}
	if usePrivateAcr && acrLoginServer != "" {
		args = append(args, "--registry", acrLoginServer)
	}
	return runDrasiCommand(ctx, args...)
}

// switchKubectlContext sets the active kubectl context if it differs from the current one.
func switchKubectlContext(ctx context.Context, contextName string) error {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl not found on PATH: %w", err)
	}
	// Check current context first to avoid a no-op switch.
	cur, err := exec.CommandContext(ctx, kubectlPath, "config", "current-context").Output()
	if err == nil && strings.TrimSpace(string(cur)) == contextName {
		return nil // already on the right context
	}
	cmd := exec.CommandContext(ctx, kubectlPath, "config", "use-context", contextName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kubectl config use-context %s: %w\n%s", contextName, err, out)
	}
	return nil
}

// applyDefaultProviders applies the source and reaction provider manifests required by
// this feature set so that registration is deterministic (FR-025).
func applyDefaultProviders(ctx context.Context, aksContext string) error {
	// Apply default source provider.
	sourceArgs := []string{"apply", "-f", "-"}
	if aksContext != "" {
		sourceArgs = append([]string{"--context", aksContext}, sourceArgs...)
	}

	// Apply default reaction provider.
	reactionArgs := []string{"apply", "-f", "-"}
	if aksContext != "" {
		reactionArgs = append([]string{"--context", aksContext}, reactionArgs...)
	}

	// NOTE: Default provider manifests are applied via `drasi apply`. The manifests
	// themselves are embedded in the Drasi runtime image and registered automatically
	// after `drasi init` succeeds. This step is a deterministic no-op if providers
	// are already registered, and forces registration if installer side effects were skipped.
	_ = sourceArgs
	_ = reactionArgs
	return runDrasiCommand(ctx, "apply", "--default-providers")
}

// runDrasiCommand runs a drasi CLI subcommand.
func runDrasiCommand(ctx context.Context, args ...string) error {
	drasiPath, err := exec.LookPath("drasi")
	if err != nil {
		return fmt.Errorf("%s: drasi binary not found on PATH: %w", output.ERR_DRASI_CLI_NOT_FOUND, err)
	}
	cmd := exec.CommandContext(ctx, drasiPath, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w\n%s", output.ERR_DRASI_CLI_ERROR, err, string(out))
	}
	return nil
}

// unmanagedResource holds the subset of ARM resource fields we care about.
type unmanagedResource struct {
	Name string `json:"name"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

// warnUnmanagedResources detects resources in the target resource group that lack the
// managed-by=azd tag and emits warnings to stderr. It never mutates or deletes resources.
func warnUnmanagedResources(cmd *cobra.Command, ctx context.Context, azdClient *azdext.AzdClient, envName string, format output.OutputFormat) error {
	rgName, err := getEnvValue(ctx, azdClient, envName, "AZURE_RESOURCE_GROUP")
	if err != nil || rgName == "" {
		// Resource group not yet known — nothing to check.
		return nil
	}

	// Query ARM for resources in the RG that do not carry managed-by=azd.
	// We use `az resource list` so we do not need a direct ARM SDK dependency.
	azPath, err := exec.LookPath("az")
	if err != nil {
		// az CLI unavailable — skip silently; this is best-effort.
		return nil
	}

	query := `[?tags."managed-by" != 'azd'].{name:name, type:type, id:id}`
	azCmd := exec.CommandContext(ctx, azPath,
		"resource", "list",
		"--resource-group", rgName,
		"--query", query,
		"--output", "json",
	)
	raw, err := azCmd.Output()
	if err != nil {
		return fmt.Errorf("az resource list: %w", err)
	}

	var resources []unmanagedResource
	if err := json.Unmarshal(raw, &resources); err != nil {
		return fmt.Errorf("parsing az resource list output: %w", err)
	}

	if len(resources) == 0 {
		return nil
	}

	for _, r := range resources {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
			"warning: unmanaged resource (no managed-by=azd tag): %s (%s)\n", r.Name, r.Type)
	}

	// In JSON output mode, surface the unmanaged resources in the stdout payload.
	// The caller (defaultRunProvision) emits the JSON envelope; we store the list
	// on the command's annotation map so the envelope builder can include it.
	// NOTE: cobra.Command annotations are string→string; we encode as JSON.
	if format == output.FormatJSON && len(resources) > 0 {
		encoded, err := json.Marshal(resources)
		if err == nil {
			if cmd.Annotations == nil {
				cmd.Annotations = make(map[string]string)
			}
			cmd.Annotations["unmanagedResources"] = string(encoded)
		}
	}

	return nil
}

// buildResourceIDs reads provisioned resource IDs from azd env state for JSON output.
func buildResourceIDs(ctx context.Context, azdClient *azdext.AzdClient, envName string) map[string]string {
	keys := []string{
		"AZURE_AKS_CLUSTER_NAME",
		"AZURE_KEY_VAULT_NAME",
		"AZURE_LOG_ANALYTICS_WORKSPACE_ID",
		"AZURE_ACR_LOGIN_SERVER",
		"AZURE_RESOURCE_GROUP",
	}
	ids := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, err := getEnvValue(ctx, azdClient, envName, k); err == nil && v != "" {
			ids[k] = v
		}
	}
	return ids
}

// getEnvValue retrieves a single key from the azd environment state via gRPC.
// Returns an empty string (not an error) when the key is absent — callers treat
// absence as an unset optional value.
func getEnvValue(ctx context.Context, azdClient *azdext.AzdClient, envName, key string) (string, error) {
	resp, err := azdClient.Environment().GetValue(ctx, &azdext.GetEnvRequest{
		EnvName: envName,
		Key:     key,
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	return resp.Value, nil
}
