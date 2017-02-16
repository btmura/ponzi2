package ponzi

import (
	"image"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartVolume struct {
	vao *vao
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
		const p = 0.1 // padding
		return (2-p)*float32(value)/float32(max) - 1
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

	return &chartVolume{createVAO(gl.TRIANGLES, vertices, colors, indices)}
}

func (v *chartVolume) render(r image.Rectangle) {
	if v == nil {
		return
	}
	setModelMatrixRectangle(r)
	v.vao.render()
}

func (v *chartVolume) close() {
	if v == nil {
		return
	}
	v.vao.close()
}
