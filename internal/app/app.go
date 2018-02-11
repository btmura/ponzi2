package app

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app/controller"
)

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.
}

// Run runs the stock chart viewer in a window.
func Run() {
	controller.NewController().Run()
}
