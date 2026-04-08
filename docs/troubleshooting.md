# Troubleshooting

Run `azd drasi diagnose` first. It checks AKS connectivity, Drasi API health, Dapr runtime, Key Vault auth, and Log Analytics in a single pass and surfaces the most common failures with remediation hints. If you use multiple azd environments, invoke with root `--environment <name>` so the command resolves `AZURE_AKS_CONTEXT` for the intended cluster.

## Error code reference

Every failure the extension emits includes a structured error code. The table below lists each code, when it occurs, the exit code, and what to do.

| Code | Exit | When it occurs | Remediation |
| ---- | ---- | -------------- | ----------- |
| `ERR_NO_AUTH` | 2 | No valid Azure credential found when the command tried to call a gRPC service or Azure API | Run `azd auth login` and try again. In CI, ensure `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, and `AZURE_SUBSCRIPTION_ID` are set and the federated credential is configured. |
| `ERR_DRASI_CLI_NOT_FOUND` | 2 | The `drasi` binary is not on `PATH` | Install the Drasi CLI (>= 0.10.0) from <https://drasi.io/docs/getting-started> and ensure it is on `PATH`. |
| `ERR_DRASI_CLI_VERSION` | 2 | The installed `drasi` binary is older than the minimum required version (0.10.0) | Upgrade the Drasi CLI. |
| `ERR_DRASI_CLI_ERROR` | 1 | The `drasi` subprocess exited with a non-zero code | Read the error message for the underlying cause. Run the same `drasi` command manually to see the full output. Check that your kubeconfig context points to the correct AKS cluster. |
| `ERR_COMPONENT_TIMEOUT` | 1 | A component did not reach `Online` state within 5 minutes | Run `azd drasi diagnose` to check Dapr and the Drasi API. Check the component pod logs: `kubectl logs -n drasi-system -l drasi.io/component-id=<id>`. |
| `ERR_TOTAL_TIMEOUT` | 1 | The entire deploy did not complete within 15 minutes | Reduce the number of components deployed at once, check cluster resource limits, or increase timeouts by splitting the deploy into separate runs. |
| `ERR_VALIDATION_FAILED` | 1 | One or more entities failed offline validation | Read the validation output for the specific fields and fix them. Run `azd drasi validate` after each fix to confirm. |
| `ERR_MISSING_REFERENCE` | 1 | A query references a source ID that does not exist in the project | Add the missing source entity or fix the `sources[].id` reference in the query file. |
| `ERR_CIRCULAR_DEPENDENCY` | 1 | Two or more entities form a circular dependency | Review the `sources` and `reactions` fields on your queries for cycles. |
| `ERR_MISSING_QUERY_LANGUAGE` | 1 | A `ContinuousQuery` entity is missing the `queryLanguage` field | Add `queryLanguage: Cypher` (or the appropriate language) to the query entity. |
| `ERR_KV_AUTH_FAILED` | 2 | The caller's Azure CLI identity does not have permission to read secrets from Key Vault | Assign the `Key Vault Secrets User` role to the caller's identity on the Key Vault. Run `azd drasi provision` again to re-apply the role assignment, or assign it manually: `az role assignment create --assignee "$(az account show --query user.name -o tsv)" --role "Key Vault Secrets User" --scope <key-vault-resource-id>`. |
| `ERR_AKS_CONTEXT_NOT_FOUND` | 2 | The kubeconfig context for the provisioned AKS cluster could not be found | Run `az aks get-credentials --resource-group <rg> --name <aks-name>` to refresh the kubeconfig, then retry. |
| `ERR_FORCE_REQUIRED` | 2 | A destructive operation was attempted without `--force` (for example `teardown` or `upgrade`) | Re-run the command with `--force` to confirm the destructive action. |
| `ERR_NO_MANIFEST` | 2 | `drasi/drasi.yaml` was not found in the current directory | Run the command from the project root that contains the `drasi/` directory, or run `azd drasi init` to scaffold one. |
| `ERR_DEPLOY_IN_PROGRESS` | 2 | A deploy is already running for this environment | Wait for the previous deploy to complete, or remove the in-progress lock from azd environment state manually. |
| `ERR_DAPR_NOT_READY` | 2 | The Dapr runtime is not running in the AKS cluster | Run `azd drasi diagnose` and check the Dapr section. Reinstall or restart Dapr with `dapr init -k`. |

## Common failure scenarios

### `azd drasi provision` fails at the AKS step

The most common causes are insufficient quota in the target region and a missing `Owner` or `Contributor` role on the subscription.

Check the deployment error in the Azure portal under the resource group's deployment history, or run:

```bash
az deployment sub show --name <deployment-name> --query properties.error
```

### Components stuck in `Pending` after deploy

If `azd drasi status` shows components in `Pending` for more than a few minutes:

1. Run `azd drasi diagnose` to confirm Dapr and the Drasi API are healthy. If targeting a non-default azd environment, run `azd drasi --environment <name> diagnose`.
2. Check pod logs in `drasi-system`: `kubectl get pods -n drasi-system` and `kubectl logs <pod-name> -n drasi-system`.
3. Confirm that all `secretRef` values in your entity files resolve to real Key Vault secrets.

### Key Vault secret reference fails silently

If a component deploys successfully but behaves as if a secret is missing, check that the secret name and vault name in the `secretRef` match exactly what is in Key Vault (including case). Key Vault secret names are case-insensitive in storage but the extension passes the name verbatim to the Key Vault API.

### Key Vault access denied during deploy

If `azd drasi deploy` fails with a Key Vault access error, the **caller's Azure CLI identity** may lack permissions. The extension uses `az keyvault secret show` under the currently logged-in Azure CLI session to fetch secrets defined in `secretMappings`. Ensure the caller has the `Key Vault Secrets User` role:

```bash
az role assignment create \
  --assignee "$(az account show --query user.name -o tsv)" \
  --role "Key Vault Secrets User" \
  --scope /subscriptions/<sub-id>/resourceGroups/<rg>/providers/Microsoft.KeyVault/vaults/<vault-name>
