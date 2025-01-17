package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	PrefixVersion         = "v1alpha1"
	WasmCloudHostXkey     = "WasmCloud-Host-Xkey"
	WasmCloudResponseXkey = "Server-Response-Xkey"
)

type APIv1alpha1 interface {
	Get(ctx context.Context, req *GetRequest) (*GetResponse, error)
}

type APIMock struct {
	GetFunc func(ctx context.Context, req *GetRequest) (*GetResponse, error)
}

func (a *APIMock) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	return a.GetFunc(ctx, req)
}

var (
	ErrInvalidServerConfig = errors.New("invalid server configuration")

	ErrSecretNotFound = newResponseError("SecretNotFound", false)
	ErrInvalidRequest = newResponseError("InvalidRequest", false)
	ErrInvalidHeaders = newResponseError("InvalidHeaders", false)
	ErrInvalidPayload = newResponseError("InvalidPayload", false)
	ErrEncryption     = newResponseError("EncryptionError", false)
	ErrDecryption     = newResponseError("DecryptionError", false)

	ErrInvalidEntityJWT = newResponseError("InvalidEntityJWT", true)
	ErrInvalidHostJWT   = newResponseError("InvalidHostJWT", true)
	ErrUpstream         = newResponseError("UpstreamError", true)
	ErrPolicy           = newResponseError("PolicyError", true)
	ErrOther            = newResponseError("Other", true)
)

type Error struct {
	Tip        string
	HasMessage bool
	Message    string
}

func (re Error) With(msg string) *Error {
	otherError := re
	otherError.Message = msg
	return &otherError
}

func (re Error) Error() string {
	return re.Tip
}

func (re *Error) UnmarshalJSON(data []byte) error {
	serdeSpecial := make(map[string]string)
	if err := json.Unmarshal(data, &serdeSpecial); err != nil {
		var msg string
		if err := json.Unmarshal(data, &msg); err != nil {
			return err
		}
		*re = *ErrOther.With(msg)
		return nil
	}
	if len(serdeSpecial) != 1 {
		return errors.New("couldn't parse ResponseError")
	}
	for k, v := range serdeSpecial {
		*re = Error{Tip: k, HasMessage: v != "", Message: v}
		break
	}

	return nil
}

func (re *Error) MarshalJSON() ([]byte, error) {
	if re == nil {
		return nil, nil
	}

	if !re.HasMessage {
		return json.Marshal(re.Tip)
	}

	serdeSpecial := make(map[string]string)
	serdeSpecial[re.Tip] = re.Message

	return json.Marshal(serdeSpecial)
}

func newResponseError(tip string, hasMessage bool) *Error {
	return &Error{Tip: tip, HasMessage: hasMessage}
}

type applicationContextPolicy struct {
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties"`
}

func (a ApplicationContext) PolicyProperties() (json.RawMessage, error) {
	policy := &applicationContextPolicy{}
	err := json.Unmarshal([]byte(a.Policy), policy)
	return policy.Properties, err
}

type GetRequest struct {
	Key     string  `json:"key"`
	Field   string  `json:"field"`
	Version string  `json:"version,omitempty"`
	Context Context `json:"context"`
	// NOTE(lxf): HostPubKey is not part of the actual request.
	// filled in by middleware.
	HostPubKey string `json:"-"`
}

// NOTE(lxf): The way we return errors here is far from optimal...
type GetResponse struct {
	Secret *SecretValue `json:"secret,omitempty"`
	Error  *Error       `json:"error,omitempty"`
}

type SecretValue struct {
	Version      string       `json:"version,omitempty"`
	StringSecret string       `json:"string_secret,omitempty"`
	BinarySecret BinarySecret `json:"binary_secret,omitempty"`
}

// NOTE(lxf): This is a rust serde special...
type BinarySecret []uint8

func (u BinarySecret) MarshalJSON() ([]byte, error) {
	var result string
	if u == nil {
		return nil, nil
	}

	result = strings.Join(strings.Fields(fmt.Sprintf("%d", u)), ",")
	return []byte(result), nil
}

