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

	req := &iex.GetTradingSessionSeriesRequest{Symbol: *symbol}
	sr, err := c.GetTradingSessionSeries(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	for i, ts := range sr.TradingSessions {
		fmt.Printf("%d: %s O: %.2f H: %.2f L: %.2f C: %.2f V: %d\n",
			i, ts.Date, ts.Open, ts.High, ts.Low, ts.Close, ts.Volume)
	}
}
