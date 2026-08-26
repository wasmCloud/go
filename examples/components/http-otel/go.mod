module github.com/wasmCloud/go/examples/components/http-otel

go 1.25

require (
	go.bytecodealliance.org/pkg v0.2.4-0.20260806154504-91f6c4863e67
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	go.wasmcloud.dev/component v0.1.0
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

// NOTE: Uncomment to build against the SDK in this repository instead of the
// released module. CI applies this replace automatically.
//replace go.wasmcloud.dev/component => ../../../component

tool github.com/bytecodealliance/componentize-go
