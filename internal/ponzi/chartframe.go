package ponzi

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartFrame struct {
	stock        *modelStock
	frameDivider *gfx.VAO2
}

func createChartFrame(stock *modelStock) *chartFrame {
	return &chartFrame{
		stock:        stock,
		frameDivider: gfx.HorizColoredLineVAO(white, white),
	}
}

func (ch *chartFrame) render(r image.Rectangle) []image.Rectangle {
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)

	//
	// Render the symbol, quote, and add button.
	//

	const pad = 5
	pt.Y -= pad + symbolQuoteTextRenderer.LineHeight()
	{
		pt := pt
		pt.X += pad
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
		ch.frameDivider.Render()
	}
	return rects
}

func (ch *chartFrame) close() {
	ch.frameDivider.Delete()
	ch.frameDivider = nil
}
