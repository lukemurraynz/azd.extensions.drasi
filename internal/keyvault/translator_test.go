package keyvault

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/azure/azd.extensions.drasi/internal/output"
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
