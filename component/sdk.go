package component

//go:generate go tool wit-bindgen-go generate --world sdk --out gen ./wit

import (
	"embed"
)

//go:embed wit/*
var Wit embed.FS