type Context struct {
	/// The application the entity belongs to.
	/// TODO: should this also be a JWT, but signed by the host?
	Application *ApplicationContext `json:"application,omitempty"`
	/// The component or provider's signed JWT.
	EntityJwt string `json:"entity_jwt"`
	/// The host's signed JWT.
	HostJwt string `json:"host_jwt"`
}

func (ctx Context) IsValid() *Error {
	if _, _, err := ctx.EntityCapabilities(); err != nil {
		return err
	}

	if _, _, err := ctx.HostCapabilities(); err != nil {
		return err
	}

	return nil
}

func (ctx Context) EntityCapabilities() (*WasCap, *ComponentClaims, *Error) {
	token, err := jwt.ParseWithClaims(ctx.EntityJwt, &WasCap{}, KeyPairFromIssuer())
	if err != nil {
		return nil, nil, ErrInvalidEntityJWT.With(err.Error())
	}

	wasCap, ok := token.Claims.(*WasCap)
	if !ok {
		return nil, nil, ErrInvalidEntityJWT.With("not wascap")
	}

	compCap := &ComponentClaims{}
	if err := json.Unmarshal(wasCap.Was, compCap); err != nil {
		return nil, nil, ErrInvalidEntityJWT.With(err.Error())
	}

	return wasCap, compCap, nil
}

func (ctx Context) HostCapabilities() (*WasCap, *HostClaims, *Error) {
	token, err := jwt.ParseWithClaims(ctx.HostJwt, &WasCap{}, KeyPairFromIssuer())
	if err != nil {
		return nil, nil, ErrInvalidHostJWT.With(err.Error())
	}

	wasCap, ok := token.Claims.(*WasCap)
	if !ok {
		return nil, nil, ErrInvalidHostJWT.With("not wascap")
	}

	hostCap := &HostClaims{}
	if err := json.Unmarshal(wasCap.Was, hostCap); err != nil {
		return nil, nil, ErrInvalidHostJWT.With(err.Error())
	}

	return wasCap, hostCap, nil
}

type ComponentClaims struct {
	jwt.RegisteredClaims

	/// A descriptive name for this component, should not include version information or public key
	Name string `json:"name"`
	/// A hash of the module's bytes as they exist without the embedded signature. This is stored so wascap
	/// can determine if a WebAssembly module's bytecode has been altered after it was signed
	ModuleHash string `json:"hash"`

	/// List of arbitrary string tags associated with the claims
	Tags []string `json:"tags"`

	/// Indicates a monotonically increasing revision number.  Optional.
	Rev int32 `json:"rev"`

	/// Indicates a human-friendly version string
	Ver string `json:"ver"`

	/// An optional, code-friendly alias that can be used instead of a public key or
	/// OCI reference for invocations
	CallAlias string `json:"call_alias"`

	/// Indicates whether this module is a capability provider
	Provider bool `json:"prov"`
}

type CapabilityProviderClaims struct {
	/// A descriptive name for the capability provider
	Name string `json:"name"`
	/// A human-readable string identifying the vendor of this provider (e.g. Redis or Cassandra or NATS etc)
	Vendor string `json:"vendor"`
	/// Indicates a monotonically increasing revision number.  Optional.
	Rev int32 `json:"rev"`
	/// Indicates a human-friendly version string. Optional.
	Ver string `json:"ver"`
	/// If the provider chooses, it can supply a JSON schma that describes its expected link configuration
	ConfigSchema json.RawMessage `json:"config_schema,omitempty"`
	/// The file hashes that correspond to the achitecture-OS target triples for this provider.
	TargetHashes map[string]string `json:"target_hashes"`
}

type HostClaims struct {
	/// Optional friendly descriptive name for the host
	Name string `json:"name"`
	/// Optional labels for the host
	Labels map[string]string `json:"labels"`
}

type WasCap struct {
	jwt.RegisteredClaims

	/// Custom jwt claims in the `wascap` namespace
	Was json.RawMessage `json:"wascap,omitempty"`

	/// Internal revision number used to aid in parsing and validating claims
	Revision int32 `json:"wascap_revision,omitempty"`
}

func (w WasCap) ParseCapability(dst interface{}) error {
	return json.Unmarshal(w.Was, dst)
}

type ApplicationContext struct {
	Policy string `json:"policy"`
	Name   string `json:"name"`
}
