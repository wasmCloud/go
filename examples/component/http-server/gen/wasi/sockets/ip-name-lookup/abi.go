// Code generated by component-bindgen-go. DO NOT EDIT.

package ipnamelookup

import (
	"go.wasmcloud.dev/component/cm"
	"unsafe"
)

// OptionIPAddressShape is used for storage in variant or result types.
type OptionIPAddressShape struct {
	_     cm.HostLayout
	shape [unsafe.Sizeof(cm.Option[IPAddress]{})]byte
}
