module github.com/wasmCloud/go/examples/components/http-server

go 1.25

require (
	github.com/julienschmidt/httprouter v1.3.0
	go.wasmcloud.dev/component v0.2.0
)

require (
	github.com/apparentlymart/go-userdirs v0.0.0-20200915174352-b0c018a67c13 // indirect
	github.com/bytecodealliance/componentize-go v0.4.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/samber/slog-common v0.22.0 // indirect
	go.bytecodealliance.org/pkg v0.2.2 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

tool github.com/bytecodealliance/componentize-go

// NOTE: Remove this line if running outside of the wasmCloud/go repository
replace go.wasmcloud.dev/component => ../../../component
