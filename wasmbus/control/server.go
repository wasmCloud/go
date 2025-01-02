package control

import (
	"context"
	"strings"

	"go.wasmcloud.dev/wasmbus"
)

type Server struct {
	*wasmbus.Server
	Lattice  string
	HostId   string
	handlers map[string]wasmbus.AnyServerHandler
	api      APIv1
}

func NewServer(bus wasmbus.Bus, lattice string, hostId string, api APIv1) *Server {
	return &Server{
		Server:   wasmbus.NewServer(bus, lattice),
		Lattice:  lattice,
		HostId:   hostId,
		handlers: make(map[string]wasmbus.AnyServerHandler),
		api:      api,
	}
}

func (s *Server) Serve() error {
	providerAuction := wasmbus.NewRequestHandler(ProviderAuctionRequest{}, ProviderAuctionResponse{}, s.api.ProviderAuction)
	if err := s.RegisterHandler(s.subject("provider", "auction"), providerAuction); err != nil {
		return err
	}

	componentAuction := wasmbus.NewRequestHandler(ComponentAuctionRequest{}, ComponentAuctionResponse{}, s.api.ComponentAuction)
	if err := s.RegisterHandler(s.subject("component", "auction"), componentAuction); err != nil {
		return err
	}

	scaleComponent := wasmbus.NewRequestHandler(ScaleComponentRequest{}, ScaleComponentResponse{}, s.api.ScaleComponent)
	if err := s.RegisterHandler(s.subject("component", "scale", s.HostId), scaleComponent); err != nil {
		return err
	}

	updateComponent := wasmbus.NewRequestHandler(UpdateComponentRequest{}, UpdateComponentResponse{}, s.api.UpdateComponent)
	if err := s.RegisterHandler(s.subject("component", "scale", s.HostId), updateComponent); err != nil {
		return err
	}

	providerStart := wasmbus.NewRequestHandler(ProviderStartRequest{}, ProviderStartResponse{}, s.api.ProviderStart)
	if err := s.RegisterHandler(s.subject("provider", "start", s.HostId), providerStart); err != nil {
		return err
	}

	providerStop := wasmbus.NewRequestHandler(ProviderStopRequest{}, ProviderStopResponse{}, s.api.ProviderStop)
	if err := s.RegisterHandler(s.subject("provider", "stop", s.HostId), providerStop); err != nil {
		return err
	}

	hostStop := wasmbus.NewRequestHandler(HostStopRequest{}, HostStopResponse{}, s.api.HostStop)
	if err := s.RegisterHandler(s.subject("host", "stop", s.HostId), hostStop); err != nil {
		return err
	}

	configPut := wasmbus.NewRequestHandler(ConfigPutRequest{}, ConfigPutResponse{}, s.api.ConfigPut)
	configPut.PreRequest = func(_ context.Context, req *ConfigPutRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("config", "put", "*"), configPut); err != nil {
		return err
	}

	configGet := wasmbus.NewRequestHandler(ConfigGetRequest{}, ConfigGetResponse{}, s.api.ConfigGet)
	configGet.PreRequest = func(_ context.Context, req *ConfigGetRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("config", "get", "*"), configGet); err != nil {
		return err
	}

	configDelete := wasmbus.NewRequestHandler(ConfigDeleteRequest{}, ConfigDeleteResponse{}, s.api.ConfigDelete)
	configDelete.PreRequest = func(_ context.Context, req *ConfigDeleteRequest, msg *wasmbus.Message) error {
		req.Name = msg.LastSubjectPart()
		return nil
	}
	if err := s.RegisterHandler(s.subject("config", "del", "*"), configDelete); err != nil {
		return err
	}

	hostLabelPut := wasmbus.NewRequestHandler(HostLabelPutRequest{}, HostLabelPutResponse{}, s.api.HostLabelPut)
	if err := s.RegisterHandler(s.subject("host", "label", "put", s.HostId), hostLabelPut); err != nil {
		return err
	}

	hostLabelDelete := wasmbus.NewRequestHandler(HostLabelDeleteRequest{}, HostLabelDeleteResponse{}, s.api.HostLabelDelete)
	if err := s.RegisterHandler(s.subject("host", "label", "delete", s.HostId), hostLabelDelete); err != nil {
		return err
	}

	linkGet := wasmbus.NewRequestHandler(LinkGetRequest{}, LinkGetResponse{}, s.api.LinkGet)
	if err := s.RegisterHandler(s.subject("link", "get"), linkGet); err != nil {
		return err
	}

	linkPut := wasmbus.NewRequestHandler(LinkPutRequest{}, LinkPutResponse{}, s.api.LinkPut)
	if err := s.RegisterHandler(s.subject("link", "put"), linkPut); err != nil {
		return err
	}

	linkDelete := wasmbus.NewRequestHandler(LinkDeleteRequest{}, LinkDeleteResponse{}, s.api.LinkDelete)
	if err := s.RegisterHandler(s.subject("link", "delete"), linkDelete); err != nil {
		return err
	}

	claimsGet := wasmbus.NewRequestHandler(ClaimsGetRequest{}, ClaimsGetResponse{}, s.api.ClaimsGet)
	if err := s.RegisterHandler(s.subject("claims", "get"), claimsGet); err != nil {
		return err
	}

	hostInventory := wasmbus.NewRequestHandler(HostInventoryRequest{}, HostInventoryResponse{}, s.api.HostInventory)
	if err := s.RegisterHandler(s.subject("host", "get", s.HostId), hostInventory); err != nil {
		return err
	}

	hostPing := wasmbus.NewRequestHandler(HostPingRequest{}, HostPingResponse{}, s.api.HostPing)
	if err := s.RegisterHandler(s.subject("host", "ping"), hostPing); err != nil {
		return err
	}

	return nil
}

func (s *Server) subject(ids ...string) string {
	parts := append([]string{wasmbus.PrefixCtlV1, s.Lattice}, ids...)
	return strings.Join(parts, ".")

}
