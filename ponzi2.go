package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
	"github.com/btmura/ponzi2/internal/log"
	"github.com/btmura/ponzi2/internal/stock/iex"
	"github.com/btmura/ponzi2/internal/stock/iexremote"
)

var (
	token            = flag.String("token", "", "IEX API token required on requests.")
	enableChartCache = flag.Bool("enable_chart_cache", true, "Whether to enable the chart cache.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

var (
	port       = flag.Int("port", 0, "Non-zero port number to start the IEX Remote server.")
	remoteAddr = flag.String("remote_addr", "", "")
)

func main() {
	flag.Parse()

	switch {
	case *port != 0:
		c, err := iex.NewClient(*token, *enableChartCache, *dumpAPIResponses)
		if err != nil {
			log.Fatal(err)
		}

		s := iexremote.NewServer(*port, c)
		log.Fatal(s.Run())

	case *remoteAddr != "":
		c := iexremote.NewClient(*remoteAddr)

		a := app.New(c)
		log.Fatal(a.Run())

	default:
		c, err := iex.NewClient(*token, *enableChartCache, *dumpAPIResponses)
		if err != nil {
			log.Fatal(err)
		}

		a := app.New(c)
		log.Fatal(a.Run())
	}
}
