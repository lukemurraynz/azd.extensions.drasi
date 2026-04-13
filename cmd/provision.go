package cmd

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"strings"
	"time"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/google/uuid"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

//go:embed network_policies.yaml
var drasiNetworkPoliciesYAML string

// runProvisionFunc can be overridden in tests to avoid a live Azure connection.
var runProvisionFunc = defaultRunProvision

func newProvisionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provision Azure infrastructure for Drasi",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProvisionFunc(cmd, args)
		},
	}
	return cmd
}

func defaultRunProvision(cmd *cobra.Command, _ []string) error {
	outputFormat, _ := cmd.Root().PersistentFlags().GetString("output")
	format := output.FormatTable
	if outputFormat == string(output.FormatJSON) {
		format = output.FormatJSON
	}

	envFlag, _ := cmd.Root().PersistentFlags().GetString("environment")

	progress, err := NewProgressHelper(cmd)
	if err != nil {
		// Spinner failure is non-fatal; degrade gracefully.
		progress = &ProgressHelper{noop: true}
	}
	_ = progress.Start()
	defer func() { _ = progress.Stop() }()

	progress.Message("Resolving environment...")

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

	envName, err := resolveEnvironmentName(ctx, cmd, azdClient, envFlag)
	if err != nil {
		return writeCommandError(cmd, output.ERR_NO_AUTH,
			fmt.Sprintf("resolving environment: %s", err),
			"Run `azd auth login` or specify --environment.",
			format, output.ExitCodes[output.ERR_NO_AUTH])
	}

	startedAt := time.Now().UTC()
	correlationID := uuid.New().String()
	var aksClusterName string
	var rgName string

	// Check for AZURE_AKS_CLUSTER_NAME and AZURE_RESOURCE_GROUP as signals
	// that Bicep infrastructure has been deployed.
	// NOTE: azd writes Bicep output names verbatim to env state.
	aksClusterName, _ = getEnvValue(ctx, azdClient, envName, "AZURE_AKS_CLUSTER_NAME")
	rgName, _ = getEnvValue(ctx, azdClient, envName, "AZURE_RESOURCE_GROUP")
	if aksClusterName == "" || rgName == "" {
		progress.Message("Provisioning Azure infrastructure...")

		// Ensure the azd environment exists before setting config. When the user
		// passes --environment <name> for a fresh project, the environment may not
		// exist yet. Running `env new` creates it; if it already exists, azd returns
		// a non-zero exit but we ignore that since the environment is already present.
		if err := ensureAzdEnvironment(cmd.Context(), envName); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not ensure environment exists: %s\n", err)
		}

		// Pre-set AZURE_SUBSCRIPTION_ID and AZURE_LOCATION before calling `azd provision`.
		// The provisioning manager prompts for subscription, location, and resource group
		// interactively, but in the gRPC extension context those prompts fail with
		// "interrupt". We use the PromptService to collect these values first, then
		// persist them so `azd provision` can proceed non-interactively.
		if err := ensureSubscriptionAndLocation(ctx, azdClient, envName, progress); err != nil {
			return writeCommandError(
				cmd,
				output.ERR_INFRA_PROVISION_FAILED,
				fmt.Sprintf("collecting subscription and location: %s", err),
				"Run `azd auth login` and ensure you have an active Azure subscription.",
				format,
				output.ExitCodes[output.ERR_INFRA_PROVISION_FAILED],
			)
		}

		progress.Message("Provisioning Azure infrastructure...")

		// Ensure infra.parameters.environmentName is configured. The Bicep template
		// declares environmentName as a required parameter. In non-interactive mode
		// (gRPC/extension context), azd cannot prompt for missing parameters, so we
		// pre-set it from the resolved environment name.
		envNameJSON, _ := json.Marshal(envName)
		if _, err := azdClient.Environment().SetConfig(ctx, &azdext.SetConfigRequest{
			Path:    "infra.parameters.environmentName",
			Value:   envNameJSON,
			EnvName: envName,
		}); err != nil {
			// Non-fatal: the parameter may already be configured or the user may have
			// set it manually. Log the warning but continue with the provisioning attempt.
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not pre-set infra.parameters.environmentName: %s\n", err)
		}

		// Resolve the caller's Entra ID object ID and pre-set it as
		// infra.parameters.principalId. The Bicep template uses this to assign
		// the AKS RBAC Cluster Admin role so the caller can run kubectl/drasi
		// commands after provisioning. Supports both user accounts and service
		// principals.
		callerOID, err := resolveCallerOID(cmd.Context())
		if err != nil {
			return writeCommandError(
				cmd,
				output.ERR_NO_AUTH,
				fmt.Sprintf("resolving caller identity OID: %s", err),
				"Run `azd auth login` as a user or service principal with directory read permissions.",
				format,
				output.ExitCodes[output.ERR_NO_AUTH],
			)
		}
		oidJSON, _ := json.Marshal(callerOID)
		if _, err := azdClient.Environment().SetConfig(ctx, &azdext.SetConfigRequest{
			Path:    "infra.parameters.principalId",
			Value:   oidJSON,
			EnvName: envName,
		}); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not pre-set infra.parameters.principalId: %s\n", err)
		}

		// Generate a secure password for PostgreSQL and pre-set it as
		// infra.parameters.postgresAdminPassword. The Bicep template stores
		// it in Key Vault so the deploy step can sync it to a Kubernetes Secret.
		// Only generate if not already set (idempotent across retries).
		existingPW, _ := getEnvValue(ctx, azdClient, envName, "POSTGRES_ADMIN_PASSWORD_SET")
		if existingPW != "true" {
			pgPassword, err := generatePassword(20)
			if err != nil {
				return writeCommandError(
					cmd,
					output.ERR_INFRA_PROVISION_FAILED,
					fmt.Sprintf("generating PostgreSQL admin password: %s", err),
					"Ensure the system has a working crypto/rand source.",
					format,
					output.ExitCodes[output.ERR_INFRA_PROVISION_FAILED],
				)
			}
			pwJSON, _ := json.Marshal(pgPassword)
			if _, err := azdClient.Environment().SetConfig(ctx, &azdext.SetConfigRequest{
				Path:    "infra.parameters.postgresAdminPassword",
				Value:   pwJSON,
				EnvName: envName,
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not pre-set infra.parameters.postgresAdminPassword: %s\n", err)
			}
			// Mark that we've generated the password so subsequent provision retries
			// reuse the same value (azd config is persisted across runs).
			if _, err := azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
				EnvName: envName,
				Key:     "POSTGRES_ADMIN_PASSWORD_SET",
				Value:   "true",
			}); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not persist POSTGRES_ADMIN_PASSWORD_SET: %s\n", err)
			}
		}

		provisionArgs := []string{"provision"}
		if envName != "" {
			provisionArgs = append(provisionArgs, "--environment", envName)
		}
		if _, err := azdClient.Workflow().Run(ctx, &azdext.RunWorkflowRequest{
			Workflow: &azdext.Workflow{
				Name: "provision",
				Steps: []*azdext.WorkflowStep{
					{Command: &azdext.WorkflowCommand{Args: provisionArgs}},
				},
			},
		}); err != nil {
			return writeCommandError(
				cmd,
				output.ERR_INFRA_PROVISION_FAILED,
				fmt.Sprintf("infrastructure provisioning failed: %s", err),
				"Run `azd auth login`, ensure you have an active Azure subscription, and check infra/ for Bicep errors.",
				format,
				output.ExitCodes[output.ERR_INFRA_PROVISION_FAILED],
			)
		}

		// Re-read cluster name and resource group after provisioning.
		aksClusterName, _ = getEnvValue(ctx, azdClient, envName, "AZURE_AKS_CLUSTER_NAME")
		rgName, _ = getEnvValue(ctx, azdClient, envName, "AZURE_RESOURCE_GROUP")
		if aksClusterName == "" || rgName == "" {
			return writeCommandError(
				cmd,
				output.ERR_INFRA_PROVISION_FAILED,
				"infrastructure provisioning completed but AZURE_AKS_CLUSTER_NAME or AZURE_RESOURCE_GROUP not found in environment state",
				"Check Bicep outputs in infra/main.bicep and ensure they include AZURE_AKS_CLUSTER_NAME.",
				format,
				output.ExitCodes[output.ERR_INFRA_PROVISION_FAILED],
			)
		}
	}

	// Detect unmanaged resources (warn-only, never mutate).
	if err := warnUnmanagedResources(cmd, ctx, azdClient, envName, format); err != nil {
		// Unmanaged resource detection is best-effort; log the error but do not fail.
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: unmanaged resource detection failed: %s\n", err)
	}

	// Set up kubectl context if not already configured.
	aksContext, _ := getEnvValue(ctx, azdClient, envName, "AZURE_AKS_CONTEXT")
	if aksContext == "" {
		progress.Message("Acquiring AKS credentials...")
		acquiredContext, err := acquireAKSCredentials(cmd.Context(), rgName, aksClusterName)
		if err != nil {
			return writeCommandError(
				cmd,
				output.ERR_AKS_CONTEXT_NOT_FOUND,
				fmt.Sprintf("acquiring AKS credentials: %s", err),
				"Ensure `az` CLI is installed and you have access to the AKS cluster.",
				format,
				output.ExitCodes[output.ERR_AKS_CONTEXT_NOT_FOUND],
			)
		}
		aksContext = acquiredContext

		// Persist the kubectl context name so subsequent commands can use it.
		if _, err := azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
			EnvName: envName,
			Key:     "AZURE_AKS_CONTEXT",
			Value:   aksContext,
		}); err != nil {
			return writeCommandError(
				cmd,
				output.ERR_NO_AUTH,
				fmt.Sprintf("writing AZURE_AKS_CONTEXT to azd env state: %s", err),
				"Ensure the azd gRPC server is running.",
				format,
				output.ExitCodes[output.ERR_NO_AUTH],
			)
		}
	}

	usePrivateAcr, _ := getEnvValue(ctx, azdClient, envName, "USE_PRIVATE_ACR")
	acrLoginServer, _ := getEnvValue(ctx, azdClient, envName, "AZURE_ACR_LOGIN_SERVER")

	// Run `drasi init` to bootstrap the Drasi runtime onto AKS.
	progress.Message("Installing Drasi runtime...")
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

	// Apply baseline Cilium NetworkPolicies to drasi-system namespace.
	progress.Message("Applying network policies...")
	if err := applyDrasiNetworkPolicies(cmd.Context(), aksContext); err != nil {
		return writeCommandError(
			cmd,
			output.ERR_DRASI_CLI_ERROR,
			fmt.Sprintf("applying Drasi NetworkPolicies failed: %s", err),
			"Ensure kubectl is installed and the AKS cluster is reachable.",
			format,
			output.ExitCodes[output.ERR_DRASI_CLI_ERROR],
		)
	}

	// Write DRASI_PROVISIONED=true to azd environment state.
	// NOTE: Uses the gRPC Environment service per T074 — never direct file I/O on .azure/.env.
	progress.Message("Finalizing environment state...")
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

	_ = progress.Stop()

	// Emit audit event to stderr.
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

