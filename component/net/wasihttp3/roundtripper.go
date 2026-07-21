package wasihttp3

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	client "go.wasmcloud.dev/component/imports/wasi_http_0_3_0_client"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_3_0_types"
)

// Transport implements [http.RoundTripper] for [wasi:http].
//
// [wasi:http]: https://github.com/WebAssembly/wasi-http/tree/v0.3.0
type Transport struct {
	ConnectTimeout time.Duration
}

var _ http.RoundTripper = (*Transport)(nil)

// DefaultTransport is the default implementation of [Transport] and is used by [DefaultClient].
// It is configured to use the same timeout value as [net/http.DefaultTransport].
var DefaultTransport = &Transport{
	ConnectTimeout: 30 * time.Second,
}

// DefaultClient is the default [net/http.Client] that uses [DefaultTransport] to adapt [net/http] to [wasi:http].
//
// [wasi:http]: https://github.com/WebAssembly/wasi-http/tree/v0.3.0
var DefaultClient = &http.Client{Transport: DefaultTransport}

func (t *Transport) requestOptions() witTypes.Option[*types.RequestOptions] {
	if t.ConnectTimeout <= 0 {
		return witTypes.None[*types.RequestOptions]()
	}
	options := types.MakeRequestOptions()
	// Both time.Duration and wasi:clocks duration are nanosecond counts.
	options.SetConnectTimeout(witTypes.Some(uint64(t.ConnectTimeout)))
	return witTypes.Some(options)
}

// RoundTrip implements the [net/http.RoundTripper] interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	request, err := t.toWASIRequest(req)
	if err != nil {
		return nil, err
	}

	result := client.Send(request)
	if result.IsErr() {
		return nil, fmt.Errorf("error sending request: %s", errorCodeString(result.Err()))
	}

	response := result.Ok()
	status := int(response.GetStatusCode())

	headerResource := response.GetHeaders()
	entries := headerResource.CopyAll()
	headerResource.Drop()

	stream, trailers := types.ResponseConsumeBody(response, unitFuture())
	body := newBodyReader(stream, trailers)

	resp := &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Request:    req,
		Header:     http.Header{},
		Body:       body,
		Trailer:    body.trailer,
	}
	toHTTPHeader(entries, &resp.Header)

	return resp, nil
}

// toWASIRequest converts an outbound [net/http.Request] into a wasi:http
// Request. The request body (if any) is streamed from a goroutine, followed by
// the request trailers, so the caller can send the request before the body has
// been fully produced.
func (t *Transport) toWASIRequest(req *http.Request) (*types.Request, error) {
	headers, err := toWASIHeaders(req.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to convert outgoing headers: %w", err)
	}

	trailersTx, trailersRx := types.MakeFutureResultOptionFieldsErrorCode()

	body := witTypes.None[*witTypes.StreamReader[uint8]]()
	if req.Body == nil {
		go writeTrailers(trailersTx, req.Trailer)
	} else {
		tx, rx := types.MakeStreamU8()
		body = witTypes.Some(rx)
		go func() {
			// For client requests, the Transport is responsible for closing the body.
			defer req.Body.Close()
			defer writeTrailers(trailersTx, req.Trailer)
			defer tx.Drop()

			buf := make([]uint8, 16*1024)
			for !tx.ReaderDropped() {
				n, err := req.Body.Read(buf)
				if n > 0 {
					tx.WriteAll(buf[:n])
				}
				if err != nil {
					if err != io.EOF {
						fmt.Fprintf(os.Stderr, "error reading request body: %v\n", err)
					}
					return
				}
			}
		}()
	}

	request, sent := types.RequestNew(headers, body, trailersRx, t.requestOptions())
	// TODO(#wasihttp3): surface body-delivery errors instead of dropping them.
	sent.Drop()

	request.SetMethod(toWASIMethod(req.Method))

	authority := req.Host
	if authority == "" {
		authority = req.URL.Host
	}
	request.SetAuthority(witTypes.Some(authority))
	request.SetPathWithQuery(witTypes.Some(req.URL.RequestURI()))

	switch req.URL.Scheme {
	case "http":
		request.SetScheme(witTypes.Some(types.MakeSchemeHttp()))
	case "https":
		request.SetScheme(witTypes.Some(types.MakeSchemeHttps()))
	default:
		request.SetScheme(witTypes.Some(types.MakeSchemeOther(req.URL.Scheme)))
	}

	return request, nil
}

// writeTrailers resolves an outbound request's trailers future once the body
// (if any) has been fully written.
func writeTrailers(
	tx *witTypes.FutureWriter[witTypes.Result[witTypes.Option[*types.Fields], types.ErrorCode]],
	trailer http.Header,
) {
	defer tx.Drop()

	if len(trailer) == 0 {
		tx.Write(witTypes.Ok[witTypes.Option[*types.Fields], types.ErrorCode](witTypes.None[*types.Fields]()))
		return
	}

	fields, err := toWASIHeaders(trailer)
	if err != nil {
		tx.Write(witTypes.Err[witTypes.Option[*types.Fields]](
			types.MakeErrorCodeInternalError(witTypes.Some(fmt.Sprintf("cannot send trailers: %v", err))),
		))
		return
	}
	tx.Write(witTypes.Ok[witTypes.Option[*types.Fields], types.ErrorCode](witTypes.Some(fields)))
}
