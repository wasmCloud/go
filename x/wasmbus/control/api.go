package control

import "context"

type APIv1 interface {
	ProviderAuction(ctx context.Context, req *ProviderAuctionRequest) (*ProviderAuctionResponse, error)
	ComponentAuction(ctx context.Context, req *ComponentAuctionRequest) (*ComponentAuctionResponse, error)

	ScaleComponent(ctx context.Context, req *ScaleComponentRequest) (*ScaleComponentResponse, error)
	UpdateComponent(ctx context.Context, req *UpdateComponentRequest) (*UpdateComponentResponse, error)

	ProviderStart(ctx context.Context, req *ProviderStartRequest) (*ProviderStartResponse, error)
	ProviderStop(ctx context.Context, req *ProviderStopRequest) (*ProviderStopResponse, error)

	HostStop(ctx context.Context, req *HostStopRequest) (*HostStopResponse, error)

	ConfigPut(ctx context.Context, req *ConfigPutRequest) (*ConfigPutResponse, error)
	ConfigGet(ctx context.Context, req *ConfigGetRequest) (*ConfigGetResponse, error)
	ConfigDelete(ctx context.Context, req *ConfigDeleteRequest) (*ConfigDeleteResponse, error)

	HostLabelPut(ctx context.Context, req *HostLabelPutRequest) (*HostLabelPutResponse, error)
	HostLabelDelete(ctx context.Context, req *HostLabelDeleteRequest) (*HostLabelDeleteResponse, error)

	LinkGet(ctx context.Context, req *LinkGetRequest) (*LinkGetResponse, error)
	LinkPut(ctx context.Context, req *LinkPutRequest) (*LinkPutResponse, error)
	LinkDelete(ctx context.Context, req *LinkDeleteRequest) (*LinkDeleteResponse, error)

	ClaimsGet(ctx context.Context, req *ClaimsGetRequest) (*ClaimsGetResponse, error)

	HostInventory(ctx context.Context, req *HostInventoryRequest) (*HostInventoryResponse, error)
	HostPing(ctx context.Context, req *HostPingRequest) (*HostPingResponse, error)
}

type Response[T any] struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Response T      `json:"response"`
}

type ProviderAuctionRequest struct {
	Constraints map[string]string `json:"constraints,omitempty"`
	ProviderId  string            `json:"provider_id,omitempty"`
	ProviderRef string            `json:"provider_ref,omitempty"`
}

type ProviderAuctionResponsePayload struct {
	HostId      string            `json:"host_id"`
	Constraints map[string]string `json:"constraints,omitempty"`
	ProviderId  string            `json:"provider_id,omitempty"`
	ProviderRef string            `json:"provider_ref,omitempty"`
}

type ProviderAuctionResponse = Response[ProviderAuctionResponsePayload]

type ComponentAuctionRequest struct {
	Constraints  map[string]string `json:"constraints,omitempty"`
	ComponentId  string            `json:"component_id,omitempty"`
	ComponentRef string            `json:"component_ref,omitempty"`
}

type ComponentAuctionResponsePayload struct {
	HostId       string            `json:"host_id"`
	Constraints  map[string]string `json:"constraints,omitempty"`
	ComponentId  string            `json:"component_id,omitempty"`
	ComponentRef string            `json:"component_ref,omitempty"`
}

type ComponentAuctionResponse = Response[ComponentAuctionResponsePayload]

type ScaleComponentRequest struct {
	ComponentId  string            `json:"component_id"`
	ComponentRef string            `json:"component_ref"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Count        int               `json:"count"`
	HostId       string            `json:"host_id"`
	Config       []string          `json:"config,omitempty"`
	AllowUpdate  bool              `json:"allow_update,omitempty"`
}

type ScaleComponentResponsePayload struct {
}

type ScaleComponentResponse = Response[ScaleComponentResponsePayload]

type UpdateComponentRequest struct {
	ComponentId     string            `json:"component_id"`
	HostId          string            `json:"host_id"`
	NewComponentRef string            `json:"new_component_ref"`
	Annotations     map[string]string `json:"annotations,omitempty"`
}

type UpdateComponentResponsePayload struct {
}

type UpdateComponentResponse = Response[UpdateComponentResponsePayload]

type ProviderStartRequest struct {
	HostId      string            `json:"host_id"`
	ProviderId  string            `json:"provider_id"`
	ProviderRef string            `json:"provider_ref"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Config      []string          `json:"config,omitempty"`
}

type ProviderStartResponsePayload struct{}

type ProviderStartResponse = Response[ProviderStartResponsePayload]

type ProviderStopRequest struct {
	HostId     string `json:"host_id"`
	ProviderId string `json:"provider_id"`
}

type ProviderStopResponsePayload struct{}

