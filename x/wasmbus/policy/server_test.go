package policy

import (
	"context"
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
	s := NewServer(bus, "subject.test", &APIMock{
		StartComponentFunc: func(ctx context.Context, req *StartComponentRequest) (*Response, error) {
			return req.Allow("passed"), nil
		},
		StartProviderFunc: func(ctx context.Context, req *StartProviderRequest) (*Response, error) {
			return req.Deny("denied"), nil
		},
		PerformInvocationFunc: func(ctx context.Context, req *PerformInvocationRequest) (*Response, error) {
			return req.Allow("passed"), nil
		},
	})
	if err := s.Serve(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	t.Run("startComponent", func(t *testing.T) {
		req := wasmbus.NewMessage("subject.test")
		req.Data = []byte(`{"requestId":"01945242-abec-71ee-f5e6-1f44eb61ad40","kind":"startComponent","version":"v1","request":{"componentId":"hello_world-http_component","imageRef":"ghcr.io/wasmcloud/components/http-hello-world-rust:0.1.0","maxInstances":1,"annotations":{"wasmcloud.dev/appspec":"hello-world","wasmcloud.dev/managed-by":"wadm","wasmcloud.dev/scaler":"a648fe966cbdb0a0dee3252a416f824858bbf0c1a24be850ef632626ddbb5133","wasmcloud.dev/spread_name":"default"},"claims":{"publicKey":"MBFFVNGFK3IA2ZXXG5DQXQNYM6TNG45PHJMJIJFVFI6YKS3XTXL3DRRK","issuer":"ADVIWF6Z3BFZNWUXJYT5NEAZZ2YX4T6NRKI3YOR3HKOSQQN7IVDGWSNO","issuedAt":"1714506509","expiresAt":null,"expired":false}},"host":{"publicKey":"NBLSJGGOETB677FQL63PWKDCVMOW4LXVI7S6WXSP55H7L5RNRKUDZKGE","lattice":"default","labels":{"hostcore.arch":"aarch64","hostcore.os":"linux","kubernetes":"true","kubernetes.hostgroup":"default","hostcore.osfamily":"unix"}}}`)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		rawResp, err := bus.Request(ctx, req)
		if err != nil {
			t.Fatal(err)
		}

		resp := &Response{}
		if err := wasmbus.Decode(rawResp, resp); err != nil {
			t.Fatal(err)
		}

		if want, got := "01945242-abec-71ee-f5e6-1f44eb61ad40", resp.Id; want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}

		if !resp.Permitted {
			t.Fatalf("expected allow, got deny")
		}
	})

	t.Run("startProvider", func(t *testing.T) {
		req := wasmbus.NewMessage("subject.test")
		req.Data = []byte(`{"requestId":"01945242-b574-5094-790d-76d823e0c948","kind":"startProvider","version":"v1","request":{"providerId":"hello_world-httpserver","imageRef":"ghcr.io/wasmcloud/http-server:0.23.0","annotations":{"wasmcloud.dev/appspec":"hello-world","wasmcloud.dev/managed-by":"wadm","wasmcloud.dev/scaler":"e82e835acda294f1a3d9cb66c0dfc8619c82fa836a1e30142d5d2b607357fc86","wasmcloud.dev/spread_name":"default"},"claims":{"publicKey":"VAG3QITQQ2ODAOWB5TTQSDJ53XK3SHBEIFNK4AYJ5RKAX2UNSCAPHA5M","issuer":"ACOJJN6WUP4ODD75XEBKKTCCUJJCY5ZKQ56XVKYK4BEJWGVAOOQHZMCW","issuedAt":"1725897949","expiresAt":null,"expired":false}},"host":{"publicKey":"NBLSJGGOETB677FQL63PWKDCVMOW4LXVI7S6WXSP55H7L5RNRKUDZKGE","lattice":"default","labels":{"hostcore.arch":"aarch64","hostcore.os":"linux","kubernetes":"true","kubernetes.hostgroup":"default","hostcore.osfamily":"unix"}}}`)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		rawResp, err := bus.Request(ctx, req)
		if err != nil {
			t.Fatal(err)
		}

		resp := &Response{}
		if err := wasmbus.Decode(rawResp, resp); err != nil {
			t.Fatal(err)
		}

		if want, got := "01945242-b574-5094-790d-76d823e0c948", resp.Id; want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}

		if resp.Permitted {
			t.Fatalf("expected deny, got allow")
		}
	})

	t.Run("performInvocation", func(t *testing.T) {
		req := wasmbus.NewMessage("subject.test")
		req.Data = []byte(`{"requestId":"01945244-9a84-a36c-a1b2-b722bd686bca","kind":"performInvocation","version":"v1","request":{"interface":"wrpc:http/incoming-handler@0.1.0","function":"handle","target":{"componentId":"hello_world-http_component","imageRef":"ghcr.io/wasmcloud/components/http-hello-world-rust:0.1.0","maxInstances":0,"annotations":{"wasmcloud.dev/appspec":"hello-world","wasmcloud.dev/managed-by":"wadm","wasmcloud.dev/scaler":"a648fe966cbdb0a0dee3252a416f824858bbf0c1a24be850ef632626ddbb5133","wasmcloud.dev/spread_name":"default"},"claims":{"publicKey":"MBFFVNGFK3IA2ZXXG5DQXQNYM6TNG45PHJMJIJFVFI6YKS3XTXL3DRRK","issuer":"ADVIWF6Z3BFZNWUXJYT5NEAZZ2YX4T6NRKI3YOR3HKOSQQN7IVDGWSNO","issuedAt":"1714506509","expiresAt":null,"expired":false}}},"host":{"publicKey":"NBLSJGGOETB677FQL63PWKDCVMOW4LXVI7S6WXSP55H7L5RNRKUDZKGE","lattice":"default","labels":{"hostcore.arch":"aarch64","hostcore.os":"linux","kubernetes":"true","kubernetes.hostgroup":"default","hostcore.osfamily":"unix"}}}`)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		rawResp, err := bus.Request(ctx, req)
		if err != nil {
			t.Fatal(err)
		}

		resp := &Response{}
		if err := wasmbus.Decode(rawResp, resp); err != nil {
			t.Fatal(err)
		}

		if want, got := "01945244-9a84-a36c-a1b2-b722bd686bca", resp.Id; want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}

		if !resp.Permitted {
			t.Fatalf("expected allow, got deny")
		}
	})

	if err := s.Drain(); err != nil {
		t.Fatalf("failed to drain server: %v", err)
	}
}
