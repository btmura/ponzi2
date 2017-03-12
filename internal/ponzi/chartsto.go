package ponzi

import (
	"fmt"
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (ch *chart) createStochasticVAOs(ss []*modelTradingSession, dColor [3]float32) (stoLines *vao) {
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

	return createVAO(gl.LINES, vertices, colors, indices)
}

func (ch *chart) renderStochastics(r image.Rectangle, vao *vao) {
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	vao.render()
}

func (ch *chart) renderStochasticLabels(r image.Rectangle) (maxLabelWidth int) {
	if ch.stock.dailySessions == nil {
		return
	}

	t1, s1 := ch.stochasticLabelText(.7)
	t2, s2 := ch.stochasticLabelText(.3)

	render := func(t string, s image.Point, yLocPercent float32) {
		x := r.Max.X - chartLabelPadding - s.X
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent) - s.Y/2
		ch.labelText.render(t, image.Pt(x, y))
	}

	render(t1, s1, .7)
	render(t2, s2, .3)

	s := s1
	if s.X < s2.X {
		s = s2
	}
	return s.X + chartLabelPadding*2
}

func (ch *chart) stochasticLabelText(percent float32) (text string, size image.Point) {
	t := fmt.Sprintf("%.f%%", percent*100)
	return t, ch.labelText.measure(t)
}
