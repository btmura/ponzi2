package ponzi

import (
	"fmt"
	"image"
	"strconv"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type chartVolume struct {
	stock     *modelStock
	labelText *dynamicText
	vao       *vao
}

func createChartVolume(stock *modelStock, labelText *dynamicText) *chartVolume {
	return &chartVolume{
		stock:     stock,
		labelText: labelText,
	}
}

func (cv *chartVolume) update() {
	if cv.stock.dailySessions == nil {
		return
	}

	if cv.vao != nil {
		return
	}

	// Find the max volume for the y-axis.
	var max int
	for _, s := range cv.stock.dailySessions {
		if s.volume > max {
			max = s.volume
		}
	}

	// Calculate vertices and indices for the volume bars.
	var vertices []float32
	var colors []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(cv.stock.dailySessions)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.2
	rightX := -1.0 + barWidth*0.8

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(max) - 1
	}

	for i, s := range cv.stock.dailySessions {
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

	cv.vao = createVAO(gl.TRIANGLES, vertices, colors, indices)
}

func (cv *chartVolume) render(r image.Rectangle) {
	if cv == nil {
		return
	}

	r = cv.renderLabels(r)
	gl.Uniform1f(colorMixAmountLocation, 1)
	setModelMatrixRectangle(r)
	cv.vao.render()
}

func (cv *chartVolume) renderLabels(r image.Rectangle) image.Rectangle {
	if cv == nil {
		return r
	}

	if cv.stock.dailySessions == nil {
		return r
	}

	var maxVol int
	for _, ds := range cv.stock.dailySessions {
		if maxVol < ds.volume {
			maxVol = ds.volume
		}
	}

	makeLabel := func(v int) string {
		switch {
		case v > 1000000000:
			return fmt.Sprintf("%dB", v/1000000000)
		case v > 1000000:
			return fmt.Sprintf("%dM", v/1000000)
		case v > 1000:
			return fmt.Sprintf("%dK", v/1000)
		}
		return strconv.Itoa(v)
	}

	labelSize := cv.labelText.measure(makeLabel(maxVol))
	labelPadX, labelPadY := 4, labelSize.Y/2

	volPerPixel := float32(maxVol) / float32(r.Dy())
	volOffset := int(float32(labelPadY+labelSize.Y/2) * volPerPixel)

	var maxTextWidth int
	render := func(v, y int) {
		l := makeLabel(v)
		s := cv.labelText.measure(l)
		if maxTextWidth < s.X {
			maxTextWidth = s.X
		}
		x := r.Max.X - s.X - labelPadX
		cv.labelText.render(l, image.Pt(x, y))
	}

	render(maxVol-volOffset, r.Max.Y-labelPadY-labelSize.Y)
	render(volOffset, r.Min.Y+labelPadY)

	r.Max.X -= maxTextWidth + labelPadX*2
	return r
}

func (cv *chartVolume) close() {
	if cv == nil {
		return
	}
	cv.vao.close()
}
