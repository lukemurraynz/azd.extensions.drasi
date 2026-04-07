package keyvault

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
)

// TranslatedValue is the result of resolving a config.Value.
type TranslatedValue struct {
	// StringValue is the resolved plain string (populated for plain strings and resolved secrets).
	StringValue string
	// IsSecretRef is true when the original value was a SecretRef.
	IsSecretRef bool
}

// SecretClient resolves a secret reference to its value.
type SecretClient interface {
	GetSecret(ctx context.Context, vaultName, secretName string) (string, error)
}

type envSecretClient struct{}

func (e *envSecretClient) GetSecret(_ context.Context, vaultName, secretName string) (string, error) {
	envKey := fmt.Sprintf("DRASI_SECRET_%s_%s", vaultName, secretName)
	value, ok := os.LookupEnv(envKey)
	if !ok {
		return "", fmt.Errorf("%s: secret %s/%s not found in env key %s", output.ERR_KV_AUTH_FAILED, vaultName, secretName, envKey)
	}
	return value, nil
}

// Translator resolves config.Value entries, replacing Key Vault SecretRef
// entries with concrete secret values.
type Translator struct {
	secretClient SecretClient
}

// NewTranslator creates a Translator.
func NewTranslator() *Translator {
	return &Translator{secretClient: &envSecretClient{}}
}

// NewTranslatorWithSecretClient creates a Translator with a custom secret client.
func NewTranslatorWithSecretClient(secretClient SecretClient) *Translator {
	if secretClient == nil {
		secretClient = &envSecretClient{}
	}
	return &Translator{secretClient: secretClient}
}

// Translate resolves a single config.Value.
// - Plain string → TranslatedValue{StringValue: value.StringValue}
// - EnvRef → value from process environment
// - SecretRef → value from secret client
func (t *Translator) Translate(ctx context.Context, v config.Value) (*TranslatedValue, error) {
	if v.SecretRef == nil && v.EnvRef == nil {
		return &TranslatedValue{StringValue: v.StringValue, IsSecretRef: false}, nil
	}

	if v.EnvRef != nil {
		slog.InfoContext(ctx, "resolving environment variable reference",
			slog.String("env_var", v.EnvRef.Name),
		)
		value, ok := os.LookupEnv(v.EnvRef.Name)
		if !ok {
			return nil, fmt.Errorf("%s: environment variable %s is not set", output.ERR_VALIDATION_FAILED, v.EnvRef.Name)
		}
		return &TranslatedValue{StringValue: value, IsSecretRef: false}, nil
	}

	// SECURITY: Log vault and secret name for audit trail, but never the secret value.
	slog.InfoContext(ctx, "resolving Key Vault secret reference",
		slog.String("vault_name", v.SecretRef.VaultName),
		slog.String("secret_name", v.SecretRef.SecretName),
	)
	secretValue, err := t.secretClient.GetSecret(ctx, v.SecretRef.VaultName, v.SecretRef.SecretName)
	if err != nil {
		return nil, err
	}
	return &TranslatedValue{StringValue: secretValue, IsSecretRef: true}, nil
}
