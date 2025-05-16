module wasitel-http

go 1.24

require (
	go.bytecodealliance.org/cm v0.2.2
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/sdk v1.34.0
	go.wasmcloud.dev/component v0.0.6
	go.wasmcloud.dev/x/wasitel v0.0.1
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/regclient/regclient v0.8.2 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/urfave/cli/v3 v3.0.0-beta1 // indirect
	go.bytecodealliance.org v0.6.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace go.wasmcloud.dev/component => ../../../component

tool go.wasmcloud.dev/component/wit-bindgen
