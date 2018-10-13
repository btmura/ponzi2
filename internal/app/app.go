// Package app exports a Run function to start the app.
package app

import (
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Run runs the app. Should be called from main.
func Run(dumpAPIResponses bool) {
	c := iex.NewClient(dumpAPIResponses)
	if err := controller.New(c).RunLoop(); err != nil {
		glog.Fatal(err)
	}
}
