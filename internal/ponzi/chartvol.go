package ponzi

import (
	"fmt"
	"image"
	"strconv"
	"time"

	"github.com/btmura/ponzi2/internal/gfx"
)

type chartVolume struct {
	stock               *modelStock
	lastStockUpdateTime time.Time
	renderable          bool
	maxVolume           int
	volRects            *gfx.VAO2
	labelLine           *gfx.VAO2
}

func createChartVolume(stock *modelStock) *chartVolume {
	return &chartVolume{
		stock: stock,
	}
}

func (ch *chartVolume) update() {
	if ch.lastStockUpdateTime == ch.stock.lastUpdateTime {
		return
	}
	ch.lastStockUpdateTime = ch.stock.lastUpdateTime

	ch.maxVolume = 0
	for _, s := range ch.stock.dailySessions {
		if ch.maxVolume < s.volume {
			ch.maxVolume = s.volume
		}
	}

	if ch.volRects != nil {
		ch.volRects.Delete()
	}
	ch.volRects = createChartVolumeBarsVAO(ch.stock.dailySessions, ch.maxVolume)

	if ch.labelLine != nil {
		ch.labelLine.Delete()
	}
	ch.labelLine = gfx.HorizColoredLineVAO(gray, gray)

	ch.renderable = true
}

func createChartVolumeBarsVAO(ds []*modelTradingSession, maxVolume int) *gfx.VAO2 {
	var vertices []float32
	var colors []float32
	var indices []uint16

	barWidth := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	leftX := -1.0 + barWidth*0.2
	rightX := -1.0 + barWidth*0.8

	calcY := func(value int) float32 {
		return 2*float32(value)/float32(maxVolume) - 1
	}

	for i, s := range ds {
		topY := calcY(s.volume)
		botY := calcY(0)

		// Add the vertices needed to create the volume bar.
		vertices = append(vertices,
			leftX, topY, 0, // UL
			rightX, topY, 0, // UR
			leftX, botY, 0, // BL
			rightX, botY, 0, // BR
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

	return gfx.NewVAO(
		gfx.Triangles,
		&gfx.VAOVertexData{
			Vertices: vertices,
			Colors:   colors,
			Indices:  indices,
		},
	)
}

func (ch *chartVolume) render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetColorMixAmount(1)
	gfx.SetModelMatrixRect(r)
	ch.volRects.Render()

	for _, yLocPercent := range []float32{0.3, 0.7} {
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent)
		gfx.SetModelMatrixRect(image.Rect(r.Min.X, y, r.Max.X, y))
		ch.labelLine.Render()
	}
}

func (ch *chartVolume) renderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	t1, s1 := ch.volumeLabelText(int(float32(ch.maxVolume) * .7))
	t2, s2 := ch.volumeLabelText(int(float32(ch.maxVolume) * .3))

	render := func(t string, s image.Point, yLocPercent float32) {
		x := r.Max.X - s.X
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent) - s.Y/2
		axisLabelTextRenderer.Render(t, image.Pt(x, y), white)
	}

	render(t1, s1, .7)
	render(t2, s2, .3)

	s := s1
	if s.X < s2.X {
		s = s2
	}
	return s.X
}

func (ch *chartVolume) volumeLabelText(v int) (text string, size image.Point) {
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
	return t, axisLabelTextRenderer.Measure(t)
}

func (ch *chartVolume) close() {
	ch.renderable = false
	ch.volRects.Delete()
	ch.volRects = nil
	ch.labelLine.Delete()
	ch.labelLine = nil
}
