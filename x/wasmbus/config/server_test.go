package config

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"go.wasmcloud.dev/x/wasmbus"
	"go.wasmcloud.dev/x/wasmbus/wasmbustest"
)

func TestServer(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("failed to connect to nats: %v", err)
	}
	bus := wasmbus.NewNatsBus(nc)
	s := NewServer(bus, "test", &APIMock{
		HostFunc: func(ctx context.Context, req *HostRequest) (*HostResponse, error) {
			return &HostResponse{
				RegistryCredentials: map[string]RegistryCredential{
					"docker.io": {
						Username: "my-username",
						Password: "hunter2",
					},
				},
			}, nil
		},
	})
	if err := s.Serve(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	req := wasmbus.NewMessage(fmt.Sprintf("%s.%s.req", wasmbus.PrefixConfig, "test"))
	req.Data = []byte(`{"labels":{"hostcore.arch":"aarch64","hostcore.os":"linux","hostcore.osfamily":"unix","kubernetes":"true","kubernetes.hostgroup":"default"}}`)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rawResp, err := bus.Request(ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	var resp HostResponse
	if err := wasmbus.Decode(rawResp, &resp); err != nil {
		t.Fatal(err)
	}

	docker, ok := resp.RegistryCredentials["docker.io"]
	if !ok {
		t.Fatalf("expected docker.io registry credentials")
	}
	if want, got := "my-username", docker.Username; want != got {
		t.Fatalf("expected username %q, got %q", want, got)
	}

	if want, got := "hunter2", docker.Password; want != got {
		t.Fatalf("expected password %q, got %q", want, got)
	}

	if err := s.Drain(); err != nil {
		t.Fatalf("failed to drain server: %v", err)
	}
}
