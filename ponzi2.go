package main

import "github.com/btmura/ponzi2/internal/app"

//go:generate go generate github.com/btmura/ponzi2/internal/app
//go:generate go generate github.com/btmura/ponzi2/internal/gfx

func main() {
	app.Run()
}
