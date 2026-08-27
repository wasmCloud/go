module http-hello-world

go 1.25

require (
	// Pseudo-version of go-pkg main (post-#10); bump to the next tagged release.
	go.bytecodealliance.org/pkg v0.2.4-0.20260806154504-91f6c4863e67
	go.wasmcloud.dev/component v0.1.0
)

require (
	// Pseudo-version of componentize-go main (post-#72, wit-bindgen 0.61.1);
	// bump to the next tagged release.
	github.com/bytecodealliance/componentize-go v0.4.2-0.20260827144128-20f3b0c2a412 // indirect
)

tool github.com/bytecodealliance/componentize-go
