package app

import (
	"flag"
	"runtime"

	"github.com/btmura/ponzi2/internal/app/controller"
)

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.

	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

// Run runs the stock chart viewer in a window.
func Run() {
	controller.NewController().Run()
}
