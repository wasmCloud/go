package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nats-io/nkeys"
	"go.wasmcloud.dev/x/wasmbus"
)

type Server struct {
	*wasmbus.Server
	Name   string
	api    APIv1alpha1
	key    KeyPair
	pubKey string
}

type ServerOption func(*Server) error

func EphemeralKey() (KeyPair, error) {
	return nkeys.CreateCurveKeys()
}

func KeyPairFromSeed(seed []byte) (KeyPair, error) {
	return nkeys.FromSeed(seed)
}

type KeyPair = nkeys.KeyPair

func NewServer(bus wasmbus.Bus, name string, kp KeyPair, api APIv1alpha1) *Server {
	server := &Server{
		Server: wasmbus.NewServer(bus),
		Name:   name,
		api:    api,
		key:    kp,
	}

	return server
}

type secretContextKey string

const hostContextKey secretContextKey = "secret"

func (s *Server) decodeCiphered(ctx context.Context, req *GetRequest, msg *wasmbus.Message) (context.Context, error) {
	hostPubKey := msg.Header.Get(WasmCloudHostXkey)
	if hostPubKey == "" {
		return ctx, fmt.Errorf("%w: missing host public key", ErrInvalidHeaders)
	}

	decrypted, err := s.key.Open(msg.Data, hostPubKey)
	if err != nil {
		return ctx, fmt.Errorf("%w: %s", ErrDecryption, err)
	}

	if err := json.Unmarshal(decrypted, req); err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, hostContextKey, hostPubKey)
	return ctx, nil
}

func (s *Server) encodeCiphered(ctx context.Context, replyTo string, resp *GetResponse) (*wasmbus.Message, error) {
	hostPubKey := ctx.Value(hostContextKey).(string)

	responseKey, err := nkeys.CreateCurveKeys()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEncryption, err)
	}

	ephemeralPubKey, err := responseKey.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEncryption, err)
	}

	msg, err := wasmbus.Encode(replyTo, resp)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEncryption, err)
	}

	msg.Data, err = responseKey.Seal(msg.Data, hostPubKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrEncryption, err)
	}

	msg.Header.Add(WasmCloudResponseXkey, ephemeralPubKey)

	return msg, nil
}

func (s *Server) serveXkey(ctx context.Context, msg *wasmbus.Message) error {
	resp := wasmbus.NewMessage(msg.Reply)
	resp.Data = []byte(s.pubKey)
	return s.Publish(resp)
}

func (s *Server) Serve() error {
	var err error
	s.pubKey, err = s.key.PublicKey()
	if err != nil {
		return err
	}

	if err := s.RegisterHandler(s.subject("server_xkey"), wasmbus.ServerHandlerFunc(s.serveXkey)); err != nil {
		return err
	}

	get := wasmbus.NewRequestHandler(GetRequest{}, GetResponse{}, s.api.Get)
	get.Decode = s.decodeCiphered
	get.Encode = s.encodeCiphered
	return s.RegisterHandler(s.subject("get"), get)
}

func (s *Server) subject(ids ...string) string {
	parts := append([]string{wasmbus.PrefixSecrets, PrefixVersion, s.Name}, ids...)
	return strings.Join(parts, ".")
}
