package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
)

var (
	alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key. Leave blank to use predefined data.")
	dumpAPIResponses   = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()
	app.Run(*alphaVantageAPIKey, *dumpAPIResponses)
}
