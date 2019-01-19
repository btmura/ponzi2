// Package app exports a Run function to start the app.
package app

import (
	"github.com/golang/glog"

	"gitlab.com/btmura/ponzi2/internal/app/controller"
	"gitlab.com/btmura/ponzi2/internal/stock/iex"
)

// Run runs the app. Should be called from main.
func Run(dumpAPIResponses bool) {
	var opts []iex.ClientOption
	if dumpAPIResponses {
		opts = append(opts, iex.DumpAPIResponses())
	}

	c, err := iex.NewClient(opts...)
	if err != nil {
		glog.Fatal(err)
	}

	if err := controller.New(c).RunLoop(); err != nil {
		glog.Fatal(err)
	}
}
