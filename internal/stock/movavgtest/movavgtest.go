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

	av := stock.NewAlphaVantage(*alphaVantageAPIKey, *dumpAPIResponses)
	req := &stock.GetMovingAverageRequest{
		Symbol:     "SPY",
		TimePeriod: 50,
	}
	resp, err := av.GetMovingAverage(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range resp.Values {
		fmt.Printf("%d: %s A: %.2f\n", i, v.Date, v.Average)
	}
}
