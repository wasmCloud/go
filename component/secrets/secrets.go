package secrets

import (
	"errors"
	"fmt"

	reveal "go.wasmcloud.dev/component/imports/wasmcloud_secrets_2_1_0_reveal"
	store "go.wasmcloud.dev/component/imports/wasmcloud_secrets_2_1_0_store"
)

// ErrNotFound is returned by [Get] when no secret exists at the requested
// key. Test for it with errors.Is.
var ErrNotFound = errors.New("secrets: secret not found")

// Secret is an opaque handle to a secret held by the host. Obtain one with
// [Get] and unwrap the underlying value with [Secret.Reveal].
type Secret struct {
	inner *store.Secret
}

// Get returns a handle to the secret stored at key.
//
// The returned handle does not carry the secret's value; call
// [Secret.Reveal] to obtain it.
func Get(key string) (*Secret, error) {
	res := store.Get(key)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &Secret{inner: res.Ok()}, nil
}

// Reveal returns the secret's underlying value.
func (s *Secret) Reveal() Value {
	return Value{v: reveal.Reveal(s.inner)}
}

// Drop releases the host-side handle. The handle is also released by the
// garbage collector if the Secret becomes unreachable.
func (s *Secret) Drop() {
	s.inner.Drop()
}

// Value is a revealed secret value: either a string or raw bytes.
type Value struct {
	v store.SecretValue
}

// IsString reports whether the secret was stored as a string (as opposed to
// raw bytes).
func (v Value) IsString() bool {
	return v.v.Tag() == store.SecretValueString
}

// String returns the value as a string. Byte-typed secrets are converted
// with string().
func (v Value) String() string {
	if v.v.Tag() == store.SecretValueString {
		return v.v.String()
	}
	return string(v.v.Bytes())
}

// Bytes returns the value as a byte slice. String-typed secrets are
// converted with []byte().
func (v Value) Bytes() []byte {
	if v.v.Tag() == store.SecretValueBytes {
		return v.v.Bytes()
	}
	return []byte(v.v.String())
}

func convertError(e store.SecretsError) error {
	switch e.Tag() {
	case store.SecretsErrorUpstream:
		return fmt.Errorf("secrets: upstream error: %s", e.Upstream())
	case store.SecretsErrorIo:
		return fmt.Errorf("secrets: io error: %s", e.Io())
	case store.SecretsErrorNotFound:
		return ErrNotFound
	default:
		return fmt.Errorf("secrets: unknown error (tag %d)", e.Tag())
	}
}
