package app

import (
	"github.com/btmura/ponzi2/internal/stock"

	"github.com/btmura/ponzi2/internal/app/controller"
)

//go:generate go generate github.com/btmura/ponzi2/internal/app/config
//go:generate go generate github.com/btmura/ponzi2/internal/app/view

// Run runs the stock chart viewer in a window.
func Run(alphaVantageAPIKey string) {
	av := stock.NewAlphaVantage(alphaVantageAPIKey)
	controller.NewController(av).Run()
}
