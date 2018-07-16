package app

import (
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Run runs the stock chart viewer in a window.
func Run(dumpAPIResponses bool) {
	c := iex.NewClient(dumpAPIResponses)
	if err := controller.New(c).Run(); err != nil {
		glog.Fatal(err)
	}
}
