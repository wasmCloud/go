package wadm

import (
	"encoding/json"
	"errors"
	"fmt"

	yaml "github.com/goccy/go-yaml"
	"go.wasmcloud.dev/wasmbus"
)

// API structure
// wadm.api.{lattice-id}.{category}.{operation}.{object}

type (
	ComponentType string
	TraitType     string
	StatusType    string
)

const (
	ComponentTypeComponent  ComponentType = "component"
	ComponentTypeCapability ComponentType = "capability"

	TraitTypeLink         TraitType = "link"
	TraitTypeSpreadScaler TraitType = "spreadscaler"
	TraitTypeDaemonScaler TraitType = "daemonscaler"

	StatusTypeWaiting     StatusType = "waiting"
	StatusTypeUndeployed  StatusType = "undeployed"
	StatusTypeReconciling StatusType = "reconciling"
	StatusTypeDeployed    StatusType = "deployed"
	StatusTypeFailed      StatusType = "failed"

	StatusResultError string = "error"
	// NOTE(lxf): inconsistency (should be succcess) ?
	StatusResultOk       string = "ok"
	StatusResultNotFound string = "notfound"

	DeployResultError        string = "error"
	DeployResultAcknowledged string = "acknowledged"
	DeployResultNotFound     string = "notfound"

	DeleteResultError   string = "error"
	DeleteResultNoop    string = "noop"
	DeleteResultDeleted string = "deleted"

	GetResultError    string = "error"
	GetResultSuccess  string = "success"
	GetResultNotFound string = "not_found"

	PutResultError      string = "error"
	PutResultCreated    string = "created"
	PutResultNewVersion string = "newversion"

	DefaultManifestApiVersion string = "core.oam.dev/v1beta1"
	DefaultManifestKind       string = "Manifest"

	// LatestVersion is a constant that represents the latest version of a model
	LatestVersion = ""
)

// RawMessage knows how to stash json & yaml
type RawMessage []byte

func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m.marshal()
}

func (m RawMessage) MarshalYAML() ([]byte, error) {
	return m.marshal()
}

func (m RawMessage) marshal() ([]byte, error) {
	return m, nil
}

func (m *RawMessage) UnmarshalJSON(data []byte) error { return m.unmarshal(data) }
func (m *RawMessage) UnmarshalYAML(data []byte) error { return m.unmarshal(data) }

