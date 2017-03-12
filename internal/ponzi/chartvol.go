package ponzi

import (
	"fmt"
	"image"
	"strconv"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func (ch *chart) createVolumeVAOs() (volRects *vao) {
	var vertices []float32
	var colors []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(ch.stock.dailySessions)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.2
	rightX := -1.0 + barWidth*0.8

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(ch.maxVolume) - 1
	}

	for i, s := range ch.stock.dailySessions {
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

	return createVAO(gl.TRIANGLES, vertices, colors, indices)
}

func (ch *chart) renderVolume(r image.Rectangle) {
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	ch.volRects.render()
}

func (ch *chart) renderVolumeLabels(r image.Rectangle) (maxLabelWidth int) {
	if ch.stock.dailySessions == nil {
		return
	}

	t1, s1 := ch.volumeLabelText(int(float32(ch.maxVolume) * .7))
	t2, s2 := ch.volumeLabelText(int(float32(ch.maxVolume) * .3))

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

func (ch *chart) volumeLabelText(v int) (text string, size image.Point) {
	var t string
	switch {
	case v > 1000000000:
		t = fmt.Sprintf("%dB", v/1000000000)
	case v > 1000000:
		t = fmt.Sprintf("%dM", v/1000000)
	case v > 1000:
		t = fmt.Sprintf("%dK", v/1000)
	default:
		t = strconv.Itoa(v)
	}
	return t, ch.labelText.measure(t)
}
