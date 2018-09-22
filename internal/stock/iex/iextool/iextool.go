// The iextool command prints stock data for a list of stock symbols.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	symbols          = flag.String("symbols", "SPY", "Comma-separated list of symbols.")
	chartLast        = flag.Int("chart_last", 0, "Last N chart elements if greater than zero.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	c := iex.NewClient(*dumpAPIResponses)

	req := &iex.GetStocksRequest{
		Symbols: strings.Split(*symbols, ","),
		Range:   iex.RangeTwoYears,
	}
	if *chartLast > 0 {
		req.ChartLast = *chartLast
	}

	stocks, err := c.GetStocks(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	for _, st := range stocks {
		fmt.Println(st.Symbol)
		for i, p := range st.Chart {
			fmt.Printf("%3d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, p.Date, p.Open, p.High, p.Low, p.Close, p.Volume)
		}
	}
}
