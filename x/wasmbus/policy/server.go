package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"go.wasmcloud.dev/x/wasmbus"
)

type Server struct {
	*wasmbus.Server
	subject  string
	api      API
	handlers map[string]wasmbus.AnyServerHandler
}

func NewServer(bus wasmbus.Bus, subject string, api API) *Server {
	return &Server{
		Server:  wasmbus.NewServer(bus),
		subject: subject,
		api:     api,
	}
}

func (s *Server) Serve() error {
	handler := wasmbus.NewTypedHandler(extractType)

	startComponent := wasmbus.NewRequestHandler(StartComponentRequest{}, Response{}, s.api.StartComponent)
	if err := handler.RegisterType("startComponent", startComponent); err != nil {
		return err
	}

	startProvider := wasmbus.NewRequestHandler(StartProviderRequest{}, Response{}, s.api.StartProvider)
	if err := handler.RegisterType("startProvider", startProvider); err != nil {
		return err
	}

	performInvocation := wasmbus.NewRequestHandler(PerformInvocationRequest{}, Response{}, s.api.PerformInvocation)
	if err := handler.RegisterType("performInvocation", performInvocation); err != nil {
		return err
	}

	return s.RegisterHandler(s.subject, handler)
}

func extractType(ctx context.Context, msg *wasmbus.Message) (string, error) {
	var baseReq BaseRequest[json.RawMessage]

	if err := wasmbus.Decode(msg, &baseReq); err != nil {
		return "", err
	}

	switch baseReq.Kind {
	case "startComponent", "startProvider", "performInvocation":
		return baseReq.Kind, nil
	default:
		return "", fmt.Errorf("unknown request kind: %s", baseReq.Kind)
	}
}
