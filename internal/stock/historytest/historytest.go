// Binary historytest exercises the stock.GetMovingAverage function.
//
// go run internal/stock/historytest/historytest.go
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
	req := &stock.GetHistoryRequest{Symbol: "SPY"}
	resp, err := av.GetHistory(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	for i, ts := range resp.TradingSessions {
		fmt.Printf("%d: %s O: %.2f H: %.2f L: %.2f C: %.2f V: %d\n",
			i, ts.Date, ts.Open, ts.High, ts.Low, ts.Close, ts.Volume)
	}
}
