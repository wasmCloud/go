module github.com/wasmCloud/go/examples/components/http-p3-streaming

go 1.25

require (
	go.bytecodealliance.org/pkg v0.2.4-0.20260806154504-91f6c4863e67
	go.wasmcloud.dev/component v0.2.0
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.1 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)

tool github.com/bytecodealliance/componentize-go

// NOTE: Remove this line if running outside of the wasmCloud/go repository
replace go.wasmcloud.dev/component => ../../../component
