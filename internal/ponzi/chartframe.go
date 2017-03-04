package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartFrame struct {
	stock     *modelStock
	titleText *dynamicText
	border    *vao
	divider   *vao
}

func createChartFrame(stock *modelStock, titleText *dynamicText) *chartFrame {
	return &chartFrame{
		stock:     stock,
		titleText: titleText,
		border:    createStrokedRectVAO(white, white, white, white),
		divider:   createLineVAO(white, white),
	}
}

func (f *chartFrame) render(r image.Rectangle) []image.Rectangle {
	// Start rendering from the top left. Track position with point.
	c := image.Pt(r.Min.X, r.Max.Y)

	//
	// Render the frame around the chart.
	//

	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	f.border.render()

	//
	// Render the symbol and its quote.
	//

	const p = 10
	s := f.titleText.measure(f.stock.symbol)
	c.Y -= p + s.Y
	{
		c := c
		c.X += p
		c = c.Add(f.titleText.render(f.stock.symbol, c))
		c = c.Add(f.titleText.render(formatQuote(f.stock.quote), c))
	}
	c.Y -= p

	//
	// Render the dividers between the sections.
	//

	r.Max.Y = c.Y
	gl.Uniform1f(colorMixAmountLocation, 1)

	rects := sliceRectangle(r, 0.13, 0.13, 0.13, 0.6)
	for _, r := range rects {
		setModelMatrixRectangle(image.Rect(r.Min.X, r.Max.Y, r.Max.X, r.Max.Y))
		f.divider.render()
	}
	return rects
}

func (f *chartFrame) close() {
	f.divider.close()
	f.border.close()
}
