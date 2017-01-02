package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartStochastics struct {
	// lineVAO is the VAO for the K and D lines.
	lineVAO uint32

	// lineCount is the number of elements to draw for lineVAO.
	lineCount int32

	// backgroundVAO is the VAO for the background behind the lines.
	backgroundVAO uint32

	// backgroundCount is the number of elements to draw for the background.
	backgroundCount int32
}

func createChartStochastics(ss []*modelTradingSession, dColor [3]float32) *chartStochastics {
	// Calculate vertices and indices for the stochastics.
	var vertices []float32
	var colors []float32
	var indices []uint16

	width := 2.0 / float32(len(ss)) // (-1 to 1) on X-axis
	calcX := func(i int) float32 {
		return -1.0 + width*0.5 + width*float32(i)
	}
	calcY := func(value float32) float32 {
		return 2*float32(value) - 1
	}

	var v uint16 // vertex index

	// Add vertices and indices for d percent lines.
	first := true
	for i, s := range ss {
		if s.d == 0.0 {
			continue
		}

		vertices = append(vertices, calcX(i), calcY(s.d))
		colors = append(colors, dColor[0], dColor[1], dColor[2])
		if !first {
			indices = append(indices, v, v-1)
		}
		v++
		first = false
	}

	// Add vertices and indices for k percent lines.
	first = true
	for i, s := range ss {
		if s.k == 0.0 {
			continue
		}

		vertices = append(vertices, calcX(i), calcY(s.k))
		colors = append(colors, 1, 0, 0)
		if !first {
			indices = append(indices, v, v-1)
		}
		v++
		first = false
	}

	// Can't create empty buffer objects. Bail out if nothing to render.
	if len(vertices) == 0 {
		return nil
	}

	vbo := createArrayBuffer(vertices)
	cbo := createArrayBuffer(colors)
	ibo := createElementArrayBuffer(indices)

	// lineVAO is the VAO for the K and D lines.
	var lineVAO uint32
	// lineVAO is the VAO for the K and D lines.
	gl.GenVertexArrays(1, &lineVAO)
	// lineVAO is the VAO for the K and D lines.
	gl.BindVertexArray(lineVAO)
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

	backgroundVAO, backgroundCount := createChartBackgroundVAO(ss, dColor)

	return &chartStochastics{
		lineVAO:         lineVAO,
		lineCount:       int32(len(indices)),
		backgroundVAO:   backgroundVAO,
		backgroundCount: backgroundCount,
	}
}

func createChartBackgroundVAO(ss []*modelTradingSession, dColor [3]float32) (uint32, int32) {
	vertices := []float32{
		-1, 1, // UL - 0
		-1, -1, // BL - 1
		1, -1, // BR - 2
		1, 1, // UR -3
	}

	colors := []float32{
		0, 0, 0,
		0, 0, 0,
		0, 0.25, 0.5,
		0, 0, 0,
	}

	indices := []uint16{
		0, 1, 3,
		1, 2, 3,
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

	return vao, int32(len(indices))
}

func (s *chartStochastics) render(r image.Rectangle) {
	if s == nil {
		return
	}

	setModelMatrixRectangle(r)

	gl.BindVertexArray(s.lineVAO)
	gl.DrawElements(gl.LINES, s.lineCount, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)

	gl.BindVertexArray(s.backgroundVAO)
	gl.DrawElements(gl.TRIANGLES, s.backgroundCount, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (s *chartStochastics) close() {
	if s == nil {
		return
	}

	gl.DeleteVertexArrays(1, &s.lineVAO)
	gl.DeleteVertexArrays(1, &s.backgroundVAO)
}
