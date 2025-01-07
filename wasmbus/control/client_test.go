package control

import (
	"context"
	"testing"
	"time"

	"go.wasmcloud.dev/wasmbus"
	"go.wasmcloud.dev/wasmbus/wasmbustest"
)

func TestClient(t *testing.T) {
	nc, teardown := wasmbustest.WithWash(t)
	defer teardown(t)

	bus := wasmbus.NewNatsBus(nc)
	c := NewClient(bus, "default")

	t.Run("component", wrapTest(testComponent, c))
	t.Run("provider", wrapTest(testProvider, c))

}

func wrapTest(f func(*testing.T, *Client), c *Client) func(*testing.T) {
	return func(t *testing.T) {
		f(t, c)
	}
}

func testProvider(t *testing.T, c *Client) {
	req := &ProviderAuctionRequest{
		ProviderId:  "test-provider",
		ProviderRef: wasmbustest.ValidProvider,
		Constraints: make(map[string]string),
	}

	resp, err := c.ProviderAuction(context.TODO(), req)
	if err != nil {
		t.Fatalf("failed to auction: %v", err)
	}

	if !resp.Success {
		t.Fatalf("auction failed: %v", resp)
	}

	if resp.Response.HostId == "" {
		t.Fatalf("host id is empty")
	}

	reqStart := &ProviderStartRequest{
		HostId:      resp.Response.HostId,
		ProviderId:  req.ProviderId,
		ProviderRef: req.ProviderRef,
	}

	startResp, err := c.ProviderStart(context.TODO(), reqStart)
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	if !startResp.Success {
		t.Fatalf("start failed: %v", startResp)
	}
}

func testComponent(t *testing.T, c *Client) {
	// we first need an auction to find the host id

	auctionReq := &ComponentAuctionRequest{
		ComponentId:  "test-component",
		ComponentRef: wasmbustest.ValidComponent,
		Constraints:  make(map[string]string),
	}

	auctionResp, err := c.ComponentAuction(context.TODO(), auctionReq)
	if err != nil {
		t.Fatalf("failed to auction: %v", err)
	}

	scaleReq := &ScaleComponentRequest{
		HostId:       auctionResp.Response.HostId,
		ComponentId:  auctionReq.ComponentId,
		ComponentRef: auctionReq.ComponentRef,
		Count:        1,
	}

	scaleResp, err := c.ScaleComponent(context.TODO(), scaleReq)
	if err != nil {
		t.Fatalf("failed to scale: %v", err)
	}

	if !scaleResp.Success {
		t.Fatalf("scale failed: %v", scaleResp)
	}

	// NOTE(lxf): it takes time for the component to be ready
	// and the only way to know is to watch for lattice events.
	// For now, we'll try in a 10 sec loop :shrug:.
	t.Run("update", func(t *testing.T) {
		attempts := 10
		for i := 0; i < attempts; i++ {
			<-time.After(1 * time.Second)
			t.Logf("attempt %d/%d: trying to update component", i, attempts)

			updateReq := &UpdateComponentRequest{
				HostId:          auctionResp.Response.HostId,
				ComponentId:     auctionReq.ComponentId,
				NewComponentRef: auctionReq.ComponentRef,
				Annotations:     map[string]string{"test": "test"},
			}

			updateResp, err := c.UpdateComponent(context.TODO(), updateReq)
			if err == nil {
				if updateResp.Success {
					t.Logf("attempt %d/%d: update succeeded", i, attempts)
					return
				}
			}

			t.Logf("attempt %d/%d: failed to update: %v", i, attempts, err)
		}
		t.Fatalf("failed to update component after %d attempts", attempts)
	})
}
