# wasmCloud Go

This repository contains Go ecosystem libraries for [wasmCloud](https://github.com/wasmcloud/wasmcloud).

Depending on what you are interested in doing with wasmCloud, you can find the relevant code under the following directories:

* [`component`](https://github.com/wasmCloud/go/tree/main/component) (available as `go.wasmcloud.dev/component`) - Component SDK for building wasmCloud applications in Go.
* [`provider`](https://github.com/wasmCloud/go/tree/main/provider) (available as `go.wasmcloud.dev/provider`) - Provider SDK for building  wasmCloud capability providers in Go.
* [`examples`](https://github.com/wasmCloud/go/tree/main/examples) - A set of example wasmCloud applications (under [`examples/component`](https://github.com/wasmCloud/go/tree/main/examples/component)) and capability providers (under [`examples/provider`](https://github.com/wasmCloud/go/tree/main/examples/provider)) that demonstrate how you can make use of the [Component SDK](https://github.com/wasmCloud/go/tree/main/component) and [Provider SDK](https://github.com/wasmCloud/go/tree/main/provider).
* [`templates`](https://github.com/wasmCloud/go/tree/main/templates) - A set of starter templates used by the wasmCloud CLI (`wash`) for starting a new wasmCloud application or capability provider.
* [`x`](https://github.com/wasmCloud/go/tree/main/x) - Experimental libraries that are made available for consumption before they are folded into one of the existing SDKs or published as a top-level library of their own.