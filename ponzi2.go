package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
)

var (
	token            = flag.String("token", "", "IEX API token required on requests.")
	enableChartCache = flag.Bool("enable_chart_cache", true, "Whether to enable the chart cache.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()
	app.Run(*token, *enableChartCache, *dumpAPIResponses)
}
