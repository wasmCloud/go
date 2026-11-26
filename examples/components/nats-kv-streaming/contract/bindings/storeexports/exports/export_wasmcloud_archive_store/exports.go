// Package export_wasmcloud_archive_store is the hand-written end of the
// generated glue in ../wit_exports, which calls the functions below by name
// for every call into `wasmcloud:archive/store`.
//
// Each one forwards to [Exports], which the store component fills in from its
// own main package during init. That keeps the implementation out of the
// bindings tree, and keeps this file — the only one under bindings/ that is
// not generated — small enough that regenerating never has to touch it.
package export_wasmcloud_archive_store

import (
	store "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/storeexports/imports/wasmcloud_archive_store"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// Done is the future a stream's writer resolves once it has finished.
type Done = witTypes.FutureReader[witTypes.Result[witTypes.Unit, string]]

// Exports holds the implementation of every function in the interface. A
// call to a function left nil panics, which traps the component.
var Exports struct {
	Put         func(name, contentType string, data *witTypes.StreamReader[uint8], done *Done) witTypes.Result[store.ObjectInfo, store.Error]
	Get         func(name string) witTypes.Result[witTypes.Tuple3[store.ObjectInfo, *witTypes.StreamReader[uint8], *Done], store.Error]
	Stat        func(name string) witTypes.Result[store.ObjectInfo, store.Error]
	ListObjects func() witTypes.Result[[]store.ObjectInfo, store.Error]
	Delete      func(name string) witTypes.Result[witTypes.Unit, store.Error]
}

func Put(name string, contentType string, data *witTypes.StreamReader[uint8], done *Done) witTypes.Result[store.ObjectInfo, store.Error] {
	return Exports.Put(name, contentType, data, done)
}

func Get(name string) witTypes.Result[witTypes.Tuple3[store.ObjectInfo, *witTypes.StreamReader[uint8], *Done], store.Error] {
	return Exports.Get(name)
}

func Stat(name string) witTypes.Result[store.ObjectInfo, store.Error] {
	return Exports.Stat(name)
}

func ListObjects() witTypes.Result[[]store.ObjectInfo, store.Error] {
	return Exports.ListObjects()
}

func Delete(name string) witTypes.Result[witTypes.Unit, store.Error] {
	return Exports.Delete(name)
}
