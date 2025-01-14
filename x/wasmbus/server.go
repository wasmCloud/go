package wasmbus

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ServerError carries information about transport & encoding errors outside Request/Response scope.
type ServerError struct {
	Context context.Context
	Err     error
	Request *Message
}

// Server is a higher-level abstraction that can be used to register handlers for specific subjects.
// See `AnyServerHandler` for more information.
type Server struct {
	Bus
	// ContextFunc is a function that returns a new context for each message.
	// Defaults to `context.Background`.
	ContextFunc func() context.Context

	subscriptions []Subscription
	lock          sync.Mutex
	errorStream   chan *ServerError
}

// NewServer returns a new server instance.
func NewServer(bus Bus) *Server {
	return &Server{
		Bus:         bus,
		ContextFunc: func() context.Context { return context.Background() },
		errorStream: make(chan *ServerError),
	}
}

// ErrorStream returns a channel that can be used to listen for Transport / Encoding level errors.
// See `ServerError` for more information.
func (s *Server) ErrorStream() <-chan *ServerError {
	return s.errorStream
}

func (s *Server) reportError(ctx context.Context, req *Message, err error) {
	select {
	// We don't want to block the server if the error stream is full or nobody is listening
	case s.errorStream <- &ServerError{Context: ctx, Err: err, Request: req}:
	default:
	}
}

// Drain walks through all subscriptions and drains them.
// It also closes the error stream.
// This is a blocking operation.
func (s *Server) Drain() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var errs []error //nolint:prealloc
	for _, sub := range s.subscriptions {
		errs = append(errs, sub.Drain())
	}
	s.subscriptions = nil
	close(s.errorStream)

	return errors.Join(errs...)
}

// AnyServerHandler is an interface that can be implemented by any handler that can be registered with a server.
// Primary implementations are `RequestHandler` and `ServerHandlerFunc`.
type AnyServerHandler interface {
	HandleMessage(ctx context.Context, msg *Message) error
}

// ServerHandlerFunc is a function type that can be used to implement a server handler from a function.
type ServerHandlerFunc func(context.Context, *Message) error

func (f ServerHandlerFunc) HandleMessage(ctx context.Context, msg *Message) error {
	return f(ctx, msg)
}

// RegisterHandler registers a handler for a given subject.
// Each handler gets their channel subscription with no backlog, and their own goroutine for queue consumption.
// Callers should handle concurrency and synchronization themselves.
func (s *Server) RegisterHandler(subject string, handler AnyServerHandler) error {
	sub, err := s.Subscribe(subject, NoBackLog)
	if err != nil {
		return err
	}
	go sub.Handle(func(msg *Message) {
		ctx := s.ContextFunc()
		if err := handler.HandleMessage(ctx, msg); err != nil {
			s.reportError(ctx, msg, err)
		}
	})

	s.lock.Lock()
	defer s.lock.Unlock()
	s.subscriptions = append(s.subscriptions, sub)

	return nil
}

// NewRequestHandler returns a new server handler instance.
// The `T` and `Y` types are used to define the Request and Response types. Both should be structs. They will be used as template for request/responses.
func NewRequestHandler[T any, Y any](req T, resp Y, handler func(context.Context, *T) (*Y, error)) *RequestHandler[T, Y] {
	return &RequestHandler[T, Y]{
		Request:  req,
		Response: resp,
		Handler:  handler,
	}
}

// RequestHandler is a generic handler that can be used to implement a server handler.
// It encodes the logic for handling a message and sending a response.
type RequestHandler[T any, Y any] struct {
	Request     T
	Response    Y
	Decode      func(context.Context, *T, *Message) (context.Context, error)
	Encode      func(context.Context, string, *Y) (*Message, error)
	PreRequest  func(context.Context, *T, *Message) error
	PostRequest func(context.Context, *Y, *Message) error
	Handler     func(context.Context, *T) (*Y, error)
}

func (s *RequestHandler[T, Y]) decode(ctx context.Context, req *T, msg *Message) (context.Context, error) {
	if s.Decode != nil {
		return s.Decode(ctx, req, msg)
	}
	return ctx, Decode(msg, req)
}

func (s *RequestHandler[T, Y]) encode(ctx context.Context, subject string, resp *Y) (*Message, error) {
	if s.Encode != nil {
		return s.Encode(ctx, subject, resp)
	}
	return Encode(subject, resp)
}

// HandleMessage implements the `AnyServerHandler` interface.
func (s *RequestHandler[T, Y]) HandleMessage(ctx context.Context, msg *Message) error {
	var err error

	req := s.Request

	ctx, err = s.decode(ctx, &req, msg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrDecode, err)
	}

	if s.PreRequest != nil {
		if err := s.PreRequest(ctx, &req, msg); err != nil {
			return fmt.Errorf("%w: %s", ErrOperation, err)
		}
	}

	resp, err := s.Handler(ctx, &req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrOperation, err)
	}

	rawResp, err := s.encode(ctx, msg.Reply, resp)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEncode, err)
	}
	rawResp.bus = msg.bus

	if s.PostRequest != nil {
		if err := s.PostRequest(ctx, resp, rawResp); err != nil {
			return fmt.Errorf("%w: %s", ErrOperation, err)
		}
	}

	if err := msg.Bus().Publish(rawResp); err != nil {
		return fmt.Errorf("%w: %s", ErrTransport, err)
	}

	return nil
}

// TypedHandler is a higher-level abstraction that can be used to register handlers for specific types.
// It uses a `TypeExtractor` function to extract the type from the message.
// Usefull when you want to handle different types of messages with different handlers based on a json field inside the message.
type TypedHandler struct {
	extractor TypeExtractor
	handlers  map[string]AnyServerHandler
	lock      sync.Mutex
}

// TypeExtractor is a function that extracts a type name from a message.
type TypeExtractor func(ctx context.Context, msg *Message) (string, error)

// NewTypedHandler returns a new typed handler instance.
func NewTypedHandler(extractor TypeExtractor) *TypedHandler {
	return &TypedHandler{extractor: extractor, handlers: make(map[string]AnyServerHandler)}
}

// HandleMessage implements the `AnyServerHandler` interface.
func (h *TypedHandler) HandleMessage(ctx context.Context, msg *Message) error {
	if h.extractor == nil {
		return fmt.Errorf("%w: no type extractor", ErrOperation)
	}

	kind, err := h.extractor(ctx, msg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrOperation, err)
	}

	h.lock.Lock()
	handler, ok := h.handlers[kind]
	h.lock.Unlock()

	if !ok {
		return fmt.Errorf("%w: no handler for type %s", ErrOperation, kind)
	}

	return handler.HandleMessage(ctx, msg)
}

// RegisterType registers a handler for a given type.
// The handler will be called when a message with the given type is received, after the type is extracted by the `TypeExtractor`.
func (h *TypedHandler) RegisterType(kind string, handler AnyServerHandler) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if _, ok := h.handlers[kind]; ok {
		return fmt.Errorf("%w: handler for type %s already registered", ErrOperation, kind)
	}

	h.handlers[kind] = handler

	return nil
}
