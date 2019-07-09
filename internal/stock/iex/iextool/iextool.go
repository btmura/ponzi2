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
	token            = flag.String("token", "", "API token required on requests.")
	dumpAPIResponses = flag.Bool("dump_api_responses", false, "Dump API responses to txt files.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	c := iex.NewCacheClient(iex.NewClient(*token, *dumpAPIResponses))

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
			log.Fatalf("GetQuotes: %v", err)
		}

		for i, q := range quotes {
			fmt.Printf("%d: %s Open: %.2f High: %.2f Low: %.2f Close: %.2f Volume: %d\n",
				i, q.Symbol, q.Open, q.High, q.Low, q.Close, q.LatestVolume)
		}
	}
}
