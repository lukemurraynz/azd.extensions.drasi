package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type DrasiManifest struct {
	APIVersion     string            `yaml:"apiVersion"`
	Includes       []IncludeSpec     `yaml:"includes"`
	Environments   map[string]string `yaml:"environments"`
	FeatureFlags   map[string]bool   `yaml:"featureFlags"`
	SecretMappings []SecretMapping   `yaml:"secretMappings,omitempty"`
}

// SecretMapping maps an Azure Key Vault secret to a key inside a Kubernetes Secret.
// During deploy, the extension fetches the value from Key Vault and upserts the
// Kubernetes Secret so Drasi components can reference it with kind: Secret.
type SecretMapping struct {
	VaultName  string `yaml:"vaultName"`
	SecretName string `yaml:"secretName"`
	K8sSecret  string `yaml:"k8sSecret"`
	K8sKey     string `yaml:"k8sKey"`
	Namespace  string `yaml:"namespace,omitempty"`
}

type IncludeSpec struct {
	Kind    string `yaml:"kind"`
	Pattern string `yaml:"pattern"`
}

// SourceSpec holds the spec block of a Source resource.
type SourceSpec struct {
	Kind       string           `yaml:"kind"`
	Properties map[string]Value `yaml:"properties,omitempty"`
}

// QuerySourceSubscription is a source reference inside a query's sources block.
type QuerySourceSubscription struct {
	ID    string      `yaml:"id"`
	Nodes []QueryNode `yaml:"nodes,omitempty"`
}

// QueryNode filters which node labels are pulled from a source.
type QueryNode struct {
	SourceLabel string `yaml:"sourceLabel"`
}

// QuerySourcesSpec holds the subscriptions block inside a query spec.
type QuerySourcesSpec struct {
	Subscriptions []QuerySourceSubscription `yaml:"subscriptions" json:",omitempty"`
}

// QuerySpec holds the spec block of a ContinuousQuery resource.
type QuerySpec struct {
	Mode      string           `yaml:"mode"                json:",omitempty"`
	Sources   QuerySourcesSpec `yaml:"sources"             json:",omitempty"`
	Query     string           `yaml:"query"`
	Reactions []string         `yaml:"reactions,omitempty" json:",omitempty"`
}

// ReactionSpec holds the spec block of a Reaction resource.
type ReactionSpec struct {
	Kind    string            `yaml:"kind"`
	Queries map[string]string `yaml:"queries,omitempty"`
}

// MiddlewareSpec holds the spec block of a Middleware resource.
type MiddlewareSpec struct {
	Kind   string           `yaml:"kind"`
	Config map[string]Value `yaml:"config,omitempty"`
}

// Source represents a Drasi Source resource.
// SourceKind and Properties are legacy fields; use Spec.Kind and Spec.Properties instead.
type Source struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	ID         string           `yaml:"name"`
	SourceKind string           `yaml:"-"`
	Properties map[string]Value `yaml:"-"`
	Spec       SourceSpec       `yaml:"spec"`
	FilePath   string           `yaml:"-"`
	Line       int              `yaml:"-"`
}

// ContinuousQuery represents a Drasi ContinuousQuery resource.
// Sources and Reactions are populated from Spec after YAML decoding.
type ContinuousQuery struct {
	APIVersion    string      `yaml:"apiVersion"`
	Kind          string      `yaml:"kind"`
	ID            string      `yaml:"name"`
	QueryLanguage string      `yaml:"-"`
	Sources       []SourceRef `yaml:"-"`
	Joins         []JoinSpec  `yaml:"-"`
	Reactions     []string    `yaml:"-"`
	AutoStart     bool        `yaml:"-"`
	Spec          QuerySpec   `yaml:"spec"`
	FilePath      string      `yaml:"-"`
	Line          int         `yaml:"-"`
}

// SourceRef is a flat source reference used by validation and tests.
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

// Reaction represents a Drasi Reaction resource.
// ReactionKind and Config are legacy fields; use Spec.Kind and Spec.Queries instead.
type Reaction struct {
	APIVersion   string           `yaml:"apiVersion"`
	Kind         string           `yaml:"kind"`
	ID           string           `yaml:"name"`
	ReactionKind string           `yaml:"-"`
	Config       map[string]Value `yaml:"-"`
	Spec         ReactionSpec     `yaml:"spec"`
	FilePath     string           `yaml:"-"`
	Line         int              `yaml:"-"`
}

