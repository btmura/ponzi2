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

	req := &iex.GetChartRequest{
		Symbols:    strings.Split(*symbols, ","),
		ChartRange: iex.ChartRangeTwoYears,
	}
	if *chartLast > 0 {
		req.ChartLast = *chartLast
	}

	chs, err := c.GetChart(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	for _, ch := range chs {
		fmt.Println(ch.Symbol)
		for i, p := range ch.Points {
			fmt.Printf("%3d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, p.Date, p.Open, p.High, p.Low, p.Close, p.Volume)
		}
	}
}
