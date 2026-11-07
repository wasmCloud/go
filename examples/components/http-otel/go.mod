module github.com/wasmCloud/go/examples/components/http-otel

go 1.27.1

require (
	go.bytecodealliance.org/pkg v0.4.1
	go.bytecodealliance.org/wasi v0.1.1
	go.opentelemetry.io/otel v1.45.0
	go.opentelemetry.io/otel/sdk v1.45.0
	go.opentelemetry.io/otel/trace v1.45.0
	go.wasmcloud.dev/component v0.1.6
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

// NOTE: Uncomment to build against the SDK in this repository instead of the
// released module. CI applies this replace automatically.
//replace go.wasmcloud.dev/component => ../../../component

tool github.com/bytecodealliance/componentize-go
