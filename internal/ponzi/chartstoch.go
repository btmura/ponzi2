package ponzi

import (
	"fmt"
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartStochasticInterval int32

const (
	dailyInterval chartStochasticInterval = iota
	weeklyInterval
)

const chartLabelPadding = 2

type chartStochastics struct {
	stock         *modelStock
	stochInterval chartStochasticInterval
	labelText     *dynamicText
	lines         *vao // lines is the VAO for the K and D lines.
}

func createChartStochastics(stock *modelStock, stochInterval chartStochasticInterval, labelText *dynamicText) *chartStochastics {
	return &chartStochastics{
		stock:         stock,
		stochInterval: stochInterval,
		labelText:     labelText,
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

	r.Max.X -= ch.renderLabels(r)
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.lines.render()
}

func (ch *chartStochastics) renderLabels(r image.Rectangle) (maxLabelWidth int) {
	if ch == nil {
		return
	}

	if ch.stock.dailySessions == nil {
		return
	}

	render := func(percent float32) (width int) {
		t := fmt.Sprintf("%.f%%", percent*100)
		s := ch.labelText.measure(t)
		p := image.Pt(r.Max.X-s.X-chartLabelPadding, r.Min.Y+int(float32(r.Dy())*percent)-s.Y/2)
		ch.labelText.render(t, p)
		return s.X + chartLabelPadding*2
	}

	w1, w2 := render(.3), render(.7)
	if w1 > w2 {
		return w1
	}
	return w2
}

func (ch *chartStochastics) close() {
	if ch == nil {
		return
	}
	ch.lines.close()
}
