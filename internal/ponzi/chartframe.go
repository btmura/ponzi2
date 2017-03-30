package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartFrame struct {
	stock           *modelStock
	symbolQuoteText *dynamicText
	frameBorder     *vao
	frameDivider    *vao
}

func createChartFrame(stock *modelStock, symbolQuoteText *dynamicText) *chartFrame {
	return &chartFrame{
		stock:           stock,
		symbolQuoteText: symbolQuoteText,
		frameBorder:     createStrokedRectVAO(white, white, white, white),
		frameDivider:    createLineVAO(white, white),
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
	ch.frameBorder.render()

	//
	// Render the symbol and its quote.
	//

	const pad = 10
	s := ch.symbolQuoteText.measure(ch.stock.symbol)
	pt.Y -= pad + s.Y
	{
		c := pt
		c.X += pad
		c = c.Add(ch.symbolQuoteText.render(ch.stock.symbol, c, white))
		c = c.Add(ch.symbolQuoteText.render(formatQuote(ch.stock.quote), c, quoteColor(ch.stock.quote)))
	}

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = pt.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13, 0.6)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		ch.frameDivider.render()
	}
	return rects
}

func (ch *chartFrame) close() {
	ch.frameDivider.close()
	ch.frameDivider = nil
	ch.frameBorder.close()
	ch.frameBorder = nil
}
