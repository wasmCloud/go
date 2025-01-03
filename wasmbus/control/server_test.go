package control

import (
	"context"
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
	"go.wasmcloud.dev/wasmbus"
)

type fakeServer struct{}

func (s *fakeServer) ProviderAuction(ctx context.Context, req *ProviderAuctionRequest) (*ProviderAuctionResponse, error) {
	return nil, nil
}

func (s *fakeServer) ComponentAuction(ctx context.Context, req *ComponentAuctionRequest) (*ComponentAuctionResponse, error) {
	return nil, nil
}

func (s *fakeServer) ScaleComponent(ctx context.Context, req *ScaleComponentRequest) (*ScaleComponentResponse, error) {
	return nil, nil
}

func (s *fakeServer) UpdateComponent(ctx context.Context, req *UpdateComponentRequest) (*UpdateComponentResponse, error) {
	return nil, nil
}

func (s *fakeServer) ProviderStart(ctx context.Context, req *ProviderStartRequest) (*ProviderStartResponse, error) {
	return nil, nil
}

func (s *fakeServer) ProviderStop(ctx context.Context, req *ProviderStopRequest) (*ProviderStopResponse, error) {
	return nil, nil
}

func (s *fakeServer) HostStop(ctx context.Context, req *HostStopRequest) (*HostStopResponse, error) {
	return nil, nil
}

func (s *fakeServer) ConfigPut(ctx context.Context, req *ConfigPutRequest) (*ConfigPutResponse, error) {
	return nil, nil
}

func (s *fakeServer) ConfigGet(ctx context.Context, req *ConfigGetRequest) (*ConfigGetResponse, error) {
	fmt.Printf("ConfigGet: %v\n", req)
	return &ConfigGetResponse{
		Success: true,
	}, nil
}

func (s *fakeServer) ConfigDelete(ctx context.Context, req *ConfigDeleteRequest) (*ConfigDeleteResponse, error) {
	return nil, nil
}

func (s *fakeServer) HostLabelPut(ctx context.Context, req *HostLabelPutRequest) (*HostLabelPutResponse, error) {
	return nil, nil
}

func (s *fakeServer) HostLabelDelete(ctx context.Context, req *HostLabelDeleteRequest) (*HostLabelDeleteResponse, error) {
	return nil, nil
}

func (s *fakeServer) LinkGet(ctx context.Context, req *LinkGetRequest) (*LinkGetResponse, error) {
	return nil, nil
}

func (s *fakeServer) LinkPut(ctx context.Context, req *LinkPutRequest) (*LinkPutResponse, error) {
	return nil, nil
}

func (s *fakeServer) LinkDelete(ctx context.Context, req *LinkDeleteRequest) (*LinkDeleteResponse, error) {
	return nil, nil
}

func (s *fakeServer) ClaimsGet(ctx context.Context, req *ClaimsGetRequest) (*ClaimsGetResponse, error) {
	return nil, nil
}

func (s *fakeServer) HostInventory(ctx context.Context, req *HostInventoryRequest) (*HostInventoryResponse, error) {
	return nil, nil
}

func (s *fakeServer) HostPing(ctx context.Context, req *HostPingRequest) (*HostPingResponse, error) {
	return nil, nil
}

func TestServer(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatal(err)
	}

	bus := wasmbus.NewNatsBus(nc)
	s := NewServer(bus, "default", "host1", &fakeServer{})
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
	for err := range s.ErrorStream() {
		t.Logf("%+v", err)
	}
}
