package config

import (
	"fmt"

	"go.wasmcloud.dev/x/wasmbus"
)

type Server struct {
	*wasmbus.Server
	Lattice string
	api     API
}

func NewServer(bus wasmbus.Bus, lattice string, api API) *Server {
	return &Server{
		Server:  wasmbus.NewServer(bus),
		Lattice: lattice,
		api:     api,
	}
}

func (s *Server) Serve() error {
	subject := fmt.Sprintf("%s.%s.req", wasmbus.PrefixConfig, s.Lattice)
	handler := wasmbus.NewRequestHandler(HostRequest{}, HostResponse{}, s.api.Host)
	return s.RegisterHandler(subject, handler)
}
