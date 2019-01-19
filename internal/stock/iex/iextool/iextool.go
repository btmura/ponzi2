// The iextool command prints stock data for a list of stock symbols.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"gitlab.com/btmura/ponzi2/internal/stock/iex"
)

var (
	symbols          = flag.String("symbols", "SPY", "Comma-separated list of symbols.")
	chartLast        = flag.Int("chart_last", 0, "Last N chart elements if greater than zero.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	var opts []iex.ClientOption
	if *dumpAPIResponses {
		opts = append(opts, iex.DumpAPIResponses())
	}

	c := iex.NewClient(opts...)

	symbols := strings.Split(*symbols, ",")

	gopts := []iex.GetStocksOption{iex.WithRange(iex.TwoYears)}
	if *chartLast > 0 {
		gopts = append(gopts, iex.WithChartLast(*chartLast))
	}

	stocks, err := c.GetStocks(ctx, symbols, gopts...)
	if err != nil {
		log.Fatal(err)
	}

	for _, st := range stocks {
		fmt.Println(st.Symbol)
		fmt.Printf("%+v\n", st.Quote)
		for i, p := range st.Chart {
			fmt.Printf("%3d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, p.Date, p.Open, p.High, p.Low, p.Close, p.Volume)
		}
	}
}
