package control

import (
	"context"
	"strings"
	"time"

	"go.wasmcloud.dev/x/wasmbus"
)

type Client struct {
	wasmbus.Bus
	lattice string
}

var _ APIv1 = (*Client)(nil)

// NewClient creates a new control client for a given lattice
func NewClient(bus wasmbus.Bus, lattice string) *Client {
	return &Client{
		Bus:     bus,
		lattice: lattice,
	}
}

func (c *Client) subject(ids ...string) string {
	parts := append([]string{wasmbus.PrefixCtlV1, c.lattice}, ids...)
	return strings.Join(parts, ".")
}

func (c *Client) ScaleComponent(ctx context.Context, req *ScaleComponentRequest) (*ScaleComponentResponse, error) {
	subject := c.subject("component", "scale", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ScaleComponentResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ConfigGet(ctx context.Context, req *ConfigGetRequest) (*ConfigGetResponse, error) {
	subject := c.subject("config", "get", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ConfigGetResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ProviderAuction(ctx context.Context, req *ProviderAuctionRequest) (*ProviderAuctionResponse, error) {
	subject := c.subject("provider", "auction")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ProviderAuctionResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ComponentAuction(ctx context.Context, req *ComponentAuctionRequest) (*ComponentAuctionResponse, error) {
	subject := c.subject("component", "auction")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ComponentAuctionResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) UpdateComponent(ctx context.Context, req *UpdateComponentRequest) (*UpdateComponentResponse, error) {
	subject := c.subject("component", "update", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, UpdateComponentResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ProviderStart(ctx context.Context, req *ProviderStartRequest) (*ProviderStartResponse, error) {
	subject := c.subject("provider", "start", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ProviderStartResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ProviderStop(ctx context.Context, req *ProviderStopRequest) (*ProviderStopResponse, error) {
	subject := c.subject("provider", "stop", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ProviderStopResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) HostStop(ctx context.Context, req *HostStopRequest) (*HostStopResponse, error) {
	subject := c.subject("host", "stop", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, HostStopResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ConfigPut(ctx context.Context, req *ConfigPutRequest) (*ConfigPutResponse, error) {
	subject := c.subject("config", "put", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req.Values, ConfigPutResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ConfigDelete(ctx context.Context, req *ConfigDeleteRequest) (*ConfigDeleteResponse, error) {
	subject := c.subject("config", "del", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ConfigDeleteResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) HostLabelPut(ctx context.Context, req *HostLabelPutRequest) (*HostLabelPutResponse, error) {
	subject := c.subject("label", "put", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, HostLabelPutResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) HostLabelDelete(ctx context.Context, req *HostLabelDeleteRequest) (*HostLabelDeleteResponse, error) {
	subject := c.subject("label", "del", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, HostLabelDeleteResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) LinkGet(ctx context.Context, req *LinkGetRequest) (*LinkGetResponse, error) {
	subject := c.subject("link", "get")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, LinkGetResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) LinkPut(ctx context.Context, req *LinkPutRequest) (*LinkPutResponse, error) {
	subject := c.subject("link", "put")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, LinkPutResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) LinkDelete(ctx context.Context, req *LinkDeleteRequest) (*LinkDeleteResponse, error) {
	subject := c.subject("link", "del")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, LinkDeleteResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ClaimsGet(ctx context.Context, req *ClaimsGetRequest) (*ClaimsGetResponse, error) {
	subject := c.subject("claims", "get")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ClaimsGetResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) HostInventory(ctx context.Context, req *HostInventoryRequest) (*HostInventoryResponse, error) {
	subject := c.subject("host", "get", req.HostId)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, HostInventoryResponse{})
	return wReq.Execute(ctx)
}

// NOTE(lxf): Why scatter/gather pattern? Why not just send a message to each host?
func (c *Client) HostPing(ctx context.Context, req *HostPingRequest) (*HostPingResponse, error) {
	reply := wasmbus.NewInbox()
	sub, err := c.Subscribe(reply, 10)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sub.Drain() }()

	resp := &HostPingResponse{
		Success: true,
	}
	var msgErrs []error
	go sub.Handle(func(msg *wasmbus.Message) {
		singleResp := &HostPingSingleResponse{}
		err := wasmbus.Decode(msg, singleResp)
		if err != nil {
			msgErrs = append(msgErrs, err)
			return
		}
		resp.Response = append(resp.Response, singleResp.Response)
	})

	subject := c.subject("host", "ping")
	pingRequest := wasmbus.NewMessage(subject)
	pingRequest.Reply = reply
	if err := c.Publish(pingRequest); err != nil {
		return nil, err
	}

	<-time.After(req.Wait)

	if err := sub.Drain(); err != nil {
		return nil, err
	}

	return resp, nil
}
