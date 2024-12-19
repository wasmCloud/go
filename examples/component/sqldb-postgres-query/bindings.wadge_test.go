//go:build !wasm && !wasi && !wasip1 && !wasip2 && !wasm_unknown && !tinygo.wasm

// Code generated by wadge-bindgen-go DO NOT EDIT

package main_test

import (
	go_bytecodealliance_org__cm "go.bytecodealliance.org/cm"
	wadge "go.wasmcloud.dev/wadge"
	"runtime"
	sqldb___postgres___query__gen__wasmcloud__postgres__query "sqldb-postgres-query/gen/wasmcloud/postgres/query"
	sqldb___postgres___query__gen__wasmcloud__postgres__types "sqldb-postgres-query/gen/wasmcloud/postgres/types"
	"unsafe"
)

const _ string = runtime.Compiler

var _ unsafe.Pointer

//go:linkname wasmimport_Log go.wasmcloud.dev/component/gen/wasi/logging/logging.wasmimport_Log
func wasmimport_Log(level0 uint32, context0 *uint8, context1 uint32, message0 *uint8, message1 uint32) {
	var __p runtime.Pinner
	defer __p.Unpin()
	if __err := wadge.WithCurrentInstance(func(__instance *wadge.Instance) error {
		return __instance.Call("wasi:logging/logging@0.1.0-draft", "log", func() unsafe.Pointer {
			ptr := unsafe.Pointer(&level0)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(context0)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(&context1)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(message0)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(&message1)
			__p.Pin(ptr)
			return ptr
		}())
	}); __err != nil {
		wadge.CurrentErrorHandler()(__err)
	}
	return
}

//go:linkname wasmimport_Query sqldb-postgres-query/gen/wasmcloud/postgres/query.wasmimport_Query
func wasmimport_Query(query0 *uint8, query1 uint32, params0 *sqldb___postgres___query__gen__wasmcloud__postgres__types.PgValue, params1 uint32, result *go_bytecodealliance_org__cm.Result[sqldb___postgres___query__gen__wasmcloud__postgres__query.QueryErrorShape, go_bytecodealliance_org__cm.List[sqldb___postgres___query__gen__wasmcloud__postgres__types.ResultRow], sqldb___postgres___query__gen__wasmcloud__postgres__types.QueryError]) {
	var __p runtime.Pinner
	defer __p.Unpin()
	if __err := wadge.WithCurrentInstance(func(__instance *wadge.Instance) error {
		return __instance.Call("wasmcloud:postgres/query@0.1.1-draft", "query", func() unsafe.Pointer {
			ptr := unsafe.Pointer(query0)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(&query1)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(params0)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(&params1)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(result)
			__p.Pin(ptr)
			return ptr
		}())
	}); __err != nil {
		wadge.CurrentErrorHandler()(__err)
	}
	return
}

//go:linkname wasmimport_QueryBatch sqldb-postgres-query/gen/wasmcloud/postgres/query.wasmimport_QueryBatch
func wasmimport_QueryBatch(query0 *uint8, query1 uint32, result *go_bytecodealliance_org__cm.Result[sqldb___postgres___query__gen__wasmcloud__postgres__types.QueryError, struct{}, sqldb___postgres___query__gen__wasmcloud__postgres__types.QueryError]) {
	var __p runtime.Pinner
	defer __p.Unpin()
	if __err := wadge.WithCurrentInstance(func(__instance *wadge.Instance) error {
		return __instance.Call("wasmcloud:postgres/query@0.1.1-draft", "query-batch", func() unsafe.Pointer {
			ptr := unsafe.Pointer(query0)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(&query1)
			__p.Pin(ptr)
			return ptr
		}(), func() unsafe.Pointer {
			ptr := unsafe.Pointer(result)
			__p.Pin(ptr)
			return ptr
		}())
	}); __err != nil {
		wadge.CurrentErrorHandler()(__err)
	}
	return
}