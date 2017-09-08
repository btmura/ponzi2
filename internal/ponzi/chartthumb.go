package ponzi

import (
	"bytes"
	"fmt"
	"image"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	thumbChartRounding = 6
	thumbChartPadding  = 2
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbFormatQuote             = func(q *modelQuote) string {
		if q.price != 0 {
			return fmt.Sprintf(" %.2f %+5.2f%% ", q.price, q.percentChange*100.0)
		}
		return ""
	}
	thumbRemoveButtonVAO = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("removeButton.ply")))
)

type chartThumbnail struct {
	stock             *modelStock
	header            *chartHeader
	lines             *chartLines
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func newChartThumbnail(stock *modelStock) *chartThumbnail {
	return &chartThumbnail{
		stock:             stock,
		header:            newChartHeader(stock, thumbSymbolQuoteTextRenderer, thumbFormatQuote, newButton(thumbRemoveButtonVAO), thumbChartRounding, thumbChartPadding),
		lines:             newChartLines(stock),
		dailyStochastics:  newChartStochastics(stock, dailySTO),
		weeklyStochastics: newChartStochastics(stock, weeklySTO),
	}
}

func (ch *chartThumbnail) update() {
	ch.lines.update()
	ch.dailyStochastics.update()
	ch.weeklyStochastics.update()
}

func (ch *chartThumbnail) render(r image.Rectangle) {
	r = ch.header.render(r)

	rects := renderHorizDividers(r, horizLine, 0.5, 0.5)

	ch.dailyStochastics.render(rects[1].Inset(thumbChartPadding))
	ch.weeklyStochastics.render(rects[0].Inset(thumbChartPadding))
	ch.lines.render(rects[1].Inset(thumbChartPadding))
	ch.lines.render(rects[0].Inset(thumbChartPadding))
}

func (ch *chartThumbnail) close() {
	ch.lines.close()
	ch.lines = nil
	ch.dailyStochastics.close()
	ch.dailyStochastics = nil
	ch.weeklyStochastics.close()
	ch.weeklyStochastics = nil
}
