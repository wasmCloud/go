module github.com/wasmCloud/go/examples/component/http-client

go 1.24

require (
	go.bytecodealliance.org v0.5.0
	go.bytecodealliance.org/cm v0.1.0
	go.wasmcloud.dev/component v0.0.6
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/regclient/regclient v0.8.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/urfave/cli/v3 v3.0.0-beta1 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

tool go.bytecodealliance.org/cmd/wit-bindgen-go

// NOTE(lxf): Remove this line if running outside of wasmCloud/go repository
replace go.wasmcloud.dev/component => ../../../component
