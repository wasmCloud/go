package wasmcloud

import (
	store "go.wasmcloud.dev/component/imports/wasi_config_0_2_0_rc_1_store"
)

// GetConfigOrDefault tries to get a configuration value by provided key using
// [wasi:config/store.get] and falling back to the provided defaultValue if a
// configuration value by provided key is not found.
//
// [wasi:config/store.get]: https://github.com/WebAssembly/wasi-config/blob/main/wit/store.wit
func GetConfigOrDefault(key string, defaultValue string) string {
	res := store.Get(key)
	if res.IsOk() {
		opt := res.Ok()
		if opt.IsSome() {
			return opt.Some()
		}
	}

	return defaultValue
}
