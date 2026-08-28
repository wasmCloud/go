// This file keeps the Go toolchain happy on non-wasm platforms: a package
// containing an assembly file may declare functions without bodies, which is
// what a `//go:wasmimport` is everywhere except the wasm build.
