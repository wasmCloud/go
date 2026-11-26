// Package streamio adapts a component-model `stream<u8>`, together with the
// `future<result<T, string>>` its writer resolves when it is finished, to
// [io.Reader] and [io.Writer].
//
// Everything past this package is ordinary Go — io.Copy, bufio, sha256 — and
// none of it ever holds more than one buffer of the data.
//
// The future is the part that is easy to leave out and expensive to miss. A
// stream ending only means its writer stopped writing; it does not say whether
// the writer finished or gave up. [Reader] therefore reports [io.EOF] only once
// the stream has ended *and* the future resolved `ok`. An `err` there becomes
// the error from Read, so io.Copy and io.ReadAll tell a complete transfer from
// a fragment without any extra bookkeeping.
package streamio

import (
	"errors"
	"io"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// Outcome is what a stream's writer reports once it is done: ok with a value
// (often [witTypes.Unit]), or an error message.
type Outcome[T any] = witTypes.Result[T, string]

// ErrReaderGone is returned by [Writer.Write] when the reading side dropped its
// end of the stream before taking everything written to it.
var ErrReaderGone = errors.New("streamio: reader dropped the stream")

// RemoteError is an error the writer on the other end of a stream reported
// through its future.
type RemoteError struct{ Message string }

func (e *RemoteError) Error() string { return e.Message }

// Reader is an [io.ReadCloser] over the readable end of a stream and the
// future that reports how it ended.
type Reader[T any] struct {
	stream *witTypes.StreamReader[uint8]
	done   *witTypes.FutureReader[Outcome[T]]
	value  T
	err    error
}

// NewReader takes ownership of both handles. Close it to release them early;
// otherwise reading to the end releases them.
func NewReader[T any](stream *witTypes.StreamReader[uint8], done *witTypes.FutureReader[Outcome[T]]) *Reader[T] {
	return &Reader[T]{stream: stream, done: done}
}

// Read reads up to len(p) bytes straight into p: the host copies the writer's
// bytes into this buffer, so nothing is allocated per read.
//
// At the end of the stream it reads the future and returns [io.EOF] if the
// writer reported ok, or a [*RemoteError] if it reported an error.
func (r *Reader[T]) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	for !r.stream.WriterDropped() {
		if n := r.stream.Read(p); n > 0 {
			return int(n), nil
		}
	}
	r.err = r.finish()
	return 0, r.err
}

func (r *Reader[T]) finish() error {
	r.stream.Drop()
	r.stream = nil
	outcome := r.done.Read() // Read releases the handle itself.
	r.done = nil
	if outcome.IsErr() {
		return &RemoteError{Message: outcome.Err()}
	}
	r.value = outcome.Ok()
	return io.EOF
}

// Value is what the writer resolved the future with. It is meaningful only
// after Read has returned [io.EOF].
func (r *Reader[T]) Value() T { return r.value }

// Close drops whatever the reader still holds. Dropping the stream before its
// end tells the writer to stop: its next write fails with [ErrReaderGone].
func (r *Reader[T]) Close() error {
	if r.stream != nil {
		r.stream.Drop()
		r.stream = nil
	}
	if r.done != nil {
		r.done.Drop()
		r.done = nil
	}
	if r.err == nil {
		r.err = io.ErrClosedPipe
	}
	return nil
}

// Writer is an [io.Writer] over the writable end of a stream and the future
// that will report how it ended.
//
// The future must be resolved exactly once, so every path through the code
// that owns a Writer has to end in [Writer.Close] or [Writer.CloseWith].
type Writer[T any] struct {
	stream *witTypes.StreamWriter[uint8]
	done   *witTypes.FutureWriter[Outcome[T]]
}

// NewWriter takes ownership of both handles.
func NewWriter[T any](stream *witTypes.StreamWriter[uint8], done *witTypes.FutureWriter[Outcome[T]]) *Writer[T] {
	return &Writer[T]{stream: stream, done: done}
}

// Write blocks until the reader has taken all of p, which is what gives a
// stream its backpressure: a slow reader slows the writer down rather than
// letting data pile up in memory.
func (w *Writer[T]) Write(p []byte) (int, error) {
	if w.stream == nil {
		return 0, io.ErrClosedPipe
	}
	n := int(w.stream.WriteAll(p))
	if n < len(p) {
		return n, ErrReaderGone
	}
	return n, nil
}

// CloseWith ends the stream and resolves the future: with value when err is
// nil, with err's message otherwise.
//
// The stream is dropped before the future is written. A reader reads the
// future only once the stream has ended, and a future write waits for its
// reader, so resolving it first would deadlock both sides.
func (w *Writer[T]) CloseWith(value T, err error) {
	if w.stream != nil {
		w.stream.Drop()
		w.stream = nil
	}
	if w.done == nil {
		return
	}
	done := w.done
	w.done = nil
	if err != nil {
		done.Write(witTypes.Err[T](err.Error()))
		return
	}
	// The result is false if the reader dropped the future without reading
	// it, which is its prerogative; there is nobody left to tell.
	done.Write(witTypes.Ok[T, string](value))
}

// UnitWriter is the common case: a stream whose future carries no value.
type UnitWriter = Writer[witTypes.Unit]

// Close ends a unit stream successfully, or with err if it is non-nil.
func Close(w *UnitWriter, err error) {
	w.CloseWith(witTypes.Unit{}, err)
}
