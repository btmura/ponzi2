// Binary stochtest exercises the stock.GetStochastics function.
//
// go run internal/stock/stochtest/stochtest.go
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

	req := &stock.GetStochasticsRequest{
		Symbol:   "SPY",
		Interval: stock.Daily,
	}

	var resp *stock.Stochastics
	var err error
	if *alphaVantageAPIKey != "" {
		av := stock.NewAlphaVantage(*alphaVantageAPIKey, *dumpAPIResponses)
		resp, err = av.GetStochastics(ctx, req)
	} else {
		d := stock.NewDemo()
		resp, err = d.GetStochastics(ctx, req)
	}

	if err != nil {
		log.Fatal(err)
	}

	for i, v := range resp.Values {
		fmt.Printf("%d: %s K: %.2f D: %.2f\n", i, v.Date, v.K, v.D)
	}
}
