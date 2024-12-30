package events

import (
	"encoding/json"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var ErrParsingEvent = fmt.Errorf("error parsing event")

type Event struct {
	CloudEvent *cloudevents.Event `json:"-"`
	BusEvent   any                `json:"-"`
}

type ComponentId string

type ComponentDescription struct {
	Id           ComponentId       `json:"id"`
	ImageRef     string            `json:"image_ref"`
	Name         string            `json:"name"`
	Annotations  map[string]string `json:"annotations"`
	Revision     int               `json:"revision"`
	MaxInstances int               `json:"max_instances"`
}

type CapabilityDescription struct {
	Id          ComponentId       `json:"id"`
	ImageRef    string            `json:"image_ref"`
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
	Revision    int               `json:"revision"`
}

type HostHeartbeat struct {
	HostId        string                  `json:"host_id"`
	UptimeSeconds int                     `json:"uptime_seconds"`
	UptimeHuman   string                  `json:"uptime_human"`
	Version       string                  `json:"version"`
	Labels        map[string]string       `json:"labels"`
	FriendlyName  string                  `json:"friendly_name"`
	Issuer        string                  `json:"issuer"`
	Components    []ComponentDescription  `json:"components"`
	Providers     []CapabilityDescription `json:"providers"`
}

type HealthCheckStatus struct {
	HostId     string `json:"host_id"`
	ProviderId string `json:"provider_id"`
}
type ComponentClaims struct {
	CallAlias      string   `json:"call_alias"`
	ExpiresHuman   string   `json:"expires_human"`
	Issuer         string   `json:"issuer"`
	Name           string   `json:"name"`
	NotBeforeHuman string   `json:"not_before_human"`
	Revision       int      `json:"revision"`
	Tags           []string `json:"tags"`
	Version        string   `json:"version"`
}

type ComponentScaled struct {
	HostId       string            `json:"host_id"`
	Annotations  map[string]string `json:"annotations"`
	ImageRef     string            `json:"image_ref"`
	MaxInstances int               `json:"max_instances"`
	ComponentId  string            `json:"component_id"`
	Claims       ComponentClaims   `json:"claims"`
	PublicKey    string            `json:"public_key"`
}

type ComponentScaleFailed struct {
	HostId       string            `json:"host_id"`
	Annotations  map[string]string `json:"annotations"`
	ImageRef     string            `json:"image_ref"`
	MaxInstances int               `json:"max_instances"`
	ComponentId  string            `json:"component_id"`
	Claims       ComponentClaims   `json:"claims"`
	PublicKey    string            `json:"public_key"`

	Error string `json:"error"`
}

type LinkDefSet struct {
	Source        ComponentId `json:"source_id"`
	Target        string      `json:"target"`
	Name          string      `json:"name"`
	WitNamespace  string      `json:"wit_namespace"`
	WitPackage    string      `json:"wit_package"`
	WitInterfaces []string    `json:"interfaces"`
	SourceConfig  []string    `json:"source_config"`
	TargetConfig  []string    `json:"target_config"`
}

type LinkDefSetFailed struct {
	Source        ComponentId `json:"source_id"`
	Target        string      `json:"target"`
	Name          string      `json:"name"`
	WitNamespace  string      `json:"wit_namespace"`
	WitPackage    string      `json:"wit_package"`
	WitInterfaces []string    `json:"interfaces"`
	SourceConfig  []string    `json:"source_config"`
	TargetConfig  []string    `json:"target_config"`
	Error         string      `json:"error"`
}

type LinkDefDeleted struct {
	Source        ComponentId `json:"source_id"`
	Target        string      `json:"target"`
	Name          string      `json:"name"`
	WitNamespace  string      `json:"wit_namespace"`
	WitPackage    string      `json:"wit_package"`
	WitInterfaces []string    `json:"interfaces"`
}

type ProviderStarted struct {
	HostId      string            `json:"host_id"`
	ImageRef    string            `json:"image_ref"`
	ProviderId  string            `json:"provider_id"`
	Annotations map[string]string `json:"annotations"`
	Claims      ComponentClaims   `json:"claims"`
}

type ProviderStartFailed struct {
	// missing from the original code
	HostId      string            `json:"host_id"`
	ImageRef    string            `json:"image_ref"`
	Annotations map[string]string `json:"annotations"`
	Claims      ComponentClaims   `json:"claims"`

	ProviderId  string `json:"provider_id"`
	ProviderRef string `json:"provider_ref"`
	Error       string `json:"error"`
}

type ProviderStopped struct {
	HostId      string            `json:"host_id"`
	ProviderId  string            `json:"provider_id"`
	Annotations map[string]string `json:"annotations"`
	Reason      string            `json:"reason"`
}

type HealthCheckPassed struct {
	HostId     string `json:"host_id"`
	ProviderId string `json:"provider_id"`
}

type HealthCheckFailed struct {
	HostId     string `json:"host_id"`
	ProviderId string `json:"provider_id"`
}

type ConfigSet struct {
	ConfigName string `json:"config_name"`
}

type ConfigDeleted struct {
	ConfigName string `json:"config_name"`
}

type LabelsChanged struct {
	HostId string            `json:"host_id"`
	Labels map[string]string `json:"labels"`
}

type HostStarted struct {
	// missing from the original code
	HostId string `json:"host_id"`

	Labels        map[string]string `json:"labels"`
	FriendlyName  string            `json:"friendly_name"`
	UptimeSeconds int               `json:"uptime_seconds"`
	Version       string            `json:"version"`
}

type HostStopped struct {
	HostId string            `json:"host_id"`
	Labels map[string]string `json:"labels"`
	Reason string            `json:"reason"`
}

func KnownEvents(typ string) any {
	switch typ {
	case "com.wasmcloud.lattice.host_heartbeat":
		return &HostHeartbeat{}
	case "com.wasmcloud.lattice.component_scaled":
		return &ComponentScaled{}
	case "com.wasmcloud.lattice.component_scale_failed":
		return &ComponentScaleFailed{}
	case "com.wasmcloud.lattice.linkdef_set":
		return &LinkDefSet{}
	case "com.wasmcloud.lattice.linkdef_set_failed":
		return &LinkDefSetFailed{}
	case "com.wasmcloud.lattice.linkdef_deleted":
		return &LinkDefDeleted{}
	case "com.wasmcloud.lattice.provider_started":
		return &ProviderStarted{}
	case "com.wasmcloud.lattice.provider_start_failed":
		return &ProviderStartFailed{}
	case "com.wasmcloud.lattice.provider_stopped":
		return &ProviderStopped{}
	case "com.wasmcloud.lattice.health_check_passed":
		return &HealthCheckPassed{}
	case "com.wasmcloud.lattice.health_check_failed":
		return &HealthCheckFailed{}
	case "com.wasmcloud.lattice.health_check_status":
		return &HealthCheckStatus{}
	case "com.wasmcloud.lattice.config_set":
		return &ConfigSet{}
	case "com.wasmcloud.lattice.config_deleted":
		return &ConfigDeleted{}
	case "com.wasmcloud.lattice.labels_changed":
		return &LabelsChanged{}
	case "com.wasmcloud.lattice.host_started":
		return &HostStarted{}
	case "com.wasmcloud.lattice.host_stopped":
		return &HostStopped{}
	default:
		return nil
	}
}

func ParseEvent(data []byte) (Event, error) {
	ce := cloudevents.NewEvent()
	ev := Event{
		CloudEvent: &ce,
	}
	err := json.Unmarshal(data, ev.CloudEvent)
	if err != nil {
		return ev, fmt.Errorf("%w: %s", ErrParsingEvent, err)
	}

	if err := ev.CloudEvent.Validate(); err != nil {
		return ev, fmt.Errorf("%w: %s", ErrParsingEvent, err)
	}

	ceType := ev.CloudEvent.Type()
	e := KnownEvents(ceType)
	if e == nil {
		return ev, fmt.Errorf("%w: unknown event type '%s'", ErrParsingEvent, ceType)
	}

	ceData := ev.CloudEvent.Data()
	if err := json.Unmarshal(ceData, e); err != nil {
		return ev, fmt.Errorf("%w: %s", ErrParsingEvent, err)
	}
	ev.BusEvent = e

	return ev, nil
}
