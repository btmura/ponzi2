package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartVolume struct {
	vao   uint32
	count int32
}

func createChartVolume(ss []*modelTradingSession) *chartVolume {
	// Find the max volume for the y-axis.
	var max int
	for _, s := range ss {
		if s.volume > max {
			max = s.volume
		}
	}

	// Calculate vertices and indices for the volume bars.
	var vertices []float32
	var colors []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(ss)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.1
	rightX := -1.0 + barWidth*0.9

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(max) - 1
	}

	for i, s := range ss {
		topY := calcY(s.volume)
		botY := calcY(0)

		// Add the vertices needed to create the volume bar.
		vertices = append(vertices,
			leftX, topY, // UL
			rightX, topY, // UR
			leftX, botY, // BL
			rightX, botY, // BR
		)

		// Add the colors corresponding to the volume bar.
		switch {
		case s.close > s.open:
			colors = append(colors,
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
			)

		case s.close < s.open:
			colors = append(colors,
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
			)

		default:
			colors = append(colors,
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
				yellow[0], yellow[1], yellow[2],
			)
		}

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*4 + j
		}

		// Use triangles for filled candlestick on lower closes.
		indices = append(indices,
			idx(0), idx(2), idx(1),
			idx(1), idx(2), idx(3),
		)

		// Move the X coordinates one bar over.
		leftX += barWidth
		rightX += barWidth
	}

	// Can't create empty buffer objects. Bail out if nothing to render.
	if len(vertices) == 0 {
		return nil
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

	return &chartVolume{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (v *chartVolume) render(r image.Rectangle) {
	if v == nil {
		return
	}

	setModelMatrixRectangle(r)

	gl.BindVertexArray(v.vao)
	gl.DrawElements(gl.TRIANGLES, v.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (v *chartVolume) close() {
	if v == nil {
		return
	}

	gl.DeleteVertexArrays(1, &v.vao)
}
