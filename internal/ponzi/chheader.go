package ponzi

import (
	"bytes"
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

var addButtonVAO = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("addButton.ply")))

type chartHeader struct {
	stock                   *modelStock
	symbolQuoteTextRenderer *gfx.TextRenderer
	quoteFormatter          func(*modelQuote) string
	roundAmount             int
	padding                 int
}

func newChartHeader(stock *modelStock, symbolQuoteTextRenderer *gfx.TextRenderer, quoteFormatter func(*modelQuote) string, roundAmount, padding int) *chartHeader {
	return &chartHeader{
		stock: stock,
		symbolQuoteTextRenderer: symbolQuoteTextRenderer,
		quoteFormatter:          quoteFormatter,
		roundAmount:             roundAmount,
		padding:                 padding,
	}
}

func (c *chartHeader) render(r image.Rectangle) (body image.Rectangle) {
	// Render the border around the chart.
	renderRoundedRect(r, c.roundAmount)

	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)
	pt.Y -= c.padding + c.symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += c.roundAmount
		pt.X += c.symbolQuoteTextRenderer.Render(c.stock.symbol, pt, white)
		pt.X += c.padding
		pt.X += c.symbolQuoteTextRenderer.Render(c.quoteFormatter(c.stock.quote), pt, quoteColor(c.stock.quote))
	}
	pt.Y -= c.padding

	// Render button in the upper right corner.
	buttonSize := image.Pt(r.Max.Y-pt.Y, r.Max.Y-pt.Y)
	buttonRect := image.Rectangle{Min: r.Max.Sub(buttonSize), Max: r.Max}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(buttonRect)
	addButtonVAO.Render()

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
