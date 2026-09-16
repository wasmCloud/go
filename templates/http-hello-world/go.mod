module http-hello-world

go 1.27.1

require (
	// Pseudo-version of go-pkg main (post-#12); bump to the next tagged release.
	go.bytecodealliance.org/pkg v0.2.4-0.20260911130647-2495ff7eca86
	go.wasmcloud.dev/component v0.2.0
)

require (
	// Pinned to a tag deliberately. componentize-go's main.go hardcodes the
	// release it downloads, and only a tag names itself; a pseudo-version names
	// the release before it and would silently build with an older wit-bindgen.
	// See BUILDING.md.
	github.com/bytecodealliance/componentize-go v0.4.3 // indirect
)

tool github.com/bytecodealliance/componentize-go
