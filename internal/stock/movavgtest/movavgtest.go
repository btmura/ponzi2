// Binary movavgtest exercises the stock.GetMovingAverage function.
//
// go run internal/stock/movavgtest/movavgtest.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/btmura/ponzi2/internal/stock"
)

var (
	alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key")
	dumpAPIResponses   = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	req := &stock.GetMovingAverageRequest{
		Symbol:     "SPY",
		TimePeriod: 50,
	}

	var resp *stock.MovingAverage
	var err error
	if *alphaVantageAPIKey != "" {
		av := stock.NewAlphaVantage(*alphaVantageAPIKey, *dumpAPIResponses)
		resp, err = av.GetMovingAverage(ctx, req)
	} else {
		d := stock.NewDemo()
		resp, err = d.GetMovingAverage(ctx, req)
	}

	if err != nil {
		log.Fatal(err)
	}

	for i, v := range resp.Values {
		fmt.Printf("%d: %s A: %.2f\n", i, v.Date, v.Average)
	}
}