func (m *RawMessage) unmarshal(data []byte) error {
	if m == nil {
		return errors.New("RawMessage: unmarshal on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type Status struct {
	Status  StatusInfo     `json:"status"`
	Scalers []ScalerStatus `json:"scalers,omitempty"`
}

type ModelStatusRequest struct {
	Name string `json:"name"`
}

type ModelStatusResponse struct {
	BaseResponse
	Status *Status `json:"status,omitempty"`
}

type ModelPutRequest struct {
	Manifest `json:",inline"`
}

type ModelPutResponse struct {
	BaseResponse
	Name           string `json:"name,omitempty"`
	TotalVersions  int    `json:"total_versions,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
}

type StatusInfo struct {
	Type    StatusType `json:"type"`
	Message string     `json:"message,omitempty"`
}

type ScalerStatus struct {
	Id     string     `json:"id"`
	Kind   string     `json:"kind"`
	Name   string     `json:"name"`
	Status StatusInfo `json:"status"`
}

type DetailedStatus struct {
	Info    StatusInfo     `json:"status"`
	Scalers []ScalerStatus `json:"scalers,omitempty"`
}

type ModelSummary struct {
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Description     string          `json:"description,omitempty"`
	DeployedVersion string          `json:"deployed_version,omitempty"`
	DetailedStatus  *DetailedStatus `json:"detailed_status,omitempty"`

	// Deprecated
	Status StatusType `json:"status,omitempty"`
}

type ManifestMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type Policy struct {
	Name       string            `json:"name" yaml:"name"`
	Type       string            `json:"type" yaml:"type"`
	Properties map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type ConfigProperty struct {
	Name       string            `json:"name" yaml:"name"`
	Properties map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type SecretSourceProperty struct {
	Policy  string `json:"policy"  yaml:"policy"`
	Key     string `json:"key"    yaml:"key"`
	Field   string `json:"field,omitempty" yaml:"field,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type SecretProperty struct {
	Name       string               `json:"name" yaml:"name"`
	Properties SecretSourceProperty `json:"properties" yaml:"properties"`
}

type SharedApplicationComponentProperties struct {
	Name      string `json:"name" yaml:"name"`
	Component string `json:"component" yaml:"component"`
}

type ComponentProperties struct {
	Image       string                                `json:"image" yaml:"image"`
	Application *SharedApplicationComponentProperties `json:"application,omitempty" yaml:"application,omitempty"`
	Id          string                                `json:"id,omitempty" yaml:"id,omitempty"`
	Config      []ConfigProperty                      `json:"config,omitempty" yaml:"config,omitempty"`
	Secrets     []SecretProperty                      `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type ConfigDefinition struct {
	Config  []ConfigProperty `json:"config,omitempty" yaml:"config,omitempty"`
	Secrets []SecretProperty `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type TargetConfigDefinition struct {
	Name    string           `json:"name" yaml:"name"`
	Config  []ConfigProperty `json:"config,omitempty" yaml:"config,omitempty"`
	Secrets []SecretProperty `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type rawTargetConfigDefinition TargetConfigDefinition

func (t *TargetConfigDefinition) UnmarshalYAML(data []byte) error {
	*t = TargetConfigDefinition{}
	if err := yaml.Unmarshal(data, &t.Name); err == nil {
		return nil
	}

	rt := &rawTargetConfigDefinition{}
	if err := yaml.Unmarshal(data, rt); err != nil {
		return err
	}
	*t = TargetConfigDefinition(*rt)

	return nil
}

type LinkProperty struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	Namespace  string                  `json:"namespace" yaml:"namespace"`
	Package    string                  `json:"package" yaml:"package"`
	Interfaces []string                `json:"interfaces" yaml:"interfaces"`
	Source     *ConfigDefinition       `json:"source,omitempty" yaml:"source,omitempty"`
	Target     *TargetConfigDefinition `json:"target,omitempty" yaml:"target,omitempty"`
}

type Spread struct {
	Name         string            `json:"name" yaml:"name"`
	Requirements map[string]string `json:"requirements,omitempty" yaml:"requirements,omitempty"`
	Weight       *int              `json:"weight,omitempty" yaml:"weight,omitempty"`
}

type SpreadScalerProperty struct {
	Instances int      `json:"instances" yaml:"instances"`
	Spread    []Spread `json:"spread,omitempty" yaml:"spread,omitempty"`
}

type Trait struct {
	Type         TraitType             `json:"type" yaml:"type"`
	Link         *LinkProperty         `json:"-" yaml:"-"`
	SpreadScaler *SpreadScalerProperty `json:"-" yaml:"-"`
}

type rawTrait struct {
	Type       TraitType  `json:"type" yaml:"type"`
	Properties RawMessage `json:"properties,omitempty" yaml:"properties,omitempty"`
}

func (t Trait) MarshalYAML() ([]byte, error) {
	return t.marshal(yaml.Marshal)
}

func (t Trait) MarshalJSON() ([]byte, error) {
	return t.marshal(json.Marshal)
}

func (t Trait) marshal(fn func(interface{}) ([]byte, error)) ([]byte, error) {
	r := rawTrait{Type: t.Type}

	var err error
	switch t.Type {
	case TraitTypeLink:
		r.Properties, err = fn(t.Link)
	case TraitTypeSpreadScaler, TraitTypeDaemonScaler:
		r.Properties, err = fn(t.SpreadScaler)
	default:
		err = wasmbus.ErrEncode
	}
	if err != nil {
		return nil, err
	}

	return fn(r)
}

func (t *Trait) unmarshal(data []byte, fn func([]byte, interface{}) error) error {
	var r rawTrait
	if err := fn(data, &r); err != nil {
		return err
	}

	*t = Trait{Type: r.Type}

	var err error
	switch r.Type {
	case TraitTypeLink:
		t.Link = &LinkProperty{}
		err = fn(r.Properties, t.Link)
	case TraitTypeSpreadScaler, TraitTypeDaemonScaler:
		t.SpreadScaler = &SpreadScalerProperty{}
		err = fn(r.Properties, t.SpreadScaler)
	default:
		err = wasmbus.ErrDecode
	}
	if err != nil {
		return err
	}

	return nil
}

func (t *Trait) UnmarshalJSON(data []byte) error {
	return t.unmarshal(data, json.Unmarshal)
}

func (t *Trait) UnmarshalYAML(data []byte) error {
	return t.unmarshal(data, yaml.Unmarshal)
}

type Component struct {
	Name       string              `json:"name"`
	Type       ComponentType       `json:"type"`
	Properties ComponentProperties `json:"properties"`
	Traits     []Trait             `json:"traits" yaml:"traits"`
}

type ManifestSpec struct {
	Components []Component `json:"components,omitempty" yaml:"components,omitempty"`
	Policies   []Policy    `json:"policies,omitempty" yaml:"policies,omitempty"`
}

type Manifest struct {
	ApiVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   ManifestMetadata `json:"metadata"`
	Spec       ManifestSpec     `json:"spec"`
}

func (m *Manifest) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Manifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(m)
}

func (m *Manifest) Validate() []error {
	var errs []error

	// no duplicate component names, check 'id' field too
	componentNames := make(map[string]bool)
	for _, c := range m.Spec.Components {
		id := c.Properties.Id
		if id == "" {
			id = c.Name
		}

		if _, ok := componentNames[id]; ok {
			errs = append(errs, fmt.Errorf("%w: duplicate component name %s", wasmbus.ErrValidation, id))
			continue
		}

		componentNames[id] = true
	}

	// no version latest
	if m.Metadata.Annotations != nil {
		if version, ok := m.Metadata.Annotations[VersionAnnotation]; ok {
			if version == "latest" {
				errs = append(errs, fmt.Errorf("%w: '%s' version is reserved", wasmbus.ErrValidation, version))
			}
		}
	}

	return errs
}

func (m *Manifest) IsValid() bool {
	return len(m.Validate()) == 0
}

type ModelListRequest struct{}

type BaseResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

func (b *BaseResponse) IsError() bool {
	return b.Result == StatusResultError || b.Result == StatusResultNotFound
}

type ModelListResponse struct {
	BaseResponse
	Models []ModelSummary `json:"models,omitempty"`
}

type ModelGetRequest struct {
	Name    string `json:"-"`
	Version string `json:"version,omitempty"`
}

type ModelGetResponse struct {
	BaseResponse
	Manifest *Manifest `json:"manifest,omitempty"`
}

type ModelVersionsRequest struct {
	Name string `json:"-"`
}

type VersionInfo struct {
	Version  string `json:"version"`
	Deployed bool   `json:"deployed"`
}

type ModelVersionsResponse struct {
	BaseResponse
	Versions []VersionInfo `json:"versions"`
}

type ModelDeleteRequest struct {
	Name    string `json:"-"`
	Version string `json:"version,omitempty"`
}

type ModelDeleteResponse struct {
	BaseResponse
	Undeploy bool `json:"undeploy"`
}

type ModelDeployRequest struct {
	Name    string `json:"-"`
	Version string `json:"version,omitempty"`
}

type ModelDeployResponse struct {
	BaseResponse
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ModelUndeployRequest struct {
	Name string `json:"-"`
}

type ModelUndeployResponse struct {
	BaseResponse
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