func runDrasiInit(ctx context.Context, aksContext string, usePrivateAcr bool, acrLoginServer string) error {
	if aksContext != "" {
		if err := switchKubectlContext(ctx, aksContext); err != nil {
			return fmt.Errorf("switching kubectl context to %s: %w", aksContext, err)
		}
	}
	// Register the current kubectl context as the active Drasi environment
	// so `drasi init` targets the right cluster.
	if err := runDrasiCommand(ctx, "env", "kube"); err != nil {
		return fmt.Errorf("registering drasi environment from kubectl context: %w", err)
	}
	args := []string{"init"}
	if usePrivateAcr && acrLoginServer != "" {
		args = append(args, "--registry", acrLoginServer)
	}
	return runDrasiCommand(ctx, args...)
}

func switchKubectlContext(ctx context.Context, contextName string) error {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl not found on PATH: %w", err)
	}
	// Check current context first to avoid a no-op switch.
	cur, err := exec.CommandContext(ctx, kubectlPath, "config", "current-context").Output() //nolint:gosec // kubectl path resolved via LookPath
	if err == nil && strings.TrimSpace(string(cur)) == contextName {
		return nil
	}
	cmd := exec.CommandContext(ctx, kubectlPath, "config", "use-context", contextName) //nolint:gosec // kubectl path resolved via LookPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kubectl config use-context %s: %w\n%s", contextName, err, out)
	}
	return nil
}

