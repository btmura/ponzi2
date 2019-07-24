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
	"time"

	"github.com/btmura/ponzi2/internal/stock/iex"
	"github.com/btmura/ponzi2/internal/stock/iexcache"
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

	c, err := iexcache.NewClient(iex.NewClient(*token, *dumpAPIResponses))
	if err != nil {
		log.Fatal(err)
	}

	for {
		expvar.Do(func(kv expvar.KeyValue) {
			if strings.HasPrefix(kv.Key, "iex") {
				fmt.Println(kv)
			}
		})

		fmt.Println()

		fmt.Printf("Enter comma-separated symbols: ")

		symbolLine := ""
		fmt.Scanln(&symbolLine)
		symbolLine = strings.ToUpper(symbolLine)

		dataRange := pick("Pick a range", iex.TwoYears, iex.OneDay).(iex.Range)

		qReq := &iex.GetQuotesRequest{
			Symbols: strings.Split(symbolLine, ","),
		}

		fmt.Printf("Quotes: %+v\n\n", qReq)

		quotes, err := c.GetQuotes(ctx, qReq)
		if err != nil {
			fmt.Println(err)
			continue
		}

		formatTime := func(t time.Time) string {
			return t.Format("1/2/06 03:04 AM")
		}

		for i, q := range quotes {
			fmt.Printf("%d: %q %q LP: %.2f LS: %v LT: %s, LU: %s, LV: %d "+
				"O: %.2f H: %.2f L: %.2f C: %.2f CH: %.2f CHP: %.2f\n",
				i,
				q.Symbol,
				q.CompanyName,
				q.LatestPrice,
				q.LatestSource,
				formatTime(q.LatestTime),
				formatTime(q.LatestUpdate),
				q.LatestVolume,
				q.Open,
				q.High,
				q.Low,
				q.Close,
				q.Change,
				q.ChangePercent)
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

		const last = 10

		for i, ch := range charts {
			fmt.Printf("%d: %s\n", i, ch.Symbol)
			for j, p := range ch.ChartPoints {
				if j >= len(ch.ChartPoints)-last {
					fmt.Printf("\t%d: %s O: %.2f H: %.2f L: %.2f C: %.2f V: %d CH: %.2f CHP: %.2f\n",
						j,
						formatTime(p.Date),
						p.Open,
						p.High,
						p.Low,
						p.Close,
						p.Volume,
						p.Change,
						p.ChangePercent)
				}
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
