package deployment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
)

type cmdRunner interface {
	RunCmd(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error)
}

type execCmdRunner struct{}

func (r *execCmdRunner) RunCmd(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
	var cmd *exec.Cmd
	switch name {
	case "az":
		// #nosec G204 -- executable is hardcoded to az; args are constructed internally for Azure CLI invocations only.
		cmd = exec.CommandContext(ctx, "az", args...)
	case "kubectl":
		// #nosec G204 -- executable is hardcoded to kubectl; args are constructed internally for Kubernetes secret apply only.
		cmd = exec.CommandContext(ctx, "kubectl", args...)
	default:
		return nil, fmt.Errorf("unsupported command %q", name)
	}
	cmd.Stdin = stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return out, nil
}

type secretGroupKey struct {
	Namespace string
	Name      string
}

type secretGroupEntry struct {
	VaultName  string
	SecretName string
	K8sKey     string
}

func syncSecrets(ctx context.Context, mappings []config.SecretMapping, envVars map[string]string, kubeContext string, runner cmdRunner) error {
	if len(mappings) == 0 {
		return nil
	}
	if runner == nil {
		runner = &execCmdRunner{}
	}

	groups := make(map[secretGroupKey][]secretGroupEntry)
	for _, mapping := range mappings {
		namespace := mapping.Namespace
		if namespace == "" {
			namespace = "drasi-system"
		}

		key := secretGroupKey{
			Namespace: namespace,
			Name:      mapping.K8sSecret,
		}

		groups[key] = append(groups[key], secretGroupEntry{
			VaultName:  expandEnvVarString(mapping.VaultName, envVars),
			SecretName: expandEnvVarString(mapping.SecretName, envVars),
			K8sKey:     mapping.K8sKey,
		})
	}

	groupKeys := make([]secretGroupKey, 0, len(groups))
	for key := range groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Slice(groupKeys, func(i, j int) bool {
		if groupKeys[i].Namespace != groupKeys[j].Namespace {
			return groupKeys[i].Namespace < groupKeys[j].Namespace
		}
		return groupKeys[i].Name < groupKeys[j].Name
	})

	for _, groupKey := range groupKeys {
		entries := groups[groupKey]
		stringData := make(map[string]string, len(entries))

		for _, entry := range entries {
			slog.InfoContext(ctx, "fetching secret from key vault",
				slog.String("vaultName", entry.VaultName),
				slog.String("secretName", entry.SecretName),
				slog.String("k8sSecret", groupKey.Name),
				slog.String("namespace", groupKey.Namespace),
				slog.String("k8sKey", entry.K8sKey))

			out, err := runner.RunCmd(ctx, nil, "az",
				"keyvault", "secret", "show",
				"--vault-name", entry.VaultName,
				"--name", entry.SecretName,
				"--query", "value",
				"-o", "tsv",
			)
			if err != nil {
				return err
			}

			stringData[entry.K8sKey] = strings.TrimSpace(string(out))
		}

		manifest, err := buildSecretManifest(groupKey.Namespace, groupKey.Name, stringData)
		if err != nil {
			return err
		}

		slog.InfoContext(ctx, "applying kubernetes secret",
			slog.String("k8sSecret", groupKey.Name),
			slog.String("namespace", groupKey.Namespace))

		kubectlArgs := []string{"apply", "-f", "-"}
		if kubeContext != "" {
			kubectlArgs = append([]string{"--context", kubeContext}, kubectlArgs...)
		}

		if _, err := runner.RunCmd(ctx, bytes.NewReader(manifest), "kubectl", kubectlArgs...); err != nil {
			return err
		}
	}

	return nil
}

func buildSecretManifest(namespace, k8sSecretName string, entries map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "apiVersion: v1\n")
	fmt.Fprintf(&buf, "kind: Secret\n")
	fmt.Fprintf(&buf, "metadata:\n")
	fmt.Fprintf(&buf, "  name: %s\n", k8sSecretName)
	fmt.Fprintf(&buf, "  namespace: %s\n", namespace)
	fmt.Fprintf(&buf, "type: Opaque\n")
	fmt.Fprintf(&buf, "stringData:\n")

	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(&buf, "  %s: %s\n", key, yamlQuote(entries[key]))
	}

	return buf.Bytes(), nil
}

func yamlQuote(s string) string {
	return strconv.Quote(s)
}

func expandEnvVarString(s string, envVars map[string]string) string {
	return string(expandEnvVars([]byte(s), envVars))
}
