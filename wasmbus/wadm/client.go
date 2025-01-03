package wadm

import (
	"context"
	"strings"

	"go.wasmcloud.dev/wasmbus"
)

var _ API = (*Client)(nil)

type Client struct {
	wasmbus.Bus
	lattice string
}

// NewClient creates a new wadm client, using the provided nats connection and lattice id (nats prefix)
func NewClient(bus wasmbus.Bus, lattice string) *Client {
	return &Client{
		Bus:     bus,
		lattice: lattice,
	}
}

func (c *Client) ModelStatus(ctx context.Context, req *ModelStatusRequest) (*ModelStatusResponse, error) {
	subject := c.subject("model", "status", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelStatusResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelPut(ctx context.Context, req *ModelPutRequest) (*ModelPutResponse, error) {
	subject := c.subject("model", "put")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelPutResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelGet(ctx context.Context, req *ModelGetRequest) (*ModelGetResponse, error) {
	subject := c.subject("model", "get", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelGetResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelVersions(ctx context.Context, req *ModelVersionsRequest) (*ModelVersionsResponse, error) {
	subject := c.subject("model", "versions", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelVersionsResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelDelete(ctx context.Context, req *ModelDeleteRequest) (*ModelDeleteResponse, error) {
	subject := c.subject("model", "del", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelDeleteResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelDeploy(ctx context.Context, req *ModelDeployRequest) (*ModelDeployResponse, error) {
	subject := c.subject("model", "deploy", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelDeployResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelUndeploy(ctx context.Context, req *ModelUndeployRequest) (*ModelUndeployResponse, error) {
	subject := c.subject("model", "undeploy", req.Name)
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelUndeployResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) ModelList(ctx context.Context, req *ModelListRequest) (*ModelListResponse, error) {
	subject := c.subject("model", "get")
	wReq := wasmbus.NewLatticeRequest(c.Bus, subject, req, ModelListResponse{})
	return wReq.Execute(ctx)
}

func (c *Client) subject(ids ...string) string {
	parts := append([]string{wasmbus.PrefixWadm, c.lattice}, ids...)
	return strings.Join(parts, ".")

}
