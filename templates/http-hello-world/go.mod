module http-hello-world

go 1.27.1

require (
	go.bytecodealliance.org/wasi v0.1.1
	go.wasmcloud.dev/component v0.1.6
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	// Pinned to a tag deliberately. componentize-go's main.go hardcodes the
	// release it downloads, and only a tag names itself; a pseudo-version names
	// the release before it and would silently build with an older wit-bindgen.
	// See BUILDING.md.
	github.com/bytecodealliance/componentize-go v0.4.3 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	// Pseudo-version of go-pkg main (post-#12); bump to the next tagged release.
	go.bytecodealliance.org/pkg v0.4.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

tool github.com/bytecodealliance/componentize-go
