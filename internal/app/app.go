package app

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app/controller"
)

//go:generate go generate github.com/btmura/ponzi2/internal/app/config
//go:generate go generate github.com/btmura/ponzi2/internal/app/view

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.
}

// Run runs the stock chart viewer in a window.
func Run() {
	controller.NewController().Run()
}
