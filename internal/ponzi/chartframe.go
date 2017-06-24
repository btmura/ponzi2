package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartFrame struct {
	stock           *modelStock
	symbolQuoteText *gfx.DynamicText
	buttonRenderer  *buttonRenderer
	frameBorder     *gfx.VAO
	frameDivider    *gfx.VAO
}

func createChartFrame(stock *modelStock, symbolQuoteText *gfx.DynamicText, br *buttonRenderer) *chartFrame {
	return &chartFrame{
		stock:           stock,
		symbolQuoteText: symbolQuoteText,
		buttonRenderer:  br,
		frameBorder:     gfx.CreateStrokedRectVAO(white, white, white, white),
		frameDivider:    gfx.CreateLineVAO(white, white),
	}
}

func (ch *chartFrame) render(r image.Rectangle) []image.Rectangle {
	// Start rendering from the top left. Track position with point.
	pt := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.frameBorder.Render()

	//
	// Render the symbol, quote, and add button.
	//

	const pad = 10
	sz := ch.symbolQuoteText.Measure(ch.stock.symbol)
	pt.Y -= pad + sz.Y
	{
		pt := pt
		pt.X += pad
		pt = pt.Add(ch.symbolQuoteText.Render(ch.stock.symbol, pt, white))
		pt = pt.Add(ch.symbolQuoteText.Render(formatQuote(ch.stock.quote), pt, quoteColor(ch.stock.quote)))
	}

	{
		barHeight := pad*2 + sz.Y
		sz := image.Pt(barHeight-pad*2, barHeight-pad*2)
		pt := pt
		pt.X = r.Max.X - pad - sz.X
		pt.Y = r.Max.Y - pad - sz.Y
		ch.buttonRenderer.render(pt, sz, addButtonIcon)
	}

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13, 0.6)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		ch.frameDivider.Render()
	}
	return rects
}

func (ch *chartFrame) close() {
	ch.frameDivider.Close()
	ch.frameDivider = nil
	ch.frameBorder.Close()
	ch.frameBorder = nil
}
