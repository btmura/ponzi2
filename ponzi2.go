package main

import (
	"flag"

	"gitlab.com/btmura/ponzi2/internal/app"
)

var dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")

func main() {
	flag.Parse()
	app.Run(*dumpAPIResponses)
}
