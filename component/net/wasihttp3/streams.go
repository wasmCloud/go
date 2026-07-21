package wasihttp3

import (
	"fmt"
	"io"
	"net/http"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_3_0_types"
)

var _ io.ReadCloser = (*bodyReader)(nil)

type trailersFutureReader = witTypes.FutureReader[witTypes.Result[witTypes.Option[*types.Fields], types.ErrorCode]]

// bodyReader adapts a wasi:http body (a stream of bytes plus a future that
// resolves to optional trailers) into an [io.ReadCloser]. When the stream is
// exhausted the trailers future is read: an error there is surfaced from Read,
// otherwise any trailers are copied into trailer and Read returns [io.EOF].
type bodyReader struct {
	stream   *witTypes.StreamReader[uint8]
	trailers *trailersFutureReader
	// trailer is populated once the body has been fully read. Callers share
	// this map via http.Request.Trailer / http.Response.Trailer.
	trailer http.Header
}

func newBodyReader(stream *witTypes.StreamReader[uint8], trailers *trailersFutureReader) *bodyReader {
	return &bodyReader{
		stream:   stream,
		trailers: trailers,
		trailer:  http.Header{},
	}
}

func (r *bodyReader) Read(p []byte) (int, error) {
	if r.stream.WriterDropped() {
		return 0, r.finish()
	}

	count := r.stream.Read(p)
	if count == 0 && r.stream.WriterDropped() {
		return 0, r.finish()
	}

	return int(count), nil
}

// finish consumes the trailers future after the body stream has ended,
// returning the transport error if there was one and io.EOF otherwise.
func (r *bodyReader) finish() error {
	if r.trailers == nil {
		return io.EOF
	}

	result := r.trailers.Read()
	r.trailers = nil
	if result.IsErr() {
		return fmt.Errorf("failed to read from HTTP body stream: %s", errorCodeString(result.Err()))
	}

	if fields := result.Ok(); fields.IsSome() {
		trailers := fields.Some()
		toHTTPHeader(trailers.CopyAll(), &r.trailer)
		trailers.Drop()
	}

	return io.EOF
}

func (r *bodyReader) Close() error {
	if r.stream != nil {
		r.stream.Drop()
		r.stream = nil
	}
	if r.trailers != nil {
		r.trailers.Drop()
		r.trailers = nil
	}
	return nil
}
