package app

import (
	"flag"
	"runtime"
)

// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to embed resources into the binary.
//go:generate go-bindata -pkg app -prefix data -ignore ".*blend.*" -ignore ".*proto" -ignore ".*xcf" data

// Generate config.pb.go. Follow setup instructions @ github.com/golang/protobuf.
//go:generate protoc -I=data --go_out=. config.proto

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.

	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

// Run runs the stock chart viewer in a window.
func Run() {
	NewController().Run()
}
