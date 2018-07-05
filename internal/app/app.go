package app

import (
	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/stock/alpha"
)

// Run runs the stock chart viewer in a window.
func Run(alphaVantageAPIKey string, dumpAPIResponses bool) {
	var sdf controller.StockDataFetcher
	if alphaVantageAPIKey != "" {
		sdf = alpha.NewAlphaVantage(alphaVantageAPIKey, dumpAPIResponses)
	} else {
		sdf = alpha.NewDemo()
	}
	controller.NewController(sdf).Run()
}
