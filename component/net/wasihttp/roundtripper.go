package wasihttp

import (
	"fmt"
	"io"
	"net/http"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	outgoinghandler "go.wasmcloud.dev/component/imports/wasi_http_0_2_8_outgoing_handler"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_2_8_types"
)

// Transport implements [http.RoundTripper] for [wasi:http].
//
// [wasi:http]: https://github.com/WebAssembly/wasi-http
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
// [wasi:http]: https://github.com/WebAssembly/wasi-http
var DefaultClient = &http.Client{Transport: DefaultTransport}

func (r *Transport) requestOptions() *types.RequestOptions {
	options := types.MakeRequestOptions()
	if r.ConnectTimeout > 0 {
		// Go’s time.Duration is a nanosecond count, and WASI’s monotonic-clock duration is also a u64 of nanoseconds
		options.SetConnectTimeout(
			witTypes.Some(types.Duration(r.ConnectTimeout)),
		)
	} else {
		options.SetConnectTimeout(
			witTypes.None[types.Duration](),
		)
	}
	return options
}

// RoundTrip implements the [net/http.RoundTripper] interface.
func (r *Transport) RoundTrip(incomingRequest *http.Request) (*http.Response, error) {
	outHeaders := types.MakeFields()
	if err := HTTPtoWASIHeader(incomingRequest.Header, outHeaders); err != nil {
		return nil, fmt.Errorf("failed to convert outgoing headers: %w", err)
	}

	outRequest := types.MakeOutgoingRequest(outHeaders)

	outRequest.SetAuthority(witTypes.Some(incomingRequest.Host))
	outRequest.SetMethod(toWasiMethod(incomingRequest.Method))

	pathWithQuery := incomingRequest.URL.Path
	if incomingRequest.URL.RawQuery != "" {
		pathWithQuery = pathWithQuery + "?" + incomingRequest.URL.Query().Encode()
	}
	outRequest.SetPathWithQuery(witTypes.Some(pathWithQuery))

	switch incomingRequest.URL.Scheme {
	case "http":
		outRequest.SetScheme(witTypes.Some(types.MakeSchemeHttp()))
	case "https":
		outRequest.SetScheme(witTypes.Some(types.MakeSchemeHttps()))
	default:
		outRequest.SetScheme(witTypes.Some(types.MakeSchemeOther(incomingRequest.URL.Scheme)))
	}

	bodyResult := outRequest.Body()
	if bodyResult.IsErr() {
		return nil, fmt.Errorf("failed to acquire resource handle to request body")
	}
	body := bodyResult.Ok()

	handleResult := outgoinghandler.Handle(outRequest, witTypes.Some(r.requestOptions()))
	if handleResult.IsErr() {
		return nil, fmt.Errorf("failed to acquire handle to outbound request: %v", handleResult.Err())
	}
	futureResponse := handleResult.Ok()

	maybeTrailers := witTypes.None[*types.Fields]()
	if len(incomingRequest.Trailer) > 0 {
		outTrailers := types.MakeFields()
		if err := HTTPtoWASIHeader(incomingRequest.Trailer, outTrailers); err != nil {
			return nil, fmt.Errorf("failed to convert outgoing trailers: %w", err)
		}
		maybeTrailers = witTypes.Some(outTrailers)
	}

	// NOTE(lxf): If request includes a body, copy it to the adapted wasi body
	if incomingRequest.Body != nil {
		// For client requests, the Transport is responsible for calling Close on request's body.
		defer incomingRequest.Body.Close()
		adaptedBody, err := NewOutgoingBody(body)
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
		return nil, fmt.Errorf("failed to finish body: %v", outFinish.Err())
	}

	// wait until resp is returned
	pollable := futureResponse.Subscribe()
	pollable.Block()
	pollable.Drop()

	incomingResponseOuterOption := futureResponse.Get()
	if incomingResponseOuterOption.IsNone() {
		// NOTE: This should never happen since we subscribe to response readiness above
		return nil, fmt.Errorf("failed to wait for future-incoming-response readiness")
	}

	// Unwrap the outer Option and the outer Result within it
	outerResult := incomingResponseOuterOption.Some()
	if outerResult.IsErr() {
		return nil, fmt.Errorf("failed to unwrap the outer result for incoming-response")
	}

	// Unwrap the inner Result
	innerResult := outerResult.Ok()
	if innerResult.IsErr() {
		return nil, fmt.Errorf("failed to unwrap the inner result for incoming-response: %v", innerResult.Err())
	}
	incomingResponse := innerResult.Ok()

	incomingBody, incomingTrailers, err := NewIncomingBodyTrailer(incomingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse incoming-response: %w", err)
	}

	incomingHeaders := http.Header{}
	headers := incomingResponse.Headers()
	WASItoHTTPHeader(headers, &incomingHeaders)
	headers.Drop()

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
