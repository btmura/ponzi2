package ponzi

import (
	"image"
	"math"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chart struct {
	vao   uint32
	count int32
}

func createChart(sessions []*modelTradingSession) *chart {
	// Find the min and max prices for the y-axis.
	min := float32(math.MaxFloat32)
	max := float32(0)
	for _, s := range sessions {
		if s.low < min {
			min = s.low
		}
		if s.high > max {
			max = s.high
		}
	}

	// Calculate vertices and indices for the candlesticks.
	var vertices []float32
	var indices []uint16

	stickWidth := 2.0 / float32(len(sessions)) // (-1 to 1) on X-axis
	leftX := -1.0 + stickWidth*0.1
	midX := -1.0 + stickWidth*0.5
	rightX := -1.0 + stickWidth*0.9

	calcY := func(value float32) float32 {
		return 2*(value-min)/(max-min) - 1
	}

	for i, s := range sessions {
		// Figure out Y coordinates of the key levels.
		lowY, highY, openY, closeY := calcY(s.low), calcY(s.high), calcY(s.open), calcY(s.close)

		// Figure out the top and bottom of the candlestick.
		topY, botY := openY, closeY
		if openY < closeY {
			topY, botY = closeY, openY
		}

		// Add the vertices needed to create the candlestick.
		vertices = append(vertices,
			midX, highY, // 0
			midX, topY, // 1
			midX, lowY, // 2
			midX, botY, // 3
			leftX, topY, // 4 - Upper left of box
			rightX, topY, // 5 - Upper right of box
			leftX, botY, // 6 - Bottom left of box
			rightX, botY, // 7 - Bottom right of box
		)

		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*8 + j
		}

		// Add the vertex indices to render the candlestick.
		indices = append(indices,
			// Top and bottom lines around the box.
			idx(0), idx(1),
			idx(2), idx(3),

			// Lines for the box.
			idx(4), idx(5),
			idx(6), idx(7),
			idx(4), idx(6),
			idx(5), idx(7),
		)

		// Move the X coordinates one stick over.
		leftX += stickWidth
		midX += stickWidth
		rightX += stickWidth
	}

	vbo := createArrayBuffer(vertices)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.EnableVertexAttribArray(positionLocation)
	gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BindVertexArray(0)

	return &chart{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (c *chart) render(r image.Rectangle) {
	m := newScaleMatrix(float32(r.Dx()/2), float32(r.Dy()/2), 1)
	m = m.mult(newTranslationMatrix(float32(r.Min.X+r.Dx()/2), float32(r.Min.Y+r.Dy()/2), 0))
	gl.UniformMatrix4fv(modelMatrixLocation, 1, false, &m[0])
	gl.BindTexture(gl.TEXTURE_2D, 0 /* dummy texture */)

	gl.BindVertexArray(c.vao)
	gl.DrawElements(gl.LINES, c.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}
