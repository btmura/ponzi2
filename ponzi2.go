package main

import "github.com/btmura/ponzi2/internal/ponzi"

//go:generate go generate github.com/btmura/ponzi2/internal/ponzi

func main() {
	ponzi.Run()
}
