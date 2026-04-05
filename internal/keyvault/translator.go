package keyvault

import (
	"context"

	"github.com/azure/azd.extensions.drasi/internal/config"
)

// TranslatedValue is the result of resolving a config.Value.
type TranslatedValue struct {
	// StringValue is the resolved plain string (populated for plain strings and resolved secrets).
	StringValue string
	// IsSecretRef is true when the original value was a SecretRef.
	IsSecretRef bool
}

// Translator resolves config.Value entries, replacing Key Vault SecretRef
// entries with Kubernetes Secret references.
type Translator struct {
	// kvClient will be injected during implementation
}

// NewTranslator creates a Translator.
func NewTranslator() *Translator {
	return &Translator{}
}

// Translate resolves a single config.Value.
// - Plain string → TranslatedValue{StringValue: value.StringValue}
// - SecretRef → resolves via Key Vault (not implemented yet)
func (t *Translator) Translate(ctx context.Context, v config.Value) (*TranslatedValue, error) {
	panic("not implemented")
}
