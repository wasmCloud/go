package wadm

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"go.wasmcloud.dev/wasmbus"
	"go.wasmcloud.dev/wasmbus/wasmbustest"
)

func TestServer(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("failed to connect to nats: %v", err)
	}
	bus := wasmbus.NewNatsBus(nc)
	s := NewServer(bus, "test", &APIMock{
		ModelListFunc: func(ctx context.Context, req *ModelListRequest) (*ModelListResponse, error) {
			return &ModelListResponse{
				BaseResponse: BaseResponse{
					Result:  GetResultSuccess,
					Message: "success",
				},
				Models: []ModelSummary{
					{
						Name:            "test",
						Version:         "abc",
						DeployedVersion: "xyz",
						Description:     "some app",
						DetailedStatus: &DetailedStatus{
							Info: StatusInfo{
								Type: StatusTypeDeployed,
							},
						},
					},
				},
			}, nil
		},
	})
	if err := s.Serve(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	c := NewClient(bus, "test")
	_, err = c.ModelList(context.Background(), &ModelListRequest{})
	if err != nil {
		t.Fatalf("failed to list models: %v", err)
	}

	if err := s.Drain(); err != nil {
		t.Fatalf("failed to drain server: %v", err)
	}
}
