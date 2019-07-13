// The iextool command prints stock data for a list of stock symbols.
package main

import (
	"context"
	"expvar"
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"strconv"
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

	// fmt.Printf("Serving metrics at http://localhost:%d/debug/vars\n", *port)
	// go func() {
	// 	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
	// 		fmt.Printf("http.ListenAndServe failed: %v\n", err)
	// 	}
	// }()

	ctx := context.Background()

	c, err := iex.NewCacheClient(iex.NewClient(*token, *dumpAPIResponses))
	if err != nil {
		log.Fatal(err)
	}

	for {
		expvar.Do(func(kv expvar.KeyValue) {
			if !strings.HasPrefix(kv.Key, "iex-") {
				return
			}
			fmt.Println(kv)
		})

		fmt.Println()

		fmt.Printf("Enter comma-separated symbols: ")

		symbolLine := ""
		fmt.Scanln(&symbolLine)
		symbolLine = strings.ToUpper(symbolLine)

		dataRange := pick("Pick a range", iex.OneDay, iex.TwoYears).(iex.Range)

		qReq := &iex.GetQuotesRequest{
			Symbols: strings.Split(symbolLine, ","),
		}

		fmt.Printf("Quotes: %+v\n\n", qReq)

		quotes, err := c.GetQuotes(ctx, qReq)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for i, q := range quotes {
			fmt.Printf("%d: %-4s O: %.2f H: %.2f L: %.2f C: %.2f V: %d\n",
				i, q.Symbol, q.Open, q.High, q.Low, q.Close, q.LatestVolume)
		}

		fmt.Println()

		cReq := &iex.GetChartsRequest{
			Symbols:   strings.Split(symbolLine, ","),
			Range:     dataRange,
			ChartLast: 3,
		}

		fmt.Printf("Charts: %+v\n\n", cReq)

		charts, err := c.GetCharts(ctx, cReq)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for i, ch := range charts {
			fmt.Printf("%d: %s\n", i, ch.Symbol)
			for j, p := range ch.ChartPoints {
				fmt.Printf("\t%d: D: %v O: %.2f H: %.2f L: %.2f C: %.2f V: %d\n",
					j, p.Date.Format("1/2/06 03:04 AM"), p.Open, p.High, p.Low, p.Close, p.Volume)
			}
		}

		fmt.Println()
	}
}

func pick(prompt string, choices ...interface{}) interface{} {
	selected := 0
	for {
		fmt.Printf("%s [%d]:\n", prompt, selected)
		for i, v := range choices {
			fmt.Printf("[%d] %v\n", i, v)
		}

		line := ""
		fmt.Scanln(&line)

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		i, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if i < 0 || i >= len(choices) {
			fmt.Printf("Enter a value from %d to %d\n", 0, len(choices)-1)
			continue
		}

		selected = i
		break
	}

	return choices[selected]
}
