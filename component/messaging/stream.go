package messaging

import (
	"io"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// streamReadCloser adapts a component-model stream<u8> reader to
// io.ReadCloser, so a message body reads like any other Go stream.
type streamReadCloser struct {
	stream *witTypes.StreamReader[uint8]
	closed bool
}

func (r *streamReadCloser) Read(p []byte) (int, error) {
	if r.closed {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := r.stream.Read(p)
	if n == 0 {
		return 0, io.EOF
	}
	return int(n), nil
}

// Close releases the host-side stream. Closing before the body is exhausted
// discards the rest of it. Close is idempotent.
func (r *streamReadCloser) Close() error {
	if !r.closed {
		r.closed = true
		r.stream.Drop()
	}
	return nil
}
