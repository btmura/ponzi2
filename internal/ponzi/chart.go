package ponzi

import (
	"math"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartPrices struct {
	lineVAO       uint32
	lineCount     int32
	triangleVAO   uint32
	triangleCount int32
}

type chartVolume struct {
	vao   uint32
	count int32
}

type chartStochastics struct {
	vao   uint32
	count int32
}

func createChartPrices(ss []*modelTradingSession) *chartPrices {
	// Find the min and max prices for the y-axis.
	min := float32(math.MaxFloat32)
	max := float32(0)
	for _, s := range ss {
		if s.low < min {
			min = s.low
		}
		if s.high > max {
			max = s.high
		}
	}

	// Calculate vertices and indices for the candlesticks.
	var vertices []float32
	var colors []float32
	var lineIndices []uint16
	var triangleIndices []uint16

	stickWidth := 2.0 / float32(len(ss)) // (-1 to 1) on X-axis
	leftX := -1.0 + stickWidth*0.1
	midX := -1.0 + stickWidth*0.5
	rightX := -1.0 + stickWidth*0.9

	calcY := func(value float32) float32 {
		return 2*(value-min)/(max-min) - 1
	}

	for i, s := range ss {
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

		// Add the colors corresponding to the vertices.
		switch {
		case s.close > s.open:
			colors = append(colors,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
			)

		case s.close < s.open:
			colors = append(colors,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
			)

		default:
			colors = append(colors,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
			)
		}
		// idx is function to refer to the vertices above.
		idx := func(j uint16) uint16 {
			return uint16(i)*8 + j
		}

		// Add the vertex indices to render the candlestick.
		lineIndices = append(lineIndices,
			// Top and bottom lines around the box.
			idx(0), idx(1),
			idx(2), idx(3),
		)

		if s.close > s.open {
			// Use lines for open candlestick on higher closes.
			lineIndices = append(lineIndices,
				idx(4), idx(5),
				idx(6), idx(7),
				idx(4), idx(6),
				idx(5), idx(7),
			)
		} else {
			// Use triangles for filled candlestick on lower closes.
			triangleIndices = append(triangleIndices,
				idx(4), idx(6), idx(5),
				idx(5), idx(6), idx(7),
			)
		}

		// Move the X coordinates one stick over.
		leftX += stickWidth
		midX += stickWidth
		rightX += stickWidth
	}

	// Can't create empty buffer objects. Bail out if nothing to render.
	if len(vertices) == 0 {
		return nil
	}

	vbo := createArrayBuffer(vertices)
	cbo := createArrayBuffer(colors)
	lineIBO := createElementArrayBuffer(lineIndices)

	var lineVAO uint32
	gl.GenVertexArrays(1, &lineVAO)
	gl.BindVertexArray(lineVAO)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
		gl.EnableVertexAttribArray(colorLocation)
		gl.VertexAttribPointer(colorLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, lineIBO)
	}
	gl.BindVertexArray(0)

	var triangleVAO uint32
	if len(triangleIndices) != 0 {
		triangleIBO := createElementArrayBuffer(triangleIndices)
		gl.GenVertexArrays(1, &triangleVAO)
		gl.BindVertexArray(triangleVAO)
		{
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			gl.EnableVertexAttribArray(positionLocation)
			gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

			gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
			gl.EnableVertexAttribArray(colorLocation)
			gl.VertexAttribPointer(colorLocation, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, triangleIBO)
		}
		gl.BindVertexArray(0)
	}

	return &chartPrices{
		lineVAO:       lineVAO,
		lineCount:     int32(len(lineIndices)),
		triangleVAO:   triangleVAO,
		triangleCount: int32(len(triangleIndices)),
	}
}

func (p *chartPrices) render() {
	if p == nil {
		return
	}

	gl.BindVertexArray(p.lineVAO)
	gl.DrawElements(gl.LINES, p.lineCount, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)

	if p.triangleCount > 0 {
		gl.BindVertexArray(p.triangleVAO)
		gl.DrawElements(gl.TRIANGLES, p.triangleCount, gl.UNSIGNED_SHORT, gl.Ptr(nil))
		gl.BindVertexArray(0)
	}
}

func (p *chartPrices) close() {
	if p == nil {
		return
	}

	gl.DeleteVertexArrays(1, &p.lineVAO)
	if p.triangleCount > 0 {
		gl.DeleteVertexArrays(1, &p.triangleVAO)
	}
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
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
				0, 1, 0,
			)

		case s.close < s.open:
			colors = append(colors,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
				1, 0, 0,
			)

		default:
			colors = append(colors,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
				1, 1, 0,
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

func (v *chartVolume) render() {
	if v == nil {
		return
	}

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

	// Add vertices and indices for k percent lines.
	first := true
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

	// Add vertices and indices for d percent lines.
	first = true
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

	return &chartStochastics{
		vao:   vao,
		count: int32(len(indices)),
	}
}

func (s *chartStochastics) render() {
	if s == nil {
		return
	}

	gl.BindVertexArray(s.vao)
	gl.DrawElements(gl.LINES, s.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (s *chartStochastics) close() {
	if s == nil {
		return
	}

	gl.DeleteVertexArrays(1, &s.vao)
}
