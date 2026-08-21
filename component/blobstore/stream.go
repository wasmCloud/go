package blobstore

import (
	"io"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// streamReadCloser adapts a component-model stream<u8> reader to
// io.ReadCloser.
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

func (r *streamReadCloser) Close() error {
	if !r.closed {
		r.closed = true
		r.stream.Drop()
	}
	return nil
}
