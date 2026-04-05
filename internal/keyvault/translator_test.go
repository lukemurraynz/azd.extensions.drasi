package keyvault

import (
	"context"
	"testing"

	"github.com/azure/azd.extensions.drasi/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslator_PlainString_PassesThrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value config.Value
		want  *TranslatedValue
	}{
		{
			name:  "plain string passes through",
			value: config.Value{StringValue: "hello"},
			want:  &TranslatedValue{StringValue: "hello", IsSecretRef: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			translator := NewTranslator()
			got, err := translator.Translate(context.Background(), tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTranslator_SecretRef_ReturnsError_NotYetImplemented(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value config.Value
	}{
		{
			name: "secret ref requires implementation",
			value: config.Value{SecretRef: &config.SecretRef{
				VaultName:  "kv-dev",
				SecretName: "reaction-token",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			translator := NewTranslator()
			_, _ = translator.Translate(context.Background(), tt.value)
		})
	}
}

func TestTranslator_EnvRef_PassesThrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value config.Value
	}{
		{
			name: "env ref not yet implemented",
			value: config.Value{EnvRef: &config.EnvRef{
				Name: "DRASI_NAMESPACE",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			translator := NewTranslator()
			_, _ = translator.Translate(context.Background(), tt.value)
		})
	}
}
