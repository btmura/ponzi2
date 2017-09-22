package ponzi

import (
	"flag"
	"runtime"
)

// Get go-bindata from github.com/jteeuwen/go-bindata. It's used to embed resources into the binary.
//go:generate go-bindata -pkg ponzi -prefix data -ignore ".*blend.*" data
//go:generate go generate github.com/btmura/ponzi2/internal/gfx

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.

	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

// Run runs the stock chart viewer in a window.
func Run() {
	c := &Controller{}
	c.model = &Model{}
	c.view = NewView(c.model)
	c.Run()
}
