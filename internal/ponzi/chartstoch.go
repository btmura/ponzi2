package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartStochasticInterval int32

const (
	dailyInterval chartStochasticInterval = iota
	weeklyInterval
)

type chartStochastics struct {
	stock         *modelStock
	stochInterval chartStochasticInterval
	lines         *vao // lines is the VAO for the K and D lines.
}

func createChartStochastics(stock *modelStock, stochInterval chartStochasticInterval) *chartStochastics {
	return &chartStochastics{
		stock:         stock,
		stochInterval: stochInterval,
	}
}

func (ch *chartStochastics) update() {
	if ch.stock.dailySessions == nil {
		return
	}

	if ch.lines != nil {
		return
	}

	ss, dColor := ch.stock.dailySessions, yellow
	if ch.stochInterval == weeklyInterval {
		ss, dColor = ch.stock.weeklySessions, purple
	}

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
		colors = append(colors, red[0], red[1], red[2])
		if !first {
			indices = append(indices, v, v-1)
		}
		v++
		first = false
	}

	ch.lines = createVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chartStochastics) render(r image.Rectangle) {
	if ch == nil {
		return
	}
	setModelMatrixRectangle(r)
	ch.lines.render()
}

func (ch *chartStochastics) close() {
	if ch == nil {
		return
	}
	ch.lines.close()
}
