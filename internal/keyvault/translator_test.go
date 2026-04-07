package keyvault

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/lukemurraynz/azd.extensions.drasi/internal/config"
	"github.com/lukemurraynz/azd.extensions.drasi/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSecretClient struct {
	value string
	err   error
}

func (f *fakeSecretClient) GetSecret(_ context.Context, _, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

func TestTranslator_PlainString_PassesThrough(t *testing.T) {
	t.Parallel()

	translator := NewTranslator()
	got, err := translator.Translate(context.Background(), config.Value{StringValue: "hello"})
	require.NoError(t, err)
	assert.Equal(t, &TranslatedValue{StringValue: "hello", IsSecretRef: false}, got)
}

func TestTranslator_SecretRef_ResolvesFromClient(t *testing.T) {
	t.Parallel()

	translator := NewTranslatorWithSecretClient(&fakeSecretClient{value: "secret-value"})
	got, err := translator.Translate(context.Background(), config.Value{SecretRef: &config.SecretRef{
		VaultName:  "kv-dev",
		SecretName: "reaction-token",
	}})
	require.NoError(t, err)
	assert.Equal(t, &TranslatedValue{StringValue: "secret-value", IsSecretRef: true}, got)
}

func TestTranslator_SecretRef_PropagatesClientError(t *testing.T) {
	t.Parallel()

	translator := NewTranslatorWithSecretClient(&fakeSecretClient{err: errors.New(output.ERR_KV_AUTH_FAILED + ": denied")})
	_, err := translator.Translate(context.Background(), config.Value{SecretRef: &config.SecretRef{
		VaultName:  "kv-dev",
		SecretName: "reaction-token",
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_KV_AUTH_FAILED)
}

func TestTranslator_SecretRef_ResolvesFromEnvironmentFallback(t *testing.T) {
	t.Setenv("DRASI_SECRET_kv_dev_reaction_token", "fallback-secret")

	translator := NewTranslatorWithSecretClient(nil)
	got, err := translator.Translate(context.Background(), config.Value{SecretRef: &config.SecretRef{
		VaultName:  "kv_dev",
		SecretName: "reaction_token",
	}})
	require.NoError(t, err)
	assert.Equal(t, &TranslatedValue{StringValue: "fallback-secret", IsSecretRef: true}, got)
}

func TestTranslator_SecretRef_MissingEnvironmentFallback_ReturnsError(t *testing.T) {
	translator := NewTranslator()
	_, err := translator.Translate(context.Background(), config.Value{SecretRef: &config.SecretRef{
		VaultName:  "kv_missing",
		SecretName: "reaction_token",
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_KV_AUTH_FAILED)
	assert.Contains(t, err.Error(), "DRASI_SECRET_kv_missing_reaction_token")
}

func TestTranslator_EnvRef_ResolvesFromEnvironment(t *testing.T) {
	t.Setenv("DRASI_NAMESPACE", "drasi-system")

	translator := NewTranslator()
	got, err := translator.Translate(context.Background(), config.Value{EnvRef: &config.EnvRef{Name: "DRASI_NAMESPACE"}})
	require.NoError(t, err)
	assert.Equal(t, &TranslatedValue{StringValue: "drasi-system", IsSecretRef: false}, got)
}

func TestTranslator_EnvRef_MissingVariable_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	const missing = "DRASI_ENV_MISSING_FOR_TEST"
	_ = os.Unsetenv(missing)

	translator := NewTranslator()
	_, err := translator.Translate(context.Background(), config.Value{EnvRef: &config.EnvRef{Name: missing}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), output.ERR_VALIDATION_FAILED)
}

func TestTranslator_SecretRef_EmitsAuditLog(t *testing.T) {
	// NOTE: Not parallel because slog.SetDefault mutates global state.
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	original := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(original) })

	translator := NewTranslatorWithSecretClient(&fakeSecretClient{value: "s3cret"})
	_, err := translator.Translate(context.Background(), config.Value{SecretRef: &config.SecretRef{
		VaultName:  "kv-audit-test",
		SecretName: "my-secret",
	}})
	require.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "resolving Key Vault secret reference")
	assert.Contains(t, logOutput, "kv-audit-test")
	assert.Contains(t, logOutput, "my-secret")
	// The secret value must never appear in logs.
	assert.NotContains(t, logOutput, "s3cret")
}

func TestTranslator_EnvRef_EmitsAuditLog(t *testing.T) {
	t.Setenv("DRASI_AUDIT_LOG_TEST", "test-value")

	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	original := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(original) })

	translator := NewTranslator()
	_, err := translator.Translate(context.Background(), config.Value{EnvRef: &config.EnvRef{Name: "DRASI_AUDIT_LOG_TEST"}})
	require.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "resolving environment variable reference")
	assert.Contains(t, logOutput, "DRASI_AUDIT_LOG_TEST")
	// The resolved value must never appear in logs.
	assert.NotContains(t, logOutput, "test-value")
}
