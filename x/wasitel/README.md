# OpenTelemetry Exporter for WebAssembly Components

> [!IMPORTANT]  
> This requires TinyGo `0.35.0` (current `dev` branch) in order to compile.

`wasitel` provides a `wasi:http`-based OpenTelemetry Go exporter implementation.

## Examples:

For usage examples, please check out the [`wasitel-http` example component](https://github.com/wasmCloud/go/tree/main/examples/component/wasitel-http).

### Acknowledgements

The `wasiteltrace/internal/convert` code has been adapted from [`opentelemetry-go`](https://github.com/open-telemetry/opentelemetry-go)'s internal packages, please see the code itself for the upstream soure references.

The `wasiteltrace/internal/types` code has been adapted from [`opentelemetry-proto-go`](https://github.com/open-telemetry/opentelemetry-proto-go)'s generated protobufs, please see the code itself for the upstream source references.