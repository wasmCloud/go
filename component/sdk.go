package component

//go:generate go tool wit-bindgen --world sdk --out internal/gen ./wit

import (
	"embed"
)

//go:embed wit/*
var Wit embed.FS
