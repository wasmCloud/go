# HTTP Password Checker

A WebAssembly component that scores password strength over HTTP, using
[go-password-validator](https://github.com/wagslane/go-password-validator).

## Develop

```shell
wash dev
curl -X POST localhost:8000/api/v1/check -d '{"value":"hunter2"}'
```

## Build & deploy

```shell
wash build
wash oci push ghcr.io/<your-org>/http-password-checker:0.1.0 build/http_password_checker.wasm
```

See the [http-server example](../http-server/deployment.yaml) for the
`WorkloadDeployment` shape.
