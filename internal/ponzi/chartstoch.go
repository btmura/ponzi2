package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartStochastics struct {
	lines      *vao // lines is the VAO for the K and D lines.
	background *vao // background is a rectangle used as a background.
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
		colors = append(colors, red[0], red[1], red[2])
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

	c := black
	if len(ss) != 0 {
		if k := ss[len(ss)-1].k; k != 0 {
			c = stochasticColor(k)
		}
	}

	return &chartStochastics{
		lines:      createVAO(gl.LINES, vertices, colors, indices),
		background: createFilledRectVAO(black, black, black, c),
	}
}

func (s *chartStochastics) render(r image.Rectangle) {
	if s == nil {
		return
	}
	s.lines.render(r)
	s.background.render(r)
}

func (s *chartStochastics) close() {
	if s == nil {
		return
	}
	s.lines.close()
	s.background.close()
}

func stochasticColor(k float32) [3]float32 {
	low := [3]float32{1, 0.4, 0}
	high := [3]float32{0, 0.2, 1}
	mix := [3]float32{}
	for i := 0; i < 3; i++ {
		mix[i] = low[i] + (high[i]-low[i])*k
	}
	return mix
}
