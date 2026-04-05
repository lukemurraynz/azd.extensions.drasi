package config

import (
	"fmt"
	"strings"
)

type DrasiManifest struct {
	APIVersion   string            `yaml:"apiVersion"`
	Includes     []IncludeSpec     `yaml:"includes"`
	Environments map[string]string `yaml:"environments"`
	FeatureFlags map[string]bool   `yaml:"featureFlags"`
}

type IncludeSpec struct {
	Kind    string `yaml:"kind"`
	Pattern string `yaml:"pattern"`
}

type Source struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	ID         string           `yaml:"id"`
	SourceKind string           `yaml:"sourceKind"`
	Properties map[string]Value `yaml:"properties"`
	FilePath   string           `yaml:"-"`
	Line       int              `yaml:"-"`
}

type ContinuousQuery struct {
	APIVersion    string      `yaml:"apiVersion"`
	Kind          string      `yaml:"kind"`
	ID            string      `yaml:"id"`
	QueryLanguage string      `yaml:"queryLanguage"`
	Sources       []SourceRef `yaml:"sources"`
	Query         string      `yaml:"query"`
	Joins         []JoinSpec  `yaml:"joins,omitempty"`
	Reactions     []string    `yaml:"reactions,omitempty"`
	AutoStart     bool        `yaml:"autoStart,omitempty"`
	FilePath      string      `yaml:"-"`
	Line          int         `yaml:"-"`
}

type SourceRef struct {
	ID string `yaml:"id"`
}

type JoinSpec struct {
	Type string    `yaml:"type"`
	Keys []JoinKey `yaml:"keys"`
}

type JoinKey struct {
	Label string `yaml:"label"`
	Field string `yaml:"field"`
}

type Reaction struct {
	APIVersion   string           `yaml:"apiVersion"`
	Kind         string           `yaml:"kind"`
	ID           string           `yaml:"id"`
	ReactionKind string           `yaml:"reactionKind"`
	Config       map[string]Value `yaml:"config,omitempty"`
	FilePath     string           `yaml:"-"`
	Line         int              `yaml:"-"`
}

type Middleware struct {
	APIVersion     string           `yaml:"apiVersion"`
	Kind           string           `yaml:"kind"`
	ID             string           `yaml:"id"`
	MiddlewareKind string           `yaml:"middlewareKind"`
	Config         map[string]Value `yaml:"config,omitempty"`
	FilePath       string           `yaml:"-"`
	Line           int              `yaml:"-"`
}

type Environment struct {
	Name       string            `yaml:"name"`
	Parameters map[string]string `yaml:"parameters,omitempty"`
}

// Value can hold a plain string, a Key Vault secret reference, or an env var reference.
type Value struct {
	StringValue string     `yaml:"value,omitempty"`
	SecretRef   *SecretRef `yaml:"secretRef,omitempty"`
	EnvRef      *EnvRef    `yaml:"envRef,omitempty"`
}

type SecretRef struct {
	VaultName  string `yaml:"vaultName"`
	SecretName string `yaml:"secretName"`
}

type EnvRef struct {
	Name string `yaml:"name"`
}

type ResolvedManifest struct {
	Sources      []Source
	Queries      []ContinuousQuery
	Reactions    []Reaction
	Middlewares  []Middleware
	Environment  Environment
	FeatureFlags map[string]bool
}

// ComponentHash stores the content hash of a deployed component.
type ComponentHash struct {
	Kind string
	ID   string
	Hash string
}

// StateKey returns the azd environment key for storing this component's hash.
// Format: DRASI_HASH_<KIND>_<ID> (KIND uppercased, ID verbatim).
func (h ComponentHash) StateKey() string {
	return fmt.Sprintf("DRASI_HASH_%s_%s", strings.ToUpper(h.Kind), h.ID)
}
