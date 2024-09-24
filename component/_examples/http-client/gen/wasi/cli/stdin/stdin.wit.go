// Code generated by wit-bindgen-go. DO NOT EDIT.

// Package stdin represents the imported interface "wasi:cli/stdin@0.2.0".
package stdin

import (
	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/wasmCloud/component-sdk-go/_examples/http-server/gen/wasi/io/streams"
)

// GetStdin represents the imported function "get-stdin".
//
//	get-stdin: func() -> input-stream
//
//go:nosplit
func GetStdin() (result streams.InputStream) {
	result0 := wasmimport_GetStdin()
	result = cm.Reinterpret[streams.InputStream]((uint32)(result0))
	return
}

//go:wasmimport wasi:cli/stdin@0.2.0 get-stdin
//go:noescape
func wasmimport_GetStdin() (result0 uint32)
