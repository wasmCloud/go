package control

import (
	"context"
	"strings"

	"go.wasmcloud.dev/wasmbus"
)

type Client struct {
	wasmbus.Bus
	lattice string
}

//var _ APIv1 = (*Client)(nil)

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
