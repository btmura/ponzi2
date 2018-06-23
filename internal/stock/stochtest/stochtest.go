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

var alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key")

func main() {
	flag.Parse()

	av := stock.NewAlphaVantage(*alphaVantageAPIKey)
	resp, err := av.GetStochastics(context.Background(), &stock.GetStochasticsRequest{Symbol: "SPY"})
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range resp.Values {
		fmt.Printf("%d: %s K: %.2f D: %.2f\n", i, v.Date, v.K, v.D)
	}
}
