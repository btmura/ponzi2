package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	iexAPIToken         = flag.String("iex_api_token", "", "IEX API Token required on requests.")
	enableIEXChartCache = flag.Bool("enable_iex_chart_cache", true, "Whether to enable the IEX chart cache.")
	dumpIEXAPIResponses = flag.Bool("dump_iex_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	switch {
	case *enableIEXChartCache:
		cache, err := iex.OpenGOBChartCache()
		if err != nil {
			logger.Fatal(err)
		}
		c := iex.NewClient(cache, *dumpIEXAPIResponses)
		a := app.New(c, *iexAPIToken)
		logger.Fatal(a.Run())

	default:
		c := iex.NewClient(new(iex.NoOpChartCache), *dumpIEXAPIResponses)
		a := app.New(c, *iexAPIToken)
		logger.Fatal(a.Run())
	}
}
