package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
)

//go:generate go generate github.com/btmura/ponzi2/internal/app
//go:generate go generate github.com/btmura/ponzi2/internal/gfx
//go:generate go generate github.com/btmura/ponzi2/internal/stock

var (
	alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key")
	dumpAPIResponses   = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()
	app.Run(*alphaVantageAPIKey, *dumpAPIResponses)
}
