package policy

import (
	"context"
	"fmt"
)

var (
	ErrProtocol = fmt.Errorf("encoding error")
	ErrInternal = fmt.Errorf("internal error")
)

type API interface {
	// PerformInvocation is called when a component is invoked
	PerformInvocation(ctx context.Context, req *PerformInvocationRequest) (*Response, error)
	// StartComponent is called when a component is started
	StartComponent(ctx context.Context, req *StartComponentRequest) (*Response, error)
	// StartProvider is called when a provider is started
	StartProvider(ctx context.Context, req *StartProviderRequest) (*Response, error)
}

var _ API = (*APIMock)(nil)

type APIMock struct {
	PerformInvocationFunc func(ctx context.Context, req *PerformInvocationRequest) (*Response, error)
	StartComponentFunc    func(ctx context.Context, req *StartComponentRequest) (*Response, error)
	StartProviderFunc     func(ctx context.Context, req *StartProviderRequest) (*Response, error)
}

func (m *APIMock) PerformInvocation(ctx context.Context, req *PerformInvocationRequest) (*Response, error) {
	return m.PerformInvocationFunc(ctx, req)
}

func (m *APIMock) StartComponent(ctx context.Context, req *StartComponentRequest) (*Response, error) {
	return m.StartComponentFunc(ctx, req)
}

func (m *APIMock) StartProvider(ctx context.Context, req *StartProviderRequest) (*Response, error) {
	return m.StartProviderFunc(ctx, req)
}

// Request is the structure of the request sent to the policy engine
type BaseRequest[T any] struct {
	Id      string `json:"requestId"`
	Kind    string `json:"kind"`
	Version string `json:"version"`
	Host    Host   `json:"host"`
	Request T      `json:"request"`
}

// Decision is a helper function to create a response
func (r BaseRequest[T]) Decision(allowed bool, msg string) *Response {
	return &Response{
		Id:        r.Id,
		Permitted: allowed,
		Message:   msg,
	}
}

// Deny is a helper function to create a response with a deny decision
func (r BaseRequest[T]) Deny(msg string) *Response {
	return r.Decision(false, msg)
}

// Allow is a helper function to create a response with an allow decision
func (r BaseRequest[T]) Allow(msg string) *Response {
	return r.Decision(true, msg)
}

// Response is the structure of the response sent by the policy engine
type Response struct {
	Id        string `json:"requestId"`
	Permitted bool   `json:"permitted"`
	Message   string `json:"message,omitempty"`
}

type Claims struct {
	PublicKey string `json:"publicKey"`
	Issuer    string `json:"issuer"`
	IssuedAt  int    `json:"issuedAt"`
	ExpiresAt int    `json:"expiresAt"`
	Expired   bool   `json:"expired"`
}

type StartComponentPayload struct {
	ComponentId  string            `json:"componentId"`
	ImageRef     string            `json:"imageRef"`
	MaxInstances int               `json:"maxInstances"`
	Annotations  map[string]string `json:"annotations"`
}

type StartComponentRequest = BaseRequest[StartComponentPayload]

type StartProviderPayload struct {
	ProviderId  string            `json:"providerId"`
	ImageRef    string            `json:"imageRef"`
	Annotations map[string]string `json:"annotations"`
}

type StartProviderRequest = BaseRequest[StartProviderPayload]

type PerformInvocationPayload struct {
	Interface string `json:"interface"`
	Function  string `json:"function"`
	// NOTE(lxf): this covers components but not providers. wut?!?
	Target InvocationTarget `json:"target"`
}

type PerformInvocationRequest = BaseRequest[PerformInvocationPayload]

type InvocationTarget struct {
	ComponentId  string            `json:"componentId"`
	ImageRef     string            `json:"imageRef"`
	MaxInstances int               `json:"maxInstances"`
	Annotations  map[string]string `json:"annotations"`
}

type Host struct {
	PublicKey string            `json:"publicKey"`
	Lattice   string            `json:"lattice"`
	Labels    map[string]string `json:"labels"`
}
