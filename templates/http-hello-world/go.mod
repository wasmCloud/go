module http-hello-world

go 1.27.1

require (
	// Pseudo-version of go-pkg main (2026-09-08); bump to the next tagged release.
	go.bytecodealliance.org/pkg v0.2.4-0.20260908212714-8fe76305a856
	go.wasmcloud.dev/component v0.1.0
)

require (
	// Pseudo-version of componentize-go main (2026-09-08), whose wrapper
	// downloads the componentize-go v0.4.2 release binary; bump to the next
	// tagged release.
	github.com/bytecodealliance/componentize-go v0.4.3-0.20260908212636-5b8eef1155ae // indirect
)

tool github.com/bytecodealliance/componentize-go
