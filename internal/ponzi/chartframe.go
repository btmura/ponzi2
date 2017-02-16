package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartFrame struct {
	symbolQuoteText *dynamicText
	border          *vaoLine
	divider         *vaoLine
}

func createChartFrame(symbolQuoteText *dynamicText) *chartFrame {
	return &chartFrame{
		symbolQuoteText: symbolQuoteText,
		border: createVAOLine(
			[]float32{
				-1, 1,
				-1, -1,
				1, -1,
				1, 1,
			},
			[]float32{
				blue[0], blue[1], blue[2],
				blue[0], blue[1], blue[2],
				blue[0], blue[1], blue[2],
				blue[0], blue[1], blue[2],
			},
			[]uint16{
				0, 1,
				1, 2,
				2, 3,
				3, 0,
			},
		),
		divider: createVAOLineSegment(blue, blue),
	}
}

func (f *chartFrame) render(stock *modelStock, r image.Rectangle) []image.Rectangle {
	// Start rendering from the top left. Track position with point.
	c := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	f.border.render(r)

	//
	// Render the symbol and its quote.
	//

	const p = 10
	s := f.symbolQuoteText.measure(stock.symbol)
	c.Y -= p + s.Y
	{
		c := c
		c.X += p
		c = c.Add(f.symbolQuoteText.render(stock.symbol, c))
		c = c.Add(f.symbolQuoteText.render(formatQuote(stock.quote), c))
	}
	c.Y -= p

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = c.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13, 0.6)
	for _, r := range rects {
		f.divider.render(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
	}
	return rects
}

func (f *chartFrame) close() {
	f.divider.close()
	f.border.close()
}