// resolveCallerOID returns the Entra ID object ID of the currently signed-in
// identity. It supports both user accounts and service principals by detecting
// the account type via `az account show --query user -o json`.
func resolveCallerOID(ctx context.Context) (string, error) {
	azPath, err := exec.LookPath("az")
	if err != nil {
		return "", fmt.Errorf("az CLI not found on PATH: %w", err)
	}

	//nolint:gosec // az CLI path resolved via LookPath
	acctOut, err := exec.CommandContext(ctx, azPath, "account", "show", "--query", "user", "-o", "json").Output()
	if err != nil {
		return "", fmt.Errorf("az account show: %w", err)
	}

	var acctUser struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(acctOut, &acctUser); err != nil {
		return "", fmt.Errorf("parsing az account show output: %w", err)
	}

	switch acctUser.Type {
	case "user":
		//nolint:gosec // az CLI path resolved via LookPath
		out, err := exec.CommandContext(ctx, azPath, "ad", "signed-in-user", "show", "--query", "id", "-o", "tsv").Output()
		if err != nil {
			return "", fmt.Errorf("az ad signed-in-user show: %w", err)
		}
		oid := strings.TrimSpace(string(out))
		if oid == "" {
			return "", fmt.Errorf("az ad signed-in-user show returned empty OID")
		}
		return oid, nil

	case "servicePrincipal":
		appID := acctUser.Name
		if appID == "" {
			return "", fmt.Errorf("az account show returned empty service principal name")
		}
		//nolint:gosec // az CLI path resolved via LookPath
		out, err := exec.CommandContext(ctx, azPath, "ad", "sp", "show", "--id", appID, "--query", "id", "-o", "tsv").Output()
		if err != nil {
			return "", fmt.Errorf("az ad sp show --id %s: %w", appID, err)
		}
		oid := strings.TrimSpace(string(out))
		if oid == "" {
			return "", fmt.Errorf("az ad sp show returned empty OID for appId %s", appID)
		}
		return oid, nil

	default:
		return "", fmt.Errorf("unsupported az account type %q", acctUser.Type)
	}
}

