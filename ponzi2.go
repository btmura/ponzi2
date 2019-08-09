package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
	"github.com/btmura/ponzi2/internal/log"
	"github.com/btmura/ponzi2/internal/server"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	token            = flag.String("token", "", "IEX API token required on requests.")
	enableChartCache = flag.Bool("enable_chart_cache", true, "Whether to enable the chart cache.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
	port             = flag.Int("port", 0, "Non-zero port number to start the server.")
)

func main() {
	flag.Parse()

	client, err := iex.NewClient(*token, *enableChartCache, *dumpAPIResponses)
	if err != nil {
		log.Fatal(err)
	}

	if *port != 0 {
		s := server.New(*port, client)
		log.Fatal(s.Run())
	} else {
		a := app.New(client)
		log.Fatal(a.Run())
	}
}
