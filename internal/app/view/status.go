package view

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
}

func join(a ...string) string {
	var b []string
	for _, s := range a {
		if s != "" {
			b = append(b, s)
		}
	}
	return strings.Join(b, " ")
}

func paren(a string) string {
	if a == "" {
		return ""
	}
	return fmt.Sprintf("(%s)", a)
}

func priceStatus(q *model.Quote) string {
	if q == nil {
		return ""
	}
	return fmt.Sprintf("%.2f %+5.2f (%+5.2f%%)", q.LatestPrice, q.Change, q.ChangePercent*100)
}

func updateStatus(q *model.Quote) string {
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
