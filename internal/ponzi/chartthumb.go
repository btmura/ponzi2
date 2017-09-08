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

func createChartThumbnail(stock *modelStock) *chartThumbnail {
	return &chartThumbnail{
		stock:             stock,
		header:            newChartHeader(stock, thumbSymbolQuoteTextRenderer, thumbFormatQuote, newButton(thumbRemoveButtonVAO), thumbChartRounding, thumbChartPadding),
		lines:             createChartLines(stock),
		dailyStochastics:  createChartStochastics(stock, dailySTO),
		weeklyStochastics: createChartStochastics(stock, weeklySTO),
	}
}

func (ct *chartThumbnail) update() {
	ct.lines.update()
	ct.dailyStochastics.update()
	ct.weeklyStochastics.update()
}

func (ct *chartThumbnail) render(r image.Rectangle) {
	r = ct.header.render(r)

	rects := renderHorizDividers(r, horizLine, 0.5, 0.5)

	ct.dailyStochastics.render(rects[1].Inset(thumbChartPadding))
	ct.weeklyStochastics.render(rects[0].Inset(thumbChartPadding))
	ct.lines.render(rects[1].Inset(thumbChartPadding))
	ct.lines.render(rects[0].Inset(thumbChartPadding))
}

func (ct *chartThumbnail) close() {
	ct.lines.close()
	ct.lines = nil
	ct.dailyStochastics.close()
	ct.dailyStochastics = nil
	ct.weeklyStochastics.close()
	ct.weeklyStochastics = nil
}
