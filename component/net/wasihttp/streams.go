package wasihttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_2_8_types"
	streams "go.wasmcloud.dev/component/imports/wasi_io_0_2_8_streams"
)

// BodyConsumer interface is implemented by [types.IncomingRequest] and [types.IncomingResponse].
// It enables the consumption of [wasi:http/types.incoming-request] and [wasi:http/types.incoming-response]
//
// [wasi:http/types.incoming-request]: https://github.com/WebAssembly/wasi-http/blob/main/wit/types.wit
// [wasi:http/types.incoming-response]: https://github.com/WebAssembly/wasi-http/blob/main/wit/types.wit
type BodyConsumer interface {
	Consume() witTypes.Result[*types.IncomingBody, witTypes.Unit]
	Headers() *types.Fields
}

type inputStreamReader struct {
	consumer    BodyConsumer
	body        *types.IncomingBody
	stream      *streams.InputStream
	trailerLock sync.Mutex
	trailers    http.Header
	trailerOnce sync.Once
}

func (r *inputStreamReader) Close() error {
	r.trailerOnce.Do(r.parseTrailers)

	if r.stream != nil {
		r.stream.Drop()
	}

	if r.body != nil {
		r.body.Drop()
		r.body = nil
	}

	return nil
}

func (r *inputStreamReader) parseTrailers() {
	r.trailerLock.Lock()
	defer r.trailerLock.Unlock()

	// if we got this far, then we release ownership from body, otherwise it is our responsibility to drop it
	r.stream.Drop()
	r.stream = nil

	futureTrailers := types.IncomingBodyFinish(r.body)
	defer futureTrailers.Drop()

	trailersResult := futureTrailers.Get()
	r.body = nil

	// unroll the future
	if trailersResult.IsNone() {
		return
	}
	if trailersResult.Some().IsErr() {
		return
	}
	if trailersResult.Some().Ok().IsErr() {
		return
	}
	maybeWasiTrailers := trailersResult.Some().Ok().Ok()

	if maybeWasiTrailers.IsNone() {
		return
	}

	wasiTrailers := maybeWasiTrailers.Some()
	for _, kv := range wasiTrailers.Entries() {
		r.trailers.Add(kv.F0, string(kv.F1))
	}

	wasiTrailers.Drop()
}

func (r *inputStreamReader) Read(p []byte) (n int, err error) {
	pollable := r.stream.Subscribe()
	pollable.Block()
	pollable.Drop()

	readResult := r.stream.Read(uint64(len(p)))
	if readResult.IsErr() {
		streamErr := readResult.Err()
		if streamErr.Tag() == streams.StreamErrorClosed {
			r.trailerOnce.Do(r.parseTrailers)
			return 0, io.EOF
		}
		return 0, fmt.Errorf("failed to read from InputStream %s", streamErr.LastOperationFailed().ToDebugString())
	}

	contents := readResult.Ok()
	copy(p, contents)
	return len(contents), nil
}

// NewIncomingBodyTrailer takes a [BodyConsumer] and parses it into corresponding [io.ReadCloser] and [net/http.Header].
func NewIncomingBodyTrailer(consumer BodyConsumer) (io.ReadCloser, http.Header, error) {
	consumeResult := consumer.Consume()
	if consumeResult.IsErr() {
		return nil, nil, errors.New("failed to consume incoming request")
	}

	body := consumeResult.Ok()
	streamResult := body.Stream()
	if streamResult.IsErr() {
		return nil, nil, errors.New("failed to consume incoming request body stream")
	}

	stream := streamResult.Ok()

	trailers := http.Header{}
	return &inputStreamReader{
		consumer: consumer,
		trailers: trailers,
		body:     body,
		stream:   stream,
	}, trailers, nil
}

type outgoingBody struct {
	body   *types.OutgoingBody
	stream *streams.OutputStream
}

// NewOutgoingBody takes a [types.OutgoingBody] and returns a [io.WriteCloser] encapsulating it.
func NewOutgoingBody(body *types.OutgoingBody) (io.WriteCloser, error) {
	stream := body.Write()
	if stream.IsErr() {
		return nil, errors.New("failed to acquire resource handle to request body")
	}
	return &outgoingBody{
		body:   body,
		stream: stream.Ok(),
	}, nil
}

func (r *outgoingBody) Close() error {
	r.stream.Drop()
	return nil
}

func (r *outgoingBody) Write(p []byte) (n int, err error) {
	// Split the input into 4096-byte chunks to avoid exceeding stream buffer limits
	const chunkSize = 4096
	totalWritten := 0

	for offset := 0; offset < len(p); offset += chunkSize {
		end := min(offset+chunkSize, len(p))

		writeResult := r.stream.BlockingWriteAndFlush(p[offset:end])
		if writeResult.IsErr() {
			streamErr := writeResult.Err()
			if streamErr.Tag() == streams.StreamErrorClosed {
				return totalWritten, io.EOF
			}
			return totalWritten, fmt.Errorf("failed to write to response body's stream: %s", streamErr.LastOperationFailed().ToDebugString())
		}

		totalWritten += end - offset
	}
	return totalWritten, nil
}
