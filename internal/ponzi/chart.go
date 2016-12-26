package ponzi

import (
	"image"
	"math"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chart struct {
	symbol      string
	prices      *chartPrices
	volume      *chartVolume
	stochastics *chartStochastics
}

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

func createChart(symbol string, sessions []*modelTradingSession) (*chart, func()) {
	prices, cleanUpPrices := createChartPrices(sessions)
	volume, cleanUpVolume := createChartVolume(sessions)
	stochastics, cleanUpStochastics := createChartStochastics(sessions)
	return &chart{
			symbol:      symbol,
			prices:      prices,
			volume:      volume,
			stochastics: stochastics,
		}, func() {
			cleanUpPrices()
			cleanUpVolume()
			cleanUpStochastics()
		}
}

func createChartPrices(sessions []*modelTradingSession) (*chartPrices, func()) {
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
	var lineIndices []uint16
	var triangleIndices []uint16

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

	vbo := createArrayBuffer(vertices)
	lineIBO := createElementArrayBuffer(lineIndices)
	triangleIBO := createElementArrayBuffer(triangleIndices)

	var lineVAO uint32
	gl.GenVertexArrays(1, &lineVAO)
	gl.BindVertexArray(lineVAO)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, lineIBO)
	}
	gl.BindVertexArray(0)

	var triangleVAO uint32
	gl.GenVertexArrays(1, &triangleVAO)
	gl.BindVertexArray(triangleVAO)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, triangleIBO)
	}
	gl.BindVertexArray(0)

	return &chartPrices{
			lineVAO:       lineVAO,
			lineCount:     int32(len(lineIndices)),
			triangleVAO:   triangleVAO,
			triangleCount: int32(len(triangleIndices)),
		}, func() {
			gl.DeleteVertexArrays(1, &lineVAO)
			gl.DeleteVertexArrays(1, &triangleVAO)
		}
}

func createChartVolume(sessions []*modelTradingSession) (*chartVolume, func()) {
	// Find the max volume for the y-axis.
	var max int
	for _, s := range sessions {
		if s.volume > max {
			max = s.volume
		}
	}

	// Calculate vertices and indices for the candlesticks.
	var vertices []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(sessions)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.1
	rightX := -1.0 + barWidth*0.9

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(max) - 1
	}

	for i, s := range sessions {
		topY := calcY(s.volume)
		botY := calcY(0)

		// Add the vertices needed to create the candlestick.
		vertices = append(vertices,
			leftX, topY, // UL
			rightX, topY, // UR
			leftX, botY, // BL
			rightX, botY, // BR
		)

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

	vbo := createArrayBuffer(vertices)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	}
	gl.BindVertexArray(0)

	return &chartVolume{
			vao:   vao,
			count: int32(len(indices)),
		}, func() {
			gl.DeleteVertexArrays(1, &vao)
		}
}

func createChartStochastics(sessions []*modelTradingSession) (*chartStochastics, func()) {
	// Calculate vertices and indices for the stochastics.
	var vertices []float32
	var indices []uint16

	segmentWidth := 2.0 / float32(len(sessions)) // (-1 to 1) on X-axis
	midX := -1.0 + segmentWidth*0.5

	calcY := func(value float32) float32 {
		return 2*float32(value) - 1
	}

	var i uint16
	for _, s := range sessions {
		if s.k > 0 {
			y := calcY(s.k)
			vertices = append(vertices, midX, y)
			if i > 0 {
				indices = append(indices, uint16(i), uint16(i-1))
			}
			i++
		}

		// Move the X coordinates one bar over.
		midX += segmentWidth
	}

	vbo := createArrayBuffer(vertices)
	ibo := createElementArrayBuffer(indices)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	{
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.EnableVertexAttribArray(positionLocation)
		gl.VertexAttribPointer(positionLocation, 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	}
	gl.BindVertexArray(0)

	return &chartStochastics{
			vao:   vao,
			count: int32(len(indices)),
		}, func() {
			gl.DeleteVertexArrays(1, &vao)
		}
}

func (c *chart) renderPrices(r image.Rectangle) {
	setModelMatrixRectangle(r)
	gl.BindTexture(gl.TEXTURE_2D, 0 /* dummy texture */)

	gl.BindVertexArray(c.prices.lineVAO)
	gl.DrawElements(gl.LINES, c.prices.lineCount, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)

	gl.BindVertexArray(c.prices.triangleVAO)
	gl.DrawElements(gl.TRIANGLES, c.prices.triangleCount, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (c *chart) renderVolume(r image.Rectangle) {
	setModelMatrixRectangle(r)
	gl.BindTexture(gl.TEXTURE_2D, 0 /* dummy texture */)

	gl.BindVertexArray(c.volume.vao)
	gl.DrawElements(gl.TRIANGLES, c.volume.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}

func (c *chart) renderStochastics(r image.Rectangle) {
	setModelMatrixRectangle(r)
	gl.BindTexture(gl.TEXTURE_2D, 0 /* dummy texture */)

	gl.BindVertexArray(c.stochastics.vao)
	gl.DrawElements(gl.LINES, c.stochastics.count, gl.UNSIGNED_SHORT, gl.Ptr(nil))
	gl.BindVertexArray(0)
}
