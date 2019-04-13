package view

import (
	"fmt"
	"image"
	"strconv"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/model"
)

var chartVolumeHorizRuleSet = horizRuleSetVAO([]float32{0.2, 0.8}, [2]float32{0, 1}, gray)

// chartVolume renders the volume bars and labels for a single stock.
type chartVolume struct {
	// renderable is whether the ChartVolume can be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// volBars is the VAO with the colored volume bars.
	volBars *gfx.VAO

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

func newChartVolume() *chartVolume {
	return &chartVolume{}
}

func (ch *chartVolume) SetData(ts *model.TradingSessionSeries) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if ts == nil {
		return
	}

	// Find the maximum volume.
	ch.maxVolume = 0
	for _, s := range ts.TradingSessions {
		if ch.maxVolume < s.Volume {
			ch.maxVolume = s.Volume
		}
	}

	// Measure the max label size by creating a label with the max value.
	ch.MaxLabelSize = makeChartVolumeLabel(ch.maxVolume, 1).size

	ch.volBars = chartVolumeBarsVAO(ts.TradingSessions, ch.maxVolume)

	ch.renderable = true
}

func (ch *chartVolume) ProcessInput(ic inputContext) {
	ch.bounds = ic.Bounds
}

func (ch *chartVolume) Render(fudge float32) {
	if !ch.renderable {
		return
	}

	gfx.SetModelMatrixRect(ch.bounds)

	// Render lines for the 20% and 80% levels.
	chartVolumeHorizRuleSet.Render()

	// Render the volume bars.
	ch.volBars.Render()
}

func (ch *chartVolume) RenderCursorLabels(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !ch.renderable {
		return
	}

	if !mousePos.In(mainRect) {
		return
	}

	perc := float32(mousePos.Y-mainRect.Min.Y) / float32(mainRect.Dy())
	l := makeChartVolumeLabel(ch.maxVolume, perc)
	tp := image.Point{
		X: labelRect.Max.X - l.size.X,
		Y: labelRect.Min.Y + int(float32(labelRect.Dy())*l.percent) - l.size.Y/2,
	}

	renderBubble(tp, l.size, chartAxisLabelBubbleSpec)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}

func (ch *chartVolume) Close() {
	ch.renderable = false
	if ch.volBars != nil {
		ch.volBars.Delete()
	}
}

// chartVolumeLabel is a right-justified Y-axis label with the volume.
type chartVolumeLabel struct {
	percent float32
	text    string
	size    image.Point
}

func makeChartVolumeLabel(maxVolume int, perc float32) chartVolumeLabel {
	v := int(float32(maxVolume) * perc)

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

	return chartVolumeLabel{
		percent: perc,
		text:    t,
		size:    chartAxisLabelTextRenderer.Measure(t),
	}
}

func chartVolumeBarsVAO(ds []*model.TradingSession, maxVolume int) *gfx.VAO {
	data := &gfx.VAOVertexData{Mode: gfx.Triangles}

	dx := 2.0 / float32(len(ds)) // (-1 to 1) on X-axis
	calcX := func(i int) (leftX, rightX float32) {
		x := -1.0 + dx*float32(i)
		return x + dx*0.2, x + dx*0.8
	}
	calcY := func(v int) (topY, botY float32) {
		return 2*float32(v)/float32(maxVolume) - 1, -1
	}

	for i, s := range ds {
		leftX, rightX := calcX(i)
		topY, botY := calcY(s.Volume)

		// Add the vertices needed to create the volume bar.
		data.Vertices = append(data.Vertices,
			leftX, topY, 0, // UL
			rightX, topY, 0, // UR
			leftX, botY, 0, // BL
			rightX, botY, 0, // BR
		)

		// Add the colors corresponding to the volume bar.
		switch {
		case s.Close > s.Open:
			data.Colors = append(data.Colors,
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
				green[0], green[1], green[2],
			)

		case s.Close < s.Open:
			data.Colors = append(data.Colors,
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
				red[0], red[1], red[2],
			)

		default:
			data.Colors = append(data.Colors,
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
		data.Indices = append(data.Indices,
			idx(0), idx(2), idx(1),
			idx(1), idx(2), idx(3),
		)
	}

	return gfx.NewVAO(data)
}
