package ponzi

import (
	"fmt"
	"image"
	"strconv"
	"time"

	"github.com/btmura/ponzi2/internal/gfx"
)

type ChartVolume struct {
	stock               *ModelStock
	lastStockUpdateTime time.Time
	renderable          bool
	maxVolume           int
	volRects            *gfx.VAO
}

func NewChartVolume(stock *ModelStock) *ChartVolume {
	return &ChartVolume{
		stock: stock,
	}
}

func (ch *ChartVolume) Update() {
	if ch.lastStockUpdateTime == ch.stock.LastUpdateTime {
		return
	}
	ch.lastStockUpdateTime = ch.stock.LastUpdateTime

	ch.maxVolume = 0
	for _, s := range ch.stock.DailySessions {
		if ch.maxVolume < s.Volume {
			ch.maxVolume = s.Volume
		}
	}

	if ch.volRects != nil {
		ch.volRects.Delete()
	}
	ch.volRects = createChartVolumeBarsVAO(ch.stock.DailySessions, ch.maxVolume)

	ch.renderable = true
}

func createChartVolumeBarsVAO(ds []*ModelTradingSession, maxVolume int) *gfx.VAO {
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
		topY := calcY(s.Volume)
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
		case s.Close > s.Open:
			colors = append(colors,
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
			)

		case s.Close < s.Open:
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

func (ch *ChartVolume) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(r)
	ch.volRects.Render()

	renderHorizDividers(r, chartGridHorizLine, 0.3, 0.4)
}

func (ch *ChartVolume) RenderLabels(r image.Rectangle) (maxLabelWidth int) {
	if !ch.renderable {
		return
	}

	t1, s1 := ch.volumeLabelText(int(float32(ch.maxVolume) * .7))
	t2, s2 := ch.volumeLabelText(int(float32(ch.maxVolume) * .3))

	render := func(t string, s image.Point, yLocPercent float32) {
		x := r.Max.X - s.X
		y := r.Min.Y + int(float32(r.Dy())*yLocPercent) - s.Y/2
		chartAxisLabelTextRenderer.Render(t, image.Pt(x, y), white)
	}

	render(t1, s1, .7)
	render(t2, s2, .3)

	s := s1
	if s.X < s2.X {
		s = s2
	}
	return s.X
}

func (ch *ChartVolume) volumeLabelText(v int) (text string, size image.Point) {
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
	return t, chartAxisLabelTextRenderer.Measure(t)
}

func (ch *ChartVolume) Close() {
	ch.renderable = false
	ch.volRects.Delete()
	ch.volRects = nil
}
