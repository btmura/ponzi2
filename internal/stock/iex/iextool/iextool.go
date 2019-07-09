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

		line := ""
		fmt.Scanln(&line)

		line = strings.ToUpper(line)

		req := &iex.GetQuotesRequest{
			Symbols: strings.Split(line, ","),
		}

		fmt.Printf("Quotes: %+v\n", req)

		quotes, err := c.GetQuotes(ctx, req)
		if err != nil {
			fmt.Printf("GetQuotes: %v\n", err)
			continue
		}

		for i, q := range quotes {
			fmt.Printf("%d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, q.Symbol, q.Open, q.High, q.Low, q.Close, q.LatestVolume)
		}
	}
}
