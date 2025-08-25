# Go WASITEL HTTP

## 📦 Dependencies

Before starting, ensure that you have the following installed in addition to the Go (1.23+) toolchain:

- [`tinygo`](https://tinygo.org/getting-started/install/) for compiling Go (always use the latest version)
- [`wasm-tools`](https://github.com/bytecodealliance/wasm-tools#installation) for Go bindings
- [wasmCloud Shell (`wash`)](https://wasmcloud.com/docs/installation) for building and running the components and wasmCloud environment

## ⚠️ Issues/FAQ

### Build errors

New releases of `wasm-tools` may introduce compatibility issues that can result in build errors. If you encounter issues, try using v1.225.0, which is currently the most consistent for Go builds. You can install `wasm-tools` [v1.225.0 from upstream releases](https://github.com/bytecodealliance/wasm-tools/releases/tag/v1.225.0), or use `cargo` ([Rust toolchain](https://doc.rust-lang.org/cargo/getting-started/installation.html)) -- (i.e. `cargo install --locked wasm-tools@1.225.0`)