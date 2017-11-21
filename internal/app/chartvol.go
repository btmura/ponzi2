package app

import (
	"fmt"
	"image"
	"strconv"

	"github.com/btmura/ponzi2/internal/gfx"
)

// ChartVolume renders the volume bars and labels for a single stock.
type ChartVolume struct {
	// renderable is whether the ChartVolume can be rendered.
	renderable bool

	// maxVolume is the maximum volume used for rendering measurements.
	maxVolume int

	// MaxLabelSize is the maximum label size useful for rendering measurements.
	MaxLabelSize image.Point

	// labels bundle rendering measurements for volume labels.
	labels []chartVolumeLabel

	// volBars is the VAO with the colored volume bars.
	volBars *gfx.VAO
}

// NewChartVolume creates a new ChartVolume.
func NewChartVolume() *ChartVolume {
	return &ChartVolume{}
}

// SetStock sets the ChartVolume's stock.
func (ch *ChartVolume) SetStock(st *ModelStock) {
	// Reset everything.
	ch.Close()

	// Bail out if there is no data yet.
	if st.LastUpdateTime.IsZero() {
		return // Stock has no data yet.
	}

	// Find the maximum volume.
	ch.maxVolume = 0
	for _, s := range st.DailySessions {
		if ch.maxVolume < s.Volume {
			ch.maxVolume = s.Volume
		}
	}

	// Measure the max label size by creating a label with the max value.
	ch.MaxLabelSize = makeChartVolumeLabel(ch.maxVolume, 1).size

	// Create Y-axis labels for key percentages.
	ch.labels = []chartVolumeLabel{
		makeChartVolumeLabel(ch.maxVolume, .7),
		makeChartVolumeLabel(ch.maxVolume, .3),
	}

	ch.volBars = chartVolumeBarsVAO(st.DailySessions, ch.maxVolume)

	ch.renderable = true
}

// Render renders the volume bars.
func (ch *ChartVolume) Render(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	// Render lines for the 30% and 70% levels.
	renderSlicedRectDividers(r, chartGridHorizLine, 0.3, 0.4)

	// Render the volume bars.
	gfx.SetModelMatrixRect(r)
	ch.volBars.Render()
}

// RenderAxisLabels renders the Y-axis labels for the volume bars.
func (ch *ChartVolume) RenderAxisLabels(r image.Rectangle) {
	if !ch.renderable {
		return
	}

	for _, l := range ch.labels {
		x := r.Max.X - l.size.X
		y := r.Min.Y + int(float32(r.Dy())*l.percent) - l.size.Y/2
		chartAxisLabelTextRenderer.Render(l.text, image.Pt(x, y), white)
	}
}

// RenderCursorLabels renders a label for the value under the mouse cursor.
func (ch *ChartVolume) RenderCursorLabels(mainRect, labelRect image.Rectangle, mousePos image.Point) {
	if !ch.renderable {
		return
	}

	if !mousePos.In(mainRect) {
		return
	}

	perc := float32(mousePos.Y-mainRect.Min.Y) / float32(mainRect.Dy())
	l := makeChartVolumeLabel(ch.maxVolume, perc)

	var tp image.Point
	tp.X = labelRect.Max.X - l.size.X
	tp.Y = labelRect.Min.Y + int(float32(labelRect.Dy())*l.percent) - l.size.Y/2

	br := image.Rectangle{Min: tp, Max: tp.Add(l.size)}
	br = br.Inset(-chartAxisLabelBubblePadding)

	fillRoundedRect(br, chartAxisLabelBubbleRounding)
	strokeRoundedRect(br, chartAxisLabelBubbleRounding)
	chartAxisLabelTextRenderer.Render(l.text, tp, white)
}

// Close frees the resources backing the ChartVolume.
func (ch *ChartVolume) Close() {
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

func chartVolumeBarsVAO(ds []*ModelTradingSession, maxVolume int) *gfx.VAO {
	data := &gfx.VAOVertexData{}

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

	return gfx.NewVAO(gfx.Triangles, data)
}
