# 🚌 Key-Value Counter with wasmCloud Bus

[This example](https://github.com/wasmCloud/go/tree/main/examples/component/http-keyvalue-bus) demonstrates how to use the `wasmcloud:bus` interface. 

The `wasmcloud:bus` interface enables users to dynamically define which runtime link to use for a particular interface. This means that a component can link to different entities under different circumstances, or even link to the same entity with a different set of configurations as needed.

The **Key-Value Counter with wasmCloud Bus** application uses the `wasi:keyvalue/store` and `wasi:keyvalue/atomics` interfaces with either the `keyvalue-redis` provider or the `keyvalue-nats` provider, depending on the endpoint to which the user makes a request. 

Apart from adding HTTP routing and link definition via `wasmcloud:bus`, the Go code is the same as the wasmCloud Quickstart in Go, making this a good intermediate step after getting acquainted with wasmCloud. If you're not yet familiar with the fundamentals of wasmCloud, we recommend completing the [Quickstart](https://wasmcloud.com/docs/tour/hello-world) first. 

## 📦 Dependencies

Before starting, ensure that you have the following installed in addition to the Go (1.23+) toolchain:

- [`tinygo`](https://tinygo.org/getting-started/install/) for compiling Go (always use the latest version)
- [`wasm-tools`](https://github.com/bytecodealliance/wasm-tools#installation) for Go bindings
- [wasmCloud Shell (`wash`)](https://wasmcloud.com/docs/installation) for building and running the components and wasmCloud environment
- A local Redis server running either [via the CLI](https://redis.io/docs/getting-started/) or Docker (`docker run -d --name redis -p 6379:6379 redis`)

## 👟 Quickstart

To run this example, clone the [wasmCloud/go repository](https://github.com/wasmcloud/go): 

```shell
git clone https://github.com/wasmCloud/go.git
```

Change directory to `examples/component/http-keyvalue-bus`:

```shell
cd examples/component/http-keyvalue-bus
```

Build the component from the Go code:

```shell
wash build
```

Make sure a Redis server is running on your local port `6379` via the Redis CLI or a container.

Run `wash up --detached` (or the short flag `-d`) to start a local wasmCloud environment in detached mode:

```shell
wash up -d
```

Deploy the application manifest in `wadm.yaml`:

```shell
wash app deploy wadm.yaml
```

**Note**: Since this example uses a more complex manifest with multiple providers for the same capability, we recommend deploying manually rather than using `wash dev`.

The user can make a request to either the `/nats` or `/redis` endpoint with a name appended as a query string:

```shell
curl 'localhost:8000/nats?name=Alice'
```
```shell
curl 'localhost:8000/redis?name=Alice'
```

Try making multiple requests with the same name to each endpoint and watch how the count increments in each key-value store.

When you're done with the example, run `wash down` to stop your local wasmCloud environment.

Read through the comments in `main.go` for step-by-step explanation of how the example works.

## 📖 Further reading 

You can learn more about runtime linking on the [Linking at Runtime](https://wasmcloud.com/docs/concepts/linking-components/linking-at-runtime) page of the wasmCloud documentation.