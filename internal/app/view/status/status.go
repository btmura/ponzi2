// Package status has functions to build a stock's status line with the symbol, price, and more.
package status

import (
	"fmt"
	"strings"

	"github.com/btmura/ponzi2/internal/app/model"
)

var displaySources = map[model.Source]string{
	model.IEXRealTimePrice:          "IEX Real Time",
	model.FifteenMinuteDelayedPrice: "15 Min Delayed",
	model.Close:                     "Close",
	model.PreviousClose:             "Previous Close",
	model.IEXPrice:                  "IEX Price",
	model.IEXLastTrade:              "IEX Last Trade",
}

// Join combines the non-empty strings in the slice together with spaces.
func Join(a ...string) string {
	var b []string
	for _, s := range a {
		if s != "" {
			b = append(b, s)
		}
	}
	return strings.Join(b, " ")
}

// Paren parethesizes a string if it is not empty.
func Paren(a string) string {
	if a == "" {
		return ""
	}
	return fmt.Sprintf("(%s)", a)
}

// PriceChange returns a status line with the quote's price information.
func PriceChange(q *model.Quote) string {
	if q == nil {
		return ""
	}
	return fmt.Sprintf("%.2f %+5.2f (%+5.2f%%)", q.LatestPrice, q.Change, q.ChangePercent*100)
}

// SourceUpdate returns a status line with the quote's source and update time information.
func SourceUpdate(q *model.Quote) string {
	if q == nil {
		return ""
	}

	ds, ok := displaySources[q.LatestSource]
	if !ok {
		ds = "?"
	}

	l := "1/2/2006"
	if q.LatestUpdate.Hour() != 0 || q.LatestUpdate.Minute() != 0 || q.LatestUpdate.Second() != 0 || q.LatestUpdate.Nanosecond() != 0 {
		l += " 3:04 PM"
	}

	return fmt.Sprintf("%s %s", ds, q.LatestUpdate.Format(l))
}
