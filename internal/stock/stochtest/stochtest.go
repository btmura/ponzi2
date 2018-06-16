// Binary stochtest exercises the stock.GetStochastics function.
//
// go run internal/stock/stochtest/stochtest.go
package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/stock"
)

var alphaVantageAPIKey = flag.String("alpha_vantage_api_key", "", "Alpha Vantage API Key")

func main() {
	flag.Parse()

	av := stock.NewAlphaVantage(*alphaVantageAPIKey)
	resp, err := av.GetStochastics(&stock.StochasticRequest{Symbol: "SPY"})
	if err != nil {
		glog.Exit(err)
	}
	// TODO(btmura): print out response properly
	fmt.Println(resp)
}
