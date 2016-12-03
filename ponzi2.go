package main

import "github.com/btmura/ponzi2/internal/ponzi"

// This generate command is here, so we can just do "go generate" in the root directory.
// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to bundle resources.
//go:generate go-bindata -pkg ponzi -o internal/ponzi/bindata.go internal/ponzi/data

func main() {
	ponzi.Run()
}
