package control

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"go.wasmcloud.dev/wasmbus"
)

func TestClient(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatal(err)
	}

	bus := wasmbus.NewNatsBus(nc)
	client := NewClient(bus, "default")
	t.Log("Client created")

	resp, err := client.ConfigGet(context.Background(), &ConfigGetRequest{
		Name: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
