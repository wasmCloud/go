module http-hello-world

go 1.27.1

require (
	go.bytecodealliance.org/wasi v0.1.1
	go.wasmcloud.dev/component v0.2.0
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	// Pinned to a tag deliberately. componentize-go's main.go hardcodes the
	// release it downloads, and only a tag names itself; a pseudo-version names
	// the release before it and would silently build with an older wit-bindgen.
	// See BUILDING.md.
	github.com/bytecodealliance/componentize-go v0.4.3 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	// v0.4.1 stops the GC running inside cabi_realloc, which traps on large
	// payloads (bytecodealliance/go-pkg#13). Do not go below it.
	go.bytecodealliance.org/pkg v0.4.1 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

tool github.com/bytecodealliance/componentize-go
