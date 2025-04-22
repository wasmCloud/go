module github.com/wasmCloud/go/examples/component/invoke

go 1.24

require (
	github.com/stretchr/testify v1.10.0
	go.bytecodealliance.org/cm v0.2.2
	go.wasmcloud.dev/component v0.0.6
	go.wasmcloud.dev/wadge v0.7.0
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/regclient/regclient v0.8.2 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/samber/slog-common v0.18.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/urfave/cli/v3 v3.0.0-beta1 // indirect
	go.bytecodealliance.org v0.6.2 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// NOTE(lxf): Remove this line if running outside of wasmCloud/go repository
replace go.wasmcloud.dev/component => ../../../component

tool (
	go.wasmcloud.dev/component/wit-bindgen
	go.wasmcloud.dev/wadge/cmd/wadge-bindgen-go
)
