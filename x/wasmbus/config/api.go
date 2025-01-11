package config

import (
	"context"
	"fmt"
)

var (
	ErrProtocol = fmt.Errorf("encoding error")
	ErrInternal = fmt.Errorf("internal error")
)

type API interface {
	// Host is currently the only method exposed by the API.
	Host(ctx context.Context, req *HostRequest) (*HostResponse, error)
}

var _ API = (*APIMock)(nil)

type APIMock struct {
	HostFunc func(ctx context.Context, req *HostRequest) (*HostResponse, error)
}

func (m *APIMock) Host(ctx context.Context, req *HostRequest) (*HostResponse, error) {
	return m.HostFunc(ctx, req)
}

type HostRequest struct {
	Labels map[string]string `json:"labels"`
}

type HostResponse struct {
	RegistryCredentials map[string]RegistryCredential `json:"registryCredentials,omitempty"`
}

type RegistryCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
