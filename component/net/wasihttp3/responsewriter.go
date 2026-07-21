package wasihttp3

import (
	"fmt"
	"net/http"
	"slices"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_3_0_types"
)

var _ http.ResponseWriter = (*responseWriter)(nil)
var _ http.Flusher = (*responseWriter)(nil)

// responseWriter implements [http.ResponseWriter] over a wasi:http Response.
// The wasi Response is constructed and delivered on channel as soon as the
// handler first writes (or when the handler returns), so the body can stream
// to the client incrementally while the handler keeps writing.
type responseWriter struct {
	// channel on which the constructed wasi Response is delivered exactly once
	channel chan witTypes.Result[*types.Response, types.ErrorCode]
	// stream to which the response body is written after send
	stream *witTypes.StreamWriter[uint8]
	// future which resolves to an error if the body could not be delivered
	streamResult *witTypes.FutureReader[witTypes.Result[witTypes.Unit, types.ErrorCode]]
	// trailersTx resolves the response's trailers future
	trailersTx *witTypes.FutureWriter[witTypes.Result[witTypes.Option[*types.Fields], types.ErrorCode]]
	headers    http.Header
	statusCode int
}

func newResponseWriter() *responseWriter {
	return &responseWriter{
		channel:    make(chan witTypes.Result[*types.Response, types.ErrorCode]),
		headers:    http.Header{},
		statusCode: http.StatusOK,
	}
}

func (w *responseWriter) Header() http.Header {
	return w.headers
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *responseWriter) Write(buf []byte) (int, error) {
	if err := w.send(); err != nil {
		return 0, err
	}

	count := w.stream.Write(buf)
	if count == 0 && w.stream.ReaderDropped() {
		return 0, w.takeError()
	}

	return int(count), nil
}

func (w *responseWriter) Flush() {}

// send constructs the wasi Response from the accumulated headers and status
// code and delivers it on channel. It is a no-op after the first call.
func (w *responseWriter) send() error {
	channel := w.channel
	if channel == nil {
		return nil
	}
	w.channel = nil

	fields, err := toWASIHeaders(w.headers)
	if err != nil {
		// Keep the channel so a later send (or the error path in wasiHandle)
		// can still deliver a result; otherwise the export would block forever.
		w.channel = channel
		return err
	}

	tx, rx := types.MakeStreamU8()
	w.stream = tx

	trailersTx, trailersRx := types.MakeFutureResultOptionFieldsErrorCode()
	w.trailersTx = trailersTx

	response, sent := types.ResponseNew(fields, witTypes.Some(rx), trailersRx)
	w.streamResult = sent

	response.SetStatusCode(uint16(w.statusCode))

	channel <- witTypes.Ok[*types.Response, types.ErrorCode](response)

	return nil
}

// writeTrailers resolves the response's trailers future with any headers the
// handler declared via the "Trailer" header, ending the response body.
func (w *responseWriter) writeTrailers() {
	if w.trailersTx == nil {
		return
	}
	trailersTx := w.trailersTx
	w.trailersTx = nil

	declared := w.headers.Values("Trailer")
	collected := make(http.Header)
	for name, vals := range w.headers {
		if slices.Contains(declared, name) {
			collected[name] = vals
		}
	}

	if len(collected) == 0 {
		trailersTx.Write(witTypes.Ok[witTypes.Option[*types.Fields], types.ErrorCode](witTypes.None[*types.Fields]()))
		return
	}

	wasiTrailers, err := toWASIHeaders(collected)
	if err != nil {
		trailersTx.Write(witTypes.Err[witTypes.Option[*types.Fields]](
			types.MakeErrorCodeInternalError(witTypes.Some(fmt.Sprintf("cannot send trailers: %v", err))),
		))
		return
	}
	trailersTx.Write(witTypes.Ok[witTypes.Option[*types.Fields], types.ErrorCode](witTypes.Some(wasiTrailers)))
}

// takeError reads the body-delivery future after the client stopped reading.
func (w *responseWriter) takeError() error {
	if w.streamResult != nil {
		result := w.streamResult.Read()
		w.streamResult = nil
		if result.IsErr() {
			return fmt.Errorf("failed to write to HTTP body stream: %s", errorCodeString(result.Err()))
		}
	}
	return nil
}

func (w *responseWriter) close() {
	if w.stream != nil {
		w.stream.Drop()
		w.stream = nil
	}
	if w.streamResult != nil {
		w.streamResult.Drop()
		w.streamResult = nil
	}
	if w.trailersTx != nil {
		w.trailersTx.Drop()
		w.trailersTx = nil
	}
}
