package ponzi

import (
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
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	// Render the border around the chart.
	const roundAmount = 10
	renderRoundedRect(r, roundAmount)

	//
	// Render the symbol and quote.
	//

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)

	const pad = 5
	pt.Y -= pad + symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += roundAmount
		pt.X += symbolQuoteTextRenderer.Render(ch.stock.symbol, pt, white)
		pt.X += pad
		pt.X += symbolQuoteTextRenderer.Render(formatQuote(ch.stock.quote), pt, quoteColor(ch.stock.quote))
	}
	pt.Y -= pad

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gfx.SetColorMixAmount(1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13)
	for _, r := range rects {
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		horizLine.Render()
	}
	return rects
}

func (ch *chartFrame) close() {}