type ProviderStopResponse = Response[ProviderStopResponsePayload]

type HostStopRequest struct {
	HostId  string `json:"host_id"`
	Timeout int    `json:"timeout,omitempty"`
}

type HostStopResponsePayload struct{}

type HostStopResponse = Response[HostStopResponsePayload]

type ConfigPutRequest struct {
	Name   string            `json:"-"`
	Values map[string]string `json:",inline"`
}

func (c *ConfigPutRequest) SetName(name string) {
	c.Name = name
}

type ConfigPutResponsePayload struct{}

type ConfigPutResponse = Response[ConfigPutResponsePayload]

type ConfigGetRequest struct {
	Name string `json:"-"`
}

func (c *ConfigGetRequest) SetName(name string) {
	c.Name = name
}

type ConfigGetResponsePayload = map[string]string

type ConfigGetResponse = Response[ConfigGetResponsePayload]

type ConfigDeleteRequest struct {
	Name string `json:"-"`
}

func (c *ConfigDeleteRequest) SetName(name string) {
	c.Name = name
}

type ConfigDeleteResponsePayload struct{}

type ConfigDeleteResponse = Response[ConfigDeleteResponsePayload]

type HostLabelPutRequest map[string]string

type HostLabelPutResponsePayload struct{}

type HostLabelPutResponse = Response[HostLabelPutResponsePayload]

type HostLabelDeleteRequest struct {
	Key string `json:"key"`
}

type HostLabelDeleteResponsePayload struct{}

type HostLabelDeleteResponse = Response[HostLabelDeleteResponsePayload]

type LinkGetRequest struct{}

type LinkGetResponsePayload struct {
	SourceId      string   `json:"source_id"`
	Target        string   `json:"target"`
	Name          string   `json:"name"`
	WitNamespace  string   `json:"wit_namespace"`
	WitPackage    string   `json:"wit_package"`
	WitInterfaces []string `json:"interfaces"`
	SourceConfig  []string `json:"source_config"`
	TargetConfig  []string `json:"target_config"`
}

type LinkGetResponse = Response[[]LinkGetResponsePayload]

type LinkPutRequest struct {
	SourceId      string   `json:"source_id"`
	Target        string   `json:"target"`
	Name          string   `json:"name"`
	WitNamespace  string   `json:"wit_namespace"`
	WitPackage    string   `json:"wit_package"`
	WitInterfaces []string `json:"interfaces"`
	SourceConfig  []string `json:"source_config"`
	TargetConfig  []string `json:"target_config"`
}

type LinkPutResponsePayload struct{}

type LinkPutResponse = Response[LinkPutResponsePayload]

type LinkDeleteRequest struct {
	SourceId     string `json:"source_id"`
	Name         string `json:"name"`
	WitNamespace string `json:"wit_namespace"`
	WitPackage   string `json:"wit_package"`
}

type LinkDeleteResponsePayload struct{}

type LinkDeleteResponse = Response[LinkDeleteResponsePayload]

type ClaimsGetRequest struct{}

type ClaimsGetResponsePayload map[string]string

type ClaimsGetResponse = Response[ClaimsGetResponsePayload]

type HostInventoryRequest struct{}

type ComponentDescription struct {
	Id           string            `json:"id"`
	ImageRef     string            `json:"image_ref"`
	Name         string            `json:"name"`
	Annotations  map[string]string `json:"annotations"`
	Revision     int               `json:"revision"`
	MaxInstances int               `json:"max_instances"`
}

type ProviderDescription struct {
	Id          string            `json:"id"`
	ImageRef    string            `json:"image_ref"`
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
	Revision    int               `json:"revision"`
}

type HostInventoryResponsePayload struct {
	Components    []ComponentDescription `json:"components"`
	Providers     []ProviderDescription  `json:"providers"`
	HostId        string                 `json:"host_id"`
	FriendlyName  string                 `json:"friendly_name"`
	Labels        map[string]string      `json:"labels"`
	Version       string                 `json:"version"`
	UptimeHuman   string                 `json:"uptime_human"`
	UptimeSeconds int                    `json:"uptime_seconds"`
}

type HostInventoryResponse = Response[HostInventoryResponsePayload]

type HostPingRequest struct{}

type HostPingResponsePayload struct {
	Id            string            `json:"id"`
	Labels        map[string]string `json:"labels"`
	FriendlyName  string            `json:"friendly_name"`
	Version       string            `json:"version"`
	Lattice       string            `json:"lattice"`
	RpcHost       string            `json:"rpc_host"`
	CtlHost       string            `json:"ctl_host"`
	UptimeSeconds int               `json:"uptime_seconds"`
	UptimeHuman   string            `json:"uptime_human"`
}

type HostPingResponse = Response[HostPingResponsePayload]
