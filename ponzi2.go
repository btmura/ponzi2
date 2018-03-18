package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
)

//go:generate go generate github.com/btmura/ponzi2/internal/app
//go:generate go generate github.com/btmura/ponzi2/internal/gfx

var alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key")

func main() {
	flag.Parse()
	app.Run(*alphaVantageAPIKey)
}
