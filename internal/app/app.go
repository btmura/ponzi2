package app

import (
	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Run runs the stock chart viewer in a window.
func Run(dumpAPIResponses bool) {
	c := iex.NewClient(dumpAPIResponses)
	controller.NewController(c).Run()
}
