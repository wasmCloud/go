module github.com/wasmCloud/go/examples/components/nats-kv-streaming/store

go 1.27.1

require (
	github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract v0.0.0
	go.bytecodealliance.org/pkg v0.4.1
	go.bytecodealliance.org/wasi v0.1.2
	go.wasmcloud.dev/component v0.2.0
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.5.3 // indirect
	github.com/gofrs/flock v0.13.2 // indirect
	golang.org/x/sys v0.49.0 // indirect
)

tool github.com/bytecodealliance/componentize-go

// The contract module (the shared WIT and its Go bindings) lives next door.
replace github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract => ../contract

// NOTE: Uncomment to build against the SDK in this repository instead of the
// released module. CI applies this replace automatically.
//replace go.wasmcloud.dev/component => ../../../../component