// acquireAKSCredentials runs `az aks get-credentials` and returns the kubectl context name.
func acquireAKSCredentials(ctx context.Context, resourceGroup, clusterName string) (string, error) {
	azPath, err := exec.LookPath("az")
	if err != nil {
		return "", fmt.Errorf("az CLI not found on PATH: %w", err)
	}

	//nolint:gosec // az CLI path resolved via LookPath
	azCmd := exec.CommandContext(ctx, azPath,
		"aks", "get-credentials",
		"--resource-group", resourceGroup,
		"--name", clusterName,
		"--overwrite-existing",
	)
	if out, err := azCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("az aks get-credentials: %w\n%s", err, string(out))
	}

	// Read back the actual kubectl context name rather than assuming it matches the cluster name.
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl not found on PATH: %w", err)
	}
	//nolint:gosec // kubectl path resolved via LookPath
	out, err := exec.CommandContext(ctx, kubectlPath, "config", "current-context").Output()
	if err != nil {
		return "", fmt.Errorf("reading current kubectl context: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func runDrasiCommand(ctx context.Context, args ...string) error {
	drasiPath, err := exec.LookPath("drasi")
	if err != nil {
		return fmt.Errorf("%s: drasi binary not found on PATH: %w", output.ERR_DRASI_CLI_NOT_FOUND, err)
	}
	cmd := exec.CommandContext(ctx, drasiPath, args...) //nolint:gosec // drasi path resolved via LookPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w\n%s", output.ERR_DRASI_CLI_ERROR, err, string(out))
	}
	return nil
}

