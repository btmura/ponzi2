// Binary movavgtest exercises the stock.GetMovingAverage function.
//
// go run internal/stock/movavgtest/movavgtest.go
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/btmura/ponzi2/internal/stock"
)

var alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key")

func main() {
	flag.Parse()

	av := stock.NewAlphaVantage(*alphaVantageAPIKey)
	resp, err := av.GetMovingAverage(&stock.GetMovingAverageRequest{Symbol: "SPY"})
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range resp.Values {
		fmt.Printf("%d: %s A: %.2f\n", i, v.Date, v.Average)
	}
}
