module github.com/wasmCloud/go/examples/components/http-keyvalue-crud

go 1.25

require (
	github.com/julienschmidt/httprouter v1.3.0
	go.bytecodealliance.org/pkg v0.2.4-0.20260806154504-91f6c4863e67
	go.wasmcloud.dev/component v0.1.0
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.1 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

tool github.com/bytecodealliance/componentize-go

// NOTE: Uncomment to build against the SDK in this repository instead of the
// released module. CI applies this replace automatically.
//replace go.wasmcloud.dev/component => ../../../component
