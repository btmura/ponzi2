package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartFrame struct {
	propText *dynamicText
	vao      uint32
	count    int32
}

func createChartFrame(propText *dynamicText) *chartFrame {
	vertices := []float32{
		-1, 1,
		-1, -1,
		1, -1,
		1, 1,
	}

	colors := []float32{
		0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,
	}

	indices := []uint16{
		0, 1,
		1, 2,
		2, 3,
		3, 0,
	}

	vbo := createArrayBuffer(vertices)
	cbo := createArrayBuffer(colors)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
		gl.EnableVertexAttribArray(colorLocation)
		gl.VertexAttribPointer(colorLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	}
	gl.BindVertexArray(0)

	return &chartFrame{
		propText: propText,
		vao:      vao,
		count:    int32(len(indices)),
	}
}

func (f *chartFrame) render(stock *modelStock, r image.Rectangle) {
	//
	// Render the frame around the chart.
	//

	setModelMatrixRectangle(r)
	gl.Uniform1f(colorMixAmountLocation, 1)

	gl.BindVertexArray(f.vao)
	gl.DrawElements(gl.LINES, f.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)

	//
	// Render the symbol and its quote.
	//

	s := f.propText.measure(stock.symbol)
	r.Max.Y -= s.Y
	c := image.Pt(r.Min.X, r.Max.Y)
	c = c.Add(f.propText.render(stock.symbol, c))
	c = c.Add(f.propText.render(formatQuote(stock.quote), c))
}

func (f *chartFrame) close() {
	gl.DeleteVertexArrays(1, &f.vao)
}
