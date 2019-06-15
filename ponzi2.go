package main

import (
	"flag"

	"github.com/btmura/ponzi2/internal/app"
)

var (
	token            = flag.String("token", "", "IEX API token required on requests.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()
	app.Run(*token, *dumpAPIResponses)
}
