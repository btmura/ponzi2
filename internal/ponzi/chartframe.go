package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (ch *chart) renderFrame(r image.Rectangle) []image.Rectangle {
	if ch == nil {
		return nil
	}

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
		c = c.Add(ch.symbolQuoteText.render(ch.stock.symbol, c))
		c = c.Add(ch.symbolQuoteText.render(formatQuote(ch.stock.quote), c))
	}
	pt.Y -= pad

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
