module wasitel-http

go 1.24

require (
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.wasmcloud.dev/component v0.0.8
	go.wasmcloud.dev/x/wasitel v0.0.1
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/trace v1.36.0 // indirect
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/regclient/regclient v0.8.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/urfave/cli/v3 v3.3.3 // indirect
	go.bytecodealliance.org v0.7.0 // indirect
	go.bytecodealliance.org/cm v0.3.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

tool go.bytecodealliance.org/cmd/wit-bindgen-go

// NOTE(lxf): Remove this line if running outside of wasmCloud/go repository
replace go.wasmcloud.dev/component => ../../../component
