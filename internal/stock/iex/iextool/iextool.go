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
	rangeFlag        = flag.String("range", "1d", "Range of data to get. Values: 1d, 2y")
	chartLast        = flag.Int("chart_last", 5, "Last N chart elements if greater than zero.")
	token            = flag.String("token", "", "API token required on requests.")
	dumpAPIResponses = flag.Bool("dump_api_responses", true, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	c := iex.NewClient(*token, *dumpAPIResponses)

	req := &iex.GetStocksRequest{
		Symbols: strings.Split(*symbols, ","),
		Range:   iex.TwoYears,
	}

	switch *rangeFlag {
	case "1d":
		req.Range = iex.OneDay
	case "2y":
		req.Range = iex.TwoYears
	default:
		log.Fatalf("iextool: unsupported range: %v", *rangeFlag)
	}

	if *chartLast > 0 {
		req.ChartLast = *chartLast
	}

	stocks, err := c.GetStocks(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Request:\n\n%+v\n\n", req)

	for _, st := range stocks {
		fmt.Printf("Symbol:\n\n%s\n\nQuote:\n\n%+v\n\n", st.Symbol, st.Quote)
		for i, p := range st.Chart {
			fmt.Printf("%3d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, p.Date, p.Open, p.High, p.Low, p.Close, p.Volume)
		}
	}
}
