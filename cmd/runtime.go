package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/azure/azd.extensions.drasi/internal/output"
	"github.com/spf13/cobra"
)

type azdEnvServiceAdapter struct {
	client *azdext.AzdClient
}

func (a *azdEnvServiceAdapter) GetValue(ctx context.Context, envName, key string) (string, error) {
	resp, err := a.client.Environment().GetValue(ctx, &azdext.GetEnvRequest{EnvName: envName, Key: key})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	return resp.Value, nil
}

func (a *azdEnvServiceAdapter) SetValue(ctx context.Context, envName, key, value string) error {
	_, err := a.client.Environment().SetValue(ctx, &azdext.SetEnvRequest{EnvName: envName, Key: key, Value: value})
	return err
}

func resolveEnvironmentName(ctx context.Context, cmd *cobra.Command, azdClient *azdext.AzdClient, explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}

	rootValue, _ := cmd.Root().PersistentFlags().GetString("environment")
	if strings.TrimSpace(rootValue) != "" {
		return rootValue, nil
	}

	resp, err := azdClient.Environment().GetCurrent(ctx, &azdext.EmptyRequest{})
	if err != nil {
		return "", err
	}
	if resp == nil || resp.Environment == nil {
		return "", errors.New("current azd environment is not set")
	}

	return resp.Environment.Name, nil
}

func outputFormatFromCommand(cmd *cobra.Command) output.OutputFormat {
	rootOutput, _ := cmd.Root().PersistentFlags().GetString("output")
	if rootOutput == string(output.FormatJSON) {
		return output.FormatJSON
	}
	return output.FormatTable
}

func errorCodeFromError(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	message := err.Error()
	for code := range output.ExitCodes {
		if strings.Contains(message, code) {
			return code
		}
	}
	return fallback
}

func resolvedKubeContextForCommand(ctx context.Context, cmd *cobra.Command, explicitEnvName string) (string, error) {
	rootEnv, _ := cmd.Root().PersistentFlags().GetString("environment")
	selectedEnv := strings.TrimSpace(explicitEnvName)
	if selectedEnv == "" {
		selectedEnv = strings.TrimSpace(rootEnv)
	}
	if selectedEnv == "" {
		return "", nil
	}

	azdClient, err := azdext.NewAzdClient()
	if err != nil {
		return "", fmt.Errorf("%s: creating azd client: %w", output.ERR_NO_AUTH, err)
	}
	defer azdClient.Close()

	resolvedEnv, err := resolveEnvironmentName(ctx, cmd, azdClient, selectedEnv)
	if err != nil {
		return "", fmt.Errorf("%s: resolving environment: %w", output.ERR_NO_AUTH, err)
	}

	kubeContext, err := getEnvValue(ctx, azdClient, resolvedEnv, "AZURE_AKS_CONTEXT")
	if err != nil {
		return "", fmt.Errorf("%s: resolving AZURE_AKS_CONTEXT: %w", output.ERR_AKS_CONTEXT_NOT_FOUND, err)
	}
	if strings.TrimSpace(kubeContext) == "" {
		return "", fmt.Errorf("%s: AZURE_AKS_CONTEXT is not set for environment %s", output.ERR_AKS_CONTEXT_NOT_FOUND, resolvedEnv)
	}

	return kubeContext, nil
}
