package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartHeader struct {
	stock                   *modelStock
	symbolQuoteTextRenderer *gfx.TextRenderer
	quoteFormatter          func(*modelQuote) string
	button                  *button
	roundAmount             int
	padding                 int
}

func newChartHeader(stock *modelStock, symbolQuoteTextRenderer *gfx.TextRenderer, quoteFormatter func(*modelQuote) string, button *button, roundAmount, padding int) *chartHeader {
	return &chartHeader{
		stock: stock,
		symbolQuoteTextRenderer: symbolQuoteTextRenderer,
		quoteFormatter:          quoteFormatter,
		button:                  button,
		roundAmount:             roundAmount,
		padding:                 padding,
	}
}

func (ch *chartHeader) render(vc viewContext) (body image.Rectangle) {
	// Render the border around the chart.
	r := vc.bounds
	renderRoundedRect(r, ch.roundAmount)

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= ch.padding + ch.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += ch.roundAmount
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.stock.symbol, pt, white)
		pt.X += ch.padding
		pt.X += ch.symbolQuoteTextRenderer.Render(ch.quoteFormatter(ch.stock.quote), pt, quoteColor(ch.stock.quote))
	}
	pt.Y -= ch.padding

	// Render button in the upper right corner.
	buttonSize := image.Pt(r.Max.Y-pt.Y, r.Max.Y-pt.Y)
	vc.bounds = image.Rectangle{r.Max.Sub(buttonSize), r.Max}
	ch.button.render(vc)

	r.Max.Y = pt.Y
	return r
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
