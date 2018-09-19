// The iextool command prints stock data for a stock symbol.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	symbol           = flag.String("symbol", "SPY", "Symbol to lookup.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	ctx := context.Background()

	c := iex.NewClient(*dumpAPIResponses)

	req := &iex.GetChartRequest{Symbol: *symbol}
	ch, err := c.GetChart(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	for i, p := range ch.Points {
		fmt.Printf("%d: %s O: %.2f H: %.2f L: %.2f C: %.2f V: %d\n",
			i, p.Date, p.Open, p.High, p.Low, p.Close, p.Volume)
	}
}