```

### `ERR_NO_AUTH` in CI

In GitHub Actions, the extension uses OIDC to authenticate. Ensure:

- `permissions: id-token: write` is set on the job.
- `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, and `AZURE_SUBSCRIPTION_ID` secrets are set.
- The federated credential on the app registration matches the repository and branch/environment.

### `azd drasi teardown --force --infrastructure` hangs

The teardown calls `az group delete` for the resource group. If the group contains resources with delete locks, Azure will block deletion until locks are removed. Check the resource group in the Azure portal for delete locks and remove them before retrying.

### `ERR_AKS_CONTEXT_NOT_FOUND` on status/logs/diagnose

This indicates the command could not resolve `AZURE_AKS_CONTEXT` for the selected environment.

1. Ensure the environment exists: `azd env list`.
2. Inspect values: `azd env get-values --environment <name>`.
3. Ensure `AZURE_AKS_CONTEXT` is present (typically written by `azd drasi provision`).
4. If missing, re-run `azd drasi provision` for that environment.

### New data source connections blocked by network policy

If a source component deploys successfully but cannot connect to its data store, the default `drasi-allow-datastores` network policy may not include the required endpoint.

The extension applies baseline Kubernetes NetworkPolicies during provisioning. The `drasi-allow-datastores` policy permits egress to common data store ports (5432 for PostgreSQL, 27017 for MongoDB, 443 for Cosmos DB). If your data source uses a non-standard port or a private endpoint with a different IP range, you need to update the policy.

To check current network policies:

```bash
kubectl get networkpolicies -n drasi-system
kubectl describe networkpolicy drasi-allow-datastores -n drasi-system
```

To add a new egress rule, edit the policy directly or update the embedded `network_policies.yaml` in the extension source and re-provision. For quick unblocking during development, you can temporarily allow all egress from the namespace:

```bash
kubectl delete networkpolicy drasi-default-deny -n drasi-system
```

Restore the default-deny policy before moving to production by re-running `azd drasi provision`.

### Deploy uses delete-then-apply for component updates

When you run `azd drasi deploy` on a component that already exists, the extension deletes it and re-applies it from the entity file. This is a limitation of the current Drasi CLI, which does not support in-place updates. The brief downtime is usually acceptable because Drasi queries resume automatically after the component restarts. If your workflow requires zero-downtime updates, deploy to a second environment first and switch traffic after validation.

## Diagnostic flow

If the error code table does not resolve the issue, work through this sequence:

1. Run `azd drasi validate` to rule out configuration errors.
2. Run `azd drasi diagnose` to check cluster health and connectivity.
3. Run `kubectl get pods -n drasi-system` and check that all Drasi pods are `Running`.
4. Check Drasi pod logs for the specific component: `kubectl logs -n drasi-system -l drasi.io/component-id=<id>`.
5. Check the azd environment state for stale values: `azd env get-values`.
6. If infrastructure is suspected, re-run `azd drasi provision` (it is idempotent).
7. If the issue persists, open an issue with the output of `azd drasi diagnose --output json`.
