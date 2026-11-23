---
name: wasmcloud-go-components
description: Pointers to the right published skill for WIT interface design, building and debugging WASI components, and running components with wash. Use when writing or changing .wit files, designing a new capability interface or world, or building, composing, or running .wasm components beyond what the examples skill covers.
---

# Component guidance

Select only the guidance for the interface or component being changed.
Use the existing examples, templates, and SDK packages as the starting point.
Load the matching published skill if your harness supports it; otherwise use the relevant repository documentation.

- WIT interfaces, worlds, and packages under `component/wit/` or `plugin/wit/`:
  [wit-interface-design](https://github.com/cosmonic-labs/skills/tree/main/wit-interface-design).
- Building, composing, testing, or debugging WASI 0.3 components:
  [webassembly-component-development](https://github.com/cosmonic-labs/skills/tree/main/webassembly-component-development).
- Building and running components with `wash`, including `examples/` and `templates/`:
  [wash](https://github.com/cosmonic-labs/skills/tree/main/wash).
- componentize-go itself, the patched Go, and offline builds:
  [BUILDING.md](../../../BUILDING.md).
