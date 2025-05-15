package wasihttp

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"go.bytecodealliance.org/cm"
	monotonicclock "go.wasmcloud.dev/component/gen/wasi/clocks/monotonic-clock"
	outgoinghandler "go.wasmcloud.dev/component/gen/wasi/http/outgoing-handler"
	"go.wasmcloud.dev/component/gen/wasi/http/types"
	poll "go.wasmcloud.dev/component/poll"
)

// Transport implements [http.RoundTripper] for [wasi:http].
//
// [wasi:http]: https://github.com/WebAssembly/wasi-http/tree/v0.2.0
type Transport struct {
	ConnectTimeout time.Duration
}

var _ http.RoundTripper = (*Transport)(nil)

// DefaultTransport is the default implementation of [Transport] and is used by [DefaultClient].
// It is configured use the same timeout value as [net/http.DefaultTransport].
var DefaultTransport = &Transport{
	ConnectTimeout: 30 * time.Second, // NOTE(lxf): Same as stdlib http.Transport
}

// DefaultClient is the default [net/http.Client] that uses [DefaultTransport] to adapt [net/http] to [wasi:http].
//
// [wasi:http]: https://github.com/WebAssembly/wasi-http/tree/v0.2.0
var DefaultClient = &http.Client{Transport: DefaultTransport}

func (r *Transport) requestOptions() types.RequestOptions {
	options := types.NewRequestOptions()
	if r.ConnectTimeout > 0 {
		// Go’s time.Duration is a nanosecond count, and WASI’s monotonicclock.Duration is also a u64 of nanoseconds
		options.SetConnectTimeout(
			cm.Some(monotonicclock.Duration(r.ConnectTimeout)),
		)
	} else {
		options.SetConnectTimeout(
			cm.None[monotonicclock.Duration](),
		)
	}
	return options
}

// RoundTrip implements the [net/http.RoundTripper] interface.
func (r *Transport) RoundTrip(incomingRequest *http.Request) (*http.Response, error) {
	var err error

	outHeaders := types.NewFields()
	if err := HTTPtoWASIHeader(incomingRequest.Header, outHeaders); err != nil {
		return nil, fmt.Errorf("failed to convert outgoing headers: %w", err)
	}

	outRequest := types.NewOutgoingRequest(outHeaders)

	outRequest.SetAuthority(cm.Some(incomingRequest.Host))
	outRequest.SetMethod(toWasiMethod(incomingRequest.Method))

	pathWithQuery := incomingRequest.URL.Path
	if incomingRequest.URL.RawQuery != "" {
		pathWithQuery = pathWithQuery + "?" + incomingRequest.URL.Query().Encode()
	}
	outRequest.SetPathWithQuery(cm.Some(pathWithQuery))

	switch incomingRequest.URL.Scheme {
	case "http":
		outRequest.SetScheme(cm.Some(types.SchemeHTTP()))
	case "https":
		outRequest.SetScheme(cm.Some(types.SchemeHTTPS()))
	default:
		outRequest.SetScheme(cm.Some(types.SchemeOther(incomingRequest.URL.Scheme)))
	}

	body, bodyErr, isErr := outRequest.Body().Result()
	if isErr {
		return nil, fmt.Errorf("failed to acquire resource handle to request body: %s", bodyErr)
	}

	futureResponse, handlerErr, isErr := outgoinghandler.Handle(outRequest, cm.Some(r.requestOptions())).Result()
	if isErr {
		return nil, fmt.Errorf("failed to acquire handle to outbound request: %s", handlerErr)
	}

	maybeTrailers := cm.None[types.Fields]()
	if len(incomingRequest.Trailer) > 0 {
		outTrailers := types.NewFields()
		if err := HTTPtoWASIHeader(incomingRequest.Trailer, outTrailers); err != nil {
			return nil, fmt.Errorf("failed to convert outgoing trailers: %w", err)
		}
		maybeTrailers = cm.Some(outTrailers)
	}

	// NOTE(lxf): If request includes a body, copy it to the adapted wasi body
	if incomingRequest.Body != nil {
		// For client requests, the Transport is responsible for calling Close on request's body.
		defer incomingRequest.Body.Close()
		adaptedBody, err := NewOutgoingBody(&body)
		if err != nil {
			return nil, fmt.Errorf("failed to adapt body: %w", err)
		}
		if _, err := io.Copy(adaptedBody, incomingRequest.Body); err != nil {
			return nil, fmt.Errorf("failed to copy body: %w", err)
		}
		if err := adaptedBody.Close(); err != nil {
			return nil, fmt.Errorf("failed to close body: %w", err)
		}
	}

	// From `outgoing-body` documentation:
	// Finalize an outgoing body, optionally providing trailers. This must be
	// called to signal that the response is complete.
	outFinish := types.OutgoingBodyFinish(body, maybeTrailers)
	if outFinish.IsErr() {
		return nil, fmt.Errorf("failed to finish body: %s", outFinish.Err())
	}

	// wait until resp is returned
	poll.PollWithBackoff(futureResponse.Subscribe())

	incomingResponseOuterOption := futureResponse.Get()
	if incomingResponseOuterOption.None() {
		// NOTE: This should never happen since we subscribe to response readiness above
		return nil, fmt.Errorf("failed to wait for future-incoming-response readiness")
	}

	// Unwrap the outer Option and the outer Result within it
	innerResult, outerResultErr, isErr := incomingResponseOuterOption.Some().Result()
	if isErr {
		return nil, fmt.Errorf("failed to unwrap the outer result for incoming-response: %s", outerResultErr)
	}

	// Unwrap the inner Result
	incomingResponse, innerResultErr, isErr := innerResult.Result()
	if isErr {
		return nil, fmt.Errorf("failed to unwrap the inner result for incoming-response: %s", innerResultErr)
	}

	incomingBody, incomingTrailers, err := NewIncomingBodyTrailer(incomingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse incoming-response: %w", err)
	}

	incomingHeaders := http.Header{}
	headers := incomingResponse.Headers()
	WASItoHTTPHeader(headers, &incomingHeaders)
	headers.ResourceDrop()

	resp := &http.Response{
		StatusCode: int(incomingResponse.Status()),
		Status:     http.StatusText(int(incomingResponse.Status())),
		Request:    incomingRequest,
		Header:     incomingHeaders,
		Body:       incomingBody,
		Trailer:    incomingTrailers,
	}

	return resp, nil
}
