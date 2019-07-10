// The iextool command prints stock data for a list of stock symbols.
package main

import (
	"context"
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	port             = flag.Int("port", 9000, "Port number to export metrics.")
	token            = flag.String("token", "", "API token required on requests.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	fmt.Printf("Serving metrics at http://localhost:%d/debug/vars\n", *port)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
			fmt.Printf("http.ListenAndServe failed: %v\n", err)
		}
	}()

	ctx := context.Background()

	c, err := iex.NewCacheClient(iex.NewClient(*token, *dumpAPIResponses))
	if err != nil {
		log.Fatal(err)
	}

	for {
		fmt.Printf("Enter comma-separated symbols: ")

		symbolLine := ""
		fmt.Scanln(&symbolLine)
		symbolLine = strings.ToUpper(symbolLine)

		rangePick := 1 // 0 is unspecified.
		for {
			fmt.Println("Pick a range:")
			for i := 1; i < 3; i++ {
				fmt.Printf("[%d] %v\n", i, iex.Range(i))
			}

			if _, err := fmt.Scanf("%d", &rangePick); err != nil {
				fmt.Println(err)
				continue
			}

			break
		}

		qReq := &iex.GetQuotesRequest{
			Symbols: strings.Split(symbolLine, ","),
		}

		fmt.Printf("Quotes: %+v\n", qReq)

		quotes, err := c.GetQuotes(ctx, qReq)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for i, q := range quotes {
			fmt.Printf("%d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, q.Symbol, q.Open, q.High, q.Low, q.Close, q.LatestVolume)
		}

		cReq := &iex.GetChartsRequest{
			Symbols:   strings.Split(symbolLine, ","),
			Range:     iex.Range(rangePick),
			ChartLast: 3,
		}

		fmt.Printf("Charts: %+v\n", cReq)

		charts, err := c.GetCharts(ctx, cReq)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for i, ch := range charts {
			fmt.Printf("%d: %s\n", i, ch.Symbol)
			for j, p := range ch.ChartPoints {
				fmt.Printf("\t%d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
					j, ch.Symbol, p.Open, p.High, p.Low, p.Close, p.Volume)
			}
		}
	}
}