// ensureSubscriptionAndLocation checks whether AZURE_SUBSCRIPTION_ID and
// AZURE_LOCATION are already set in the azd environment. If either is missing,
// the user is prompted via the azd PromptService (which routes through the host's
// interactive terminal). The selected values are persisted so that `azd provision`
// can proceed without further interactive prompts.
func ensureSubscriptionAndLocation(
	ctx context.Context,
	azdClient *azdext.AzdClient,
	envName string,
	progress *ProgressHelper,
) error {
	subID, _ := getEnvValue(ctx, azdClient, envName, "AZURE_SUBSCRIPTION_ID")
	location, _ := getEnvValue(ctx, azdClient, envName, "AZURE_LOCATION")

	if subID == "" {
		progress.Message("Selecting Azure subscription...")

		resp, err := azdClient.Prompt().PromptSubscription(ctx, &azdext.PromptSubscriptionRequest{
			Message: "Select an Azure Subscription to use",
		})
		if err != nil {
			return fmt.Errorf("prompting for subscription: %w", err)
		}
		subID = resp.Subscription.Id

		if _, err := azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
			EnvName: envName,
			Key:     "AZURE_SUBSCRIPTION_ID",
			Value:   subID,
		}); err != nil {
			return fmt.Errorf("persisting AZURE_SUBSCRIPTION_ID: %w", err)
		}
	}

	if location == "" {
		progress.Message("Selecting Azure location...")

		azureContext := &azdext.AzureContext{
			Scope: &azdext.AzureScope{
				SubscriptionId: subID,
			},
		}
		resp, err := azdClient.Prompt().PromptLocation(ctx, &azdext.PromptLocationRequest{
			AzureContext: azureContext,
		})
		if err != nil {
			return fmt.Errorf("prompting for location: %w", err)
		}
		location = resp.Location.Name

		if _, err := azdClient.Environment().SetValue(ctx, &azdext.SetEnvRequest{
			EnvName: envName,
			Key:     "AZURE_LOCATION",
			Value:   location,
		}); err != nil {
			return fmt.Errorf("persisting AZURE_LOCATION: %w", err)
		}
	}

	return nil
}

// ensureAzdEnvironment creates the named azd environment if it does not already exist.
// It runs `azd env new <name> --no-prompt` which is idempotent — if the environment
// already exists azd returns an error, which we ignore.
func ensureAzdEnvironment(ctx context.Context, envName string) error {
	azdPath, err := exec.LookPath("azd")
	if err != nil {
		return fmt.Errorf("azd not found on PATH: %w", err)
	}
	//nolint:gosec // azd path resolved via LookPath
	cmd := exec.CommandContext(ctx, azdPath, "env", "new", envName, "--no-prompt")
	_ = cmd.Run() // Ignore error: env may already exist.
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
	azCmd := exec.CommandContext(ctx, azPath, //nolint:gosec // az CLI path resolved via LookPath
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

// generatePassword creates a cryptographically random password suitable for
// Azure Database for PostgreSQL. The password contains a mix of uppercase,
// lowercase, digits, and special characters to satisfy Azure complexity requirements.
func generatePassword(length int) (string, error) {
	const (
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lower   = "abcdefghijklmnopqrstuvwxyz"
		digits  = "0123456789"
		special = "!@#$%^&*"
	)
	charset := upper + lower + digits + special

	if length < 8 {
		length = 16
	}

	for attempts := 0; attempts < 10; attempts++ {
		buf := make([]byte, length)
		for i := range buf {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", fmt.Errorf("generating random byte: %w", err)
			}
			buf[i] = charset[idx.Int64()]
		}
		pw := string(buf)

		// Verify complexity: at least one char from each required set.
		hasUpper := strings.ContainsAny(pw, upper)
		hasLower := strings.ContainsAny(pw, lower)
		hasDigit := strings.ContainsAny(pw, digits)
		hasSpecial := strings.ContainsAny(pw, special)
		if hasUpper && hasLower && hasDigit && hasSpecial {
			return pw, nil
		}
	}

	return "", fmt.Errorf("failed to generate compliant password after 10 attempts")
}

// applyDrasiNetworkPolicies applies baseline Kubernetes NetworkPolicy manifests
// to the drasi-system namespace (default-deny, intra-namespace allow, DNS, Azure API egress).
func applyDrasiNetworkPolicies(ctx context.Context, aksContext string) error {
	if aksContext != "" {
		if err := switchKubectlContext(ctx, aksContext); err != nil {
			return fmt.Errorf("switching kubectl context: %w", err)
		}
	}
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return fmt.Errorf("kubectl not found on PATH: %w", err)
	}
	cmd := exec.CommandContext(ctx, kubectlPath, "apply", "-f", "-") //nolint:gosec // kubectl path resolved via LookPath
	cmd.Stdin = strings.NewReader(drasiNetworkPoliciesYAML)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("kubectl apply NetworkPolicies: %w\n%s", err, out)
	}
	return nil
}
