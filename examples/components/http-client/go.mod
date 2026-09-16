module github.com/wasmCloud/go/examples/components/http-client

go 1.27.1

require (
	go.bytecodealliance.org/pkg v0.2.4-0.20260911130647-2495ff7eca86
	go.wasmcloud.dev/component v0.1.5
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.3 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

tool github.com/bytecodealliance/componentize-go

// NOTE: Uncomment to build against the SDK in this repository instead of the
// released module. CI applies this replace automatically.
//replace go.wasmcloud.dev/component => ../../../component