// Middleware represents a Drasi Middleware resource.
// MiddlewareKind and Config are legacy fields; use Spec.Kind and Spec.Config instead.
type Middleware struct {
	APIVersion     string           `yaml:"apiVersion"`
	Kind           string           `yaml:"kind"`
	ID             string           `yaml:"name"`
	MiddlewareKind string           `yaml:"-"`
	Config         map[string]Value `yaml:"-"`
	Spec           MiddlewareSpec   `yaml:"spec"`
	FilePath       string           `yaml:"-"`
	Line           int              `yaml:"-"`
}

// ComponentRef identifies a specific component by kind and ID.
type ComponentRef struct {
	Kind string `yaml:"kind" json:"kind"`
	ID   string `yaml:"id"   json:"id"`
}

// Components defines inclusion/exclusion rules for environment overlays.
type Components struct {
	Exclude []ComponentRef `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

type Environment struct {
	Name       string            `yaml:"name" json:"name"`
	Parameters map[string]string `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Components Components        `yaml:"components,omitempty" json:"components,omitempty"`
}

// Value can hold a plain string, a Key Vault secret reference, or an env var reference.
type Value struct {
	StringValue string     `yaml:"value,omitempty"`
	SecretRef   *SecretRef `yaml:"secretRef,omitempty"`
	EnvRef      *EnvRef    `yaml:"envRef,omitempty"`
}

// UnmarshalYAML accepts both plain scalar strings and the {value: "..."} struct form.
// Drasi YAML files may use either format:
//   - Plain scalar:      inCluster: "true"      → Value{StringValue: "true"}
//   - Struct form:       inCluster: {value: "true"} → Value{StringValue: "true"}
//   - SecretRef:         endpoint: {secretRef: {...}} → Value{SecretRef: &SecretRef{...}}
func (v *Value) UnmarshalYAML(node *yaml.Node) error {
	// Plain string scalar: inCluster: "true"
	if node.Kind == yaml.ScalarNode {
		v.StringValue = node.Value
		return nil
	}
	// Mapping node: {value: "...", secretRef: {...}, envRef: {...}}
	// Decode into a shadow struct to avoid infinite recursion.
	type valueShadow struct {
		StringValue string     `yaml:"value,omitempty"`
		SecretRef   *SecretRef `yaml:"secretRef,omitempty"`
		EnvRef      *EnvRef    `yaml:"envRef,omitempty"`
	}
	var shadow valueShadow
	if err := node.Decode(&shadow); err != nil {
		return err
	}
	v.StringValue = shadow.StringValue
	v.SecretRef = shadow.SecretRef
	v.EnvRef = shadow.EnvRef
	return nil
}

type SecretRef struct {
	VaultName  string `yaml:"vaultName"`
	SecretName string `yaml:"secretName"`
}

type EnvRef struct {
	Name string `yaml:"name"`
}

type ResolvedManifest struct {
	Sources        []Source
	Queries        []ContinuousQuery
	Reactions      []Reaction
	Middlewares    []Middleware
	Environment    Environment
	FeatureFlags   map[string]bool
	SecretMappings []SecretMapping
	// ManifestDir is the absolute directory of the drasi.yaml file.
	// Used by the deploy engine to locate original YAML files for each component.
	ManifestDir string
}

const (
	WarningUndeclaredOverlayParameter = "WARN_UNDECLARED_PARAMETER"
	WarningInvalidComponentExclusion  = "WARN_INVALID_COMPONENT_EXCLUSION"
	WarningMissingComponentExclusion  = "WARN_MISSING_COMPONENT_EXCLUSION"
)

// OverlayWarning captures a non-fatal issue found while resolving environment overlays.
type OverlayWarning struct {
	Code    string
	Message string
}

// ComponentHash stores the content hash of a deployed component.
type ComponentHash struct {
	Kind string
	ID   string
	Hash string
}

// StateKey returns the azd environment key for storing this component's hash.
// Format: DRASI_HASH_<KIND>_<ID> (KIND uppercased, ID sanitized for .env).
// Hyphens in ID are replaced with underscores because .env variable names
// only support letters, digits, and underscores.
func (h ComponentHash) StateKey() string {
	sanitized := strings.ReplaceAll(h.ID, "-", "_")
	return fmt.Sprintf("DRASI_HASH_%s_%s", strings.ToUpper(h.Kind), sanitized)
}
