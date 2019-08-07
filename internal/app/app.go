// Package app exports a Run function to start the app.
package app

import (
	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/log"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Run runs the app. Should be called from main.
func Run(token string, enableChartCache, dumpAPIResponses bool) {
	c, err := iex.NewClient(token, enableChartCache, dumpAPIResponses)
	if err != nil {
		log.Fatal(err)
	}
	if err := controller.New(c).RunLoop(); err != nil {
		log.Fatal(err)
	}
}
