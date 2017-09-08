package ponzi

import (
	"fmt"
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	chartRounding = 10
	chartPadding  = 5
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartFormatQuote             = func(q *modelQuote) string {
		if q.price != 0 {
			return fmt.Sprintf("%.2f %+5.2f %+5.2f%%", q.price, q.change, q.percentChange*100.0)
		}
		return ""
	}
)

// Shared variables used by multiple chart components.
var (
	chartAxisLabelTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	chartGridHorizLine         = gfx.HorizColoredLineVAO(gray, gray)
)

type chart struct {
	stock             *modelStock
	header            *chartHeader
	lines             *chartLines
	prices            *chartPrices
	volume            *chartVolume
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createChart(stock *modelStock) *chart {
	return &chart{
		stock:             stock,
		header:            newChartHeader(stock, chartSymbolQuoteTextRenderer, chartFormatQuote, &button{}, chartRounding, chartPadding),
		lines:             createChartLines(stock),
		prices:            createChartPrices(stock),
		volume:            createChartVolume(stock),
		dailyStochastics:  createChartStochastics(stock, dailySTO),
		weeklyStochastics: createChartStochastics(stock, weeklySTO),
	}
}

func (ch *chart) update() {
	ch.lines.update()
	ch.prices.update()
	ch.volume.update()
	ch.dailyStochastics.update()
	ch.weeklyStochastics.update()
}

func (ch *chart) render(r image.Rectangle) {
	r = ch.header.render(r)

	rects := renderHorizDividers(r, horizLine, 0.13, 0.13, 0.13, 0.61)
	pr, vr, dr, wr := rects[3], rects[2], rects[1], rects[0]

	pr = pr.Inset(chartPadding)
	vr = vr.Inset(chartPadding)
	dr = dr.Inset(chartPadding)
	wr = wr.Inset(chartPadding)

	maxWidth := ch.prices.renderLabels(pr)
	if w := ch.volume.renderLabels(vr); w > maxWidth {
		maxWidth = w
	}
	if w := ch.dailyStochastics.renderLabels(dr); w > maxWidth {
		maxWidth = w
	}
	if w := ch.weeklyStochastics.renderLabels(wr); w > maxWidth {
		maxWidth = w
	}

	pr.Max.X -= maxWidth + chartPadding
	vr.Max.X -= maxWidth + chartPadding
	dr.Max.X -= maxWidth + chartPadding
	wr.Max.X -= maxWidth + chartPadding

	ch.prices.render(pr)
	ch.volume.render(vr)
	ch.dailyStochastics.render(dr)
	ch.weeklyStochastics.render(wr)

	ch.lines.render(pr)
	ch.lines.render(vr)
	ch.lines.render(dr)
	ch.lines.render(wr)
}

func (ch *chart) close() {
	ch.prices.close()
	ch.prices = nil
	ch.volume.close()
	ch.volume = nil
	ch.dailyStochastics.close()
	ch.dailyStochastics = nil
	ch.weeklyStochastics.close()
	ch.weeklyStochastics = nil
}
