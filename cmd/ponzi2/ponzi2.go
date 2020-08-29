package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/server"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	token            = flag.String("token", "", "IEX API token required on requests.")
	enableChartCache = flag.Bool("enable_chart_cache", true, "Whether to enable the chart cache.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
	remoteAddr       = flag.String("remote_addr", "", "")
)

func main() {
	flag.Parse()

	switch {
	case *remoteAddr != "":
		c := server.NewClient(*remoteAddr)
		a := app.New(c, *token)
		logger.Fatal(a.Run())

	case *enableChartCache:
		cache, err := iex.OpenGOBChartCache()
		if err != nil {
			logger.Fatal(err)
		}
		c := iex.NewClient(cache, *dumpAPIResponses)
		a := app.New(c, *token)
		logger.Fatal(a.Run())

	default:
		c := iex.NewClient(new(iex.NoOpChartCache), *dumpAPIResponses)
		a := app.New(c, *token)
		logger.Fatal(a.Run())
	}
}
