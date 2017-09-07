package ponzi

import (
	"fmt"
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartFrame struct {
	stock *modelStock
}

func createChartFrame(stock *modelStock) *chartFrame {
	return &chartFrame{
		stock: stock,
	}
}

func (ch *chartFrame) render(r image.Rectangle) []image.Rectangle {
	r = renderChartFrame(r, ch.stock, symbolQuoteTextRenderer, mainChartRounding, mainChartPadding)

	gfx.SetColorMixAmount(1)
	rects := sliceRect(r, 0.13, 0.13, 0.13)
	for _, r := range rects {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		horizLine.Render()
	}
	return rects
}

func (ch *chartFrame) close() {}

func renderChartFrame(r image.Rectangle, stock *modelStock, symbolQuoteTextRenderer *gfx.TextRenderer, roundAmount, padding int) (body image.Rectangle) {
	// Render the border around the chart.
	renderRoundedRect(r, roundAmount)

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= padding + symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += roundAmount
		pt.X += symbolQuoteTextRenderer.Render(stock.symbol, pt, white)
		pt.X += padding
		pt.X += symbolQuoteTextRenderer.Render(formatQuote(stock.quote), pt, quoteColor(stock.quote))
	}
	pt.Y -= padding

	r.Max.Y = pt.Y
	return r
}

func formatQuote(q *modelQuote) string {
	if q.price != 0 {
		return fmt.Sprintf("%.2f %+5.2f %+5.2f%%", q.price, q.change, q.percentChange*100.0)
	}
	return ""
}

func shortFormatQuote(q *modelQuote) string {
	if q.price != 0 {
		return fmt.Sprintf(" %.2f %+5.2f%% ", q.price, q.percentChange*100.0)
	}
	return ""
}

func quoteColor(q *modelQuote) [3]float32 {
	switch {
	case q.percentChange > 0:
		return green

	case q.percentChange < 0:
		return red
	}
	return white
}
