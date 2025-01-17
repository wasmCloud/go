package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"go.wasmcloud.dev/x/wasmbus"
	"go.wasmcloud.dev/x/wasmbus/wasmbustest"
)

func keyPairForTest(t *testing.T) KeyPair {
	t.Helper()

	kp, err := EphemeralKey()
	if err != nil {
		t.Fatal(err)
	}

	return kp
}

type apiMock struct {
	APIMock
	t *testing.T
}

func TestServerXKey(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()

	kp := keyPairForTest(t)

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("failed to connect to nats: %v", err)
	}

	bus := wasmbus.NewNatsBus(nc)
	mock := &apiMock{}
	s := NewServer(bus, "test", kp, mock)
	if err := s.Serve(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	serverPubKey, err := kp.PublicKey()
	if err != nil {
		t.Fatal(err)
	}

	req := wasmbus.NewMessage(fmt.Sprintf("%s.%s.%s.server_xkey", wasmbus.PrefixSecrets, PrefixVersion, "test"))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rawResp, err := bus.Request(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := serverPubKey, string(rawResp.Data); want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestGet(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()

	kp := keyPairForTest(t)

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("failed to connect to nats: %v", err)
	}

	bus := wasmbus.NewNatsBus(nc)
	mock := &apiMock{}
	s := NewServer(bus, "test", kp, mock)
	if err := s.Serve(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	// log errors
	go func() {
		for c := range s.ErrorStream() {
			t.Log(c)
		}
	}()

	serverPubKey, err := kp.PublicKey()
	if err != nil {
		t.Fatal(err)
	}

	reqCtx := Context{
		Application: &ApplicationContext{
			Name: "appname",
		},
		EntityJwt: "eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJxdmVOakZjcW51dWhQaVJUMkU1YWJXIiwiaWF0IjoxNzIxODM0ODg5LCJpc3MiOiJBQk9HQjRXNURPWDNVTzNSVldXUUdZU01WWEhSUFFZWFZaUDVVNFZGTUpEQ1lDV0FSN1M1Q1lNTyIsInN1YiI6Ik1DNUNDNFVENUxQRFo0QzdaTkFFQTRPWlEzQkVGTFNWUTc0MlczVEVUM09OS1M0RFJCVk5NNUlDIiwid2FzY2FwIjp7Im5hbWUiOiJodHRwLWhlbGxvLXdvcmxkIiwiaGFzaCI6IkNFOTAxOTJDOTlDMEIyQzYwOEIyRTJDQjYxOUE5MjUxRkI2ODE4NTZDMTU2ODFCMUJDRDYyRUVEQTJENTEyOEUiLCJ0YWdzIjpbIndhc21jbG91ZC5jb20vZXhwZXJpbWVudGFsIl0sInJldiI6MCwidmVyIjoiMC4xLjAiLCJwcm92IjpmYWxzZX0sIndhc2NhcF9yZXZpc2lvbiI6M30.8awbkvrBnRKLpz88s7GXYCW0onpKf_nNfsj7pXhCyvq8pm4y2IotrIPCdBvWqDvDouX4VAM6DQQUHuI-VdKYAA",
		HostJwt:   "eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJuTGdta2Zud2p2Nkw1R28xSlNUdU0zIiwiaWF0IjoxNzIyMDE5OTk1LCJpc3MiOiJBQzNGU0IzT0VSQ1IzVU00WVNWUjJUQURFVlFWUTNITVpQQUtHS082QkNRSTRSNEFITFY2SVhSMiIsInN1YiI6Ik5ETlBUM0QzWVNUQzVKR0g2QVBKUDZBTVZYUVk2QklETVVXWkdTU1FXMjZWSjNINFBDRjJTU0ZSIiwid2FzY2FwIjp7Im5hbWUiOiJkZWxpY2F0ZS1icmVlemUtOTc4NSIsImxhYmVscyI6eyJzZWxmX3NpZ25lZCI6InRydWUifX0sIndhc2NhcF9yZXZpc2lvbiI6M30.5LM_GOpo-6qg0kDrIP_jswI_ZQfOILzHT-FHixvUeAf-1isamLg81S-rb84w6topfvevI6quyV3b-uHZt6q9BQ",
	}

	tt := []struct {
		name     string
		req      *GetRequest
		getFunc  func(context.Context, *GetRequest) (*GetResponse, error)
		validate func(*testing.T, *GetResponse)
	}{
		{
			name: "get string",
			req: &GetRequest{
				Key: "key",
			},
			getFunc: func(ctx context.Context, r *GetRequest) (*GetResponse, error) {
				if want, got := "key", r.Key; want != got {
					mock.t.Errorf("want %v, got %v", want, got)
				}
				return &GetResponse{
					Secret: &SecretValue{
						StringSecret: "hunter2",
					},
				}, nil
			},
			validate: func(t *testing.T, resp *GetResponse) {
				if want, got := "hunter2", resp.Secret.StringSecret; want != got {
					t.Errorf("want %v, got %v", want, got)
				}
			},
		},
		{
			name: "get binary",
			req: &GetRequest{
				Key: "keybin",
			},
			getFunc: func(ctx context.Context, r *GetRequest) (*GetResponse, error) {
				if want, got := "keybin", r.Key; want != got {
					mock.t.Errorf("want %v, got %v", want, got)
				}
				return &GetResponse{
					Secret: &SecretValue{
						BinarySecret: BinarySecret([]byte("hunter2")),
					},
				}, nil
			},
			validate: func(t *testing.T, resp *GetResponse) {
				if want, got := "hunter2", string(resp.Secret.BinarySecret); want != got {
					t.Errorf("want %v, got %v", want, got)
				}
			},
		},
		{
			name: "validate context",
			req: &GetRequest{
				Key:     "test",
				Context: reqCtx,
			},
			getFunc: func(ctx context.Context, r *GetRequest) (*GetResponse, error) {
				if err := r.Context.IsValid(); err != nil {
					mock.t.Errorf("context validation failed: %v", err)
				}
				return &GetResponse{}, nil
			},
			validate: func(t *testing.T, resp *GetResponse) {
			},
		},
		{
			name: "internal error",
			req: &GetRequest{
				Key: "test",
			},
			getFunc: func(ctx context.Context, r *GetRequest) (*GetResponse, error) {
				return &GetResponse{
					Error: ErrOther,
				}, nil
			},
			validate: func(t *testing.T, resp *GetResponse) {
				if want, got := "Other", resp.Error.Tip; want != got {
					t.Errorf("want %v, got %v", want, got)
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mock.t = t
			mock.GetFunc = tc.getFunc
			hostKey := keyPairForTest(t)
			hostPubKey, err := hostKey.PublicKey()
			if err != nil {
				t.Fatal(err)
			}

			rawSreq, err := wasmbus.EncodeMimetype(tc.req, "application/json")
			if err != nil {
				t.Fatal(err)
			}

			req := wasmbus.NewMessage(fmt.Sprintf("%s.%s.%s.get", wasmbus.PrefixSecrets, PrefixVersion, "test"))
			req.Header.Add(WasmCloudHostXkey, hostPubKey)
			req.Data, err = hostKey.Seal(rawSreq, serverPubKey)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			rawResp, err := bus.Request(ctx, req)
			if err != nil {
				t.Fatal(err)
			}
			decrypted, err := hostKey.Open(rawResp.Data, rawResp.Header.Get(WasmCloudResponseXkey))
			if err != nil {
				t.Fatal(err)
			}

			resp := &GetResponse{}
			if err := json.Unmarshal(decrypted, resp); err != nil {
				t.Fatal(err)
			}

			tc.validate(t, resp)
		})
	}

	if err := s.Drain(); err != nil {
		t.Fatalf("failed to drain server: %v", err)
	}
}

/*
	kp := keyPairForTest(t)
	hostPubKey, err := kp.PublicKey()
	if err != nil {
		t.Fatal(err)
	}

	reqCtx := Context{
		Application: &ApplicationContext{
			Name: "appname",
		},
		EntityJwt: "eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJxdmVOakZjcW51dWhQaVJUMkU1YWJXIiwiaWF0IjoxNzIxODM0ODg5LCJpc3MiOiJBQk9HQjRXNURPWDNVTzNSVldXUUdZU01WWEhSUFFZWFZaUDVVNFZGTUpEQ1lDV0FSN1M1Q1lNTyIsInN1YiI6Ik1DNUNDNFVENUxQRFo0QzdaTkFFQTRPWlEzQkVGTFNWUTc0MlczVEVUM09OS1M0RFJCVk5NNUlDIiwid2FzY2FwIjp7Im5hbWUiOiJodHRwLWhlbGxvLXdvcmxkIiwiaGFzaCI6IkNFOTAxOTJDOTlDMEIyQzYwOEIyRTJDQjYxOUE5MjUxRkI2ODE4NTZDMTU2ODFCMUJDRDYyRUVEQTJENTEyOEUiLCJ0YWdzIjpbIndhc21jbG91ZC5jb20vZXhwZXJpbWVudGFsIl0sInJldiI6MCwidmVyIjoiMC4xLjAiLCJwcm92IjpmYWxzZX0sIndhc2NhcF9yZXZpc2lvbiI6M30.8awbkvrBnRKLpz88s7GXYCW0onpKf_nNfsj7pXhCyvq8pm4y2IotrIPCdBvWqDvDouX4VAM6DQQUHuI-VdKYAA",
		HostJwt:   "eyJ0eXAiOiJqd3QiLCJhbGciOiJFZDI1NTE5In0.eyJqdGkiOiJuTGdta2Zud2p2Nkw1R28xSlNUdU0zIiwiaWF0IjoxNzIyMDE5OTk1LCJpc3MiOiJBQzNGU0IzT0VSQ1IzVU00WVNWUjJUQURFVlFWUTNITVpQQUtHS082QkNRSTRSNEFITFY2SVhSMiIsInN1YiI6Ik5ETlBUM0QzWVNUQzVKR0g2QVBKUDZBTVZYUVk2QklETVVXWkdTU1FXMjZWSjNINFBDRjJTU0ZSIiwid2FzY2FwIjp7Im5hbWUiOiJkZWxpY2F0ZS1icmVlemUtOTc4NSIsImxhYmVscyI6eyJzZWxmX3NpZ25lZCI6InRydWUifX0sIndhc2NhcF9yZXZpc2lvbiI6M30.5LM_GOpo-6qg0kDrIP_jswI_ZQfOILzHT-FHixvUeAf-1isamLg81S-rb84w6topfvevI6quyV3b-uHZt6q9BQ",
	}

	tests := map[string]struct {
		plainText     bool
		req           Request
		protocolError bool
		hostKey       string
		getFunc       func(ctx context.Context, r *Request) (*SecretValue, error)
		checkResponse func(*testing.T, Response)
	}{
		"blank": {
			plainText:     true,
			protocolError: true,
		},
		"happyPath": {
			req: Request{
				Key:     "secret",
				Context: reqCtx,
			},
		},
		"upstreamError": {
			req: Request{
				Key:     "secret",
				Context: reqCtx,
			},
			protocolError: true,
			getFunc: func(context.Context, *Request) (*SecretValue, error) {
				return nil, ErrUpstream.With("boom")
			},
			checkResponse: func(t *testing.T, resp Response) {
				if want, got := ErrUpstream.Error(), resp.Error.Error(); want != got {
					t.Errorf("want %v, got %v", want, got)
				}
			},
		},
		"badSecret": {
			req: Request{
				Key:     "secret",
				Context: reqCtx,
			},
			hostKey:       "badkey",
			protocolError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.getFunc != nil {
				handler.getFunc = test.getFunc
			} else {
				handler.getFunc = basicGetFunc
			}
			rawData, err := json.Marshal(&test.req)
			if err != nil {
				t.Fatal(err)
			}

			rawReq := nats.NewMsg(server.subjectMapper.SecretsSubject() + ".get")

			if !test.plainText {
				sealedData, err := kp.Seal(rawData, serverPubKey)
				if err != nil {
					t.Fatal(err)
				}

				rawReq.Data = sealedData
				hostKey := hostPubKey
				if test.hostKey != "" {
					hostKey = test.hostKey
				}
				rawReq.Header.Add(WasmCloudHostXkey, hostKey)
			}

			rawReply, err := nc.RequestMsg(rawReq, time.Second)
			if err != nil {
				t.Fatal(err)
			}

			var resp Response

			// the presence of the response header indicates if this is an encrypted response or not
			// plain responses are protocol errors
			responseKey := rawReply.Header.Get(WasmCloudResponseXkey)
			if test.protocolError {
				if responseKey != "" {
					t.Error("saw encryption header on protocol error")
				}

				if err := json.Unmarshal(rawReply.Data, &resp); err != nil {
					t.Fatal(err)
				}

				if resp.Error == nil {
					t.Fatal("Expected an error but got none")
				}

				if test.checkResponse != nil {
					test.checkResponse(t, resp)
				}
				return
			}

			if !test.protocolError && responseKey == "" {
				t.Error("missing encryption header")
			}

			rawResponse, err := kp.Open(rawReply.Data, responseKey)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(rawResponse, &resp); err != nil {
				t.Fatal(err)
			}

			if test.checkResponse != nil {
				test.checkResponse(t, resp)
			} else {
				basicCheckResponse(t, resp)
			}
		})
	}
}
*/
